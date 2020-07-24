package dto

import (
	"github.com/gin-gonic/gin"
	"go-micro-gateway/go_gateway/public"
	"time"
)

/*
	validate  默认的为v9参数绑定解析的 tag 标记字段
	comment   middleware.TranslationMiddleware 中我们自己定义注册的 tag 标记字段
	example	  go-swagger 中用于文档生成参数值示例解析用的 tag 标记字段
*/
type AdminLoginInput struct {
	UserName string `json:"username" form:"username" comment:"管理员用户名" example:"admin" validate:"required,valid_username"` //管理员用户名
	Password string `json:"password" form:"password" comment:"密码" example:"123456" validate:"required"`                   //密码
}

func (param *AdminLoginInput) BindValidParam(c *gin.Context) error {
	return public.DefaultGetValidParams(c, param)
}

type AdminLoginOutput struct {
	Token string `json:"token" form:"token" comment:"管理员token" example:"admin_tests" validate:""` // token
}

type AdminSessionInfo struct {
	ID        int       `json:"id"`
	UserName  string    `json:"user_name"`
	LoginTime time.Time `json:"login_time"`
}
