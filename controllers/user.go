package controllers

import (
	"github.com/astaxie/beego"
	"regexp"
	"github.com/astaxie/beego/orm"
	"dailyfresh/models"
	"github.com/astaxie/beego/utils"
	"strconv"
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



