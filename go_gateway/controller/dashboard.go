package controller

import (
	"github.com/e421083458/golang_common/lib"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"go-micro-gateway/go_gateway/dao"
	"go-micro-gateway/go_gateway/dto"
	"go-micro-gateway/go_gateway/middleware"
	"go-micro-gateway/go_gateway/public"
	"time"
)

type DashBoardController struct {
}

func DashBoardRegister(group *gin.RouterGroup) {
	dashBoard := &DashBoardController{}
	group.GET("/panel_group_data", dashBoard.PanelGroupData)
	group.GET("/flow_stat", dashBoard.FlowStat)
	group.GET("/service_stat", dashBoard.ServiceStat)
}

// @Summary 指标统计
// @Description 指标统计
// @Tags 首页大盘
// @Produce  json
// @Success 200 {object} middleware.Response{data=dto.PanelGroupDataOutput} "success"
// @Router /dashboard/panel_group_data [get]
func (d *DashBoardController) PanelGroupData(ctx *gin.Context) {
	gDB := lib.GORMDefaultPool
	serviceInfo := &dao.ServiceInfo{}
	_, serviceNum, err := serviceInfo.PageList(ctx, gDB, &dto.ServiceListInput{PageSize: 1, PageNo: 1})
	if err != nil {
		middleware.ResponseError(ctx, 2001, err)
		return
	}

	info := &dao.App{}
	_, appNum, err := info.APPList(ctx, gDB, &dto.APPListInput{PageNo: 1, PageSize: 1})
	if err != nil {
		middleware.ResponseError(ctx, 2002, err)
		return
	}

	middleware.ResponseSuccess(ctx, &dto.PanelGroupDataOutput{
		ServiceNum:      serviceNum,
		AppNum:          appNum,
		TodayRequestNum: 0,
		CurrentQPS:      0,
	})
}

// @Summary 流量统计
// @Description 流量统计
// @Tags 首页大盘
// @Produce  json
// @Success 200 {object} middleware.Response{data=dto.ServiceStatOutput} "success"
// @Router /dashboard/flow_stat [get]
func (d *DashBoardController) FlowStat(ctx *gin.Context) {
	var todayList []int64
	for i := 0; i < time.Now().Hour(); i++ {
		todayList = append(todayList, 0)
	}

	var yesTodayList []int64
	for i := 0; i < 23; i++ {
		yesTodayList = append(yesTodayList, 0)
	}

	middleware.ResponseSuccess(ctx, &dto.ServiceStatOutput{
		Yesterday: yesTodayList,
		Today:     todayList,
	})
}

// @Summary 服务统计
// @Description 服务统计
// @Tags 首页大盘
// @Produce  json
// @Success 200 {object} middleware.Response{data=dto.DashServiceStatOutput} "success"
// @Router /dashboard/service_stat [get]
func (d *DashBoardController) ServiceStat(ctx *gin.Context) {
	gDB := lib.GORMDefaultPool
	serviceInfo := &dao.ServiceInfo{}

	list, err := serviceInfo.GroupByLoadType(ctx, gDB)
	if err != nil {
		middleware.ResponseError(ctx, 2001, err)
		return
	}

	var legend []string
	for index, value := range list {
		value, ok := public.LoadTypeMap[value.LoadType]
		if !ok {
			middleware.ResponseError(ctx, 2002, errors.New("load_type not found"))
			return
		}
		list[index].Name = value
		legend = append(legend, value)
	}

	middleware.ResponseSuccess(ctx, &dto.DashServiceStatOutput{
		Legend: legend,
		Data:   list,
	})
}
