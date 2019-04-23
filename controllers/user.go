package controllers

import (
	"github.com/astaxie/beego"
	"regexp"
	"github.com/astaxie/beego/orm"
	"dailyfresh/models"
	"github.com/astaxie/beego/utils"
	"strconv"
	"encoding/base64"
	"github.com/gomodule/redigo/redis"
)

type UserController struct {
	beego.Controller
}

/*显示注册页面*/
func (self *UserController) ShowRegister()  {
	self.TplName = "register.html"
}

/*处理注册数据*/
func (self *UserController)HandleRegister()  {
	// 获取数据
	userName := self.GetString("user_name")
	pwd := self.GetString("pwd")
	cpwd := self.GetString("cpwd")
	email := self.GetString("email")

	// 校验数据
	if userName == "" || pwd == "" || cpwd == "" || email == "" {
		self.Data["errmsg"] = "数据不完整"
		self.TplName = "register.html"
		return
	}

	if pwd != cpwd {
		self.Data["errmsg"] = "两次输入密码不一致,请重新输入"
		self.TplName = "register.html"
		return
	}
	reg, _ := regexp.Compile("^[A-Za-z0-9\u4e00-\u9fa5]+@[a-zA-Z0-9_-]+(\\.[a-zA-Z0-9_-]+)+$") // 邮箱的正则
	res := reg.FindString(email)
	if res == "" {
		self.Data["errmsg"] = "邮箱格式不正确"
		self.TplName = "register.html"
		return
	}

	//
	o := orm.NewOrm()
	var user models.User
	user.Name = userName
	user.PassWord = pwd
	user.Email = email
	_, err := o.Insert(&user)
	if err != nil {
		self.Data["errmsg"] = "用户名已经存在，请重新注册"
		self.TplName = "register.html"
		return
	}

	// 发送邮件
	emailConfig := `{"username":"18607970065@163.com","password":"gobackinba","host":"smtp.163.com","port":25}`
	emailConn := utils.NewEMail(emailConfig)
	emailConn.From = "天天生鲜系统注册服务"
	emailConn.To = []string{email}
	emailConn.Subject = "天天生鲜新用户激活" // 标题
	// 注意发送给用户的是激活请求地址
	emailConn.Text = "127.0.0.1:8080/active?id=" + strconv.Itoa(user.Id)
	emailConn.Send()

	// 4.返回页面
	self.Ctx.WriteString("注册成功，请去相应邮箱激活用户!")
}

/*激活处理*/
func (self *UserController) ActiveUser() {
	// 1、获取数据
	id, err := self.GetInt("id")

	// 2、校验数据
	if err != nil {
		self.Data["errmsg"] = "要激活的用户不存在"
		self.TplName = "register.html"
		return
	}

	// 3、处理数据
	o := orm.NewOrm()
	var user models.User
	user.Id = id
	err = o.Read(&user)
	if err != nil {
		self.Data["errmsg"] = "要激活的用户不存在"
		self.TplName = "register.html"
		return
	}
	user.Active = true
	o.Update(&user)

	// 4、返回视图
	self.Redirect("/login",302)
}

/*展示登录页面*/
func (self *UserController)ShowLogin()  {
	userName := self.Ctx.GetCookie("userName")
	temp,_ := base64.StdEncoding.DecodeString(userName)

	if string(temp) == "" {
		self.Data["userName"] = ""
		self.Data["checked"] = ""
	} else  {
		self.Data["userName"] = string(temp)
		self.Data["checked"] = "checked"
	}

	self.TplName = "login.html"
}

/*处理登录数据*/
func (self *UserController) HandleLogin() {
	// 1、获取数据
	userName := self.GetString("username")
	pwd := self.GetString("pwd")
	// 2、校验数据
	if userName == "" || pwd == "" {
		 self.Data["errmsg"] = "登录数据不完整，请重新输入"
		self.TplName = "login.html"
		 return
	}


	// 3、处理数据
	o := orm.NewOrm()
	var user models.User
	user.Name = userName
	err := o.Read(&user, "Name")
	if err != nil || user.PassWord != pwd {
		self.Data["errmsg"] = "用户名或者密码错误"
		self.TplName = "login.html"
		return
	}

	if user.Active != true {

		self.Data["errmsg"] = "用户名未激活，请前往邮箱激活"
		self.TplName = "login.html"
		return
	}

	// 4、返回页面
	remember := self.GetString("remember")
	if remember == "on" {
		temp := base64.StdEncoding.EncodeToString([]byte(userName))
		self.Ctx.SetCookie("userName",temp,24 * 3600 * 30)
	} else  {
		self.Ctx.SetCookie("userName",userName,-1)
	}
	// 保存登陆状态到Redis里面
	saveName(userName)
	self.Redirect("/",302)
}

/*用户中心*/

