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
	"time"
)

func ServiceRegister(group *gin.RouterGroup) {
	service := &ServiceController{}
	group.GET("/service_list", service.ServiceList)
	group.GET("/service_delete", service.ServiceDelete)
	group.POST("/service_add_http", service.ServiceAddHTTP)
	group.GET("/service_detail", service.ServiceDetail)
	group.GET("/service_stat", service.ServiceStat)
	group.POST("/service_update_http", service.ServiceUpdateHTTP)

	group.POST("/service_add_tcp", service.ServiceAddTcp)
	group.POST("/service_update_tcp", service.ServiceUpdateTcp)
	group.POST("/service_add_grpc", service.ServiceAddGrpc)
	group.POST("/service_update_grpc", service.ServiceUpdateGrpc)
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

// @Summary 修改HTTP服务
// @Description 修改HTTP服务
// @Tags 服务管理
// @Accept  json
// @Produce  json
// @Param body body dto.ServiceUpdateInput true "body"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /service/service_update_http [POST]
func (c *ServiceController) ServiceUpdateHTTP(ctx *gin.Context) {
	params := &dto.ServiceUpdateInput{}
	if err := params.BindValidParam(ctx); err != nil {
		middleware.ResponseError(ctx, 2000, err)
		return
	}

	gDB, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(ctx, 2001, err)
		return
	}
	if len(strings.Split(params.IpList, ",")) != len(strings.Split(params.WeightList, ",")) {
		middleware.ResponseError(ctx, 2002, errors.New("IP列表与权重列表数量不一致"))
		return
	}

	gDB = gDB.Begin()
	serviceInfo := &dao.ServiceInfo{ServiceName: params.ID}
	serviceInfo, err = serviceInfo.Find(ctx, gDB, serviceInfo)
	if err != nil {
		gDB.Rollback()
		middleware.ResponseError(ctx, 2003, errors.New("服务未查询到："+err.Error()))
		return
	}

	serviceDetail, err := serviceInfo.ServiceDetail(ctx, gDB, serviceInfo)
	if err != nil {
		gDB.Rollback()
		middleware.ResponseError(ctx, 2004, errors.New("服务详情未查询到："+err.Error()))
		return
	}

	info := serviceDetail.Info
	info.ServiceName = params.ServiceName
	info.ServiceDesc = params.ServiceDesc
	if err := info.Save(ctx, gDB); err != nil {
		gDB.Rollback()
		middleware.ResponseError(ctx, 2005, errors.New("服务描述更新失败："+err.Error()))
		return
	}

	httpRule := serviceDetail.HTTPRule
	httpRule.NeedHttps = params.NeedHttps
	httpRule.NeedStripUri = params.NeedStripUri
	httpRule.NeedWebsocket = params.NeedWebsocket
	httpRule.UrlRewrite = params.UrlRewrite
	httpRule.HeaderTransfor = params.HeaderTransfor
	if err := httpRule.Save(ctx, gDB); err != nil {
		gDB.Rollback()
		middleware.ResponseError(ctx, 2006, err)
		return
	}

	accessControl := serviceDetail.AccessControl
	accessControl.OpenAuth = params.OpenAuth
	accessControl.BlackList = params.BlackList
	accessControl.WhiteList = params.WhiteList
	accessControl.ClientIPFlowLimit = params.ClientipFlowLimit
	accessControl.ServiceFlowLimit = params.ServiceFlowLimit
	if err := accessControl.Save(ctx, gDB); err != nil {
		gDB.Rollback()
		middleware.ResponseError(ctx, 2007, err)
		return
	}

	loadBalance := serviceDetail.LoadBalance
	loadBalance.RoundType = params.RoundType
	loadBalance.IpList = params.IpList
	loadBalance.WeightList = params.WeightList
	loadBalance.UpstreamConnectTimeout = params.UpstreamConnectTimeout
	loadBalance.UpstreamHeaderTimeout = params.UpstreamHeaderTimeout
	loadBalance.UpstreamIdleTimeout = params.UpstreamIdleTimeout
	loadBalance.UpstreamMaxIdle = params.UpstreamMaxIdle
	if err := loadBalance.Save(ctx, gDB); err != nil {
		gDB.Rollback()
		middleware.ResponseError(ctx, 2008, err)
		return
	}

	gDB.Commit()
	middleware.ResponseSuccess(ctx, "修改成功")
}

