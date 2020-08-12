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

func AppRegister(group *gin.RouterGroup) {
	admin := AppController{}
	group.GET("/app_list", admin.AppList)
	group.GET("/app_detail", admin.AppDetail)
	group.GET("/app_stat", admin.AppStat)
	group.GET("/app_delete", admin.AppDelete)
	group.POST("/app_add", admin.AppAdd)
	group.POST("/app_update", admin.AppUpdate)

}

type AppController struct {
}

// @Summary 租户列表
// @Description 获取租户列表
// @Tags 租户管理
// @Produce  json
// @Param info query string false "关键词"
// @Param page_no query int true "页数"
// @Param page_size query int true "每页个数"
// @Success 200 {object} middleware.Response{data=dto.APPListOutput} "success"
// @Router /app/app_list [get]
func (c *AppController) AppList(ctx *gin.Context) {
	params := &dto.APPListInput{}
	if err := params.BindValidParam(ctx); err != nil {
		middleware.ResponseError(ctx, 2000, err)
		return
	}
	gDB := lib.GORMDefaultPool

	info := &dao.App{}
	list, total, err := info.APPList(ctx, gDB, params)
	if err != nil {
		middleware.ResponseError(ctx, 2001, err)
		return
	}

	var outputList []dto.APPListItemOutput
	for _, item := range list {
		var realQps int64 = 0
		var realQpd int64 = 0
		outputList = append(outputList, dto.APPListItemOutput{
			ID:       item.ID,
			AppID:    item.AppID,
			Name:     item.Name,
			Secret:   item.Secret,
			WhiteIPS: item.WhiteIPS,
			Qpd:      item.ID,
			Qps:      item.ID,
			RealQpd:  realQpd,
			RealQps:  realQps,
		})
	}

	output := dto.APPListOutput{
		List:  outputList,
		Total: total,
	}
	middleware.ResponseSuccess(ctx, output)
	return
}

// @Summary 获取租户详情
// @Description 获取租户详情
// @Tags 租户管理
// @Accept  json
// @Produce  json
// @Param id query string true "租户ID"
// @Success 200 {object} middleware.Response{data=dao.App} "success"
// @Router /app/app_detail [get]
func (c *AppController) AppDetail(ctx *gin.Context) {
	params := &dto.APPDetailInput{}
	if err := params.BindValidParam(ctx); err != nil {
		middleware.ResponseError(ctx, 2000, err)
		return
	}
	gDB := lib.GORMDefaultPool

	info := &dao.App{ID: params.ID}
	info, err := info.Find(ctx, gDB, info)
	if err != nil {
		middleware.ResponseError(ctx, 2001, err)
		return
	}
	middleware.ResponseSuccess(ctx, info)
	return
}

// @Summary 租户统计
// @Description 租户统计
// @Tags 租户管理
// @Accept  json
// @Produce  json
// @Param id query string true "租户ID"
// @Success 200 {object} middleware.Response{data=dto.StatisticsOutput} "success"
// @Router /app/app_stat [get]
func (c *AppController) AppStat(ctx *gin.Context) {
	params := &dto.APPDetailInput{}
	if err := params.BindValidParam(ctx); err != nil {
		middleware.ResponseError(ctx, 2000, err)
		return
	}

	var todayList []int64
	for i := 0; i < time.Now().Hour(); i++ {
		todayList = append(todayList, 0)
	}

	var yesTodayList []int64
	for i := 0; i < 23; i++ {
		yesTodayList = append(yesTodayList, 0)
	}

	middleware.ResponseSuccess(ctx, &dto.StatisticsOutput{
		Yesterday: yesTodayList,
		Today:     todayList,
	})
}

// @Summary 租户添加
// @Description 租户添加
// @Tags 租户管理
// @Accept  json
// @Produce  json
// @Param body body dto.APPAddHttpInput true "body"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /app/app_add [post]
func (c *AppController) AppAdd(ctx *gin.Context) {
	params := &dto.APPAddHttpInput{}
	err := params.BindValidParam(ctx)
	if err != nil {
		middleware.ResponseError(ctx, 2000, err)
		return
	}
	gDB := lib.GORMDefaultPool
	info := &dao.App{AppID: params.AppID}
	if _, err := info.Find(ctx, gDB, info); err == nil {
		middleware.ResponseError(ctx, 2001, errors.New("租户ID被占用，请重新输入"))
		return
	}

	if params.Secret == "" {
		params.Secret = public.MD5(params.AppID)
	}
	info.Secret = params.Secret
	info.Name = params.Name
	info.WhiteIPS = params.WhiteIPS
	info.Qps = params.Qps
	info.Qpd = params.Qpd
	if err := info.Save(ctx, gDB); err != nil {
		middleware.ResponseError(ctx, 2002, err)
		return
	}

	middleware.ResponseSuccess(ctx, "添加成功")
	return
}

// @Summary 更新租户详情
// @Description 更新租户详情
// @Tags 租户管理
// @Accept  json
// @Produce  json
// @Param body body dto.APPUpdateHttpInput true "body"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /app/app_update [post]
func (c *AppController) AppUpdate(ctx *gin.Context) {
	params := &dto.APPUpdateHttpInput{}
	err := params.BindValidParam(ctx)
	if err != nil {
		middleware.ResponseError(ctx, 2000, err)
		return
	}
	gDB := lib.GORMDefaultPool
	search := &dao.App{ID: params.ID}
	info, err := search.Find(ctx, gDB, search)
	if err != nil {
		middleware.ResponseError(ctx, 2001, err)
		return
	}

	if params.Secret == "" {
		params.Secret = public.MD5(params.AppID)
	}
	info.AppID = params.AppID
	info.Secret = params.Secret
	info.Name = params.Name
	info.WhiteIPS = params.WhiteIPS
	info.Qps = params.Qps
	info.Qpd = params.Qpd
	if err := info.Save(ctx, gDB); err != nil {
		middleware.ResponseError(ctx, 2002, err)
		return
	}

	middleware.ResponseSuccess(ctx, "修改成功")
	return
}

// @Summary 删除租户详情
// @Description 删除租户详情
// @Tags 租户管理
// @Accept  json
// @Produce  json
// @Param id query string true "租户ID"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /app/app_delete [get]
func (c *AppController) AppDelete(ctx *gin.Context) {
	params := &dto.APPDetailInput{}
	if err := params.BindValidParam(ctx); err != nil {
		middleware.ResponseError(ctx, 2000, err)
		return
	}
	gDB := lib.GORMDefaultPool

	info := &dao.App{ID: params.ID}
	info, err := info.Find(ctx, gDB, info)
	if err != nil {
		middleware.ResponseError(ctx, 2001, err)
		return
	}
	if info.IsDelete == public.IsDelete {
		middleware.ResponseError(ctx, 2002, errors.New("该用户已删除"))
		return
	}
	info.IsDelete = public.IsDelete
	if err = info.Save(ctx, gDB); err != nil {
		middleware.ResponseError(ctx, 2003, err)
		return
	}

	middleware.ResponseSuccess(ctx, "删除成功")
	return
}
