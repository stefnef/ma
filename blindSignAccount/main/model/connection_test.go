package model

import "testing"

func TestUtConnection_Reset(t *testing.T) {
	s := setupServer()
	utCon := newUtConnection()
	utCon.server = s

	// chang any value of s
	s.CntReqRegister++

	// a reset hast to clear that value
	if err := utCon.Reset(); err != nil {
		t.Error(err)
		t.FailNow()
	}
	if s.CntReqRegister != 0 {
		t.Errorf("value not changed: %d", s.CntReqRegister)
		t.Fail()
	}
}

func TestUtConnection_Register(t *testing.T) {
	s := setupServer()
	utCon := newUtConnection()
	utCon.server = s

	if _, err := utCon.Register(); err != nil {
		t.Error(err)
		t.FailNow()
	}
}

func TestRecoveryStatus_String(t *testing.T) {
	strRep := RecoveryStatus(RecoveryTestAfterAccess).String()
	if strRep != "RecoveryTestAfterAccess" {
		t.Errorf("wrong string representation: %s", strRep)
	}

	strRep = RecoveryStatus(RecoveryTestAfterFirstAdrUpd).String()
	if strRep != "RecoveryTest After 1st Adr Upd" {
		t.Errorf("wrong string representation: %s", strRep)
	}

	strRep = RecoveryStatus(RecoveryTestPenultimateAdr).String()
	if strRep != "RecoveryTest Penultimate Adr" {
		t.Errorf("wrong string representation: %s", strRep)
	}

	strRep = RecoveryStatus(RecoveryTest).String()
	if strRep != "RecoveryTest" {
		t.Errorf("wrong string representation: %s", strRep)
	}

	strRep = RecoveryStatus(Failure).String()
	if strRep != "Failure" {
		t.Errorf("wrong string representation: %s", strRep)
	}

	strRep = RecoveryStatus(-1).String()
	if strRep != "unknown status" {
		t.Errorf("wrong string representation: %s", strRep)
	}
	strRep = RecoveryStatus(5).String()
	if strRep != "unknown status" {
		t.Errorf("wrong string representation: %s", strRep)
	}

}
