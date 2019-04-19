package controllers

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"dailyfresh/models"
	"github.com/gomodule/redigo/redis"
	"strconv"
	"math"
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

func GetUser(self *beego.Controller) string {
	userName := self.GetSession("userName")
	if userName == nil {
		self.Data["userName"] = ""
	} else {
		self.Data["userName"] = userName.(string)
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
	goods := make([]map[string]interface{}, len(goodsType))

	// 向切片interface中插入类型数据
	for index, value := range goodsType {
		temp := make(map[string]interface{})
		temp["type"] = value
		goods[index] = temp
	}
	// 获取对应类型的首页展示商品

	for _, value := range goods {
		var textGoods []models.IndexTypeGoodsBanner
		var imageGoods []models.IndexTypeGoodsBanner
		o.QueryTable(&models.IndexTypeGoodsBanner{}).RelatedSel("GoodsType", "GoodsSKU").OrderBy("Index").Filter("GoodsType", value["type"]).Filter("DisplayType", 0).All(&textGoods)
		o.QueryTable(&models.IndexTypeGoodsBanner{}).RelatedSel("GoodsType", "GoodsSKU").OrderBy("Index").Filter("GoodsType", value["type"]).Filter("DisplayType", 1).All(&imageGoods)
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
		beego.Error("id获取错误", err)
		self.Redirect("/", 302)
		return
	}

	// 处理数据
	o := orm.NewOrm()
	var goodsSKU models.GoodsSKU
	goodsSKU.Id = id
	o.QueryTable(&goodsSKU).RelatedSel("GoodsType", "Goods").Filter("Id", id).One(&goodsSKU)

	// 获取同类型时间靠前的两条商品数据
	var newGoods []models.GoodsSKU
	o.QueryTable(&models.GoodsSKU{}).RelatedSel("GoodsType").Filter("GoodsType", goodsSKU.GoodsType).OrderBy("Time").Limit(2, 0).All(&newGoods)

	self.Data["goodsSku"] = goodsSKU
	self.Data["goodsNew"] = newGoods

	// 添加历史浏览记录
	// 判断用户是否登录
	userName := self.GetSession("userName")
	beego.Info(userName)
	if userName != nil {
		// 查询
		var user models.User
		user.Name = userName.(string)
		o.Read(&user, "Name")
		redisServer := beego.AppConfig.String("redisServer")
		// Redis存储
		conn, err := redis.Dial("tcp", redisServer)
		defer conn.Close()
		if err != nil {
			beego.Info("redis连接错误", err)
		}
		// 把以前相同的浏览历史记录删除
		_, err = conn.Do("lrem", "history_"+strconv.Itoa(user.Id), 0, id)

		// 添加新的浏览历史记录到商品中
		_, err = conn.Do("lpush", "history_"+strconv.Itoa(user.Id), id)
	}

	// 添加历史浏览记录

	ShowLayout(&self.Controller)
	self.TplName = "detail.html"
}

func PageTool(pageCount int, pageIndex int) []int {
	var pages []int
	if pageCount <= 5 {
		pages = make([]int, pageCount)

		for i,_ := range pages{
			pages[i] = i + 1
		}
	} else if pageIndex <= 3 {
		pages = []int{1,2,3,4,5}
	} else if pageIndex > pageCount - 3 {
		pages = []int{pageCount - 4, pageCount - 3, pageCount - 2, pageCount - 1, pageCount}
	} else {
		pages = []int{pageIndex - 2, pageIndex - 1, pageIndex, pageIndex + 1, pageIndex + 2}
	}
	return pages
}

func (self *GoodsController) ShowGoodsList() {
	// 获取数据
	id, err := self.GetInt("typeId")
	if err != nil {
		beego.Info("请求路径错误")
		self.Redirect("/", 302)
		return
	}

	// 处理数据
	ShowLayout(&self.Controller)

	// 获取新品
	o := orm.NewOrm()
	var goodsNew []models.GoodsSKU
	o.QueryTable(&models.GoodsSKU{}).RelatedSel("GoodsType").Filter("GoodsType__Id",id).OrderBy("Time").Limit(2,0).All(&goodsNew)
	self.Data["goodsNew"] = goodsNew



	// 分页实现
	// 获取PageCount
	pageSize := 2
	count, _ := o.QueryTable(&models.GoodsSKU{}).RelatedSel("GoodsType").Filter("GoodsType__Id",id).Count()

	pageCount := math.Ceil(float64(count) / float64(pageSize))

	pageIndex, err := self.GetInt("pageIndex")
	if err != nil {
		pageIndex = 1
	}

	pages := PageTool(int(pageCount), pageIndex)

	//获取商品
	var goods []models.GoodsSKU
	start := (pageIndex - 1) * pageSize

	self.Data["typeId"] = id
	self.Data["pages"] = pages
	self.Data["pageIndex"] = pageIndex

	// 获取上一页
	prePage := pageIndex - 1
	if pageIndex <= 1 {
		prePage = 1
	}
	self.Data["prePage"] = prePage

	// 获取下一页
	nextPage := pageIndex + 1
	if nextPage > int(pageCount) {
		nextPage = int(pageCount)
	}
	self.Data["nextPage"] = nextPage

	sort := self.GetString("sort")
	if sort == "" {
		o.QueryTable(&models.GoodsSKU{}).RelatedSel("GoodsType").Filter("GoodsType__Id",id).Limit(pageSize,start).All(&goods)
		self.Data["goods"]= goods
	} else if sort == "price" {
		o.QueryTable(&models.GoodsSKU{}).RelatedSel("GoodsType").Filter("GoodsType__Id",id).OrderBy("Price").Limit(pageSize,start).All(&goods)
		self.Data["goods"]= goods
	} else {
		o.QueryTable(&models.GoodsSKU{}).RelatedSel("GoodsType").Filter("GoodsType__Id",id).OrderBy("Sales").Limit(pageSize,start).All(&goods)
		self.Data["goods"]= goods
	}

	self.Data["sort"] = sort

	self.TplName = "list.html"
}

// 处理搜索
func (self *GoodsController) HandleGoodsSearch() {

	// 获取数据
	goodsName := self.GetString("goodsName")

	o := orm.NewOrm()
	var goods []models.GoodsSKU

	// 校验数据
	if goodsName == "" {
		o.QueryTable(&models.GoodsSKU{}).All(&goods)
		self.Data["goods"] = goods
		self.TplName = "search.html"
		return
	}

	// 处理数据
	o.QueryTable(&models.GoodsSKU{}).Filter("Name__icontains",goodsName).All(&goods)

	// 返回视图
	self.Data["goods"] = goods
	ShowLayout(&self.Controller)
	self.TplName = "search.html"

	return
}


