package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"shop/pkg/proto"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

// --- stub clients ---

// stubGoodsClient 只实现 BatchGetGoods，其余方法通过嵌入接口留空（测试不会调用）。
type stubGoodsClient struct {
	proto.GoodsClient
	goods map[int32]*proto.GoodsInfoResponse
}

func (c *stubGoodsClient) BatchGetGoods(ctx context.Context, in *proto.BatchGoodsIdInfo, opts ...grpc.CallOption) (*proto.GoodsListResponse, error) {
	rsp := &proto.GoodsListResponse{}
	for _, id := range in.Id {
		if g, ok := c.goods[id]; ok {
			rsp.Data = append(rsp.Data, g)
		}
	}
	return rsp, nil
}

// stubInventoryClient 记录扣减/归还调用，可注入错误以测试失败路径。
type stubInventoryClient struct {
	proto.InventoryClient
	sellErr     error
	rebackErr   error
	sellCalls   []*proto.OrderStockDetail
	rebackCalls []*proto.OrderStockDetail
}

func (c *stubInventoryClient) StockSellDetail(ctx context.Context, in *proto.OrderStockDetail, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	c.sellCalls = append(c.sellCalls, in)
	if c.sellErr != nil {
		return nil, c.sellErr
	}
	return &emptypb.Empty{}, nil
}

func (c *stubInventoryClient) RebackDetail(ctx context.Context, in *proto.OrderStockDetail, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	c.rebackCalls = append(c.rebackCalls, in)
	if c.rebackErr != nil {
		return nil, c.rebackErr
	}
	return &emptypb.Empty{}, nil
}

// --- test DB harness ---

const testDBName = "shop_order_srv_test"

func mysqlCreds() (host, user, pass string) {
	host = envOr("SHOP_TEST_MYSQL_HOST", "127.0.0.1:3306")
	user = envOr("SHOP_TEST_MYSQL_USER", "root")
	pass = envOr("SHOP_TEST_MYSQL_PASS", "root123456")
	return
}

func dsn(dbName string) string {
	host, user, pass := mysqlCreds()
	return fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=true&loc=Local", user, pass, host, dbName)
}

// setupTestDB 在 MySQL 中重建一个独立的测试库并迁移 order_srv 的全部表。
// MySQL 不可达时跳过测试（t.Skip），保证环境缺失时不报错。
func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	cfg := &gorm.Config{
		NamingStrategy: schema.NamingStrategy{SingularTable: true},
		DisableForeignKeyConstraintWhenMigrating: true,
		Logger: logger.Default.LogMode(logger.Silent),
	}

	srv, err := gorm.Open(mysql.Open(dsn("")), cfg)
	if err != nil {
		t.Skipf("mysql 不可达，跳过集成测试: %v", err)
	}
	if err := srv.Exec("DROP DATABASE IF EXISTS `" + testDBName + "`").Error; err != nil {
		t.Fatalf("drop test db: %v", err)
	}
	if err := srv.Exec("CREATE DATABASE `" + testDBName + "` DEFAULT CHARACTER SET utf8mb4").Error; err != nil {
		t.Fatalf("create test db: %v", err)
	}
	sqlSrv, _ := srv.DB()
	_ = sqlSrv.Close()

	db, err := gorm.Open(mysql.Open(dsn(testDBName)), cfg)
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	if err := db.AutoMigrate(&ShoppingCart{}, &OrderInfo{}, &OrderGoods{}); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}

	t.Cleanup(func() {
		if sqlDB, err := db.DB(); err == nil {
			_ = sqlDB.Close()
		}
		cleanup, err := gorm.Open(mysql.Open(dsn("")), cfg)
		if err == nil {
			_ = cleanup.Exec("DROP DATABASE IF EXISTS `" + testDBName + "`")
			if c, e := cleanup.DB(); e == nil {
				_ = c.Close()
			}
		}
	})
	return db
}

