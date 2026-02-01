package form

type CategoryForm struct {
	Name           string `form:"name" json:"name" binding:"required,min=2,max=20"`
	Level          int32  `form:"level" json:"level" binding:"required,oneof=1 2 3"`
	ParentCategory int32  `form:"parent_category" json:"parent_category" binding:"required,min=1"`
	IsTab          bool   `form:"is_tab" json:"is_tab" binding:"required"`
}

type UpdateCategoryForm struct {
	Name  string `form:"name" json:"name" binding:"required,min=2,max=20"`
	IsTab bool   `form:"is_tab" json:"is_tab"`
}
