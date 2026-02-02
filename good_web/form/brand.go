package form

type BrandForm struct {
	Name string `form:"name" json:"name" binding:"required,min=2,max=20"`
	Logo string `form:"logo" json:"logo" binding:"required,url"`
}

type CategoryBrandForm struct {
	CategoryId int32 `form:"category_id" json:"category_id" binding:"required,min=1"`
	BrandId    int32 `form:"brand_id" json:"brand_id" binding:"required,min=1"`
}
