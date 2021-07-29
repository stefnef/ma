package model

import (
	"blindSignAccount/main/crypt"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
)

func setupClient(t *testing.T) *Client {
	client := NewClient(1, utMnemonic, 2)
	if client == nil {
		t.Error("could not create client")
		t.FailNow()
	}
	client.con = newUtConnection()
	if err := client.GetSystemInformation(); err != nil {
		t.Error(err)
		t.Fail()
	}

	return client
}

func TestNewClient(t *testing.T) {
	setupClient(t)
}

func TestClient_Booking(t *testing.T) {
	client := setupClient(t)
	if err := client.Booking(1, utLowLevelID); err != nil {
		t.Error(err)
		t.Fail()
	}
	// check that the bonus code has a new entry
	if len(client.BonusCodes) != 1 {
		t.Error("wrong test setup")
		t.Fail()
	}
	if client.BonusCodes[0].ValidFor != client.BonusLevels[utLowLevelID] {
		t.Error("wrong mapping bonus code <--> bonus level")
		t.Fail()
	}
}

func TestClient_Booking_Fail(t *testing.T) {
	client := setupClient(t)
	// try to book a non-existing flight
	if err := client.Booking(-1, utLowLevelID); err == nil {
		t.Error("success for booking with invalid flight id")
		t.Fail()
	}
	if len(client.BonusCodes) != 0 {
		t.Error("there exists a bonus code")
		t.Fail()
	}
	// try to book an existing flight, but with non-existing bonus level
	if err := client.Booking(0, utLowLevelID+"_unknown"); err == nil {
		t.Error("success for booking with invalid bonus level")
		t.Fail()
	}
	if len(client.BonusCodes) != 0 {
		t.Error("there exists a bonus code")
		t.Fail()
	}
}

func TestClient_AccessBonusSystem(t *testing.T) {
	client := setupClient(t)

	// create 5 bookings for bonus level 'middle'
	for i := 0; i < 5; i++ {
		if err := client.Booking(i, utMiddleLevelID); err != nil {
			t.Error(err)
			t.Fail()
		}
	}

	err := client.AccessBonusSystem()
	if err != nil {
		t.Error(err)
		t.Fail()
	}

	// check that the client has a correct mapping
	if len(client.BLevelToTokens) != 2 ||
		client.BLevelToTokens[utLowLevelID] == "" || client.BLevelToTokens[utMiddleLevelID] == "" {
		t.Error("wrong mapping of bonus levels to tokens")
		t.Fail()
	}

	// received recovery tokens?
	if len(client.BLevelToRecovery) != 2 ||
		client.BLevelToRecovery[utLowLevelID] == "" || client.BLevelToRecovery[utMiddleLevelID] == "" {
		t.Error("no recovery tokens received")
		t.Fail()
	}
	// check that both recovery tokens differ
	if client.BLevelToRecovery[utLowLevelID] == client.BLevelToRecovery[utMiddleLevelID] {
		t.Error("recovery tokens do not differ")
		t.Fail()
	}
}

