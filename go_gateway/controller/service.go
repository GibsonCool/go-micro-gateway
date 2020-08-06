package controller

import (
	"fmt"
	"github.com/e421083458/golang_common/lib"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"go-micro-gateway/go_gateway/dao"
	"go-micro-gateway/go_gateway/dto"
	"go-micro-gateway/go_gateway/middleware"
	"go-micro-gateway/go_gateway/public"
	"strconv"
	"strings"
)

func ServiceRegister(group *gin.RouterGroup) {
	service := &ServiceController{}
	group.GET("/service_list", service.ServiceList)
	group.GET("/service_delete", service.ServiceDelete)
	group.POST("/service_add_http", service.ServiceAddHTTP)
}

type ServiceController struct {
}

// @Summary 服务列表
// @Description 获取服务列表
// @Tags 服务管理
// @Produce  json
// @Param info query string false "关键词"
// @Param page_no query int true "页数"
// @Param page_size query int true "每页个数"
// @Success 200 {object} middleware.Response{data=dto.ServiceListOutput} "success"
// @Router /service/service_list [get]
func (c *ServiceController) ServiceList(ctx *gin.Context) {
	params := &dto.ServiceListInput{}
	if err := params.BindValidParam(ctx); err != nil {
		middleware.ResponseError(ctx, 2000, err)
		return
	}

	gDB, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(ctx, 2001, err)
		return
	}

	serviceInfo := &dao.ServiceInfo{}
	list, total, err := serviceInfo.PageList(ctx, gDB, params)
	if err != nil {
		middleware.ResponseError(ctx, 2002, err)
		return
	}

	// 格式化输出数据
	var outList []dto.ServiceListItemOutput
	for _, listItem := range list {
		serviceDetail, err := listItem.ServiceDetail(ctx, gDB, &listItem)
		if err != nil {
			middleware.ResponseError(ctx, 2003, err)
			return
		}

		//1、http后缀接入 clusterIP+clusterPort+path
		//2、http域名接入 domain
		//3、tcp、grpc接入 clusterIP+servicePort
		serviceAddr := "unknow"
		clusterIP := lib.GetStringConf("base.cluster.cluster_ip")
		clusterPort := lib.GetStringConf("base.cluster.cluster_port")
		clusterSSLPort := lib.GetStringConf("base.cluster.cluster_ssl_port")
		if serviceDetail.Info.LoadType == public.LoadTypeHTTP &&
			serviceDetail.HTTPRule.RuleType == public.HTTPRuleTypePrefixURL &&
			serviceDetail.HTTPRule.NeedHttps == 1 {
			serviceAddr = fmt.Sprintf("%s:%s%s", clusterIP, clusterSSLPort, serviceDetail.HTTPRule.Rule)
		}
		if serviceDetail.Info.LoadType == public.LoadTypeHTTP &&
			serviceDetail.HTTPRule.RuleType == public.HTTPRuleTypePrefixURL &&
			serviceDetail.HTTPRule.NeedHttps == 0 {
			serviceAddr = fmt.Sprintf("%s:%s%s", clusterIP, clusterPort, serviceDetail.HTTPRule.Rule)
		}
		if serviceDetail.Info.LoadType == public.LoadTypeHTTP &&
			serviceDetail.HTTPRule.RuleType == public.HTTPRuleTypeDomain {
			serviceAddr = serviceDetail.HTTPRule.Rule
		}
		if serviceDetail.Info.LoadType == public.LoadTypeTCP {
			serviceAddr = fmt.Sprintf("%s:%d", clusterIP, serviceDetail.TCPRule.Port)
		}
		if serviceDetail.Info.LoadType == public.LoadTypeGRPC {
			serviceAddr = fmt.Sprintf("%s:%d", clusterIP, serviceDetail.GRPCRule.Port)
		}
		ipList := serviceDetail.LoadBalance.GetIPListByModel()

		outItem := dto.ServiceListItemOutput{
			Id:          listItem.ID,
			ServiceName: listItem.ServiceName,
			ServiceDesc: listItem.ServiceDesc,
			ServiceAddr: serviceAddr,
			Qps:         0,
			Qpd:         0,
			TotalNode:   len(ipList),
		}

		outList = append(outList, outItem)
	}

	out := &dto.ServiceListOutput{
		Total: strconv.FormatInt(total, 10),
		List:  outList,
	}
	middleware.ResponseSuccess(ctx, out)
}

