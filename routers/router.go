package routers

import (
	"manageChain/controllers"

	"github.com/astaxie/beego"
)

func init() {
	// beego.InsertFilter("*", beego.BeforeRouter, cors.Allow(&cors.Options{
	// 	// AllowAllOrigins:  true,
	// 	AllowMethods:     []string{"POST", "GET"},
	// 	AllowHeaders:     []string{"Origin", "Authorization", "Access-Control-Allow-Origin", "Access-Control-Allow-Headers", "Content-Type", "x-requested-with", "no-referrer-when-downgrade"},
	// 	ExposeHeaders:    []string{"Content-Length", "Access-Control-Allow-Origin", "Access-Control-Allow-Headers", "Content-Type", "Access-Control-Allow-Origin"},
	// 	AllowCredentials: true,
	// 	AllowOrigins:     beego.AppConfig.Strings("Allowip"),
	// }))

	beego.Router("/", &controllers.MainController{})
	beego.Router("/gencrypto", &controllers.ChannelController{}, "post:GenCrypto")
	beego.Router("/gengenesisblock", &controllers.ChannelController{}, "post:GenGenesisBlock")
	// beego.Router("/genchannelconfig", &controllers.ChannelController{}, "post:GenChannelConfig")
	beego.Router("/channel/identity", &controllers.ChannelController{}, "post:Identity")
	beego.Router("/channel/addorg", &controllers.ChannelController{}, "post:AddOrg")
	beego.Router("/channel/deleteorg", &controllers.ChannelController{}, "post:DeleteOrg")
	beego.Router("/channel/create", &controllers.ChannelController{}, "post:CreateChannel")
	beego.Router("/channel/join", &controllers.ChannelController{}, "post:JoinChannel")

	beego.Router("/chaincode/install", &controllers.ChaincodeController{}, "post:InstallChaincode")
	beego.Router("/chaincode/instantiate", &controllers.ChaincodeController{}, "post:InstantiateChaincode")
	beego.Router("/chaincode/invoke ", &controllers.ChaincodeController{}, "post:Invoke")

}
