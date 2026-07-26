package web

import (
	"context"
	"net/http"
	"strconv"

	"shop/pkg/proto"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/smartwalle/alipay/v3"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// HandleGrpcErrorToHttp maps a gRPC status error onto an HTTP response.
func HandleGrpcErrorToHttp(err error, ctx *gin.Context) {
	if err == nil {
		return
	}
	if e, ok := status.FromError(err); ok {
		switch e.Code() {
		case codes.NotFound:
			ctx.JSON(http.StatusNotFound, gin.H{"msg": e.Message()})
		case codes.Internal:
			ctx.JSON(http.StatusInternalServerError, gin.H{"msg": e.Message()})
		case codes.InvalidArgument:
			ctx.JSON(http.StatusBadRequest, gin.H{"msg": e.Message()})
		case codes.FailedPrecondition:
			ctx.JSON(http.StatusBadRequest, gin.H{"msg": e.Message()})
		case codes.ResourceExhausted:
			ctx.JSON(http.StatusBadRequest, gin.H{"msg": e.Message()})
		case codes.AlreadyExists:
			ctx.JSON(http.StatusConflict, gin.H{"msg": e.Message()})
		case codes.PermissionDenied:
			ctx.JSON(http.StatusForbidden, gin.H{"msg": e.Message()})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"msg": e.Message()})
		}
		return
	}
	ctx.JSON(http.StatusInternalServerError, gin.H{"msg": err.Error()})
}

func handleValidatorError(err error, ctx *gin.Context) {
	errs, ok := err.(validator.ValidationErrors)
	if !ok {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": err.Error()})
		return
	}
	ctx.JSON(http.StatusBadRequest, gin.H{"msg": removeTopStruct(errs.Translate(translator))})
}

// ==================== 购物车 Handlers ====================

// CartItemList 获取当前用户的购物车列表，返回 snake_case JSON，
// 并通过 goods_srv 批量补全商品名/图/价。
func (s *Server) CartItemList(ctx *gin.Context) {
	rsp, err := s.orderSrv.CartItemList(ctx.Request.Context(), &proto.CartItemListRequest{
		UserId: currentUserID(ctx),
	})
	if err != nil {
		s.log.Error("CartItemList grpc error: ", err)
		HandleGrpcErrorToHttp(err, ctx)
		return
	}

	// 批量拉取商品信息（名/封面图/本店价）补全购物车项
	goodsMap := map[int32]*proto.GoodsInfoResponse{}
	if len(rsp.Data) > 0 {
		ids := make([]int32, 0, len(rsp.Data))
		for _, c := range rsp.Data {
			ids = append(ids, c.GoodsId)
		}
		if gRsp, gErr := s.goodsSrv.BatchGetGoods(ctx.Request.Context(), &proto.BatchGoodsIdInfo{Id: ids}); gErr == nil {
			for _, g := range gRsp.Data {
				goodsMap[g.Id] = g
			}
		} else {
			s.log.Error("CartItemList BatchGetGoods error: ", gErr)
		}
	}

	out := &cartItemListResponse{
		Total: rsp.Total,
		Data:  make([]*cartItemResponse, 0, len(rsp.Data)),
	}
	for _, c := range rsp.Data {
		out.Data = append(out.Data, cartItemToResponse(c, goodsMap[c.GoodsId]))
	}
	ctx.JSON(http.StatusOK, out)
}

