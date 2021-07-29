package model

import (
	"blindSignAccount/main/crypt"
	"errors"
	"testing"
)

func testSetupForRestTests() (connection *RestConnection, err error) {
	connection = NewRestConnection()
	if connection == nil || connection.netClient == nil {
		return nil, errors.New("no rest connection available")
	}

	// check if server is up
	if err = connection.Reset(); err != nil {
		return nil, err
	}

	return connection, nil
}

func TestNewRestConnection(t *testing.T) {
	con := NewRestConnection()
	var err error

	if con == nil || con.netClient == nil {
		t.Error("no rest connection available")
		t.FailNow()
	}

	// check if server is up
	if err = con.Reset(); err != nil {
		t.Log(err)
		t.Skip("server is down")
	}
}

func TestRestConnection_GetSystemInformation(t *testing.T) {
	var con *RestConnection
	var flights []*Flight
	var bLevels []*BonusLevel
	var err error

	if con, err = testSetupForRestTests(); err != nil {
		t.Log(err)
		t.Skip("server is down")
	}

	// check if server is up
	if flights, bLevels, err = con.GetSystemInformation(); err != nil {
		t.Error(err)
		t.Fail()
	}
	if len(flights) == 0 || len(bLevels) == 0 {
		t.Error("flights or bonus levels not loaded")
		t.Fail()
	}
}

func TestRestConnection_Register(t *testing.T) {
	var con *RestConnection
	var clientID int
	var err error
	if con, err = testSetupForRestTests(); err != nil {
		t.Log(err)
		t.Skip("server is down")
	}
	clientID = -1
	if clientID, err = con.Register(); err != nil {
		t.Error(err)
		t.Fail()
	}
	if clientID != 0 {
		t.Errorf("wrong client id %d", clientID)
		t.Fail()
	}
	if clientID, err = con.Register(); err != nil {
		t.Error(err)
		t.Fail()
	}
	if clientID != 1 {
		t.Errorf("wrong client id %d", clientID)
		t.Fail()
	}
}

func TestRestConnection_SendBooking(t *testing.T) {
	var con *RestConnection
	var token string
	var err error

	if con, err = testSetupForRestTests(); err != nil {
		t.Log(err)
		t.Skip("server is down")
	}

	// check if server is up
	if token, err = con.SendBooking(1, 2, "low"); err != nil {
		t.Error(err)
		t.Fail()
	}
	if token == "" {
		t.Error("no Token for booking received")
		t.Fail()
	}
}

func TestRestConnection_GetBookingCode(t *testing.T) {
	var con *RestConnection
	var bCode string
	var hash, signature []byte
	var err error

	if con, err = testSetupForRestTests(); err != nil {
		t.Log(err)
		t.Skip("server is down")
	}

	// has to fail since no stupid hash and signature values are used
	hash = []byte{1, 2}
	signature = []byte{3, 4}
	if bCode, err = con.GetBookingCode("low", hash, signature); err == nil {
		t.Error("no error received")
		t.FailNow()
	}
	if err.Error() != "crypto/rsa: verification error" {
		t.Error("unexpected error: " + err.Error())
		t.Fail()
	}
	if bCode != "" {
		t.Error("received bonus code '" + bCode + "'")
		t.Fail()
	}
}

func TestRestConnection_GetBlindSignature(t *testing.T) {
	var con *RestConnection
	var signature string
	var blindToken = []byte{1, 2}
	var err error

	if con, err = testSetupForRestTests(); err != nil {
		t.Log(err)
		t.Skip("server is down")
	}

	// has to fail since a an invalid Token is used
	if signature, err = con.GetBlindSignature("low", "token", blindToken, ActionParticipate); err == nil {
		t.Error("no error received")
		t.FailNow()
	}
	if err.Error() != "Token is not valid" {
		t.Error("unexpected error: " + err.Error())
		t.Fail()
	}
	if signature != "" {
		t.Error("received signature '" + signature + "'")
		t.Fail()
	}
}

