package controllers

import (
	"github.com/astaxie/beego"
	"github.com/gomodule/redigo/redis"
	"github.com/astaxie/beego/orm"
	"dailyfresh/models"
	"strconv"
)

type CartController struct {
	beego.Controller
}

func (self *CartController) HandleAddCart() {
	// 获取数据
	skuid, err1 := self.GetInt("skuid")
	count, err2 := self.GetInt("count")
	resp := make(map[string]interface{})
	defer self.ServeJSON()

	// 校验数据
	if err1 != nil || err2 != nil{
		resp["code"] = 1
		resp["msg"] = "传递的数据不正确"
		self.Data["json"] = resp

		return
	}

	// 处理数据
	userName := GetUser(&self.Controller)
	if userName == "" {
		resp["code"] = 2
		resp["msg"] = "当前用户未登录"
		self.Data["json"] = resp
		return
	}

	o := orm.NewOrm()
	var user models.User
	user.Name = userName
	o.Read(&user,"Name")

	conn,err := redis.Dial("tcp",beego.AppConfig.String("redisServer"))
	if err != nil {
		beego.Info("Redis连接错误")
		return
	}
	defer conn.Close()
	preCount, err := redis.Int(conn.Do("hget","cart_"+strconv.Itoa(user.Id),skuid,count))
	conn.Do("hset","cart_"+strconv.Itoa(user.Id),skuid,count + preCount)
	rep, err := conn.Do("hlen","cart_"+strconv.Itoa(user.Id) )
	carCount, _ := redis.Int(rep, err)
	resp["code"] = 5
	resp["msg"] = "OK"
	resp["cartCount"] = carCount

	// 返回json数据
	self.Data["json"] = resp

}

// 展示购物车页面
func (self *CartController)ShowCart()  {

	// 用户信息
	userName := GetUser(&self.Controller)

	// 从Redis中获取数据
	conn, err := redis.Dial("tcp",beego.AppConfig.String("redisServer"))
	if err != nil {
		beego.Info("Redis连接失败")
		return
	}

	defer conn.Close()

	o := orm.NewOrm()
	var user models.User
	user.Name = userName
	o.Read(&user, "Name")

	goodsMap, err := redis.IntMap( conn.Do("hgetall","cart_"+strconv.Itoa(user.Id))) // map[string]int
	if err != nil {
		beego.Info("没有购物车数据")
		return
	}
	totalPrice := 0
	totalCount := 0
	i := 0
	goods := make([]map[string]interface{},len(goodsMap))
	for index, value := range goodsMap {
		skuid, _ := strconv.Atoi(index)
		var goodsSku models.GoodsSKU
		goodsSku.Id = skuid
		o.Read(&goodsSku)
		temp := make(map[string]interface{})
		temp["goodsSku"] = goodsSku
		temp["count"] = value
		goods[i] = temp
		temp["addPrice"] = goodsSku.Price * value
		totalPrice += goodsSku.Price * value
		totalCount += value
		i += 1
	}
	self.Data["goods"] = goods
	self.Data["totalPrice"] = totalPrice
	self.Data["totalCount"] = totalCount
	self.Data["nginxHost"] = beego.AppConfig.String("nginxHost")
	self.TplName = "cart.html"
}

// 获取购物车数量的函数
func GetCartCount(self *beego.Controller) int {
	// 从Redis中获取购物车数量
	userName := GetUser(self)
	if userName == "" {
		return 0
	}

	o := orm.NewOrm()
	var user models.User
	user.Name = userName
	o.Read(&user,"Name")

	conn, err := redis.Dial("tcp",beego.AppConfig.String("redisServer"))
	if err != nil {
		return 0
	}
	defer conn.Close()

	rep, err := conn.Do("hlen","cart_"+strconv.Itoa(user.Id))
	cartCount, _ := redis.Int(rep,err)
	// cart_userId
	return cartCount
}

// 更新购物车数据
func (self *CartController)HandleUpdateCart()  {
	// 获取数据
	skuid, err := self.GetInt("skuid")
	count, err2 := self.GetInt("count")

	resp := make(map[string]interface{})
	defer self.ServeJSON()

	// 校验数据
	if err != nil || err2 != nil {
		beego.Info(err)
		beego.Info(err2)

		resp["code"] = 1
		resp["msg"] = "请求数据不正确"
		self.Data["json"] = resp
		return
	}

	userName := GetUser(&self.Controller)

	if userName == "" {
		 resp["code"] = 3
		resp["msg"] = "当前用户未登录"
		self.Data["json"] = resp
		return
	}

	o := orm.NewOrm()
	var user models.User
	user.Name = userName
	o.Read(&user,"Name")

	// 处理数据
	conn, err := redis.Dial("tcp",beego.AppConfig.String("redisServer"))
	if err != nil {
		resp["code"] = 2
		resp["msg"] = "redis数据库连接失败"
		self.Data["json"] = resp
		return
	}
	defer conn.Close()

	conn.Do("hset","cart_"+strconv.Itoa(user.Id),skuid,count)
	resp["code"] = 5
	resp["msg"] = "ok"

	self.Data["json"] = resp
}

func (self *CartController)DeleteCart()  {

	// 获取数据
	skuid, err := self.GetInt("skuid")

	defer self.ServeJSON()

	resp := make(map[string]interface{})
	if err != nil {
		resp["code"] = 1
		resp["errmsg"] = "请求数据不正确"
		self.Data["json"] = resp
		return
	}

	conn, err := redis.Dial("tcp",beego.AppConfig.String("redisServer"))
	if err != nil {
		resp["code"] = 2
		resp["errmsg"] = "Redis数据库连接失败"
		self.Data["json"] = resp
		return
	}
	userName := GetUser(&self.Controller)
	user := models.User{}
	user.Name = userName
	o := orm.NewOrm()
	o.Read(&user, "Name")

	_, err = conn.Do("hdel","cart_"+strconv.Itoa(user.Id),skuid)
	if err != nil {
		beego.Info(err)
	}

	resp["code"] = 5
	resp["errmsg"] = "ok"
	self.Data["json"] = resp
}