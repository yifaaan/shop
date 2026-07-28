package web

type PasswordLoginForm struct {
	Mobile    string `form:"mobile" json:"mobile" binding:"required,mobile"`
	Password  string `form:"password" json:"password" binding:"required,min=6,max=15"`
	Captcha   string `form:"captcha" json:"captcha" binding:"required,len=5"`
	CaptchaId string `form:"captcha_id" json:"captcha_id" binding:"required"`
}

type SendSmsForm struct {
	Mobile string `form:"mobile" json:"mobile" binding:"required,mobile"`
	Type   int    `form:"type" json:"type" binding:"required,oneof=1 2"` // 1:注册 2:登录
}

type RegisterForm struct {
	Mobile   string `form:"mobile" json:"mobile" binding:"required,mobile"`
	Password string `form:"password" json:"password" binding:"required,min=6,max=15"`
	Code     string `form:"code" json:"code" binding:"required,min=5"`
}

// UpdateUserForm 更新个人资料（mobile 由后端 UpdateUser RPC 不支持修改，忽略）。
type UpdateUserForm struct {
	NickName string `form:"name" json:"name" binding:"required"`
	Gender   string `form:"gender" json:"gender"`
	Birthday string `form:"birthday" json:"birthday"` // YYYY-MM-DD
}