func (self *UserController) ShowUserCenterInfo() {
	userName := GetUser(&self.Controller)


	// 查询地址表的内容
	o := orm.NewOrm()

	var addr models.Address

	// 高级查询
	err := o.QueryTable(&addr).RelatedSel("User").Filter("User__Name",userName).Filter("Isdefault",true).One(&addr)
	beego.Info(err)
	if addr.Id == 0	 {
		self.Data["addr"] = ""
	} else {
		self.Data["addr"] = addr
	}

	// 获取历史浏览记录
	conn, err := redis.Dial("tcp",beego.AppConfig.String("redisServer"))
	defer conn.Close()
	if err != nil {
		beego.Info("连接Redis出错")
		return
	}

	var user models.User
	user.Name = userName
	o.Read(&user,"Name")
	rep, err := conn.Do("lrange","history_"+strconv.Itoa(user.Id),0,4)
	// 回复助手函数
	goodsIds, _ := redis.Ints(rep, err)
	var goodsSKUs []models.GoodsSKU
	for _, value := range goodsIds {
		var goods models.GoodsSKU
		goods.Id = value
		o.Read(&goods)
		goodsSKUs = append(goodsSKUs, goods)
	}
	beego.Info(goodsSKUs)
	self.Data["goodsSKUs"] = goodsSKUs
	self.Data["nginxHost"] = beego.AppConfig.String("nginxHost")
	self.Layout = "userCenterLayout.html"
	self.TplName = "user_center_info.html"
}

/* 展示用户中心订单页*/
func (self *UserController) ShowUserCenterOrder() {
	userName := GetUser(&self.Controller)
	var user models.User
	user.Name = userName
	// 获取订单表的数据
	o := orm.NewOrm()
	o.Read(&user,"Name")

	var orderInfos []models.OrderInfo
	o.QueryTable(&models.OrderInfo{}).Filter("User__Id",user.Id).All(&orderInfos)
	beego.Error(orderInfos)

	goodsBuffer := make([]map[string]interface{},len(orderInfos))
	for index, orderInfo := range orderInfos {
		var orderGoods []models.OrderGoods
		o.QueryTable(&models.OrderGoods{}).RelatedSel("OrderInfo","GoodsSKU").Filter("OrderInfo__Id",orderInfo.Id).All(&orderGoods)
		temp := make(map[string]interface{})
		temp["orderInfo"] = orderInfo
		temp["orderGoods"] = orderGoods
		goodsBuffer[index] = temp
	}
	beego.Error(goodsBuffer)
	self.Data["goodsBuffer"] = goodsBuffer
	self.Data["nginxHost"] = beego.AppConfig.String("nginxHost")
	self.Layout = "userCenterLayout.html"
	self.TplName = "user_center_order.html"
}

/*展示用户中心地址页*/
func (self *UserController) ShowUserCenterSite() {
	userName := GetUser(&self.Controller)


	//self.Data["userName"] = userName

	o := orm.NewOrm()
	var addr models.Address
	o.QueryTable(&addr).RelatedSel("User").Filter("User__Name",userName).Filter("Isdefault",true).One(&addr)
	self.Data["addr"] = addr

	self.Layout = "userCenterLayout.html"
	self.TplName = "user_center_site.html"
}

/*处理提交的地址*/
func (self *UserController) HandleUserCenterSite() {
	// 获取数据
	receiver := self.GetString("receiver")
	addr := self.GetString("addr")
	zipCode := self.GetString("zipCode")
	phone := self.GetString("phone")

	// 校验数据
	if receiver == "" || addr == "" || zipCode == "" || phone == "" {
		beego.Info("添加数据不完整")
		self.Redirect("/user/userCenterSite",302)
		return
	}

	// 处理数据
	o := orm.NewOrm()
	var userAddr models.Address
	userAddr.Isdefault = true
	err := o.Read(&userAddr, "Isdefault")

	// 添加默认地址之前，需要把原来的默认地址更新成非默认地址
	if err == nil {
		userAddr.Isdefault = false
		o.Update(&userAddr)
	}

	// 关联
	userName := GetUser(&self.Controller)
	var user models.User
	user.Name = userName

	o.Read(&user,"Name") // 通过Id查找时可以省略，其他时候必须指定字段
	var newUserAddr models.Address
	newUserAddr.Receiver = receiver
	newUserAddr.Zipcode = zipCode
	newUserAddr.Addr = addr
	newUserAddr.Phone = phone
	newUserAddr.Isdefault = true
	newUserAddr.User = &user
	_, err = o.Insert(&newUserAddr)
	beego.Info(err)

	// 返回视图
	self.Redirect("/user/userCenterSite", 302)
}

/*退出登录*/
func (self *UserController) Logout() {
	beego.Info("11111111")

	RemoveName()
	self.Redirect("/login",302)
}

func RemoveName()  {
	conn, err := redis.Dial("tcp",beego.AppConfig.String("redisServer"))
	if err != nil {
		beego.Info("连接错误")
		return
	}
	beego.Info("00000")
	defer conn.Close()
	_, err = conn.Do("del","userName")

	if err != nil {
		beego.Info("删除登陆信息错误",err)
		return
	}
}

func saveName(userName string) {
	conn, err := redis.Dial("tcp",beego.AppConfig.String("redisServer"))
	if err != nil {
		beego.Info("Redis连接错误")
		return
	}

	defer conn.Close()

	_, err = conn.Do("set","userName",userName)

	if err != nil {
		beego.Info("Redis插入错误",err)
		return
	}

}




