package controllers

import (
	// "fmt"
	"manageChain/protocols"

	"github.com/astaxie/beego"
	logger "github.com/astaxie/beego/logs"
)

type BaseController struct {
	beego.Controller
}

func (c *BaseController) ReturnErrorCode(code string, msg string) {
	logger.Error("Code: ", code)
	c.Ctx.Output.SetStatus(500)
	c.Data["json"] = &protocols.ErrorMessage{
		Code:    code,
		Message: msg,
	}
	c.ServeJSON()
}

// ReturnErrorMsg return given message to the front end
func (c *BaseController) ReturnErrorMsg(err error) {
	logger.Error("Got error: ", err)
	c.Ctx.Output.SetStatus(500)
	c.Data["json"] = &protocols.ErrorMessage{
		Message: err.Error(),
	}
	c.ServeJSON()
}

func (c *BaseController) ReturnOKMsg(data interface{}) {
	// logger.Debug("Got normal response: ", data)
	c.Ctx.Output.SetStatus(200)
	c.Data["json"] = data
	c.ServeJSON()
}
