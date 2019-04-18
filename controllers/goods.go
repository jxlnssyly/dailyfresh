package controllers

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"dailyfresh/models"
)

type GoodsController struct {
	beego.Controller
}

func ShowLayout(self *beego.Controller) {
	// 查询类型
	o := orm.NewOrm()
	var types []models.GoodsType
	o.QueryTable(&models.GoodsType{}).All(&types)

	self.Data["goodsTypes"] = types
	nginxHost := beego.AppConfig.String("nginxHost")
	self.Data["nginxHost"] = nginxHost
	// 获取用户信息

	GetUser(self)

	// 指定layout
	self.Layout = "goodsLayout.html"
}

func GetUser(self *beego.Controller) string  {
	userName := self.GetSession("userName")
	if userName == nil {
		self.Data["userName"] = ""
	} else  {
		self.Data["userName"]= userName.(string)
		return userName.(string)

	}


	return ""
}

/*展示首页*/
func (self *GoodsController) ShowIndex() {
	GetUser(&self.Controller)
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
	goods := make([]map[string]interface{},len(goodsType))

	// 向切片interface中插入类型数据
	for index, value := range goodsType {
		temp := make(map[string]interface{})
		temp["type"] = value
		goods[index] = temp
	}
	// 获取对应类型的首页展示商品

	for _,value := range goods {
		var textGoods []models.IndexTypeGoodsBanner
		var imageGoods []models.IndexTypeGoodsBanner
		o.QueryTable(&models.IndexTypeGoodsBanner{}).RelatedSel("GoodsType","GoodsSKU").OrderBy("Index").Filter("GoodsType",value["type"]).Filter("DisplayType",0).All(&textGoods)
		o.QueryTable(&models.IndexTypeGoodsBanner{}).RelatedSel("GoodsType","GoodsSKU").OrderBy("Index").Filter("GoodsType",value["type"]).Filter("DisplayType",1).All(&imageGoods)
		value["textGoods"] = textGoods
		value["imageGoods"] = imageGoods
	}
	self.Data["goods"] = goods

	self.TplName = "index.html"
}

// 展示商品详情
func (self *GoodsController) ShowGoodsDetail() {
	// 获取数据
	id, err := self.GetInt("id")
	if err != nil {
		beego.Error("id获取错误",err)
		self.Redirect("/",302)
		return
	}

	GetUser(&self.Controller)
	// 处理数据
	o := orm.NewOrm()
	var goodsSKU models.GoodsSKU
	goodsSKU.Id = id
	o.QueryTable(&goodsSKU).RelatedSel("GoodsType","Goods").Filter("Id",id).One(&goodsSKU)

	// 获取同类型时间靠前的两条商品数据
	var newGoods []models.GoodsSKU
	o.QueryTable(&models.GoodsSKU{}).RelatedSel("GoodsType").Filter("GoodsType",goodsSKU.GoodsType).OrderBy("Time").Limit(2,0).All(&newGoods)

	self.Data["goodsSku"] = goodsSKU
	self.Data["goodsNew"] = newGoods
	ShowLayout(&self.Controller)
	self.TplName = "detail.html"

}
