package form

type BannerForm struct {
	Index int    `form:"index" json:"index" binding:"required,min=0"`
	Image string `form:"image" json:"image" binding:"required,url"`
	Url   string `form:"url" json:"url" binding:"required,url"`
}