func newTestServer(t *testing.T, db *gorm.DB, goods map[int32]*proto.GoodsInfoResponse, inv *stubInventoryClient) *OrderServer {
	t.Helper()
	if goods == nil {
		goods = map[int32]*proto.GoodsInfoResponse{}
	}
	if inv == nil {
		inv = &stubInventoryClient{}
	}
	return NewOrderServer(db, &stubGoodsClient{goods: goods}, inv, zap.NewNop().Sugar())
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// --- pure function ---

func TestNewOrderSn(t *testing.T) {
	sn := newOrderSn(1001)
	// 结构：年月日时分秒(14) + userId(4) + 2位随机(2) = 20
	if len(sn) != 20 {
		t.Fatalf("newOrderSn(1001) = %q (len %d), 期望长度 20", sn, len(sn))
	}
	// 时间戳段为纯数字且以 20 开头
	ts := sn[:14]
	for _, c := range ts {
		if c < '0' || c > '9' {
			t.Fatalf("时间戳段非数字: %q", ts)
		}
	}
	if !strings.HasPrefix(ts, "20") {
		t.Fatalf("时间戳应以 20 开头: %q", ts)
	}
	// userId 段
	if sn[14:18] != "1001" {
		t.Fatalf("userId 段 = %q, 期望 1001", sn[14:18])
	}
	// 2 位随机
	rand := sn[18:]
	if len(rand) != 2 || rand[0] < '0' || rand[0] > '9' || rand[1] < '0' || rand[1] > '9' {
		t.Fatalf("随机段异常: %q", rand)
	}
}

// --- 购物车 CRUD ---

func TestCartItemCRUD(t *testing.T) {
	db := setupTestDB(t)
	srv := newTestServer(t, db, nil, nil)
	ctx := context.Background()

	// 新增
	must(t, errOf(srv.AddCartItem(ctx, &proto.AddCartItemRequest{UserId: 1, GoodsId: 101, Num: 2, Checked: true})))
	// 同商品再添加 -> 累加数量（选中状态不变，Add 不传 checked 时为 false 但已存在则不改动）
	must(t, errOf(srv.AddCartItem(ctx, &proto.AddCartItemRequest{UserId: 1, GoodsId: 101, Num: 3})))
	list, err := srv.CartItemList(ctx, &proto.CartItemListRequest{UserId: 1})
	must(t, err)
	if len(list.Data) != 1 || list.Data[0].Num != 5 || !list.Data[0].Checked {
		t.Fatalf("累加后购物车 = %+v, 期望 num=5 checked=true", list.Data)
	}

	// 更新（数量 + 取消选中）
	must(t, errOf(srv.UpdateCartItem(ctx, &proto.UpdateCartItemRequest{Id: list.Data[0].Id, Num: 1, Checked: false})))
	list, err = srv.CartItemList(ctx, &proto.CartItemListRequest{UserId: 1})
	must(t, err)
	if list.Data[0].Num != 1 || list.Data[0].Checked {
		t.Fatalf("更新后购物车 = %+v, 期望 num=1 checked=false", list.Data)
	}

	// 删除
	must(t, errOf(srv.DeleteCartItem(ctx, &proto.DeleteCartItemRequest{Id: list.Data[0].Id})))
	list, err = srv.CartItemList(ctx, &proto.CartItemListRequest{UserId: 1})
	must(t, err)
	if len(list.Data) != 0 {
		t.Fatalf("删除后购物车应为空, got %+v", list.Data)
	}

	// 删除不存在的记录 -> NotFound
	if _, err := srv.DeleteCartItem(ctx, &proto.DeleteCartItemRequest{Id: 9999}); status.Code(err) != codes.NotFound {
		t.Fatalf("删除不存在期望 NotFound, got %v", err)
	}
}

// --- CreateOrder 编排 ---

func sampleGoods() map[int32]*proto.GoodsInfoResponse {
	return map[int32]*proto.GoodsInfoResponse{
		101: {Id: 101, Name: "苹果", ShopPrice: 5.0, GoodsFrontImage: "a.jpg"},
		102: {Id: 102, Name: "梨", ShopPrice: 3.0, GoodsFrontImage: "p.jpg"},
	}
}

func addCheckedCart(t *testing.T, srv *OrderServer, ctx context.Context, userID int32, items ...struct{ gid, num int32 }) {
	t.Helper()
	for _, it := range items {
		must(t, errOf(srv.AddCartItem(ctx, &proto.AddCartItemRequest{UserId: userID, GoodsId: it.gid, Num: it.num, Checked: true})))
	}
}

func TestCreateOrder_HappyPath(t *testing.T) {
	db := setupTestDB(t)
	inv := &stubInventoryClient{}
	srv := newTestServer(t, db, sampleGoods(), inv)
	ctx := context.Background()

	addCheckedCart(t, srv, ctx, 1, struct{ gid, num int32 }{101, 2}, struct{ gid, num int32 }{102, 1})
	// 一件未选中商品，下单不应包含、也不应删除
	must(t, errOf(srv.AddCartItem(ctx, &proto.AddCartItemRequest{UserId: 1, GoodsId: 999, Num: 1, Checked: false})))

	rsp, err := srv.CreateOrder(ctx, &proto.OrderInfoRequest{
		UserId: 1, Address: "addr", Name: "tom", Mobile: "13800000000", PayType: 1, PostFee: 0,
	})
	must(t, err)

	// 总价 = 5*2 + 3*1 = 13
	if !nearlyEqual(rsp.Total, 13.0) {
		t.Fatalf("total = %v, 期望 13", rsp.Total)
	}
	if len(rsp.OrderGoods) != 2 {
		t.Fatalf("订单商品数 = %d, 期望 2", len(rsp.OrderGoods))
	}
	if rsp.OrderSn == "" {
		t.Fatal("orderSn 为空")
	}
	// 商品快照来自 goods_srv
	for _, og := range rsp.OrderGoods {
		if og.GoodsName == "" || og.GoodsImage == "" || og.GoodsPrice == 0 {
			t.Fatalf("订单商品快照缺失: %+v", og)
		}
	}

	// 库存扣减被调用一次，明细为 2 件
	if len(inv.sellCalls) != 1 || len(inv.sellCalls[0].OrderGoods) != 2 {
		t.Fatalf("库存扣减调用异常: %+v", inv.sellCalls)
	}
	if len(inv.rebackCalls) != 0 {
		t.Fatalf("成功路径不应归还库存, got %d 次", len(inv.rebackCalls))
	}

	// 购物车：已购(选中)的删除，未选中的保留
	list, err := srv.CartItemList(ctx, &proto.CartItemListRequest{UserId: 1})
	must(t, err)
	if len(list.Data) != 1 || list.Data[0].GoodsId != 999 {
		t.Fatalf("购物车应仅剩未选中商品 999, got %+v", list.Data)
	}
}

func TestCreateOrder_NoCheckedItems(t *testing.T) {
	db := setupTestDB(t)
	srv := newTestServer(t, db, sampleGoods(), nil)
	ctx := context.Background()

	_, err := srv.CreateOrder(ctx, &proto.OrderInfoRequest{UserId: 1})
	if status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("空购物车期望 FailedPrecondition, got %v", err)
	}
}

