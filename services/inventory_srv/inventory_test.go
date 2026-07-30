package main

import (
	"context"
	"fmt"
	"os"
	"testing"

	"shop/pkg/proto"

	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

func TestNormalizeOrderGoods(t *testing.T) {
	tests := []struct {
		name    string
		goods   []*proto.OrderGoodsDetail
		want    []stockDeductionGoods
		wantErr bool
	}{
		{
			name: "aggregates duplicates and sorts",
			goods: []*proto.OrderGoodsDetail{
				{GoodsId: 20, Num: 1},
				{GoodsId: 10, Num: 2},
				{GoodsId: 20, Num: 3},
			},
			want: []stockDeductionGoods{{GoodsID: 10, Num: 2}, {GoodsID: 20, Num: 4}},
		},
		{
			name:    "empty goods",
			wantErr: true,
		},
		{
			name:    "invalid goods id",
			goods:   []*proto.OrderGoodsDetail{{GoodsId: 0, Num: 1}},
			wantErr: true,
		},
		{
			name:    "invalid quantity",
			goods:   []*proto.OrderGoodsDetail{{GoodsId: 1, Num: 0}},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := normalizeOrderGoods(tt.goods)
			if (err != nil) != tt.wantErr {
				t.Fatalf("normalizeOrderGoods() error = %v, wantErr %t", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if len(got) != len(tt.want) {
				t.Fatalf("normalizeOrderGoods() = %+v, want %+v", got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Fatalf("normalizeOrderGoods()[%d] = %+v, want %+v", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestRebackDetailIsIdempotent(t *testing.T) {
	db := setupInventoryTestDB(t)
	if err := db.Create(&Inventory{GoodsID: 10, Stocks: 8}).Error; err != nil {
		t.Fatalf("create inventory: %v", err)
	}
	deduction := StockDeduction{OrderSn: "order-1", Status: stockDeductionStatusDeducted}
	if err := db.Create(&deduction).Error; err != nil {
		t.Fatalf("create deduction: %v", err)
	}
	if err := db.Create(&StockDeductionItem{
		StockSellDetailID: deduction.ID,
		GoodsID:           10,
		Num:               3,
	}).Error; err != nil {
		t.Fatalf("create deduction item: %v", err)
	}

	server := NewInventoryServer(db, nil, zap.NewNop().Sugar())
	req := &proto.OrderStockDetail{OrderSn: "order-1"}
	start := make(chan struct{})
	errs := make(chan error, 2)
	for range 2 {
		go func() {
			<-start
			_, err := server.RebackDetail(context.Background(), req)
			errs <- err
		}()
	}
	close(start)
	for range 2 {
		if err := <-errs; err != nil {
			t.Fatalf("concurrent RebackDetail() error = %v", err)
		}
	}

	if _, err := server.RebackDetail(t.Context(), req); err != nil {
		t.Fatalf("repeated RebackDetail() error = %v", err)
	}

	var inv Inventory
	if err := db.Where("goods_id = ?", 10).First(&inv).Error; err != nil {
		t.Fatalf("query inventory: %v", err)
	}
	if inv.Stocks != 11 {
		t.Fatalf("inventory stocks = %d, want 11", inv.Stocks)
	}
	if err := db.Where("order_sn = ?", "order-1").First(&deduction).Error; err != nil {
		t.Fatalf("query deduction: %v", err)
	}
	if deduction.Status != stockDeductionStatusReturned {
		t.Fatalf("deduction status = %d, want %d", deduction.Status, stockDeductionStatusReturned)
	}
}

func setupInventoryTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dbName := fmt.Sprintf("shop_inventory_srv_test_%d", os.Getpid())
	host := inventoryTestEnv("SHOP_TEST_MYSQL_HOST", "127.0.0.1:3306")
	user := inventoryTestEnv("SHOP_TEST_MYSQL_USER", "root")
	pass := inventoryTestEnv("SHOP_TEST_MYSQL_PASS", "root123456")
	dsn := func(dbName string) string {
		return fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=true&loc=Local", user, pass, host, dbName)
	}
	cfg := &gorm.Config{
		NamingStrategy:                           schema.NamingStrategy{SingularTable: true},
		DisableForeignKeyConstraintWhenMigrating: true,
		Logger:                                   logger.Default.LogMode(logger.Silent),
	}

	serverDB, err := gorm.Open(mysql.Open(dsn("")), cfg)
	if err != nil {
		t.Skipf("mysql unavailable, skipping integration test: %v", err)
	}
	if err := serverDB.Exec("DROP DATABASE IF EXISTS `" + dbName + "`").Error; err != nil {
		t.Fatalf("drop test database: %v", err)
	}
	if err := serverDB.Exec("CREATE DATABASE `" + dbName + "` DEFAULT CHARACTER SET utf8mb4").Error; err != nil {
		t.Fatalf("create test database: %v", err)
	}
	if sqlDB, err := serverDB.DB(); err == nil {
		_ = sqlDB.Close()
	}

	db, err := gorm.Open(mysql.Open(dsn(dbName)), cfg)
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	if err := db.AutoMigrate(&Inventory{}, &StockDeduction{}, &StockDeductionItem{}); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}
	t.Cleanup(func() {
		if sqlDB, err := db.DB(); err == nil {
			_ = sqlDB.Close()
		}
		cleanup, err := gorm.Open(mysql.Open(dsn("")), cfg)
		if err == nil {
			_ = cleanup.Exec("DROP DATABASE IF EXISTS `" + dbName + "`").Error
			if sqlDB, dbErr := cleanup.DB(); dbErr == nil {
				_ = sqlDB.Close()
			}
		}
	})
	return db
}

func inventoryTestEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