func TestClient_Participate(t *testing.T) {
	client := setupClient(t)
	// the client needs an initial recovery Token
	initialRecoveryToken := "INITIAL RECOVERY TOKEN"
	client.BLevelToRecovery[utMiddleLevelID] = initialRecoveryToken

	// setup a valid Token that can be used for participation
	initialToken := "INITIAL_TOKEN"
	client.BLevelToTokens[utMiddleLevelID] = initialToken
	serverParticipate := client.con.(*utConnection).server.BonusList[utMiddleLevelID].ActionVariants[ActionParticipate]
	serverParticipate.ValidTokens[initialToken] = false
	serverParticipate.TokenToSeed[initialToken] = hex.EncodeToString(client.Seed)
	serverParticipate.SeedToAddress[hex.EncodeToString(client.Seed)] = string("INITIAL_ADDRESS")
	serverParticipate.SeedToAccountID[hex.EncodeToString(client.Seed)] = 2
	serverParticipate.AddressToRecovery["INITIAL_ADDRESS"] = initialRecoveryToken

	// try to participate
	if bonusData, err := client.Participate(utMiddleLevelID); err != nil || bonusData == "" {
		t.Error(err)
		t.Fail()
	}

	if len(serverParticipate.PkrToAdrUpd) != 1 {
		t.Errorf("wrong number of bonus data")
		t.Fail()
	}
	if len(serverParticipate.PkrToBonusData) != 1 {
		t.Errorf("wrong number of bonus data")
		t.Fail()
	}
	if len(serverParticipate.SeedToAccountID) != 1 {
		t.Errorf("wrong number of account IDs")
		t.Fail()
	}

	for tmpPkr, blindSigPair := range serverParticipate.PkrToBonusData {
		if _, ok := serverParticipate.PkrToAdrUpd[tmpPkr]; ok == true {
			t.Errorf("same pkr in both maps")
			t.Fail()
		}
		if blindSigPair.RecoveryToken != client.BLevelToRecovery[utMiddleLevelID] {
			t.Error("wrong mapping of recovery Token")
			t.Fail()
		}
		if blindSigPair.Token != client.BLevelToTokens[utMiddleLevelID] {
			t.Error("wrong mapping of Token")
			t.Fail()
		}
	}

	// 2nd time has to be successful
	if bonusData, err := client.Participate(utMiddleLevelID); err != nil || bonusData == "" {
		t.Error(err)
		t.Fail()
	}
	if len(serverParticipate.PkrToAdrUpd) != 2 {
		t.Errorf("wrong number of bonus data")
		t.Fail()
	}
	if len(serverParticipate.PkrToBonusData) != 2 {
		t.Errorf("wrong number of bonus data")
		t.Fail()
	}
}

func TestClient_RestoreAfterAccess(t *testing.T) {
	// a client which accesses the bonus system
	// restore before address update
	client := setupClient(t)
	if err := clientAccessesBonusSystem(client); err != nil {
		t.Error()
		t.Fail()
	}

	// client looses data
	testClientDeleteHistory(client)

	// participation has to fail
	if _, err := client.Participate(utMiddleLevelID); err == nil {
		t.Error("participation did not fail")
		t.Fail()
	}

	// restore must not fail
	if bData, err := client.Restore(utMiddleLevelID); err != nil || bData != "" {
		t.Error("restore directly after bonus access")
		t.Fail()
	}

	// participation must not fail
	if _, err := client.Participate(utMiddleLevelID); err != nil {
		t.Error(err)
		t.Fail()
	}
}

func TestClient_RestoreAfterFirstAdrUpd(t *testing.T) {
	// a client which accesses the bonus system
	// restore after the first address update
	client := setupClient(t)
	if err := clientAccessesBonusSystem(client); err != nil {
		t.Error()
		t.Fail()
	}
	// execute one address update
	bLevel := client.BonusLevels[utMiddleLevelID]
	//Token, RecoveryToken, _, _, err := client.AdrUpdate(bLevel)
	_, _, _, _, err := client.AdrUpdate(bLevel)
	if err != nil {
		t.Error(err)
		t.Fail()
	}

	// client looses access data
	testClientDeleteHistory(client)

	// participation has to fail
	if _, err := client.Participate(utMiddleLevelID); err == nil {
		t.Error("participation did not fail")
		t.Fail()
	}

	// restore must not fail
	if bData, err := client.Restore(utMiddleLevelID); err != nil || bData != "" {
		t.Error("restore after first address update ")
		t.Fail()
	}

	// participation must not fail
	if _, err := client.Participate(utMiddleLevelID); err != nil {
		t.Error(err)
		t.Fail()
	}
}

func TestClient_RestoreAfterParticipation(t *testing.T) {
	// a client which accesses the bonus system
	// restore after the first participation
	testClientRestoreAfterNthParticipation(t, 1)
}

