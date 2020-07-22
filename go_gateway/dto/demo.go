package dto

import (
	"github.com/gin-gonic/gin"
	"go-micro-gateway/go_gateway/public"
)

type DemoInput struct {
	Name   string `json:"name" form:"name" comment:"姓名" example:"姓名" validate:"required"`
	Age    int64  `json:"age" form:"age" comment:"年龄" example:"20" validate:"required"`
	Passwd string `json:"passwd" form:"passwd" comment:"密码" example:"123456" validate:"required"`
}

func (params *DemoInput) BindingValidParams(c *gin.Context) error {
	return public.DefaultGetValidParams(c, params)
}