func TestRestConnection_AccessBonusSystem(t *testing.T) {
	var con *RestConnection
	var adrBundle = &crypt.AddressBundle{Seed: []byte{1, 2, 3, 4, 5, 6, 4, 7, 8, 9, 1, 2, 3, 4, 5, 6},
		Address: "adr", AccountID: 1, AddressID: 2}
	var codes = []string{"code1", "code2"}
	var tokens, recoveries map[string]string
	var err error

	if con, err = testSetupForRestTests(); err != nil {
		t.Log(err)
		t.Skip("server is down")
	}

	// has to fail since a an invalid Token is used
	if tokens, recoveries, err = con.AccessBonusSystem(codes, adrBundle); err == nil {
		t.Error("no error received")
		t.FailNow()
	}
	if err.Error() != "address does not fit to given seed" {
		t.Error("unexpected error: " + err.Error())
		t.Fail()
	}
	if len(tokens) != 0 || len(recoveries) != 0 {
		t.Error("received tokens or recoveries")
		t.Fail()
	}
}

func TestRestConnection_SetAddress(t *testing.T) {
	var con *RestConnection
	var adrBundle = &crypt.AddressBundle{Seed: []byte{1, 2, 3, 4, 5, 6, 4, 7, 8, 9, 1, 2, 3, 4, 5, 6},
		Address: "adr", AccountID: 1, AddressID: 2}
	var hash = []byte{1, 2, 3}
	var signature = []byte{5, 6}
	var token, recovery string
	var err error

	if con, err = testSetupForRestTests(); err != nil {
		t.Log(err)
		t.Skip("server is down")
	}

	// has to fail since seed is unknown
	if token, recovery, err = con.SetAddress("low", hash, signature, adrBundle, ActionBooking, "pkr"); err == nil {
		t.Error("no error received")
		t.FailNow()
	}
	if err.Error() != "seed unknown" {
		t.Error("unexpected error: " + err.Error())
		t.Fail()
	}
	if token != "" || recovery != "" {
		t.Error("received tokens or recoveries")
		t.Fail()
	}
}

func TestRestConnection_Participate(t *testing.T) {
	var con *RestConnection
	var hash = []byte{1, 2, 3}
	var signature = []byte{5, 6}
	var token, recovery, bData string
	var err error

	if con, err = testSetupForRestTests(); err != nil {
		t.Log(err)
		t.Skip("server is down")
	}

	// has to fail since hash and signature do not fit
	if token, recovery, bData, err = con.Participate("low", hash, signature, "pkr"); err == nil {
		t.Error("no error received")
		t.FailNow()
	}
	if err.Error() != "crypto/rsa: verification error" {
		t.Error("unexpected error: " + err.Error())
		t.Fail()
	}
	if token != "" || recovery != "" || bData != "" {
		t.Error("received tokens or recoveries or bonus data")
		t.Fail()
	}
}

func TestRestConnection_CanBeUsedForRecovery(t *testing.T) {
	var con *RestConnection
	var adrBundle = &crypt.AddressBundle{Seed: []byte{1, 2, 3, 4, 5, 6, 4, 7, 8, 9, 1, 2, 3, 4, 5, 6},
		Address: "adr", AccountID: 1, AddressID: 2}
	var status RecoveryStatus
	var token string
	var err error

	if con, err = testSetupForRestTests(); err != nil {
		t.Log(err)
		t.Skip("server is down")
	}

	// has to fail since hash and signature do not fit
	if status, token, err = con.CanBeUsedForRecovery("low", adrBundle); status != Failure || err == nil {
		t.Error("no failure received: " + status.String())
		t.FailNow()
	}
	if err.Error() != "address invalid" {
		t.Error(err)
		t.Fail()
	}
	if token != "" {
		t.Error("received a Token: " + token)
		t.Fail()
	}
}

func TestRestConnection_RecoveryTest(t *testing.T) {
	var con *RestConnection
	var adrBundle = &crypt.AddressBundle{Seed: []byte{1, 2, 3, 4, 5, 6, 4, 7, 8, 9, 1, 2, 3, 4, 5, 6},
		Address: "adr", AccountID: 1, AddressID: 2}
	var token, recoveryToken, bData string
	var err error

	if con, err = testSetupForRestTests(); err != nil {
		t.Log(err)
		t.Skip("server is down")
	}

	// has to fail since hash and signature do not fit
	token, recoveryToken, bData, err = con.RecoveryTest("low", "recoveryToken", "pkr", adrBundle)
	if err == nil {
		t.Error("no failure received")
		t.Fail()
	} else if err.Error() != "address invalid" {
		t.Error("wrong error message: " + err.Error())
		t.Fail()
	}
	if token != "" || recoveryToken != "" || bData != "" {
		t.Error("received tokens or recoveries or bonus data")
		t.Fail()
	}
}