// @Summary 获取服务详情
// @Description 获取服务详情
// @Tags 服务管理
// @Accept  json
// @Produce  json
// @Param id query string true "服务ID"
// @Success 200 {object} middleware.Response{data=dao.ServiceDetail} "success"
// @Router /service/service_detail [get]
func (c *ServiceController) ServiceDetail(ctx *gin.Context) {
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
		gDB.Rollback()
		middleware.ResponseError(ctx, 2002, errors.New("服务未查询到："+err.Error()))
		return
	}

	serviceDetail, err := serviceInfo.ServiceDetail(ctx, gDB, serviceInfo)
	if err != nil {
		gDB.Rollback()
		middleware.ResponseError(ctx, 2003, errors.New("服务详情未查询到："+err.Error()))
		return
	}

	middleware.ResponseSuccess(ctx, serviceDetail)
}

// @Summary 服务统计
// @Description 服务统计
// @Tags 服务管理
// @Accept  json
// @Produce  json
// @Param id query string true "服务ID"
// @Success 200 {object} middleware.Response{data=dto.ServiceStatOutput} "success"
// @Router /service/service_stat [get]
func (c *ServiceController) ServiceStat(ctx *gin.Context) {
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
	//if err != nil {
	//	gDB.Rollback()
	//	middleware.ResponseError(ctx, 2002, errors.New("服务未查询到："+err.Error()))
	//	return
	//}

	//serviceDetail, err := serviceInfo.ServiceDetail(ctx, gDB, serviceInfo)
	//if err != nil {
	//	gDB.Rollback()
	//	middleware.ResponseError(ctx, 2003, errors.New("服务详情未查询到："+err.Error()))
	//	return
	//}

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

// @Summary 添加TCP服务
// @Description 添加TCP服务
// @Tags 服务管理
// @Accept  json
// @Produce  json
// @Param body body dto.ServiceAddTcpInput true "body"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /service/service_add_tcp [POST]
func (c *ServiceController) ServiceAddTcp(ctx *gin.Context) {
	params := &dto.ServiceAddTcpInput{}
	err := params.BindValidParam(ctx)
	if err != nil {
		middleware.ResponseError(ctx, 2000, err)
		return
	}

	serviceInfo := &dao.ServiceInfo{ServiceName: params.ServiceName}
	if _, err = serviceInfo.Find(ctx, lib.GORMDefaultPool, serviceInfo); err == nil {
		middleware.ResponseError(ctx, 2002, errors.New("服务已被占用，请重新输入"))
		return
	}

	tcpUrl := &dao.TcpRule{Port: params.Port}
	if tcpUrl, err = tcpUrl.Find(ctx, lib.GORMDefaultPool, tcpUrl); err == nil {
		middleware.ResponseError(ctx, 2003, errors.New("tcp服务端口被占用，请重新输入"))
		return
	}

	grpcUrl := &dao.GrpcRule{Port: params.Port}
	if grpcUrl, err = grpcUrl.Find(ctx, lib.GORMDefaultPool, grpcUrl); err == nil {
		middleware.ResponseError(ctx, 2004, errors.New("grpc服务端口被占用，请重新输入"))
		return
	}

	if len(strings.Split(params.IpList, ",")) != len(strings.Split(params.WeightList, ",")) {
		middleware.ResponseError(ctx, 2005, errors.New("IP列表与权重列表数量不一致"))
		return
	}

	// 数据插入涉及多张表，开启事物
	gDB := lib.GORMDefaultPool.Begin()
	serviceModel := &dao.ServiceInfo{
		LoadType:    public.LoadTypeTCP,
		ServiceName: params.ServiceName,
		ServiceDesc: params.ServiceDesc,
	}
	if err = serviceModel.Save(ctx, gDB); err != nil {
		gDB.Rollback()
		middleware.ResponseError(ctx, 2006, err)
		return
	}

	// serviceModel.ID
	loadBalance := &dao.LoadBalance{
		ServiceID:  serviceModel.ID,
		RoundType:  params.RoundType,
		IpList:     params.IpList,
		WeightList: params.WeightList,
		ForbidList: params.ForbidList,
	}
	if err := loadBalance.Save(ctx, gDB); err != nil {
		gDB.Rollback()
		middleware.ResponseError(ctx, 2007, err)
		return
	}

	tcpRule := &dao.TcpRule{
		ServiceID: serviceModel.ID,
		Port:      params.Port,
	}
	if err := tcpRule.Save(ctx, gDB); err != nil {
		gDB.Rollback()
		middleware.ResponseError(ctx, 2008, err)
		return
	}

	accessControl := &dao.AccessControl{
		ServiceID:         serviceModel.ID,
		OpenAuth:          params.OpenAuth,
		BlackList:         params.BlackList,
		WhiteList:         params.WhiteList,
		ClientIPFlowLimit: params.ClientIPFlowLimit,
		ServiceFlowLimit:  params.ServiceFlowLimit,
	}
	if err := accessControl.Save(ctx, gDB); err != nil {
		gDB.Rollback()
		middleware.ResponseError(ctx, 2007, err)
		return
	}

	gDB.Commit()
	middleware.ResponseSuccess(ctx, "添加 tcp 服务成功")
}

// @Summary 修改Tcp服务
// @Description 修改Tcp服务
// @Tags 服务管理
// @Accept  json
// @Produce  json
// @Param body body dto.ServiceUpdateTcpInput true "body"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /service/service_update_tcp [POST]
func (c *ServiceController) ServiceUpdateTcp(ctx *gin.Context) {
	params := &dto.ServiceUpdateTcpInput{}
	err := params.BindValidParam(ctx)
	if err != nil {
		middleware.ResponseError(ctx, 2000, err)
		return
	}

	if len(strings.Split(params.IpList, ",")) != len(strings.Split(params.WeightList, ",")) {
		middleware.ResponseError(ctx, 2002, errors.New("IP列表与权重列表数量不一致"))
		return
	}

	gDB := lib.GORMDefaultPool.Begin()
	serviceInfo := &dao.ServiceInfo{ID: params.ID}
	serviceInfo, err = serviceInfo.Find(ctx, gDB, serviceInfo)
	if err != nil {
		gDB.Rollback()
		middleware.ResponseError(ctx, 2003, errors.New("服务未查询到："+err.Error()))
		return
	}

	serviceDetail, err := serviceInfo.ServiceDetail(ctx, gDB, serviceInfo)
	if err != nil {
		gDB.Rollback()
		middleware.ResponseError(ctx, 2004, errors.New("服务详情未查询到："+err.Error()))
		return
	}

	info := serviceDetail.Info
	info.ServiceDesc = params.ServiceDesc
	info.ServiceName = params.ServiceName
	if err := info.Save(ctx, gDB); err != nil {
		gDB.Rollback()
		middleware.ResponseError(ctx, 2005, errors.New("服务描述更新失败："+err.Error()))
		return
	}

	tcpRule := serviceDetail.TCPRule
	tcpRule.Port = params.Port
	if err := tcpRule.Save(ctx, gDB); err != nil {
		gDB.Rollback()
		middleware.ResponseError(ctx, 2006, errors.New("tcp端口更新失败："+err.Error()))
		return
	}

	accessControl := serviceDetail.AccessControl
	accessControl.OpenAuth = params.OpenAuth
	accessControl.BlackList = params.BlackList
	accessControl.WhiteList = params.WhiteList
	accessControl.WhiteHostName = params.WhiteHostName
	accessControl.ClientIPFlowLimit = params.ClientIPFlowLimit
	accessControl.ServiceFlowLimit = params.ServiceFlowLimit
	if err := accessControl.Save(ctx, gDB); err != nil {
		gDB.Rollback()
		middleware.ResponseError(ctx, 2007, err)
		return
	}

	loadBalance := serviceDetail.LoadBalance
	loadBalance.RoundType = params.RoundType
	loadBalance.IpList = params.IpList
	loadBalance.WeightList = params.WeightList
	loadBalance.ForbidList = params.ForbidList
	if err := loadBalance.Save(ctx, gDB); err != nil {
		gDB.Rollback()
		middleware.ResponseError(ctx, 2008, err)
		return
	}

	gDB.Commit()
	middleware.ResponseSuccess(ctx, "tcp修改成功")
}

// @Summary 添加Grpc服务
// @Description 添加Grpc服务
// @Tags 服务管理
// @Accept  json
// @Produce  json
// @Param body body dto.ServiceAddGrpcInput true "body"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /service/service_add_grpc [POST]
func (c *ServiceController) ServiceAddGrpc(ctx *gin.Context) {
	params := &dto.ServiceAddGrpcInput{}
	err := params.BindValidParam(ctx)
	if err != nil {
		middleware.ResponseError(ctx, 2000, err)
		return
	}

	serviceInfo := &dao.ServiceInfo{ServiceName: params.ServiceName}
	if _, err = serviceInfo.Find(ctx, lib.GORMDefaultPool, serviceInfo); err == nil {
		middleware.ResponseError(ctx, 2002, errors.New("服务已被占用，请重新输入"))
		return
	}

	tcpUrl := &dao.TcpRule{Port: params.Port}
	if tcpUrl, err = tcpUrl.Find(ctx, lib.GORMDefaultPool, tcpUrl); err == nil {
		middleware.ResponseError(ctx, 2003, errors.New("tcp服务端口被占用，请重新输入"))
		return
	}

	grpcUrl := &dao.GrpcRule{Port: params.Port}
	if grpcUrl, err = grpcUrl.Find(ctx, lib.GORMDefaultPool, grpcUrl); err == nil {
		middleware.ResponseError(ctx, 2004, errors.New("grpc服务端口被占用，请重新输入"))
		return
	}

	if len(strings.Split(params.IpList, ",")) != len(strings.Split(params.WeightList, ",")) {
		middleware.ResponseError(ctx, 2005, errors.New("IP列表与权重列表数量不一致"))
		return
	}

	// 数据插入涉及多张表，开启事物
	gDB := lib.GORMDefaultPool.Begin()
	serviceModel := &dao.ServiceInfo{
		LoadType:    public.LoadTypeGRPC,
		ServiceName: params.ServiceName,
		ServiceDesc: params.ServiceDesc,
	}
	if err = serviceModel.Save(ctx, gDB); err != nil {
		gDB.Rollback()
		middleware.ResponseError(ctx, 2006, err)
		return
	}

	// serviceModel.ID
	loadBalance := &dao.LoadBalance{
		ServiceID:  serviceModel.ID,
		RoundType:  params.RoundType,
		IpList:     params.IpList,
		WeightList: params.WeightList,
		ForbidList: params.ForbidList,
	}
	if err := loadBalance.Save(ctx, gDB); err != nil {
		gDB.Rollback()
		middleware.ResponseError(ctx, 2007, err)
		return
	}

	grpcRule := &dao.GrpcRule{
		ServiceID:      serviceModel.ID,
		Port:           params.Port,
		HeaderTransfor: params.HeaderTransfor,
	}
	if err := grpcRule.Save(ctx, gDB); err != nil {
		gDB.Rollback()
		middleware.ResponseError(ctx, 2008, err)
		return
	}

	accessControl := &dao.AccessControl{
		ServiceID:         serviceModel.ID,
		OpenAuth:          params.OpenAuth,
		BlackList:         params.BlackList,
		WhiteList:         params.WhiteList,
		ClientIPFlowLimit: params.ClientIPFlowLimit,
		ServiceFlowLimit:  params.ServiceFlowLimit,
	}
	if err := accessControl.Save(ctx, gDB); err != nil {
		gDB.Rollback()
		middleware.ResponseError(ctx, 2009, err)
		return
	}

	gDB.Commit()
	middleware.ResponseSuccess(ctx, "添加 grpc 服务成功")
}

// @Summary 修改Grpc服务
// @Description 修改Grpc服务
// @Tags 服务管理
// @Accept  json
// @Produce  json
// @Param body body dto.ServiceUpdateGrpcInput true "body"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /service/service_update_grpc [POST]
func (c *ServiceController) ServiceUpdateGrpc(ctx *gin.Context) {
	params := &dto.ServiceUpdateGrpcInput{}
	err := params.BindValidParam(ctx)
	if err != nil {
		middleware.ResponseError(ctx, 2000, err)
		return
	}

	if len(strings.Split(params.IpList, ",")) != len(strings.Split(params.WeightList, ",")) {
		middleware.ResponseError(ctx, 2002, errors.New("IP列表与权重列表数量不一致"))
		return
	}

	gDB := lib.GORMDefaultPool.Begin()
	serviceInfo := &dao.ServiceInfo{ID: params.ID}
	serviceInfo, err = serviceInfo.Find(ctx, gDB, serviceInfo)
	if err != nil {
		gDB.Rollback()
		middleware.ResponseError(ctx, 2003, errors.New("服务未查询到："+err.Error()))
		return
	}

	serviceDetail, err := serviceInfo.ServiceDetail(ctx, gDB, serviceInfo)
	if err != nil {
		gDB.Rollback()
		middleware.ResponseError(ctx, 2004, errors.New("服务详情未查询到："+err.Error()))
		return
	}

	info := serviceDetail.Info
	info.ServiceDesc = params.ServiceDesc
	info.ServiceName = params.ServiceName
	if err := info.Save(ctx, gDB); err != nil {
		gDB.Rollback()
		middleware.ResponseError(ctx, 2005, errors.New("服务描述更新失败："+err.Error()))
		return
	}

	grpcRule := serviceDetail.GRPCRule
	grpcRule.Port = params.Port
	grpcRule.HeaderTransfor = params.HeaderTransfor
	if err := grpcRule.Save(ctx, gDB); err != nil {
		gDB.Rollback()
		middleware.ResponseError(ctx, 2006, errors.New("grpc 端口更新失败："+err.Error()))
		return
	}

	accessControl := serviceDetail.AccessControl
	accessControl.OpenAuth = params.OpenAuth
	accessControl.BlackList = params.BlackList
	accessControl.WhiteList = params.WhiteList
	accessControl.WhiteHostName = params.WhiteHostName
	accessControl.ClientIPFlowLimit = params.ClientIPFlowLimit
	accessControl.ServiceFlowLimit = params.ServiceFlowLimit
	if err := accessControl.Save(ctx, gDB); err != nil {
		gDB.Rollback()
		middleware.ResponseError(ctx, 2007, err)
		return
	}

	loadBalance := serviceDetail.LoadBalance
	loadBalance.RoundType = params.RoundType
	loadBalance.IpList = params.IpList
	loadBalance.WeightList = params.WeightList
	loadBalance.ForbidList = params.ForbidList
	if err := loadBalance.Save(ctx, gDB); err != nil {
		gDB.Rollback()
		middleware.ResponseError(ctx, 2008, err)
		return
	}

	gDB.Commit()
	middleware.ResponseSuccess(ctx, "grpc修改成功")
}
