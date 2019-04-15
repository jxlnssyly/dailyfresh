package routers

import (
	"dailyfresh/controllers"
	"github.com/astaxie/beego"
	_ "dailyfresh/models"
)

func init() {
    beego.Router("/", &controllers.MainController{})
    beego.Router("/register",&controllers.UserController{},"get:ShowRegister;post:HandleRegister")
}