func TestClient_RestoreAfter2ndAdrUpd(t *testing.T) {
	// restore after the second address update

	client := setupClient(t)
	if err := clientAccessesBonusSystem(client); err != nil {
		t.Error()
		t.Fail()
	}
	// execute a participation
	if _, err := client.Participate(utMiddleLevelID); err != nil {
		t.Error(err)
		t.Fail()
	}
	// execute one address update
	bLevel := client.BonusLevels[utMiddleLevelID]
	//Token, RecoveryToken, _, _, err := client.AdrUpdate(bLevel)
	_, _, _, _, err := client.AdrUpdate(bLevel)
	if err != nil {
		t.Error(err)
		t.Fail()
	}

	// client looses access data
	testClientDeleteHistory(client)
	// participation has to fail
	if _, err := client.Participate(utMiddleLevelID); err == nil {
		t.Error("participation did not fail")
		t.Fail()
	}
	// restore must not fail
	if bData, err := client.Restore(utMiddleLevelID); err != nil || bData == "" {
		t.Error("restore after 2nd address update ")
		t.Fail()
	}
	// participation must not fail
	if _, err := client.Participate(utMiddleLevelID); err != nil {
		t.Error(err)
		t.Fail()
	}
}

func TestClient_RestoreAfter2ndParticipation(t *testing.T) {
	// restore after the second participation
	testClientRestoreAfterNthParticipation(t, 2)
}

func TestClient_RestoreAfter10thParticipation(t *testing.T) {
	testClientRestoreAfterNthParticipation(t, 10)
}

func TestClient_RestoreAfter20thParticipation(t *testing.T) {
	testClientRestoreAfterNthParticipation(t, 20)
}

func TestClient_RestoreAfter11thParticipation(t *testing.T) {
	testClientRestoreAfterNthParticipation(t, 24)
	testClientRestoreAfterNthParticipation(t, 25)
	testClientRestoreAfterNthParticipation(t, 26)
}

func testClientRestoreAfterNthParticipation(t *testing.T, n int) {
	// restore after the nTh participation

	client := setupClient(t)
	if err := clientAccessesBonusSystem(client); err != nil {
		t.Error()
		t.Fail()
	}
	// execute n times participation
	for i := 0; i < n; i++ {
		if _, err := client.Participate(utMiddleLevelID); err != nil {
			t.Error(err)
			t.Fail()
		}
	}
	// client looses access data
	if uint32(n) >= maxAdrID {
		if uint32(n) < 2*maxAdrID {
			client.AccountID = 87
		} else {
			client.AccountID = 47
		}
	} else {
		client.AccountID = 81
	}
	testClientDeleteHistory(client)
	// participation has to fail
	if _, err := client.Participate(utMiddleLevelID); err == nil {
		t.Error("participation did not fail")
		t.Fail()
	}
	// restore must not fail
	if bData, err := client.Restore(utMiddleLevelID); err != nil || bData == "" {
		t.Errorf("restore after %d. participation", n)
		t.Fail()
	}
	// participation must not fail
	if _, err := client.Participate(utMiddleLevelID); err != nil {
		t.Error(err)
		t.Fail()
	}
}

func TestClient_RestoreDifferentBLevels(t *testing.T) {
	// restore after the nTh participation
	client := setupClient(t)
	if err := clientAccessesBonusSystem(client); err != nil {
		t.Error()
		t.Fail()
	}
	// participate with middle level
	if _, err := client.Participate(utMiddleLevelID); err != nil {
		t.Error(err)
		t.Fail()
	}
	// participate with low level
	if _, err := client.Participate(utLowLevelID); err != nil {
		t.Error(err)
		t.Fail()
	}

	// client looses access data
	testClientDeleteHistory(client)
	// participation has to fail for both levels
	if _, err := client.Participate(utMiddleLevelID); err == nil {
		t.Error("participation did not fail for middle level")
		t.Fail()
	}
	// participation has to fail
	if _, err := client.Participate(utLowLevelID); err == nil {
		t.Error("participation did not fail for low level")
		t.Fail()
	}
	// restore must not fail
	if bData, err := client.Restore(utMiddleLevelID); err != nil || bData == "" {
		t.Error("restore after participation")
		t.Fail()
	}
	// participation must not fail for middle but for low level
	if _, err := client.Participate(utMiddleLevelID); err != nil {
		t.Errorf("middle level: %s", err)
		t.Fail()
	}
	if _, err := client.Participate(utLowLevelID); err == nil {
		t.Error("participation did not fail for low level")
		t.Fail()
	}
}

