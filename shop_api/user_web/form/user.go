package form

type LoginForm struct {
	Mobile   string `form:"mobile" json:"mobile" binding:"required,min=11,max=11"`
	Password string `form:"password" json:"password" binding:"required,min=3,max=11"`
}
