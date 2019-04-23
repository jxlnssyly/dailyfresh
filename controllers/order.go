package controllers

import (
	"github.com/astaxie/beego"
	"dailyfresh/models"
	"strconv"
	"github.com/astaxie/beego/orm"
	"github.com/gomodule/redigo/redis"
	"time"
	"strings"
)

type OrderController struct {
	beego.Controller
}

func (self *OrderController) ShowOrder() {
	skuids := self.GetStrings("skuid")
	if len(skuids) == 0 {
		beego.Info("购物车为空,请选择商品")
		self.Redirect("/user/cart", 302)
		return
	}
	userName := GetUser(&self.Controller)
	o := orm.NewOrm()
	var user models.User
	user.Name = userName
	o.Read(&user, "Name")
	conn, _ := redis.Dial("tcp", beego.AppConfig.String("redisServer"))
	defer conn.Close()

	totalPrice := 0
	totalCount := 0
	// 处理数据
	goodsBuffer := make([]map[string]interface{}, len(skuids))
	for index, skuid := range skuids {
		temp := make(map[string]interface{})
		id, _ := strconv.Atoi(skuid)
		var goodsSku models.GoodsSKU
		goodsSku.Id = id
		o.Read(&goodsSku)
		temp["goods"] = goodsSku
		// 获取商品数量
		count, _ := redis.Int(conn.Do("hget", "cart_"+strconv.Itoa(user.Id), id))
		temp["count"] = count
		amount := goodsSku.Price * count
		totalPrice += amount
		totalCount += count
		temp["amount"] = amount
		goodsBuffer[index] = temp

	}
	self.Data["nginxHost"] = beego.AppConfig.String("nginxHost")

	self.Data["goodsBuffer"] = goodsBuffer
	self.Data["totalPrice"] = totalPrice
	self.Data["totalCount"] = totalCount
	transferPrice := 10
	self.Data["transferPrice"] = transferPrice
	// 获取地址数据
	var addrs []models.Address
	o.QueryTable("Address").RelatedSel("User").Filter("User__Id", user.Id).All(&addrs)
	self.Data["address"] = addrs

	self.Data["realPrice"] = totalPrice + transferPrice

	// 传递所有商品的Id
	self.Data["skuids"] = skuids

	self.TplName = "place_order.html"
}

// 提交订单
func (self *OrderController) AddOrder() {
	// 获取数据
	addr, _ := self.GetInt("addrid")
	payId, _ := self.GetInt("payId")
	ids := self.GetString("skuids")

	skuid := ids[1 : len(ids)-1]
	skuids := strings.Split(skuid, " ")
	beego.Error(skuids)
	totalCount, _ := self.GetInt("totalCount")
	//totalPrice, _ := self.GetInt("totalPrice")
	transferPrice, _ := self.GetInt("transferPrice")
	realPrice, _ := self.GetInt("realPrice")
	resp := make(map[string]interface{})
	defer self.ServeJSON()
	// 校验数据
	if len(skuids) == 0 {
		beego.Info("未获取到商品id")
		resp["code"] = 1
		resp["msg"] = "数据库连接错误"
		self.Data["json"] = resp
		return
	}

	// 处理数据
	// 向订单表中插入数据
	o := orm.NewOrm()
	o.Begin()
	var user models.User
	user.Name = GetUser(&self.Controller)
	o.Read(&user, "Name")

	var order models.OrderInfo
	order.OrderId = time.Now().Format("2006010215030405") + strconv.Itoa(user.Id)
	order.User = &user
	order.Orderstatus = 1
	order.PayMethod = payId
	order.TotalCount = totalCount
	order.TotalPrice = realPrice
	order.TransitPrice = transferPrice

	var address models.Address
	address.Id = addr
	o.Read(&address)
	order.Address = &address

	o.Insert(&order)

	conn, _ := redis.Dial("tcp", beego.AppConfig.String("redisServer"))
	defer conn.Close()
	// 向订单商品表插入数据
	for _, skuid := range skuids {
		var goodsSku models.GoodsSKU
		id, _ := strconv.Atoi(skuid)
		goodsSku.Id = id
		i := 3
		for i > 0 {

			o.Read(&goodsSku)
			var orderGoods models.OrderGoods
			orderGoods.GoodsSKU = &goodsSku
			orderGoods.OrderInfo = &order
			count, _ := redis.Int(conn.Do("hget", "cart_"+strconv.Itoa(user.Id), id))
			if count > goodsSku.Stock {
				resp["code"] = 2
				resp["msg"] = "商品库存不足"
				self.Data["json"] = resp
				o.Rollback()
				return
			}
			preCount := goodsSku.Stock

			orderGoods.Count = count
			orderGoods.Price = count * goodsSku.Price
			o.Insert(&orderGoods)
			goodsSku.Stock -= count
			goodsSku.Sales += count

			updateCount, _ := o.QueryTable("GoodsSKU").Filter("Id", goodsSku.Id).Filter("Stock", preCount).Update(orm.Params{"Stock": goodsSku.Stock, "Sales": goodsSku.Sales})
			if updateCount == 0 {

				if i > 0 {
					i -= 1
					continue
				}
				resp["code"] = 3
				resp["msg"] = "商品库存改变，订单提交失败"
				self.Data["json"] = resp
				o.Rollback()
				return
			} else {
				conn.Do("hdel","cart_"+strconv.Itoa(user.Id),id)
				break
			}
		}
	}
	o.Commit()
	// 返回数据
	resp["code"] = 5
	resp["msg"] = "ok"
	self.Data["json"] = resp
}