// AddCartItem 添加商品到购物车。先经 goods_srv 校验商品存在且在售，再写入购物车。
func (s *Server) AddCartItem(ctx *gin.Context) {
	var form AddCartItemForm
	if err := ctx.ShouldBind(&form); err != nil {
		handleValidatorError(err, ctx)
		return
	}
	// 校验商品存在且在售
	goods, err := s.goodsSrv.GetGoodsDetail(ctx.Request.Context(), &proto.GoodInfoRequest{Id: form.GoodsId})
	if err != nil {
		s.log.Error("AddCartItem GetGoodsDetail error: ", err)
		HandleGrpcErrorToHttp(err, ctx) // 商品不存在 -> 404
		return
	}
	if !goods.OnSale {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "商品已下架"})
		return
	}
	// 校验库存是否足够（无库存记录视为不足）
	inv, err := s.invSrv.GetInvDetail(ctx.Request.Context(), &proto.GoodsInvInfo{GoodsId: form.GoodsId})
	if err != nil {
		if status.Code(err) == codes.NotFound {
			ctx.JSON(http.StatusBadRequest, gin.H{"msg": "库存不足"})
			return
		}
		s.log.Error("AddCartItem GetInvDetail error: ", err)
		HandleGrpcErrorToHttp(err, ctx)
		return
	}
	if inv.Stocks < form.Num {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "库存不足"})
		return
	}
	_, err = s.orderSrv.AddCartItem(ctx.Request.Context(), &proto.AddCartItemRequest{
		UserId:  currentUserID(ctx),
		GoodsId: form.GoodsId,
		Num:     form.Num,
		Checked: form.Checked,
	})
	if err != nil {
		s.log.Error("AddCartItem grpc error: ", err)
		HandleGrpcErrorToHttp(err, ctx)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"msg": "添加成功"})
}

// UpdateCartItem 更新购物车商品（数量 / 选中状态）
func (s *Server) UpdateCartItem(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "invalid id"})
		return
	}
	var form UpdateCartItemForm
	if err := ctx.ShouldBind(&form); err != nil {
		handleValidatorError(err, ctx)
		return
	}
	_, err = s.orderSrv.UpdateCartItem(ctx.Request.Context(), &proto.UpdateCartItemRequest{
		Id:      int32(id),
		Num:     form.Num,
		Checked: form.Checked,
		UserId:  currentUserID(ctx),
	})
	if err != nil {
		s.log.Error("UpdateCartItem grpc error: ", err)
		HandleGrpcErrorToHttp(err, ctx)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"msg": "更新成功"})
}

// DeleteCartItem 删除购物车商品
func (s *Server) DeleteCartItem(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "invalid id"})
		return
	}
	_, err = s.orderSrv.DeleteCartItem(ctx.Request.Context(), &proto.DeleteCartItemRequest{
		Id:     int32(id),
		UserId: currentUserID(ctx),
	})
	if err != nil {
		s.log.Error("DeleteCartItem grpc error: ", err)
		HandleGrpcErrorToHttp(err, ctx)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"msg": "删除成功"})
}

// ==================== 订单 Handlers ====================

// CreateOrder 创建订单（商品来自购物车已选项）
func (s *Server) CreateOrder(ctx *gin.Context) {
	var form CreateOrderForm
	if err := ctx.ShouldBind(&form); err != nil {
		handleValidatorError(err, ctx)
		return
	}
	rsp, err := s.orderSrv.CreateOrder(ctx.Request.Context(), &proto.OrderInfoRequest{
		UserId:  currentUserID(ctx),
		Address: form.Address,
		Name:    form.Name,
		Mobile:  form.Mobile,
		Post:    form.Post,
		PayType: form.PayType,
		PostFee: form.PostFee,
	})
	if err != nil {
		s.log.Error("CreateOrder grpc error: ", err)
		HandleGrpcErrorToHttp(err, ctx)
		return
	}

	// 生成支付宝url
	client, err := alipay.New(s.cfg.AliPay.AppId, s.cfg.AliPay.PrivateKey, s.cfg.AliPay.IsProduction)
	if err != nil {
		s.log.Error("alipay.New error: ", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"msg": "支付服务初始化失败"})
		return
	}
	err = client.LoadAliPayPublicKey(s.cfg.AliPay.PublicKey)
	if err != nil {
		s.log.Error("client.LoadAliPayPublicKey error: ", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"msg": "支付公钥加载失败"})
		return
	}

	var p = alipay.TradePagePay{}
	p.NotifyURL = s.cfg.AliPay.NotifyUrl
	p.ReturnURL = s.cfg.AliPay.ReturnUrl
	p.Subject = "shop订单支付-" + rsp.OrderSn
	p.OutTradeNo = rsp.OrderSn
	p.TotalAmount = strconv.FormatFloat(float64(rsp.Total), 'f', 2, 64)
	p.ProductCode = "FAST_INSTANT_TRADE_PAY"
	url, err := client.TradePagePay(p)
	if err != nil {
		s.log.Error("client.TradePagePay error: ", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"msg": "支付请求url生成失败"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"id": rsp.Id, "alipay_url": url.String()})
}

