package controllers

import (
	"github.com/astaxie/beego"
)

type LogoutController struct {
	beego.Controller
}

func (this *LogoutController) Get() {
	this.DelSession(tokenName)
	this.Redirect("/home", 302)
}
