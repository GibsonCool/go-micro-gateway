package controller

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
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
// @Param body body dto.AdminLoginInput true "body"
// @Success 200 {object} middleware.Response{data=dto.AdminLoginOutput} "success"
// @Router /admin/change_pwd [get]
func (c *AdminController) ChangePwd(context *gin.Context) {

}