// OrderList 订单列表（角色感知）：
//   - 管理员：默认查全部订单，可用 user_id 按用户过滤
//   - 普通用户：只看本人订单（忽略 user_id 入参，防越权）
func (s *Server) OrderList(ctx *gin.Context) {
	var userId int32
	if isAdmin(ctx) {
		uid, _ := strconv.Atoi(ctx.DefaultQuery("user_id", "0"))
		userId = int32(uid) // 0 = 全部
	} else {
		userId = currentUserID(ctx) // 强制本人
	}
	status, _ := strconv.Atoi(ctx.DefaultQuery("status", "0"))
	pages, _ := strconv.Atoi(ctx.DefaultQuery("p", "1"))
	pagePerNums, _ := strconv.Atoi(ctx.DefaultQuery("pnum", "10"))
	rsp, err := s.orderSrv.OrderList(ctx.Request.Context(), &proto.OrderFilterRequest{
		UserId:      userId,
		Status:      int32(status),
		Pages:       int32(pages),
		PagePerNums: int32(pagePerNums),
	})
	if err != nil {
		s.log.Error("OrderList grpc error: ", err)
		HandleGrpcErrorToHttp(err, ctx)
		return
	}
	out := &orderListResponse{
		Total: rsp.Total,
		Data:  make([]*orderResponse, 0, len(rsp.Data)),
	}
	for _, o := range rsp.Data {
		out.Data = append(out.Data, orderToResponse(o))
	}
	ctx.JSON(http.StatusOK, out)
}

// GetOrderDetail 订单详情（带 userId 归属校验，非本人不可见）
func (s *Server) GetOrderDetail(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "invalid id"})
		return
	}
	rsp, err := s.orderSrv.GetOrderDetail(ctx.Request.Context(), &proto.OrderInfoRequest{
		Id:     int32(id),
		UserId: currentUserID(ctx),
	})
	if err != nil {
		s.log.Error("GetOrderDetail grpc error: ", err)
		HandleGrpcErrorToHttp(err, ctx)
		return
	}

	// 生成支付宝url
	client, err := alipay.New(s.cfg.AliPay.AppId, s.cfg.AliPay.PrivateKey, s.cfg.AliPay.IsProduction)
	if err != nil {
		s.log.Error("alipay.New error: ", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"msg": "支付服务初始化失败"})
		return
	}
	err = client.LoadAliPayPublicKey(s.cfg.AliPay.PublicKey)
	if err != nil {
		s.log.Error("client.LoadAliPayPublicKey error: ", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"msg": "支付公钥加载失败"})
		return
	}

	var p = alipay.TradePagePay{}
	p.NotifyURL = s.cfg.AliPay.NotifyUrl
	p.ReturnURL = s.cfg.AliPay.ReturnUrl
	p.Subject = "shop订单支付-" + rsp.OrderSn
	p.OutTradeNo = rsp.OrderSn
	p.TotalAmount = strconv.FormatFloat(float64(rsp.Total), 'f', 2, 64)
	p.ProductCode = "FAST_INSTANT_TRADE_PAY"
	url, err := client.TradePagePay(p)
	if err != nil {
		s.log.Error("client.TradePagePay error: ", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"msg": "支付请求url生成失败"})
		return
	}
	// 单次响应：订单详情 + 支付宝支付链接
	resp := orderToResponse(rsp)
	resp.AlipayUrl = url.String()
	ctx.JSON(http.StatusOK, resp)
}

// UpdateOrderStatus 更新订单状态（支付 / 取消）
func (s *Server) UpdateOrderStatus(ctx *gin.Context) {
	var form UpdateOrderStatusForm
	if err := ctx.ShouldBind(&form); err != nil {
		handleValidatorError(err, ctx)
		return
	}
	_, err := s.orderSrv.UpdateOrderStatus(ctx.Request.Context(), &proto.UpdateOrderStatusInfo{
		OrderSn: form.OrderSn,
		Status:  form.Status,
	})
	if err != nil {
		s.log.Error("UpdateOrderStatus grpc error: ", err)
		HandleGrpcErrorToHttp(err, ctx)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"msg": "更新成功"})
}

