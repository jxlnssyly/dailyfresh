package controllers

import "github.com/astaxie/beego"

type GoodsController struct {
	beego.Controller
}

func GetUser(self *beego.Controller) string  {
	userName := self.GetSession("userName")
	if userName == nil {
		self.Data["userName"] = ""
	} else  {
		self.Data["userName"]= userName.(string)
	}
	return userName.(string)
}

func (self *GoodsController) ShowIndex() {
	GetUser(&self.Controller)
	self.TplName = "index.html"
}
