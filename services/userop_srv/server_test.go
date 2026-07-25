package main

import (
	"fmt"
	"os"
	"testing"

	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

const testDBName = "shop_userop_srv_test"

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

// setupTestDB 重建独立测试库并迁移 userop_srv 全部表；MySQL 不可达则 t.Skip。
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
	if err := db.AutoMigrate(&UserFav{}, &Address{}, &Message{}); err != nil {
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

// newTestServer 用测试库装配一个 UserOpServer。
func newTestServer(t *testing.T, db *gorm.DB) *UserOpServer {
	t.Helper()
	return NewUserOpServer(db, zap.NewNop().Sugar())
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// --- helpers shared by domain tests ---

func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// errOf 取 (T, error) 中的 error，配合 must 使用；T 任意（*emptypb.Empty / *UserFavInfoResponse …）。
func errOf[T any](_ T, err error) error { return err }

func count(db *gorm.DB, model any) int64 {
	var n int64
	db.Model(model).Count(&n)
	return n
}

// --- 迁移冒烟 ---

func TestAutoMigrate(t *testing.T) {
	db := setupTestDB(t)
	for _, tbl := range []string{"user_fav", "address", "message"} {
		if !db.Migrator().HasTable(tbl) {
			t.Fatalf("表 %s 未创建", tbl)
		}
	}
}
