package routers

import (
	"dailyfresh/controllers"
	"github.com/astaxie/beego"
	_ "dailyfresh/models"
	"github.com/astaxie/beego/context"
	"github.com/gomodule/redigo/redis"
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
	beego.Router("/logout",&controllers.UserController{},"get:Logout")

	// 用户中心
	beego.Router("/user/userCenterInfo",&controllers.UserController{},"get:ShowUserCenterInfo")

	// 用户中心订单页
	beego.Router("/user/userCenterOrder",&controllers.UserController{},"get:ShowUserCenterOrder")
	// 用户中心地址页
	beego.Router("/user/userCenterSite",&controllers.UserController{},"get:ShowUserCenterSite;post:HandleUserCenterSite")

	// 商品详情
	beego.Router("/goodsDetail",&controllers.GoodsController{},"get:ShowGoodsDetail")

	// 商品列表页
	beego.Router("/goodsList",&controllers.GoodsController{},"get:ShowGoodsList")

	// 商品搜索
	beego.Router("/goodsSearch",&controllers.GoodsController{},"post:HandleGoodsSearch")

	// 添加购物车
	beego.Router("/user/addCart",&controllers.CartController{},"post:HandleAddCart")

	// 展示购物车页面
	beego.Router("/user/cart", &controllers.CartController{},"get:ShowCart")

	// 添加购物车数量
	beego.Router("/user/updateCart", &controllers.CartController{},"post:HandleUpdateCart")

	// 删除购物车数据
	beego.Router("/user/deleteCart",&controllers.CartController{},"post:DeleteCart")
}

var filterFunc = func(ctx *context.Context) {
	conn, err := redis.Dial("tcp",beego.AppConfig.String("redisServer"))
	if err != nil {
		beego.Info("Redis连接错误")
		ctx.Redirect(302,"/login")
		return
	}

	defer conn.Close()

	userName, err := redis.String(conn.Do("get","userName"))
	if userName == "" {
		ctx.Redirect(302,"/login")
		return
	}
}