func TestCreateOrder_GoodsMissing(t *testing.T) {
	db := setupTestDB(t)
	inv := &stubInventoryClient{}
	srv := newTestServer(t, db, sampleGoods(), inv) // 仅 101/102
	ctx := context.Background()

	addCheckedCart(t, srv, ctx, 1, struct{ gid, num int32 }{101, 1}, struct{ gid, num int32 }{777, 1})

	if _, err := srv.CreateOrder(ctx, &proto.OrderInfoRequest{UserId: 1}); status.Code(err) != codes.NotFound {
		t.Fatalf("商品缺失期望 NotFound, got %v", err)
	}
	// 商品校验在库存扣减之前，库存不应被调用
	if len(inv.sellCalls) != 0 {
		t.Fatalf("商品缺失时不应扣减库存, got %d 次", len(inv.sellCalls))
	}
	// 订单未创建、购物车未被清空
	if count(db, &OrderInfo{}) != 0 {
		t.Fatal("不应有订单产生")
	}
	if c := count(db, &ShoppingCart{}); c != 2 {
		t.Fatalf("购物车应保留 2 条, got %d", c)
	}
}

func TestCreateOrder_InventoryFail(t *testing.T) {
	db := setupTestDB(t)
	inv := &stubInventoryClient{sellErr: status.Error(codes.ResourceExhausted, "库存不足")}
	srv := newTestServer(t, db, sampleGoods(), inv)
	ctx := context.Background()

	addCheckedCart(t, srv, ctx, 1, struct{ gid, num int32 }{101, 1})

	_, err := srv.CreateOrder(ctx, &proto.OrderInfoRequest{UserId: 1})
	if err == nil {
		t.Fatal("库存扣减失败应返回错误")
	}
	// 库存失败 -> 不创建订单、不清购物车
	if count(db, &OrderInfo{}) != 0 {
		t.Fatal("库存失败不应产生订单")
	}
	if c := count(db, &ShoppingCart{}); c != 1 {
		t.Fatalf("库存失败购物车应保留 1 条, got %d", c)
	}
}