func clientAccessesBonusSystem(client *Client) error {
	// create 5 bookings for bonus level 'middle'
	for i := 0; i < 5; i++ {
		if err := client.Booking(i, utMiddleLevelID); err != nil {
			return err
		}
	}

	return client.AccessBonusSystem()
}

func testClientDeleteHistory(client *Client) {
	client.BLevelToTokens = map[string]string{}
	client.BLevelToRecovery = map[string]string{}
	client.BonusCodes = []*BonusCode{}
	client.BonusLevels = map[string]*BonusLevel{}
	client.flightList = map[int]*Flight{}
	client.AddressID = 0
	_ = client.GetSystemInformation()
	_, client.Keys, _ = crypt.GetWalletKeys(client.Mnemonic, client.AccountID, false)
}

func TestClient_RestoreFromMnemonic(t *testing.T) {
	// restore after participation

	// this time: use another recovery id
	client := NewClient(1, utMnemonic, 22)
	utConnection := newUtConnection()
	oldAccountID := client.AccountID
	if client == nil {
		t.Error("could not create client")
		t.Fail()
		return
	}
	client.con = utConnection
	if err := client.GetSystemInformation(); err != nil {
		t.Error(err)
		t.Fail()
	}

	if err := clientAccessesBonusSystem(client); err != nil {
		t.Error()
		t.Fail()
	}
	// execute participation
	if _, err := client.Participate(utMiddleLevelID); err != nil {
		t.Error(err)
		t.Fail()
	}
	// 3 times low level
	for iteration := 0; iteration < 2; iteration++ {
		if _, err := client.Participate(utLowLevelID); err != nil {
			t.Error(err)
			t.Fail()
		}
	}

	// remember the last address id
	oldAdrID := client.AddressID
	// remember the tokens and recovery tokens
	tokLow := client.BLevelToTokens[utLowLevelID]
	tokMiddle := client.BLevelToTokens[utMiddleLevelID]
	tokHigh := client.BLevelToTokens[utHighLevelID]
	recLow := client.BLevelToRecovery[utLowLevelID]
	recMiddle := client.BLevelToRecovery[utMiddleLevelID]
	recHigh := client.BLevelToRecovery[utHighLevelID]

	// ******************************************************************
	// ************* client looses access data **************************
	// ******************************************************************
	client = NewClient(1, utMnemonic, 22)
	client.con = utConnection

	// participation has to fail
	if _, err := client.Participate(utMiddleLevelID); err == nil {
		t.Error("participation did not fail")
		t.Fail()
	}
	// restore must not fail
	if bData, err := client.RestoreFromMnemonic(1, utMnemonic, 22); err != nil || bData[utMiddleLevelID] == "" {
		t.Error(err)
		t.Fail()
	}
	if client.AccountID != oldAccountID {
		t.Errorf("could not restore the account id: act(%d) exp(%d)", client.AccountID, oldAccountID)
		t.Fail()
	}
	if client.AddressID != oldAdrID {
		t.Errorf("could not restore the address id: act(%d) exp(%d)", client.AddressID, oldAdrID)
		t.Fail()
	}
	if tokLow != client.BLevelToTokens[utLowLevelID] || tokMiddle != client.BLevelToTokens[utMiddleLevelID] ||
		tokHigh != client.BLevelToTokens[utHighLevelID] {
		t.Errorf("tokens not recovered correctly\nact\t\texp")
		t.Errorf("%s\t%s", tokLow, client.BLevelToTokens[utLowLevelID])
		t.Errorf("%s\t%s", tokMiddle, client.BLevelToTokens[utMiddleLevelID])
		t.Errorf("%s\t%s", tokHigh, client.BLevelToTokens[utHighLevelID])
		t.Fail()
	}
	if recLow != client.BLevelToRecovery[utLowLevelID] || recMiddle != client.BLevelToRecovery[utMiddleLevelID] ||
		recHigh != client.BLevelToRecovery[utHighLevelID] {
		t.Errorf("recovery tokens not recovered correctly\nact\t\texp")
		t.Errorf("%s\t%s", recLow, client.BLevelToRecovery[utLowLevelID])
		t.Errorf("%s\t%s", recMiddle, client.BLevelToRecovery[utMiddleLevelID])
		t.Errorf("%s\t%s", recHigh, client.BLevelToRecovery[utHighLevelID])
		t.Fail()
	}

	// participation must not fail
	if _, err := client.Participate(utMiddleLevelID); err != nil {
		t.Error(err)
		t.Fail()
	}
	// participation must not fail for low level
	if _, err := client.Participate(utLowLevelID); err != nil {
		t.Error(err)
		t.Fail()
	}
}

