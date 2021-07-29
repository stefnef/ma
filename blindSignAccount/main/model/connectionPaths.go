package model

type RoutePath int

const (
	PathSendBooking = iota
	PathLastAdrBdl
	PathGetBookingCode
	PathGetSystemInformation
	PathBlindSignature
	PathSetAddress
	PathAccessBonusSystem
	PathParticipate
	PathCanBesUsedForRecovery
	PathRecoveryTest
	PathRegister
	PathExit
	PathStatistic
	PathDebugInfos
	PathReset
)

var ServerAddress string

func (path RoutePath) String() string {
	names := [...]string{
		"/booking/send", "/recovery/lastAdrBdl", "/booking/code", "/system/info", "/blindSignature",
		"/setAddress", "/accessBonusSystem", "/participate",
		"/recovery/canBeUsedForRecovery", "/recovery/test", "/system/register", "/system/exit",
		"/system/statistic", "/system/debug", "/system/reset",
	}
	if path < PathSendBooking || path > PathReset {
		return "unknown path"
	}
	return names[path]
}
