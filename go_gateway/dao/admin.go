package dao

import (
	"github.com/e421083458/gorm"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"go-micro-gateway/go_gateway/dto"
	"go-micro-gateway/go_gateway/public"
	"time"
)

type Admin struct {
	Id        int       `json:"id" gorm:"primary_key" description:"自增主键"`
	UserName  string    `json:"user_name" gorm:"column:user_name" description:"管理员用户名"`
	Salt      string    `json:"salt" gorm:"column:salt" description:"盐"`
	Password  string    `json:"password" gorm:"column:password" description:"密码"`
	UpdatedAt time.Time `json:"update_at" gorm:"column:update_at" description:"更新时间"`
	CreatedAt time.Time `json:"create_at" gorm:"column:create_at" description:"创建时间"`
	IsDelete  int       `json:"is_delete" gorm:"column:is_delete" description:"是否删除"`
}

// 实现 gorm 获取重定义表名
func (t *Admin) TableName() string {
	return "gateway_admin"
}

func (t *Admin) LoginCheck(c *gin.Context, db *gorm.DB, param *dto.AdminLoginInput) (*Admin, error) {
	adminInfo, err := t.Find(c, db, &Admin{UserName: param.UserName, IsDelete: 0})
	if err != nil {
		return nil, errors.New("用户信息不存在：" + err.Error())
	}
	saltPassword := public.GenSaltPassword(adminInfo.Salt, param.Password)
	if adminInfo.Password != saltPassword {
		return nil, errors.New("密码错误，请重新输入")
	}
	return adminInfo, nil
}

func (t *Admin) Find(c *gin.Context, db *gorm.DB, search *Admin) (*Admin, error) {
	out := &Admin{}
	if err := db.SetCtx(public.GetGinTraceContext(c)).Where(search).Find(out).Error; err != nil {
		return nil, err
	}
	return out, nil
}