func TestClient_SaveDebugInfos(t *testing.T) {
	var matches []string
	var err error

	client := setupClient(t)
	if client == nil {
		t.Error("could not create client")
		t.Fail()
		return
	}
	if err := client.Booking(1, "high"); err != nil {
		t.Error(err)
		t.FailNow()
	}
	if err := client.AccessBonusSystem(); err != nil {
		t.Error(err)
		t.FailNow()
	}
	if err := client.SaveDebugInfos(); err != nil {
		t.Error(err)
		t.Fail()
	}

	if matches, err = filepath.Glob("1*_debug.json"); err != nil {
		t.Error(err)
		t.FailNow()
	}
	// check if there exists a debug file
	if len(matches) != 2 {
		t.Error("wrong number of found files")
		t.FailNow()
	}

	// try to parse the data
	if sameServer, err := NewServerFromFile(matches[1]); err != nil || sameServer == nil {
		t.Error("could not parse the server ")
		if err != nil {
			t.Error(err)
		}
		t.FailNow()
	}

	//remove the debug file
	if err := os.Remove(matches[0]); err != nil {
		t.Error(err)
		t.FailNow()
	}
	if err := os.Remove(matches[1]); err != nil {
		t.Error(err)
		t.FailNow()
	}

}

func TestClient_Restore_debug_2(t *testing.T) {
	clientID := 12
	mnemonic := "air air air air air air air air air air air air air air air air air air air air air air blind car"
	recoveryID := 3

	connection := NewRestConnection()
	if connection == nil || connection.netClient == nil {
		t.Error("no rest connection available")
		t.FailNow()
	}

	client := NewClient(clientID, mnemonic, recoveryID)
	client.con = connection
	if err := client.GetSystemInformation(); err != nil {
		t.Error(err)
		t.FailNow()
	}

	if err := client.Booking(1, "high"); err != nil {
		t.Error(err)
		t.FailNow()
	}

	if err := client.AccessBonusSystem(); err != nil {
		t.Error(err)
		t.FailNow()
	}

	if _, err := client.RestoreFromMnemonic(clientID, mnemonic, recoveryID); err != nil {
		t.Error(err)
		t.FailNow()
	}
	if client.AccountID != 9 {
		t.Errorf("accountID is not 9, but %d", client.AccountID)
		t.Fail()
	}

	//if err := client.SaveDebugInfos(); err != nil {
	//	t.Error(err)
	//	t.Fail()
	//}
}

func TestClient_GetAccountID(t *testing.T) {
	client := NewClientWithAccountID(1, utMnemonic, 4, uint32(23))
	if client.GetAccountID() != uint32(23) {
		t.Error("wrong account id of client")
		t.FailNow()
	}
}

func TestClient_SetAdrFromString(t *testing.T) {
	client := NewClientWithAccountID(1, utMnemonic, 4, uint32(23))
	client.AddressID = 2
	nextAdr := crypt.GetAddress(client.Keys[4], client.AddressID).String()
	client.AddressID = 0
	client.AddressID = client.getAdrIDFromString(nextAdr)
	if client.AddressID != 2 {
		t.Error("wrong address id from string rep")
		t.Fail()
	}

	client = NewClient(1, utMnemonic, 22)
	utConnection := newUtConnection()
	client.SetConnection(utConnection)
	client.AccountID = 81
	client.AddressID = 3
	lastAdr := crypt.GetAddress(client.Keys[4], client.AddressID).String()
	client.AddressID = 0
	client.AddressID = client.getAdrIDFromString(lastAdr)
	if client.AddressID != 3 {
		t.Errorf("wrong address id %d", client.AddressID)
		t.Fail()
	}
}
