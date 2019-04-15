package routers

import (
	"dailyfresh/controllers"
	"github.com/astaxie/beego"
	_ "dailyfresh/models"
)

func init() {
    beego.Router("/", &controllers.MainController{})
    // 注册
    beego.Router("/register",&controllers.UserController{},"get:ShowRegister;post:HandleRegister")
    // 激活
    beego.Router("/active",&controllers.UserController{},"get:ActiveUser")

    // 登录页
    beego.Router("/login",&controllers.UserController{},"get:ShowLogin;post:HandleLogin")
}
