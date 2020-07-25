package dto

import (
	"github.com/gin-gonic/gin"
	"go-micro-gateway/go_gateway/public"
	"time"
)

type AdminInfoOutput struct {
	ID           int       `json:"id"`
	Name         string    `json:"name"`
	LoginTime    time.Time `json:"login_time"`
	Avatar       string    `json:"avatar"`
	Introduction string    `json:"introduction"`
	Roles        []string  `json:"roles"`
}

type ChangePwdInput struct {
	Password string `json:"password" form:"password" comment:"密码" example:"123123" validate:"required"` // 修改密码
}

func (param *ChangePwdInput) BindValidParam(ctx *gin.Context) error {
	return public.DefaultGetValidParams(ctx, param)
}
