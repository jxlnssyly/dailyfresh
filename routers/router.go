package routers

import (
	"dailyfresh/controllers"
	"github.com/astaxie/beego"
	_ "dailyfresh/models"
	"github.com/astaxie/beego/context"
)

func init() {
	beego.InsertFilter("/user/*",beego.BeforeExec,filterFunc)
	// 跳转首页
	beego.Router("/", &controllers.GoodsController{},"get:ShowIndex")
    // 注册
    beego.Router("/register",&controllers.UserController{},"get:ShowRegister;post:HandleRegister")
    // 激活
    beego.Router("/active",&controllers.UserController{},"get:ActiveUser")

    // 登录页
    beego.Router("/login",&controllers.UserController{},"get:ShowLogin;post:HandleLogin")

	// 退出登录
	beego.Router("/user/logout",&controllers.UserController{},"get:Logout")

	// 用户中心
	beego.Router("/user/userCenterInfo",&controllers.UserController{},"get:ShowUserCenterInfo")

	// 用户中心订单页
	beego.Router("/user/userCenterOrder",&controllers.UserController{},"get:ShowUserCenterOrder")
	// 用户中心地址页
	beego.Router("/user/userCenterSite",&controllers.UserController{},"get:ShowUserCenterSite;post:HandleUserCenterSite")

	// 商品详情
	beego.Router("/goodsDetail",&controllers.GoodsController{},"get:ShowGoodsDetail")
}

var filterFunc = func(ctx *context.Context) {
	userName := ctx.Input.Session("userName")
	if userName == nil {
		ctx.Redirect(302,"/login")
		return
	}
}
