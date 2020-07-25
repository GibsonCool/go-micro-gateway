package controller

import (
	"encoding/json"
	"github.com/e421083458/golang_common/lib"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"go-micro-gateway/go_gateway/dao"
	"go-micro-gateway/go_gateway/dto"
	"go-micro-gateway/go_gateway/middleware"
	"go-micro-gateway/go_gateway/public"
	"time"
)

func AdminLoginRegister(group *gin.RouterGroup) {
	adminLogin := &AdminLoginController{}
	group.POST("/login", adminLogin.AdminLogin)
	group.GET("/login_out", adminLogin.AdminLogOut)
}

type AdminLoginController struct {
}

// @Summary 管理员登录
// @Description 管理员登录
// @Tags 管理员接口
// @Accept  json
// @Produce  json
// @Param body body dto.AdminLoginInput true "body"
// @Success 200 {object} middleware.Response{data=dto.AdminLoginOutput} "success"
// @Router /admin_login/login [POST]
func (c *AdminLoginController) AdminLogin(ctx *gin.Context) {
	params := &dto.AdminLoginInput{}
	if err := params.BindValidParam(ctx); err != nil {
		middleware.ResponseError(ctx, 2000, err)
		return
	}
	gDB, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(ctx, 2001, err)
		return
	}

	admin := &dao.Admin{}
	admin, err = admin.LoginCheck(ctx, gDB, params)
	if err != nil {
		middleware.ResponseError(ctx, 2002, err)
		return
	}

	// 登录校验通过，设置 session
	sessInfo := &dto.AdminSessionInfo{
		ID:        admin.Id,
		UserName:  admin.UserName,
		LoginTime: time.Now(),
	}
	marshal, err := json.Marshal(sessInfo)
	if err != nil {
		middleware.ResponseError(ctx, 2003, err)
		return
	}
	session := sessions.Default(ctx)
	session.Set(public.AdminSessionInfoKey, string(marshal))
	_ = session.Save()

	out := &dto.AdminLoginOutput{Token: params.UserName}
	middleware.ResponseSuccess(ctx, out)
}

// @Summary 管理员退出
// @Description 管理员退出
// @Tags 管理员接口
// @Accept  json
// @Produce  json
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /admin_login/login_out [get]
func (c *AdminLoginController) AdminLogOut(ctx *gin.Context) {
	session := sessions.Default(ctx)
	session.Delete(public.AdminSessionInfoKey)
	session.Save()
	middleware.ResponseSuccess(ctx, "退出成功")
}
