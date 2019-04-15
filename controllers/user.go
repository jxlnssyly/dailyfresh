package controllers

import (
	"github.com/astaxie/beego"
	"regexp"
	"github.com/astaxie/beego/orm"
	"dailyfresh/models"
	"github.com/astaxie/beego/utils"
	"strconv"
	"encoding/base64"
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
	user.Actice = true
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

	if user.Actice != true {

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
	self.SetSession("userName",userName)
	self.Redirect("/",302)
}

/*用户中心*/

func (self *UserController) ShowUserCenterInfo() {
	userName := GetUser(&self.Controller)


	// 查询地址表的内容
	o := orm.NewOrm()

	var addr models.Address

	// 高级查询
	o.QueryTable(&models.Address{}).RelatedSel("User").Filter("User__Name",userName).Filter("Isdefault",true).One(&addr)

	if addr.Id == 0	 {
		self.Data["addr"] = ""
	} else {
		self.Data["addr"] = addr
	}

	self.Layout = "userCenterLayout.html"
	self.TplName = "user_center_info.html"
}

/*退出登录*/
func (self *UserController) Logout() {
	self.DelSession("userName")

	self.Redirect("/login",302)
}




