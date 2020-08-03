package dao

import (
	"github.com/e421083458/gorm"
	"github.com/gin-gonic/gin"
	"go-micro-gateway/go_gateway/dto"
	"go-micro-gateway/go_gateway/public"
	"time"
)

type ServiceInfo struct {
	ID          int64     `json:"id" gorm:"primary_key" description:"自增主键"`
	ServiceName string    `json:"service_name" gorm:"column:service_name" description:"服务器名称"`
	ServiceDesc string    `json:"service_desc" gorm:"column:service_desc" description:"服务器描述"`
	LoadType    int       `json:"load_type" gorm:"column:load_type" description:"负载均衡类型"`
	UpdatedAt   time.Time `json:"update_at" gorm:"column:update_at" description:"更新时间"`
	CreatedAt   time.Time `json:"create_at" gorm:"column:create_at" description:"创建时间"`
	IsDelete    int       `json:"is_delete" gorm:"column:is_delete" description:"是否删除"`
}

func (t *ServiceInfo) TableName() string {
	return "gateway_service_info"
}

func (t *ServiceInfo) Find(c *gin.Context, db *gorm.DB, search *ServiceInfo) (*ServiceInfo, error) {
	out := &ServiceInfo{}
	if err := db.SetCtx(public.GetGinTraceContext(c)).Where(search).Find(out).Error; err != nil {
		return nil, err
	}
	return out, nil
}

func (t *ServiceInfo) Save(c *gin.Context, db *gorm.DB) error {
	return db.SetCtx(public.GetGinTraceContext(c)).Save(t).Error
}

func (t *ServiceInfo) PageList(
	c *gin.Context, db *gorm.DB, param *dto.ServiceListInput) (list []ServiceInfo, total int64, err error) {
	offset := (param.PageNo - 1) * param.PageSize

	query := db.SetCtx(public.GetGinTraceContext(c))

	// 确定查询表以及过滤已逻辑删除的数据
	query = query.Table(t.TableName()).Where("is_delete=0")

	// 模糊查询更关键词相同的数据
	if param.Info != "" {
		value := "%" + param.Info + "%"
		query = query.Where("(service_name like ? or service_desc like ?)", value, value)
	}

	// 分页
	query = query.Limit(param.PageSize).Offset(offset)
	query.Count(&total)
	if err := query.Find(&list).Error; err != nil && err != gorm.ErrRecordNotFound {
		return nil, 0, err
	}
	return
}

// 根据 serviceInfo 查询对应的各种 http,grpc,tcp,accessControl 信息
func (t *ServiceInfo) ServiceDetail(c *gin.Context, db *gorm.DB, search *ServiceInfo) (*ServiceDetail, error) {
	if search.ServiceName == "" {
		info, err := t.Find(c, db, search)
		if err != nil {
			return nil, err
		}
		search = info
	}

	httpRule := &HttpRule{ServiceID: search.ID}
	httpRule, err := httpRule.Find(c, db, httpRule)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	tcpRule := &TcpRule{ServiceID: search.ID}
	tcpRule, err = tcpRule.Find(c, db, tcpRule)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	grpcRule := &GrpcRule{ServiceID: search.ID}
	grpcRule, err = grpcRule.Find(c, db, grpcRule)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	accessControl := &AccessControl{ServiceID: search.ID}
	accessControl, err = accessControl.Find(c, db, accessControl)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	loadBalance := &LoadBalance{ServiceID: search.ID}
	loadBalance, err = loadBalance.Find(c, db, loadBalance)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	detail := &ServiceDetail{
		Info:          search,
		HTTPRule:      httpRule,
		TCPRule:       tcpRule,
		GRPCRule:      grpcRule,
		LoadBalance:   loadBalance,
		AccessControl: accessControl,
	}
	return detail, nil
}