// alipayClient 按配置构造并加载公钥的支付宝客户端，notify / return 回调共用。
func (s *Server) alipayClient() (*alipay.Client, error) {
	client, err := alipay.New(s.cfg.AliPay.AppId, s.cfg.AliPay.PrivateKey, s.cfg.AliPay.IsProduction)
	if err != nil {
		return nil, err
	}
	if err := client.LoadAliPayPublicKey(s.cfg.AliPay.PublicKey); err != nil {
		return nil, err
	}
	return client, nil
}

// AlipayNotify 支付宝支付结果异步通知（POST），由支付宝服务器主动回调。
func (s *Server) AlipayNotify(ctx *gin.Context) {
	client, err := s.alipayClient()
	if err != nil {
		s.log.Error("alipayClient error: ", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"msg": "支付服务初始化失败"})
		return
	}

	ctx.Request.ParseForm()

	noti, err := client.DecodeNotification(context.Background(), ctx.Request.Form)
	if err != nil {
		s.log.Error("AlipayNotify DecodeNotification error: ", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "支付通知解析失败"})
		return
	}
	// 将支付宝交易状态映射为订单状态码：1-待支付 2-已支付 3-已取消
	// WAIT_BUYER_PAY 仍属待支付，无需更新，直接回 success 让支付宝停止重试。
	status, ok := alipayTradeStatusToOrderStatus(noti.TradeStatus)
	if !ok {
		s.log.Infof("AlipayNotify 忽略状态 %q，订单 %s 无需更新", noti.TradeStatus, noti.OutTradeNo)
		ctx.String(http.StatusOK, "success")
		return
	}
	_, err = s.orderSrv.UpdateOrderStatus(ctx.Request.Context(), &proto.UpdateOrderStatusInfo{
		OrderSn: noti.OutTradeNo,
		Status:  status,
	})
	if err != nil {
		s.log.Error("AlipayNotify UpdateOrderStatus error: ", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"msg": "订单状态更新失败"})
		return
	}
	ctx.String(http.StatusOK, "success")
}

// AlipayReturn 支付宝支付完成后的同步跳转（GET），用户浏览器回到此端点。
// 仅校验签名并展示结果，订单状态以 AlipayNotify 异步通知为准。
func (s *Server) AlipayReturn(ctx *gin.Context) {
	client, err := s.alipayClient()
	if err != nil {
		s.log.Error("alipayClient error: ", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"msg": "支付服务初始化失败"})
		return
	}

	// GET 回调参数在 query string，ParseForm 后 ctx.Request.Form 即可读取。
	ctx.Request.ParseForm()

	noti, err := client.DecodeNotification(context.Background(), ctx.Request.Form)
	if err != nil {
		s.log.Error("AlipayReturn DecodeNotification error: ", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "支付回跳校验失败"})
		return
	}

	ctx.Header("Content-Type", "text/html; charset=utf-8")
	ctx.String(http.StatusOK, `<!DOCTYPE html>
<html lang="zh-CN"><head><meta charset="utf-8"><title>支付结果</title></head>
<body>
  <h2>支付完成</h2>
  <p>订单号：%s</p>
  <p>交易状态：%s</p>
  <p><a href="/">返回首页</a></p>
</body></html>`, noti.OutTradeNo, noti.TradeStatus)
}

// alipayTradeStatusToOrderStatus 把支付宝异步通知的交易状态映射为订单状态码。
// 返回 ok=false 表示该状态无需推进订单（如待付款），调用方应直接回 success。
func alipayTradeStatusToOrderStatus(ts alipay.TradeStatus) (int32, bool) {
	switch ts {
	case alipay.TradeStatusSuccess, alipay.TradeStatusFinished:
		return 2, true // 已支付
	case alipay.TradeStatusClosed:
		return 3, true // 已取消（超时关闭 / 全额退款）
	default:
		return 0, false // WAIT_BUYER_PAY 等无需更新
	}
}