// --- 订单生命周期 ---

func TestOrderLifecycle(t *testing.T) {
	db := setupTestDB(t)
	srv := newTestServer(t, db, sampleGoods(), nil)
	ctx := context.Background()

	addCheckedCart(t, srv, ctx, 7, struct{ gid, num int32 }{101, 2})
	created, err := srv.CreateOrder(ctx, &proto.OrderInfoRequest{
		UserId: 7, Address: "addr", Name: "jerry", Mobile: "13900000000", PayType: 2,
	})
	must(t, err)

	// 列表
	lst, err := srv.OrderList(ctx, &proto.OrderFilterRequest{UserId: 7})
	must(t, err)
	if lst.Total == 0 {
		t.Fatal("OrderList 应包含该用户订单")
	}

	// 详情（带 userId 校验）
	detail, err := srv.GetOrderDetail(ctx, &proto.OrderInfoRequest{Id: created.Id, UserId: 7})
	must(t, err)
	if len(detail.OrderGoods) != 1 {
		t.Fatalf("详情商品数 = %d, 期望 1", len(detail.OrderGoods))
	}

	// 归属校验：其他用户看不到
	if _, err := srv.GetOrderDetail(ctx, &proto.OrderInfoRequest{Id: created.Id, UserId: 999}); status.Code(err) != codes.NotFound {
		t.Fatalf("非归属用户查询期望 NotFound, got %v", err)
	}

	// 更新状态（按 orderSn）
	must(t, errOf(srv.UpdateOrderStatus(ctx, &proto.UpdateOrderStatusInfo{OrderSn: created.OrderSn, Status: StatusPaid})))
	detail, err = srv.GetOrderDetail(ctx, &proto.OrderInfoRequest{Id: created.Id, UserId: 7})
	must(t, err)
	if detail.Status != StatusPaid {
		t.Fatalf("状态未更新, got %d", detail.Status)
	}

	// 非法状态码
	if _, err := srv.UpdateOrderStatus(ctx, &proto.UpdateOrderStatusInfo{OrderSn: created.OrderSn, Status: 99}); status.Code(err) != codes.InvalidArgument {
		t.Fatalf("非法状态期望 InvalidArgument, got %v", err)
	}

	// 不存在的订单号
	if _, err := srv.UpdateOrderStatus(ctx, &proto.UpdateOrderStatusInfo{OrderSn: "bogus", Status: StatusCancel}); status.Code(err) != codes.NotFound {
		t.Fatalf("不存在订单号期望 NotFound, got %v", err)
	}

	// 删除
	must(t, errOf(srv.DeleteOrder(ctx, &proto.DeleteOrderInfo{Id: created.Id})))
	if _, err := srv.GetOrderDetail(ctx, &proto.OrderInfoRequest{Id: created.Id, UserId: 7}); status.Code(err) != codes.NotFound {
		t.Fatalf("删除后查询期望 NotFound, got %v", err)
	}
}

// --- helpers ---

func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// errOf 取 (*emptypb.Empty, error) 中的 error，配合 must 使用：must(t, errOf(srv.RPC(...)))。
// srv.RPC(...) 作为 errOf 的唯一实参，满足 Go 多值返回透传。
func errOf(_ *emptypb.Empty, err error) error { return err }

func count(db *gorm.DB, model any) int64 {
	var n int64
	db.Model(model).Count(&n)
	return n
}

func nearlyEqual(a, b float32) bool {
	d := a - b
	if d < 0 {
		d = -d
	}
	return d < 1e-3
}
