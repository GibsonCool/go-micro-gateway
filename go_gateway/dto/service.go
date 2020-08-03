package dto

import (
	"github.com/gin-gonic/gin"
	"go-micro-gateway/go_gateway/public"
)

type ServiceListInput struct {
	Info     string `json:"info" form:"info" comment:"关键词" example:"test" validate:""`                  // 关键词
	PageNo   int    `json:"page_no" form:"page_no" comment:"页数" example:"1" validate:"required"`        // 页数
	PageSize int    `json:"page_size" form:"page_size" comment:"每页条数" example:"20" validate:"required"` // 每页条数
}

func (param *ServiceListInput) BindValidParam(ctx *gin.Context) error {
	return public.DefaultGetValidParams(ctx, param)
}

type ServiceListOutput struct {
	Total string                  `json:"total" form:"total" comment:"总数" example:"" validate:""` // 总数
	List  []ServiceListItemOutput `json:"list" form:"list" comment:"数据列表" example:"" validate:""` // 总数
}

type ServiceListItemOutput struct {
	Id          int64  `json:"id" form:"id"`                     // id
	ServiceName string `json:"service_name" form:"service_name"` //服务名称
	ServiceDesc string `json:"service_desc" form:"service_desc"` //服务描述
	LoadType    int    `json:"load_type" form:"load_type"`       //类型
	ServiceAddr string `json:"service_addr" form:"service_addr"` //服务地址
	Qps         int64  `json:"qps" form:"qps"`                   //qps
	Qpd         int64  `json:"qpd" form:"qpd"`                   //qpd
	TotalNode   int    `json:"total_node" form:"total_node"`     //节点数
}