// @Summary 服务删除
// @Description 服务删除
// @Tags 服务管理
// @Produce  json
// @Param id query string true "服务ID"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /service/service_delete [get]
func (c *ServiceController) ServiceDelete(ctx *gin.Context) {
	params := &dto.ServiceDelete{}
	if err := params.BindValidParam(ctx); err != nil {
		middleware.ResponseError(ctx, 2000, err)
		return
	}

	gDB, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(ctx, 2001, err)
		return
	}

	serviceInfo := &dao.ServiceInfo{ID: params.ID}
	serviceInfo, err = serviceInfo.Find(ctx, gDB, serviceInfo)
	if err != nil {
		middleware.ResponseError(ctx, 2002, err)
		return
	}

	serviceInfo.IsDelete = public.IsDelete
	if err := serviceInfo.Save(ctx, gDB); err != nil {
		middleware.ResponseError(ctx, 2003, err)
		return
	}
	middleware.ResponseSuccess(ctx, "删除成功")
}

// @Summary 添加HTTP服务
// @Description 添加HTTP服务
// @Tags 服务管理
// @Accept  json
// @Produce  json
// @Param body body dto.ServiceAddHTTPInput true "body"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /service/service_add_http [POST]
func (c *ServiceController) ServiceAddHTTP(ctx *gin.Context) {
	params := &dto.ServiceAddHTTPInput{}
	if err := params.BindValidParam(ctx); err != nil {
		middleware.ResponseError(ctx, 2000, err)
		return
	}

	gDB, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(ctx, 2001, err)
		return
	}
	serviceInfo := &dao.ServiceInfo{ServiceName: params.ServiceName}
	if _, err = serviceInfo.Find(ctx, gDB, serviceInfo); err == nil {
		middleware.ResponseError(ctx, 2002, errors.New("服务已经存在"))
		return
	}

	httpUrl := &dao.HttpRule{RuleType: params.RuleType, Rule: params.Rule}
	if httpUrl, err = httpUrl.Find(ctx, gDB, httpUrl); err == nil {
		middleware.ResponseError(ctx, 2003, errors.New("服务接入前缀或域名已存在"))
		return
	}

	if len(strings.Split(params.IpList, ",")) != len(strings.Split(params.WeightList, ",")) {
		middleware.ResponseError(ctx, 2004, errors.New("IP列表与权重列表数量不一致"))
		return
	}

	// 数据插入涉及多张表，开启事物
	gDB = gDB.Begin()
	serviceModel := &dao.ServiceInfo{
		ServiceName: params.ServiceName,
		ServiceDesc: params.ServiceDesc,
	}
	if err = serviceModel.Save(ctx, gDB); err != nil {
		gDB.Rollback()
		middleware.ResponseError(ctx, 2005, err)
		return
	}

	// serviceModel.ID
	httpRule := &dao.HttpRule{
		ServiceID:      serviceModel.ID,
		RuleType:       params.RuleType,
		Rule:           params.Rule,
		NeedHttps:      params.NeedHttps,
		NeedStripUri:   params.NeedStripUri,
		NeedWebsocket:  params.NeedWebsocket,
		UrlRewrite:     params.UrlRewrite,
		HeaderTransfor: params.HeaderTransfor,
	}
	if err := httpRule.Save(ctx, gDB); err != nil {
		gDB.Rollback()
		middleware.ResponseError(ctx, 2006, err)
		return
	}

	accessControl := &dao.AccessControl{
		ServiceID:         serviceModel.ID,
		OpenAuth:          params.OpenAuth,
		BlackList:         params.BlackList,
		WhiteList:         params.WhiteList,
		ClientIPFlowLimit: params.ClientipFlowLimit,
		ServiceFlowLimit:  params.ServiceFlowLimit,
	}
	if err := accessControl.Save(ctx, gDB); err != nil {
		gDB.Rollback()
		middleware.ResponseError(ctx, 2007, err)
		return
	}

	loadbalance := &dao.LoadBalance{
		ServiceID:              serviceModel.ID,
		RoundType:              params.RoundType,
		IpList:                 params.IpList,
		WeightList:             params.WeightList,
		UpstreamConnectTimeout: params.UpstreamConnectTimeout,
		UpstreamHeaderTimeout:  params.UpstreamHeaderTimeout,
		UpstreamIdleTimeout:    params.UpstreamIdleTimeout,
		UpstreamMaxIdle:        params.UpstreamMaxIdle,
	}
	if err := loadbalance.Save(ctx, gDB); err != nil {
		gDB.Rollback()
		middleware.ResponseError(ctx, 2008, err)
		return
	}

	gDB.Commit()
	middleware.ResponseSuccess(ctx, "添加成功")
}