func TestRestConnection_ClientBooking(t *testing.T) {
	if _, err := testSetupForRestTests(); err != nil {
		t.Log(err)
		t.Skip("server is down")
	}

	client := NewClient(1, utMnemonic, 2)
	client.con = NewRestConnection()
	if err := client.GetSystemInformation(); err != nil {
		t.Error(err)
		t.Fail()
	}

	if err := client.Booking(1, utLowLevelID); err != nil {
		t.Error(err)
		t.Fail()
	}
	if len(client.BonusCodes) != 1 {
		t.Errorf("wrong number of bonus codes: %d", len(client.BonusCodes))
		t.Fail()
	}
}

func TestRestConnection_ClientAccess(t *testing.T) {
	if _, err := testSetupForRestTests(); err != nil {
		t.Log(err)
		t.Skip("server is down")
	}

	client := NewClient(1, utMnemonic, 2)
	client.con = NewRestConnection()
	if err := client.GetSystemInformation(); err != nil {
		t.Error(err)
		t.Fail()
	}
	// create bookings
	_ = client.Booking(1, utMiddleLevelID)
	_ = client.Booking(1, utMiddleLevelID)
	_ = client.Booking(1, utMiddleLevelID)
	// try to access
	if err := client.AccessBonusSystem(); err != nil {
		t.Error(err)
		t.Fail()
	}
	if len(client.BLevelToTokens) != 2 {
		t.Errorf("wrong number of accessed bonus levels: %d", len(client.BLevelToTokens))
		t.Fail()
	}
	if client.BLevelToTokens[utLowLevelID] == "" || client.BLevelToTokens[utMiddleLevelID] == "" ||
		client.BLevelToTokens[utLowLevelID] == client.BLevelToTokens[utMiddleLevelID] {
		t.Error("wrong mapping of bonus levels")
		t.Fail()
	}
}

func TestRestConnection_ClientParticipate(t *testing.T) {
	var bData string
	var err error

	if _, err = testSetupForRestTests(); err != nil {
		t.Log(err)
		t.Skip("server is down")
	}

	client := NewClient(1, utMnemonic, 2)
	client.con = NewRestConnection()
	if err := client.GetSystemInformation(); err != nil {
		t.Error(err)
		t.Fail()
	}
	// create bookings
	_ = client.Booking(1, utMiddleLevelID)
	_ = client.Booking(1, utMiddleLevelID)
	_ = client.Booking(1, utMiddleLevelID)
	// access bonus system 'low' and 'middle'
	_ = client.AccessBonusSystem()

	// try to participate with level low
	if bData, err = client.Participate(utLowLevelID); err != nil {
		t.Error(err)
		t.Fail()
	}
	if bData == "" {
		t.Error("no bonus data received")
		t.Fail()
	}

	// try to participate with level middle
	bData = ""
	err = nil
	if bData, err = client.Participate(utMiddleLevelID); err != nil {
		t.Error(err)
		t.Fail()
	}
	if bData == "" {
		t.Error("no bonus data received")
		t.Fail()
	}

	// try to participate with level high => has to fail!
	bData = ""
	err = nil
	if bData, err = client.Participate(utHighLevelID); err == nil {
		t.Error("no error received")
		t.Fail()
	}
	if bData != "" {
		t.Error("no bonus data received")
		t.Fail()
	}

	// another participation in level middle must not fail
	// try to participate with level middle
	bData = ""
	err = nil
	if bData, err = client.Participate(utMiddleLevelID); err != nil {
		t.Error(err)
		t.Fail()
	}
	if bData == "" {
		t.Error("no bonus data received")
		t.Fail()
	}
}

