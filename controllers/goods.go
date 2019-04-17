package controllers

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"dailyfresh/models"
)

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

/*展示首页*/
func (self *GoodsController) ShowIndex() {
	//GetUser(&self.Controller)
	o := orm.NewOrm()

	// 获取nginx配置
	nginxHost := beego.AppConfig.String("nginxHost")
	self.Data["nginxHost"] = nginxHost

	// 获取类型数据
	var goodsType []models.GoodsType
	o.QueryTable(&models.GoodsType{}).All(&goodsType)
	self.Data["goodsTypes"] = goodsType

	// 获取轮播图数据
	var indexGoods []models.IndexGoodsBanner
	o.QueryTable(&models.IndexGoodsBanner{}).OrderBy("Index").All(&indexGoods)
	self.Data["indexGoodsBanner"] = indexGoods

	// 获取促销商品数据
	var promotionGoods []models.IndexPromotionBanner
	o.QueryTable(&models.IndexPromotionBanner{}).OrderBy("Index").All(&promotionGoods)
	self.Data["promotionGoods"] = promotionGoods

	// 获取首页商品数据

	self.TplName = "index.html"
}
