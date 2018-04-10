package routers

import (
	"../controllers"

	"github.com/astaxie/beego"
)

func init() {
	beego.Router("/", &controllers.MainController{})
	beego.Router("/home", &controllers.HomeController{})
	beego.Router("/livingList", &controllers.LivingListController{})
	beego.Router("/logout", &controllers.LogoutController{})
	beego.Router("/opt", &controllers.OPTController{})
}