func TestRestConnection_ClientRecovery(t *testing.T) {
	var bData, bData2nd string
	var bDataRecovered map[string]string
	var err error

	if _, err = testSetupForRestTests(); err != nil {
		t.Log(err)
		t.Skip("server is down")
	}

	client := NewClientWithAccountID(1, utMnemonic, 2, 2)
	client.con = NewRestConnection()
	if err := client.GetSystemInformation(); err != nil {
		t.Error(err)
		t.Fail()
	}
	// create bookings
	_ = client.Booking(1, utMiddleLevelID)
	_ = client.Booking(1, utMiddleLevelID)
	_ = client.Booking(1, utMiddleLevelID)
	// access bonus system 'low' and 'middle'
	_ = client.AccessBonusSystem()

	// try to participate with level low
	bData, err = client.Participate(utLowLevelID)
	bData2nd, err = client.Participate(utLowLevelID)

	// client forgets everything
	testClientDeleteHistory(client)

	// try to recover
	if bDataRecovered, err = client.RestoreFromMnemonic(1, utMnemonic, 2); err != nil {
		t.Error(err)
		t.Fail()
	}

	if bData == bDataRecovered[utLowLevelID] {
		t.Errorf("bonus data from 1st participation is equal. Exp: %s\t\tact: %s", bData, bDataRecovered[utLowLevelID])
		t.Fail()
	}
	if bData2nd != bDataRecovered[utLowLevelID] {
		t.Errorf("2nd bonus data differs. Exp: %s\t\tact: %s", bData, bDataRecovered[utLowLevelID])
		t.Fail()
	}
	if bDataRecovered[utMiddleLevelID] != "" {
		t.Errorf("unexpected bonus data: %s", bDataRecovered[utMiddleLevelID])
		t.Fail()
	}
}

func TestRestConnection_GetDebugInfos(t *testing.T) {
	var con *RestConnection
	var clientID int
	var server *Server
	var err error
	if con, err = testSetupForRestTests(); err != nil {
		t.Log(err)
		t.Skip("server is down")
	}
	// first, make a registration
	if clientID, err = con.Register(); err != nil {
		t.Error(err)
		t.FailNow()
	}

	// try to get server information, now
	if server, err = con.GetDebugInfos(); err != nil {
		t.Error(err)
		t.FailNow()
	}

	// the received client id has to be in the debug information
	if len(server.ClientIDs) != 1 {
		t.Error("wrong length of clients in debug information")
		t.FailNow()
	}
	if server.ClientIDs[0] != clientID {
		t.Error("wrong client id in debug information")
		t.FailNow()
	}

}

func TestRestConnection_GetLastAdrBdl(t *testing.T) {
	var con *RestConnection
	var adr string
	var accountID uint32
	var err error
	if con, err = testSetupForRestTests(); err != nil {
		t.Log(err)
		t.Skip("server is down")
	}
	client := NewClientWithAccountID(1, utMnemonic, 2, 2)
	client.con = NewRestConnection()
	if err := client.GetSystemInformation(); err != nil {
		t.Error(err)
		t.Fail()
	}
	// create bookings
	_ = client.Booking(1, utMiddleLevelID)
	_ = client.Booking(1, utMiddleLevelID)
	_ = client.Booking(1, utMiddleLevelID)
	// access bonus system 'low' and 'middle'
	_ = client.AccessBonusSystem()
	lastAdrMiddle := crypt.GetAddress(client.Keys[4], client.AddressID-1).String()

	// try to participate with level low
	_, err = client.Participate(utLowLevelID)
	_, err = client.Participate(utLowLevelID)
	lastAdrLow := crypt.GetAddress(client.Keys[4], client.AddressID-1).String()

	// the last address of the bonus level "low"
	if adr, accountID, err = con.GetLastAdrBdl(client.Seed, utLowLevelID); err != nil {
		t.Error(err)
		t.Fail()
	}
	if adr != lastAdrLow || accountID != client.AccountID {
		t.Error("wrong last address or account id received")
		t.Fail()
	}

	// the last address of the bonus level "middle"
	if adr, accountID, err = con.GetLastAdrBdl(client.Seed, utMiddleLevelID); err != nil {
		t.Error(err)
		t.Fail()
	}
	if adr != lastAdrMiddle || accountID != client.AccountID {
		t.Error("wrong last address or account id received")
		t.Fail()
	}

}
