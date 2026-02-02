package model

// 库存表
type Inventory struct {
	BaseModel
	Good    int32 `gorm:"type:int;not null;index:idx_inventory_good,unique;comment:'商品ID'"`
	Stock   int32 `gorm:"type:int;not null;comment:'库存数量'"`
	Version int32 `gorm:"type:int;not null;comment:'版本号'"` // 分布式锁的乐观锁
}

// type InventoryHistory struct {
// 	BaseModel
// 	user   int32 `gorm:"type:int;not null;comment:'用户ID'"`
// 	good   int32 `gorm:"type:int;not null;comment:'商品ID'"`
// 	nums   int32 `gorm:"type:int;not null;comment:'扣减数量'"`
// 	order  int32 `gorm:"type:int;not null;comment:'订单ID'"`
// 	status int32 `gorm:"type:int;not null;comment:'状态,1:预扣减,2:已支付后扣减'"`
// }
