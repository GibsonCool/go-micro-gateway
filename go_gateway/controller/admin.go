package controller

import (
	"encoding/json"
	"fmt"
	"github.com/e421083458/golang_common/lib"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"go-micro-gateway/go_gateway/dao"
	"go-micro-gateway/go_gateway/dto"
	"go-micro-gateway/go_gateway/middleware"
	"go-micro-gateway/go_gateway/public"
)

func AdminRegister(group *gin.RouterGroup) {
	adminLogin := &AdminController{}
	group.GET("/admin_info", adminLogin.AdminInfo)
	group.POST("/change_pwd", adminLogin.ChangePwd)
}

type AdminController struct {
}

// @Summary 获取管理员信息
// @Description 获取管理员信息
// @Tags 管理员接口
// @Accept  json
// @Produce  json
// @Success 200 {object} middleware.Response{data=dto.AdminInfoOutput} "success"
// @Router /admin/admin_info [get]
func (c *AdminController) AdminInfo(ctx *gin.Context) {
	// 1.读取 sessionKey 对应的 json 转换为结构体
	sess := sessions.Default(ctx)
	sessInfo := sess.Get(public.AdminSessionInfoKey)
	adminSessionInfo := &dto.AdminSessionInfo{}
	if err := json.Unmarshal([]byte(fmt.Sprint(sessInfo)), adminSessionInfo); err != nil {
		middleware.ResponseError(ctx, 2000, err)
		return
	}

	// 2.取出数据封装输出
	out := &dto.AdminInfoOutput{
		ID:           adminSessionInfo.ID,
		Name:         adminSessionInfo.UserName,
		LoginTime:    adminSessionInfo.LoginTime,
		Avatar:       "https://avatars0.githubusercontent.com/u/12468166?s=460&u=d212d36b54219f73d11dc16d444742b75996a3bf&v=4",
		Introduction: "test info",
		Roles:        []string{"admin"},
	}
	middleware.ResponseSuccess(ctx, out)
}

// @Summary 改变密码
// @Description 改变密码
// @Tags 管理员接口
// @Accept  json
// @Produce  json
// @Param body body dto.ChangePwdInput true "body"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /admin/change_pwd [post]
func (c *AdminController) ChangePwd(ctx *gin.Context) {
	// 参数解析
	inputParam := &dto.ChangePwdInput{}
	if err := inputParam.BindValidParam(ctx); err != nil {
		middleware.ResponseError(ctx, 2000, err)
		return
	}

	// 1.通过 session 获取用户信息
	sess := sessions.Default(ctx)
	sessInfo := sess.Get(public.AdminSessionInfoKey)
	adminSessionInfo := &dto.AdminSessionInfo{}
	if err := json.Unmarshal([]byte(fmt.Sprint(sessInfo)), adminSessionInfo); err != nil {
		middleware.ResponseError(ctx, 2000, err)
		return
	}

	// 2.根据用户名从数据库查询出完整用户信息
	gDB, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(ctx, 2001, err)
		return
	}

	adminInfo := &dao.Admin{}
	adminInfo, err = adminInfo.Find(ctx, gDB, &dao.Admin{UserName: adminSessionInfo.UserName})
	if err != nil {
		middleware.ResponseError(ctx, 2002, err)
		return
	}

	// 更具传入新密码，加原有 盐  生成新密码
	adminInfo.Password = public.GenSaltPassword(adminInfo.Salt, inputParam.Password)

	// 保存更新数据到数据库，返回结果
	if err := adminInfo.Save(ctx, gDB); err != nil {
		middleware.ResponseError(ctx, 2003, err)
		return
	}
	middleware.ResponseSuccess(ctx, "密码更新成功")
}
