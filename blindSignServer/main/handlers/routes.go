package handlers

import (
	"blindSignAccount/main/model"
	"github.com/gin-gonic/gin"
)

func InitRoutes(r *gin.Engine) {
	r.GET(model.RoutePath(model.PathGetSystemInformation).String(), GetSystemInformation)
	r.GET(model.RoutePath(model.PathReset).String(), GetReset)
	r.POST(model.RoutePath(model.PathSendBooking).String(), SendBooking)
	r.POST(model.RoutePath(model.PathGetBookingCode).String(), HdlGetBookingCode)
	r.POST(model.RoutePath(model.PathBlindSignature).String(), PostBlindSignature)
	r.POST(model.RoutePath(model.PathSetAddress).String(), HdlSetAddress)
	r.POST(model.RoutePath(model.PathAccessBonusSystem).String(), HdlAccessBonusSystem)
	r.POST(model.RoutePath(model.PathParticipate).String(), HdlParticipate)
	r.POST(model.RoutePath(model.PathCanBesUsedForRecovery).String(), HdlCanBeUsedForRecovery)
	r.POST(model.RoutePath(model.PathRecoveryTest).String(), HdlRecoveryTest)
	r.GET(model.RoutePath(model.PathRegister).String(), GetSystemRegister)
	r.POST(model.RoutePath(model.PathExit).String(), PostSystemExit)
	r.GET(model.RoutePath(model.PathStatistic).String(), GetSystemStatistic)
	r.GET(model.RoutePath(model.PathDebugInfos).String(), GetDebugInformation)
	r.POST(model.RoutePath(model.PathLastAdrBdl).String(), HdlGetLastAdrBundle)
}
