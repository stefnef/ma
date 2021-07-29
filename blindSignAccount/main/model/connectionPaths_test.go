package model

import "testing"

func TestRoutePath_String(t *testing.T) {
	strRep := RoutePath(PathSendBooking).String()
	if strRep != "/booking/send" {
		t.Errorf("wrong string representation: %s", strRep)
	}

	strRep = RoutePath(PathLastAdrBdl).String()
	if strRep != "/recovery/lastAdrBdl" {
		t.Errorf("wrong string representation: %s", strRep)
	}

	strRep = RoutePath(PathGetBookingCode).String()
	if strRep != "/booking/code" {
		t.Errorf("wrong string representation: %s", strRep)
	}

	strRep = RoutePath(PathGetSystemInformation).String()
	if strRep != "/system/info" {
		t.Errorf("wrong string representation: %s", strRep)
	}

	strRep = RoutePath(PathBlindSignature).String()
	if strRep != "/blindSignature" {
		t.Errorf("wrong string representation: %s", strRep)
	}

	strRep = RoutePath(PathSetAddress).String()
	if strRep != "/setAddress" {
		t.Errorf("wrong string representation: %s", strRep)
	}
	strRep = RoutePath(PathAccessBonusSystem).String()
	if strRep != "/accessBonusSystem" {
		t.Errorf("wrong string representation: %s", strRep)
	}

	strRep = RoutePath(PathParticipate).String()
	if strRep != "/participate" {
		t.Errorf("wrong string representation: %s", strRep)
	}

	strRep = RoutePath(PathCanBesUsedForRecovery).String()
	if strRep != "/recovery/canBeUsedForRecovery" {
		t.Errorf("wrong string representation: %s", strRep)
	}

	strRep = RoutePath(PathRecoveryTest).String()
	if strRep != "/recovery/test" {
		t.Errorf("wrong string representation: %s", strRep)
	}

	strRep = RoutePath(PathRegister).String()
	if strRep != "/system/register" {
		t.Errorf("wrong string representation: %s", strRep)
	}

	strRep = RoutePath(PathExit).String()
	if strRep != "/system/exit" {
		t.Errorf("wrong string representation: %s", strRep)
	}

	strRep = RoutePath(PathStatistic).String()
	if strRep != "/system/statistic" {
		t.Errorf("wrong string representation: %s", strRep)
	}

	strRep = RoutePath(PathDebugInfos).String()
	if strRep != "/system/debug" {
		t.Errorf("wrong string representation: %s", strRep)
	}

	strRep = RoutePath(PathReset).String()
	if strRep != "/system/reset" {
		t.Errorf("wrong string representation: %s", strRep)
	}

	strRep = RoutePath(-1).String()
	if strRep != "unknown path" {
		t.Errorf("wrong string representation: %s", strRep)
	}
	strRep = RoutePath(15).String()
	if strRep != "unknown path" {
		t.Errorf("wrong string representation: %s", strRep)
	}
}
