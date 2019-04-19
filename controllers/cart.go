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
	userName := self.GetSession("userName")
	if userName == nil {
		resp["code"] = 2
		resp["msg"] = "当前用户未登录"
		self.Data["json"] = resp
		return
	}

	o := orm.NewOrm()
	var user models.User
	user.Name = userName.(string)
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

// 获取购物车数量的函数
func GetCartCount(self *beego.Controller) int {
	// 从Redis中获取购物车数量
	userName := self.GetSession("userName")
	if userName == nil {
		return 0
	}

	o := orm.NewOrm()
	var user models.User
	user.Name = userName.(string)
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
