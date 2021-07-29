package model

import (
	"blindSignAccount/main/crypt"
	"crypto"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"github.com/btcsuite/btcutil/hdkeychain"
	"github.com/cryptoballot/fdh"
	"github.com/cryptoballot/rsablind"
	"sort"
	"strconv"
	"testing"
)

var utMnemonic = "coil early bronze maze battle any core sweet burger busy cotton impact evoke oven jeans glance clock final eight crowd tool okay mushroom shrimp"

func setupServer() *Server {
	return NewServer()
}

func fail(t *testing.T, errMsg string) {
	t.Error(errMsg)
	t.Fail()
}

func TestNewServer(t *testing.T) {
	s := setupServer()
	// the setup is done with default bonus system

	// check that there are 3 bonus levels "low" - "middle" - "high"
	if s.BonusList["low"] == nil || s.BonusList["middle"] == nil || s.BonusList["high"] == nil {
		t.Error("bonus hierarchy system changed (level ids)!")
		t.Fail()
	}
	// check the hierarchy
	if len(s.Hierarchy) != 3 ||
		s.Hierarchy[2] != s.BonusList["low"] ||
		s.Hierarchy[1] != s.BonusList["middle"] ||
		s.Hierarchy[0] != s.BonusList["high"] {
		t.Error("bonus hierarchy system changed (priorities)!")
		t.Fail()
	}

	if s.flightMap == nil || len(s.flightMap) == 0 {
		t.Error("flightMap nil or initial")
		t.Fail()
	}
}

// Creates valid bonus codes for the highest bonus level
func createValidTestCodes(t *testing.T, server *Server) (codes []string) {
	for i := 0; i < 5; i++ {
		bc := NewBonusCode(server.Hierarchy[0])
		server.BonusCodes[bc.CodeID] = bc
		codes = append(codes, bc.CodeID)
	}
	if len(codes) == 0 {
		t.Fail()
	}
	return
}

func TestServer_Booking(t *testing.T) {
	s := setupServer()
	flightID := 0
	customerID := 12
	token, err := s.Booking(flightID, customerID, utLowLevelID)
	if err != nil {
		t.Error(err)
		t.Fail()
	}
	if token == "" {
		t.Error("no Token returned")
		t.Fail()
	}
	// the server has to know the Token
	if used, known := s.BonusList[utLowLevelID].ActionVariants[ActionBooking].ValidTokens[token]; !known || used {
		t.Errorf("Token is unknown (%t) or was marked as used (%t)", known, used)
		t.Fail()
	}
	if len(s.BonusCodes) != 0 {
		t.Error("code was created")
		t.Fail()
	}
	// the flight has to know the booking for the given flight
	booking := s.flightMap[flightID].Bookings[0]
	if booking == nil {
		t.Error("booking does not exist")
		t.Fail()
	} else {
		// check booking parameters
		if booking.CustomerID != customerID {
			t.Fail()
		}
		if booking.BonusLevel.BonusID != utLowLevelID {
			t.Fail()
		}
	}
}

func TestServer_Booking_Fails(t *testing.T) {
	s := setupServer()
	// unknown flight id
	token, err := s.Booking(-1, 0, utLowLevelID)
	if token != "" || err == nil {
		t.Error(err)
		t.Fail()
	}
	// unknown bonus level
	token, err = s.Booking(0, 0, utLowLevelID+"_unknown")
	if token != "" || err == nil {
		t.Error(err)
		t.Fail()
	}
}

func TestServer_GetBookingCode(t *testing.T) {
	s := setupServer()
	_, _, hashValue, signature, _ := crypt.GetBlindSignatureTestData("test123456", s.BonusList[utLowLevelID].ActionVariants[ActionBooking].SkKey)
	code, err := s.GetBookingCode(utLowLevelID, hashValue, signature)
	if err != nil {
		t.Error(err)
		t.Fail()
	}
	if code == "" {
		t.Error("no code created")
		t.Fail()
	}
	// the bonus code has to appear in the list of bonus codes
	if s.BonusCodes[code] == nil {
		t.Error("code does not appear in the list of bonus codes")
		t.Fail()
	}
}

func TestAccessBonusSystem(t *testing.T) {
	server := setupServer()
	codes := createValidTestCodes(t, server)
	accountID := uint32(0)
	addressID := uint32(11)

	// generate a seed and keys
	seed, keys, err := crypt.GetWalletKeys(utMnemonic, accountID, false)
	if err != nil {
		fail(t, err.Error())
	}

	// calculate the 12th address
	address := crypt.GetAddress(keys[4], addressID)
	addressBundle := &crypt.AddressBundle{Seed: seed, AccountID: accountID, AddressID: addressID, Address: address.String()}
	tokens, recoveries, err := server.AccessBonusSystem(codes, addressBundle)

	if err != nil {
		fail(t, err.Error())
	}

	if len(tokens) == 0 {
		fail(t, "no Token generated")
	}
	if len(recoveries) == 0 {
		fail(t, "no recoveries generated")
	}

	// check that every valid bonus level mapped the seed to the Token
	for bLevel, token := range tokens {
		if server.BonusList[bLevel].ActionVariants[ActionParticipate].TokenToSeed[token] != hex.EncodeToString(seed) {
			fail(t, "seed not mapped to Token (action = participate)")
		}
		if server.BonusList[bLevel].ActionVariants[ActionBooking].TokenToSeed[token] != "" {
			fail(t, "seed mapped to Token (action = booking)")
		}
		// check for correct mapping of address
		if server.BonusList[bLevel].ActionVariants[ActionParticipate].AddressToToken[address.String()] != token {
			fail(t, "address not mapped to Token (action = participate)")
		}
		if server.BonusList[bLevel].ActionVariants[ActionBooking].AddressToToken[address.String()] != "" {
			fail(t, "address mapped to Token (action = booking)")
		}
		if server.BonusList[bLevel].ActionVariants[ActionParticipate].SeedToAddress[hex.EncodeToString(seed)] != address.String() {
			fail(t, "seed not mapped to address (action = participate)")
		}
		if server.BonusList[bLevel].ActionVariants[ActionBooking].SeedToAddress[hex.EncodeToString(seed)] != "" {
			fail(t, "seed mapped to address (action = booking)")
		}
		if server.BonusList[bLevel].ActionVariants[ActionParticipate].SeedToAccountID[hex.EncodeToString(seed)] != accountID {
			fail(t, "seed not mapped to account id (action = participate)")
		}
		if len(server.BonusList[bLevel].ActionVariants[ActionBooking].SeedToAccountID) != 0 {
			fail(t, "seed mapped to account id (action = booking)")
		}
	}

	// check that the address is mapped to the recoveries Token
	for bLevel, recoveryToken := range recoveries {
		actionVariant := server.BonusList[bLevel].ActionVariants[ActionParticipate]
		if actionVariant.AddressToRecovery[address.String()] != recoveryToken {
			fail(t, "recoveries unknown")
		}
		actionVariant = server.BonusList[bLevel].ActionVariants[ActionBooking]
		if _, ok := actionVariant.AddressToRecovery[recoveryToken]; ok != false {
			fail(t, "recoveries found")
		}
	}

	// check that the address was marked as the one for accessing
	if server.BonusList[utLowLevelID].ActionVariants[ActionParticipate].SeedToAccessAdr[hex.EncodeToString(seed)] != address.String() {
		t.Error("address was not marked as the accessed one")
		t.Fail()
	}
	if server.BonusList[utMiddleLevelID].ActionVariants[ActionParticipate].SeedToAccessAdr[hex.EncodeToString(seed)] != address.String() {
		t.Fail()
	}
	if server.BonusList[utLowLevelID].ActionVariants[ActionParticipate].SeedToAccessAdr[hex.EncodeToString(seed)] != address.String() {
		t.Fail()
	}
	if server.BonusList[utHighLevelID].ActionVariants[ActionParticipate].SeedToAccessAdr[hex.EncodeToString(seed)] != address.String() {
		t.Fail()
	}

	// codes cannot be used again
	tokens, recoveries, err = server.AccessBonusSystem(codes, addressBundle)
	if err == nil || len(tokens) != 0 || len(recoveries) != 0 {
		t.Error(err)
		t.Fail()
	}
}

func TestServer_AccessBonusSystem_failures(t *testing.T) {
	server := setupServer()
	codes := createValidTestCodes(t, server)
	accountID := uint32(0)
	addressID := uint32(11)

	// generate a seed and keys
	seed, keys, err := crypt.GetWalletKeys(utMnemonic, accountID, false)
	if err != nil {
		fail(t, err.Error())
	}

	// calculate the 12th address
	address := crypt.GetAddress(keys[4], addressID)

	// try to set an invalid seed to the bundle
	addressBundle := &crypt.AddressBundle{Seed: []byte{1, 2, 3}, AccountID: accountID, AddressID: addressID, Address: address.String()}
	_, _, err = server.AccessBonusSystem(codes, addressBundle)
	if err == nil {
		t.Error("access was granted, but an invalid seed was given")
		t.Fail()
	}

	// try to set another account id to the bundle
	addressBundle = &crypt.AddressBundle{Seed: seed, AccountID: 10, AddressID: addressID, Address: address.String()}
	_, _, err = server.AccessBonusSystem(codes, addressBundle)
	if err == nil {
		t.Error("access was granted, but wrong ACCOUNT id was given")
		t.Fail()
	}
	// try to set another address id to the bundle
	addressBundle = &crypt.AddressBundle{Seed: seed, AccountID: accountID, AddressID: 999, Address: address.String()}
	_, _, err = server.AccessBonusSystem(codes, addressBundle)
	if err == nil {
		t.Error("access was granted, but wrong ADDRESS id was given")
		t.Fail()
	}
}

func TestServer_verifyCodes(t *testing.T) {
	server := setupServer()
	codes := createValidTestCodes(t, server)

	accessible := server.verifyCodes(codes)
	sort.Slice(accessible, func(i, j int) bool { return accessible[i].BonusID < accessible[j].BonusID })
	if len(accessible) != 3 || accessible[0] != server.BonusList["high"] ||
		accessible[1] != server.BonusList["low"] ||
		accessible[2] != server.BonusList["middle"] {
		t.Error("wrong bonus level accessible")
		t.Fail()
	}

	// choose other codes which are not valid for any bonus levels
	codes = []string{"code1", "code2", "code3", "code4", "code5", "code6"}
	accessible = server.verifyCodes(codes)
	if len(accessible) != 0 {
		t.Error("a bonus level is accessible")
		t.Fail()
	}

	// change the date 'createdAt' of all codes, s.t. they are invalid w.r.t.
	// the highest bonus level
	codes = createValidTestCodes(t, server)
	for _, code := range server.BonusCodes {
		if code != nil {
			code.CreatedAt = code.CreatedAt.AddDate(0, 0, -11)
		}
	}
	accessible = server.verifyCodes(codes)
	// no bonus level is accessible since all generated test codes
	// were only valid for the highest bonus level.
	// Due to the modification of the expiration date, all code are invalid
	// for the highest level and therefore for the lower level also.
	if len(accessible) != 0 {
		t.Error("a bonus level is accessible")
		t.Fail()
	}
}

func TestGetBlindSignatureSmallExample(t *testing.T) {
	server := setupServer()
	token := "testToken"
	blindToken := []byte("blindToken")

	// insert initial Token for seed
	server.BonusList[utLowLevelID].ActionVariants[ActionBooking].ValidTokens = map[string]bool{token: false}
	// try to use it for wrong action variant
	blindSignature, err := server.GetBlindSignature(utLowLevelID, token, blindToken, ActionParticipate)
	if err == nil {
		fail(t, "no error raised")
	}
	if blindSignature != "" {
		fail(t, "blind signature created")
	}
	// try to use it for correct action variant
	blindSignature, err = server.GetBlindSignature(utLowLevelID, token, blindToken, ActionBooking)
	if err != nil {
		fail(t, err.Error())
	}
	if blindSignature == "" {
		fail(t, "no blind signature created")
	}

	// check that the first Token is not valid anymore
	if len(server.BonusList[utLowLevelID].ActionVariants[ActionBooking].ValidTokens) != 1 {
		fail(t, "Token was not deleted")
	}
	if server.BonusList[utLowLevelID].ActionVariants[ActionBooking].ValidTokens[token] != true {
		fail(t, "Token not marked as used")
	}
}

func TestServer_GetBlindSignature(t *testing.T) {
	var action = ActionParticipate
	server := setupServer()
	token := generateToken()
	// hash it
	hashed := fdh.Sum(crypto.SHA256, 768, []byte(token))
	// blind and unBlind
	blindedToken, unBlinder, err := rsablind.Blind(&server.BonusList["low"].ActionVariants[action].SkKey.PublicKey, hashed)
	if err != nil {
		fail(t, err.Error())
	}
	bLevel := server.BonusList[utLowLevelID]
	bLevel.ActionVariants[action].ValidTokens = map[string]bool{token: false}

	// call with unknown bonus level
	blindSignatureHex, err := server.GetBlindSignature(utLowLevelID+"_unknown", token, blindedToken, action)
	if err == nil {
		t.Error("blind sign for unknown bonus level")
		t.Fail()
	}

	// call with invalid Token
	blindSignatureHex, err = server.GetBlindSignature(utLowLevelID, token+"_invalid", blindedToken, action)
	if err == nil {
		t.Error("blind sign with invalid Token ")
		t.Fail()
	}

	// success expected
	blindSignatureHex, err = server.GetBlindSignature(utLowLevelID, token, blindedToken, action)
	if err != nil {
		fail(t, err.Error())
	}
	if blindSignatureHex == "" {
		fail(t, "could not generate blind signature")
	}

	//////////// CLIENT SITE //////////////////////////
	blindSig, _ := base64.URLEncoding.DecodeString(blindSignatureHex)
	signature := rsablind.Unblind(&bLevel.ActionVariants[action].SkKey.PublicKey, blindSig, unBlinder)
	///////////////////////////////////////////////////
	// try to verify
	err = rsablind.VerifyBlindSignature(&bLevel.ActionVariants[action].SkKey.PublicKey, hashed, signature)
	if err != nil {
		fail(t, err.Error())
	}
}

func TestServer_SetAddress(t *testing.T) {
	var action = ActionParticipate
	var recovery string

	server := setupServer()
	token := generateToken()
	// generate a valid hash and signature
	_, _, hashed, signature, err := crypt.GetBlindSignatureTestData(token, server.BonusList[utLowLevelID].ActionVariants[action].SkKey)
	if err != nil {
		t.Error(err)
		t.Fail()
	}

	accountID := uint32(0)
	addressID := uint32(11)

	// generate a seed and keys
	seed, keys, err := crypt.GetWalletKeys(utMnemonic, accountID, false)
	if err != nil {
		fail(t, err.Error())
	}

	// calculate the 12th address
	address := crypt.GetAddress(keys[4], addressID)
	if address == nil {
		fail(t, "couldn't calculate address")
		t.FailNow()
	}
	adrBundle := &crypt.AddressBundle{Seed: seed, AccountID: accountID, AddressID: addressID, Address: address.String()}

	// the seed has to be known => set it manually for this test
	server.BonusList[utLowLevelID].ActionVariants[action].SeedToAddress[hex.EncodeToString(seed)] = "oldAddress"
	server.BonusList[utLowLevelID].ActionVariants[action].SeedToAccountID[hex.EncodeToString(seed)] = accountID
	// create a blinded recovery Token
	pkr := "pkr"
	token, recovery, err = server.SetAddress(utLowLevelID, hashed, signature, adrBundle, action, pkr)
	if err != nil {
		fail(t, err.Error())
	}
	if token == "" || recovery == "" {
		fail(t, "no Token was generated")
	}

	// check that the new Token was marked as valid
	bLevel := server.BonusList[utLowLevelID]
	if len(bLevel.ActionVariants[action].ValidTokens) != 1 || bLevel.ActionVariants[action].ValidTokens[token] != false {
		fail(t, "new Token was not marked as valid")
	}

	// check that the blinded recovery Token was mapped to the given address
	if bLevel.ActionVariants[action].PkrToAdrUpd[pkr] != adrBundle.Address {
		t.Error("pkr was not mapped to the correct address")
		t.Fail()
	}
	// check that a new recovery Token was generated
	if len(bLevel.ActionVariants[action].AddressToRecovery) != 1 ||
		bLevel.ActionVariants[action].AddressToRecovery[address.String()] != recovery {
		fail(t, "no new recovery Token was generated")
	}

	// the account id has to be correct
	if bLevel.ActionVariants[action].SeedToAccountID[hex.EncodeToString(seed)] != accountID {
		fail(t, "not correct mapping seed -> account id 2")
	}
}

func TestServer_SetAddress_failure(t *testing.T) {
	var action = ActionParticipate
	var unknownMnemonic = "noble fire perfect garlic nasty maid invite relief august orient doll profit search huge impose rare fade suffer legend audit announce can lottery drum"

	server := setupServer()
	token := generateToken()
	// generate a valid hash and signature
	_, _, hashed, signature, _ := crypt.GetBlindSignatureTestData(token, server.BonusList[utLowLevelID].ActionVariants[action].SkKey)
	accountID := uint32(0)
	addressID := uint32(11)
	// generate a seed and keys
	seed, keys, _ := crypt.GetWalletKeys(utMnemonic, accountID, false)
	// calculate the 12th address
	address := crypt.GetAddress(keys[4], addressID)
	adrBundle := &crypt.AddressBundle{Seed: seed, AccountID: accountID, AddressID: addressID, Address: address.String()}
	// the seed has to be known => set it manually for this test
	server.BonusList[utLowLevelID].ActionVariants[action].SeedToAddress[hex.EncodeToString(seed)] = "oldAddress"
	// create a blinded recovery Token
	pkr := "pkr"

	// call with wrong bonus level id
	_, _, err := server.SetAddress(utLowLevelID+"_unknown", hashed, signature, adrBundle, action, pkr)
	if err == nil {
		t.Error("set address possible with unknown bonus level id")
		t.Fail()
	}

	// call with unknown seed
	unknownSeed, keys, _ := crypt.GetWalletKeys(unknownMnemonic, accountID, false)
	// calculate the 12th address
	unknownAddress := crypt.GetAddress(keys[4], addressID)
	adrBundle = &crypt.AddressBundle{Seed: unknownSeed, AccountID: accountID, AddressID: addressID, Address: unknownAddress.String()}
	_, _, err = server.SetAddress(utLowLevelID, hashed, signature, adrBundle, action, pkr)
	if err == nil {
		t.Error("set address possible with unknown seed")
		t.Fail()
	}

	// call with empty signature
	adrBundle = &crypt.AddressBundle{Seed: seed, AccountID: accountID, AddressID: addressID, Address: address.String()}
	_, _, err = server.SetAddress(utLowLevelID, hashed, []byte{}, adrBundle, action, pkr)
	if err == nil {
		t.Error("set address possible with empty signature")
		t.Fail()
	}

	// hash value and signature do not fit
	_, _, err = server.SetAddress(utLowLevelID, []byte{1, 2, 3, 4, 5, 6}, signature, adrBundle, action, pkr)
	if err == nil {
		t.Error("set address possible with non-fitting hash and signature")
		t.Fail()
	}

	// seed and address do not fit
	adrBundle = &crypt.AddressBundle{Seed: seed, AccountID: accountID, AddressID: addressID, Address: unknownAddress.String()}
	_, _, err = server.SetAddress(utLowLevelID, hashed, signature, adrBundle, action, pkr)
	if err == nil {
		t.Error("set address possible with non-fitting seed and address")
		t.Fail()
	}

	// check that failure were not caused due to wrong setup data
	adrBundle = &crypt.AddressBundle{Seed: seed, AccountID: accountID, AddressID: addressID, Address: address.String()}
	_, _, err = server.SetAddress(utLowLevelID, hashed, signature, adrBundle, action, pkr)
	if err != nil {
		t.Error("set address not possible with well formed data")
		t.Fail()
	}
}

func TestServer_Participate(t *testing.T) {
	var action = ActionParticipate
	server := setupServer()
	serverActionVariant := server.BonusList[utLowLevelID].ActionVariants[action]
	pkr := "pkr"

	// generate tokens, a blind hash value and its corresponding signature
	token := generateToken()
	_, _, hashValue, sig, err := crypt.GetBlindSignatureTestData(token, serverActionVariant.SkKey)
	if err != nil {
		fail(t, err.Error())
	}

	newToken, recoveryToken, bonusData, err := server.Participate(utLowLevelID, hashValue, sig, pkr)
	if err != nil {
		fail(t, err.Error())
	}
	if newToken == "" || recoveryToken == "" {
		fail(t, "no Token was generated")
	}
	if bonusData == "" {
		t.Error("no bonus data received")
		t.Fail()
	}

	// the new Token has to be marked as valid
	bLevel := server.BonusList[utLowLevelID]
	if bLevel.ActionVariants[action].ValidTokens[newToken] != false {
		fail(t, "new Token was not added")
	}

	// the blinded Token has to be mapped to the hashed value and signature
	if len(serverActionVariant.PkrToBonusData) != 1 {
		fail(t, "wrong number of recovery tokens in server map")
	}
	for _, bData := range serverActionVariant.PkrToBonusData {
		if bData.Token != newToken || bData.RecoveryToken != recoveryToken {
			t.Errorf("wrong tokens! Token act:%s exp:%s\nrecoveryToken act:%s exp:%s", newToken, bData.Token, recoveryToken, bData.RecoveryToken)
			t.Fail()
		}
	}
}

func TestServer_Participate_Long(t *testing.T) {
	var action = ActionParticipate
	var recoveryToken string
	var bonusData string

	// large example of participating
	// multiple iterations for participating
	server := setupServer()
	bLevel := server.BonusList[utLowLevelID]
	initialToken := generateToken()
	token := generateToken()
	// create a new hd wallet and save the seed and Token
	seed, keys, err := crypt.GetWalletKeys(utMnemonic, 0, false)
	if err != nil {
		t.Error(err)
		t.Fail()
	}
	bLevel.ActionVariants[action].TokenToSeed[initialToken] = hex.EncodeToString(seed)
	bLevel.ActionVariants[action].SeedToAddress[hex.EncodeToString(seed)] = "initAddress"
	bLevel.ActionVariants[action].SeedToAccountID[hex.EncodeToString(seed)] = 2
	bLevel.ActionVariants[action].ValidTokens = map[string]bool{initialToken: false}

	////////// step 1: Get blind Token and signature for address update ////////////
	hashValue := fdh.Sum(crypto.SHA256, 768, []byte(token))
	blindToken, unBlind, err := rsablind.Blind(&bLevel.ActionVariants[action].SkKey.PublicKey, hashValue)
	blindSig64, err := server.GetBlindSignature(utLowLevelID, initialToken, blindToken, action)
	if err != nil {
		t.Error(err)
		t.Fail()
	}
	// the initial Token has to be invalid now
	if valid := bLevel.isTokenValid(initialToken, action); valid != false {
		t.Error("step 1: Initial Token is not invalid")
		t.Fail()
	}
	blindSig, _ := base64.URLEncoding.DecodeString(blindSig64)
	signature := rsablind.Unblind(&bLevel.ActionVariants[action].SkKey.PublicKey, blindSig, unBlind)

	////////// step 2: Update address ////////////
	// calculate a new address
	address := crypt.GetAddress(keys[4], 0)
	adrBundle := &crypt.AddressBundle{Seed: seed, AccountID: 0, AddressID: 0, Address: address.String()}
	// create a blinded recovery Token
	pkr := recoveryToken
	initialToken, recoveryToken, err = server.SetAddress(utLowLevelID, hashValue, signature, adrBundle, action, pkr)
	if err != nil {
		t.Error(err.Error())
		t.Fail()
	}
	if initialToken == "" || recoveryToken == "" {
		t.Error("empty tokens after address update")
		t.Fail()
	}
	// the initialToken has to be valid
	if valid := bLevel.isTokenValid(initialToken, action); valid != true {
		t.Error("step 2: Initial Token is invalid")
		t.Fail()
	}

	// the blinded Token has to be mapped to the address
	serverActionVariant := server.BonusList[utLowLevelID].ActionVariants[action]
	if len(serverActionVariant.PkrToAdrUpd) != 1 {
		t.Error("wrong length of pkr map")
		t.Fail()
	}
	if serverActionVariant.PkrToAdrUpd[pkr] != adrBundle.Address {
		t.Error("pkr was not mapped to the correct address")
		t.Fail()
	}

	////////// step 3: Get blind Token and signature for 'real' participation ////////////
	token = generateToken()
	hashValue = fdh.Sum(crypto.SHA256, 768, []byte(token))
	blindToken, unBlind, err = rsablind.Blind(&bLevel.ActionVariants[action].SkKey.PublicKey, hashValue)
	blindSig64, err = server.GetBlindSignature(utLowLevelID, initialToken, blindToken, action)
	if err != nil {
		t.Error(err)
		t.Fail()
	}
	blindSig, _ = base64.URLEncoding.DecodeString(blindSig64)
	signature = rsablind.Unblind(&bLevel.ActionVariants[action].SkKey.PublicKey, blindSig, unBlind)

	// calculate blind participate Token
	pkr = recoveryToken

	initialToken, recoveryToken, bonusData, err = server.Participate(utLowLevelID, hashValue, signature, pkr)
	if err != nil {
		t.Error(err)
		t.Fail()
	}
	if initialToken == "" || recoveryToken == "" {
		t.Error("step 3: Initial Token is empty")
		t.Fail()
	}
	if valid := bLevel.isTokenValid(initialToken, action); valid != true {
		t.Error("step 3: Initial Token is invalid")
		t.Fail()
	}
	if bonusData == "" {
		t.Error("no bonus data received")
		t.Fail()
	}

	if len(serverActionVariant.PkrToAdrUpd) != 1 || len(serverActionVariant.PkrToBonusData) != 1 {
		t.Error("wrong length of pkr maps")
		t.Fail()
	}
	bData := serverActionVariant.PkrToBonusData[pkr]
	if bData.Token != initialToken || bData.RecoveryToken != recoveryToken {
		t.Error("wrong bonus data received")
		t.Fail()
	}
}

func TestServer_Participate_failure(t *testing.T) {
	var action = ActionParticipate
	var pkr = "pkr"
	server := setupServer()

	// generate tokens, a blind hash value and its corresponding signature
	token := generateToken()
	_, _, hashValue, sig, _ := crypt.GetBlindSignatureTestData(token, server.BonusList[utLowLevelID].ActionVariants[action].SkKey)

	// try to call with unknown bonus level id
	if _, _, _, err := server.Participate(utLowLevelID+"_unknown", hashValue, sig, pkr); err == nil {
		t.Error("unknown bonus level not detected")
		t.Fail()
	}

	// call with empty signature
	if _, _, _, err := server.Participate(utLowLevelID, hashValue, []byte{}, pkr); err == nil {
		t.Error("empty signature not detected")
		t.Fail()
	}

	// hash and signature do not fit together
	// call with empty signature
	if _, _, _, err := server.Participate(utLowLevelID, hashValue, []byte{1, 2, 3, 4, 5}, pkr); err == nil {
		t.Error("participation, but hash and signature do not fit")
		t.Fail()
	}

}

func TestServer_GenerateNewBonusCode(t *testing.T) {
	server := setupServer()
	code := server.GenerateNewBonusCode(utLowLevelID)
	if code == nil {
		t.Error("could not generate code")
		t.Fail()
	} else if code.ValidFor.BonusID != utLowLevelID {
		t.Error("wrong bonus level connected to code")
		t.Fail()
	}
	// generate a new date AFTER the bonusLevel expiration date
	//expireAt := time.Now().AddDate(0,0, server.BonusList[utLowLevelID].ValidDuration)
}

// helper function for generating a new random Token
func generateToken() string {
	// generate tokens, a blind hash value and its corresponding signature
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

func TestServer_GetSystemInformation(t *testing.T) {
	s := setupServer()
	flightList, bonusList, err := s.GetSystemInformation()
	if err != nil {
		t.Error(err)
		t.Fail()
	}
	if len(flightList) != len(s.flightMap) {
		t.Error("wrong number of flights received")
		t.Fail()
	}
	if len(bonusList) != len(s.BonusList) {
		t.Error("wrong number of bonus levels received")
		t.Fail()
	}

	// check the flight map
	for fID := range s.flightMap {
		found := false
		for _, flight := range flightList {
			if flight.ID == fID {
				found = true
			}
		}
		if !found {
			t.Error("flight with id " + strconv.Itoa(fID) + " is not in flight list")
			t.Fail()
		}
	}

	// check the bonus map
	for bID := range s.BonusList {
		found := false
		for _, bLevel := range bonusList {
			if bID == bLevel.BonusID {
				found = true
			}
		}
		if !found {
			t.Error("bonus with id " + bID + " is not in bonus list")
			t.Fail()
		}
	}
}

func TestServer_CanBeUsedForRecovery(t *testing.T) {
	// Loosing Token directly after access

	client := setupClient(t)
	server := client.con.(*utConnection).server
	accountID := uint32(0)
	addressID := uint32(11)

	// simulate an access to the bonus system
	adrBdl, keys, tokens, recoveries := testExecuteAccessBonusSystem(server, accountID, addressID, t)

	// the airline has to return the values of the address bundle if asked for
	if lastAdr, lastAccount, err := server.GetLastAdrBundle(client.Seed, utLowLevelID); err != nil || lastAdr != adrBdl.Address || lastAccount != adrBdl.AccountID {
		t.Error(err)
		t.Error("wrong address or account received")
		t.Fail()
	}

	// now, the client looses his tokens
	status, token, err := server.CanBeUsedForRecovery(utMiddleLevelID, adrBdl)
	if err != nil {
		t.Error(err)
		t.Fail()
	}

	if status != RecoveryTestAfterAccess {
		t.Errorf("Wrong status '%s': Recovery directly after access", status)
		t.Fail()
	}
	if token != tokens[utMiddleLevelID] {
		t.Error("Wrong Token: Recovery directly after access")
		t.Fail()
	}

	foundToken, foundRecoveryToken, bData, err := server.RecoveryTest(utMiddleLevelID, "", "", adrBdl)
	if err != nil {
		t.Error(err)
		t.Fail()
	}
	if foundToken != "" || bData != "" {
		t.Errorf("not initial: Token ('%s') or bonus data ('%s')", foundToken, bData)
	}
	recoveryToken := recoveries[utMiddleLevelID]
	if foundRecoveryToken != recoveryToken {
		t.Errorf("wrong recovery Token found:'%s'", foundRecoveryToken)
		t.Fail()
	}

	// An address update has to be possible now
	nextAddressID := uint32(12)
	pkr, _ := client.blindRecoveryToken(recoveryToken)
	testExecuteSetAddress(client, server, utMiddleLevelID, token, adrBdl, keys, nextAddressID, pkr, t)
}

func TestServer_CanBeUsedForRecovery2(t *testing.T) {
	// Loosing Token  after first address update

	client := setupClient(t)
	server := client.con.(*utConnection).server
	accountID := uint32(0)
	addressID := uint32(11)

	// simulate an access to the bonus system
	adrBdl, keys, tokens, recoveries := testExecuteAccessBonusSystem(server, accountID, addressID, t)

	// An address update has to be possible now
	nextAddressID := uint32(12)
	token := tokens[utMiddleLevelID]
	recoveryToken := recoveries[utMiddleLevelID]
	pkr, _ := client.blindRecoveryToken(recoveryToken)
	nextAdrBdl, tokenAfterUpd, recTokenAfterUpd := testExecuteSetAddress(client, server, utMiddleLevelID, token, adrBdl, keys, nextAddressID, pkr, t)

	// now, the client looses his tokens
	// The initial address bundle has to fail
	status, token, err := server.CanBeUsedForRecovery(utMiddleLevelID, adrBdl)
	if status != Failure || token != "" || err == nil {
		t.Errorf("Wrong status '%s' or Token '%s': Recovery after first update", status, token)
		t.Fail()
	}
	status, token, err = server.CanBeUsedForRecovery(utMiddleLevelID, nextAdrBdl)
	if err != nil {
		t.Error(err)
		t.Fail()
	}
	// The last updated address bundle must not fail
	if status != RecoveryTestAfterFirstAdrUpd || token != recoveryToken {
		t.Errorf("Wrong status '%s' or Token '%s': Recovery after first update", status, token)
		t.Fail()
	}

	token, foundRecoveryToken, _, err := server.RecoveryTest(utMiddleLevelID, recoveryToken, pkr, nextAdrBdl)
	if err != nil {
		t.Error(err)
		t.Fail()
	}
	if token != tokenAfterUpd {
		t.Errorf("Wrong Token recovered: '%s'", token)
		t.Fail()
	}
	if foundRecoveryToken != recTokenAfterUpd {
		t.Errorf("wrong recovery Token received: %s\texp:%s", foundRecoveryToken, recTokenAfterUpd)
		t.Fail()
	}
}

func TestServer_CanBeUsedForRecovery3(t *testing.T) {
	// Loosing Token after participation

	client := setupClient(t)
	server := client.con.(*utConnection).server
	accountID := uint32(0)
	addressID := uint32(11)

	// simulate an access to the bonus system
	accessAdrBdl, keys, accessTokens, accessRecoveries := testExecuteAccessBonusSystem(server, accountID, addressID, t)

	// An address update
	addressIDBeforeAdrUpd := uint32(12)
	tokenBeforeAdrUpd := accessTokens[utMiddleLevelID]
	recTokenBeforeAdrUpd := accessRecoveries[utMiddleLevelID]
	pkrBeforeAdrUpd, _ := client.blindRecoveryToken(recTokenBeforeAdrUpd)
	adrBdlAfterAdrUpd, tokenAfterAdrUpd, recTokenAfterAdrUpd := testExecuteSetAddress(client, server, utMiddleLevelID, tokenBeforeAdrUpd, accessAdrBdl, keys, addressIDBeforeAdrUpd, pkrBeforeAdrUpd, t)
	pkrAfterAdrUpd, _ := client.blindRecoveryToken(recTokenAfterAdrUpd)

	// Execute a participation
	tokenAfterPart, recTokenAfterPart, bonusData := testExecuteParticipation(client, server, utMiddleLevelID, tokenAfterAdrUpd, pkrAfterAdrUpd, t)
	if tokenAfterPart == "" || recTokenAfterPart == "" || bonusData == "" {
		t.Error("initial tokens")
		t.Fail()
	}

	// now, the client looses his tokens
	// The initial address bundle has to fail
	status, token, err := server.CanBeUsedForRecovery(utMiddleLevelID, accessAdrBdl)
	if status != Failure || token != "" || err == nil {
		t.Errorf("Wrong status '%s' or Token '%s': access bundle did not fail", status, token)
		t.Fail()
	}

	// The address of the first update must not fail
	status, token, err = server.CanBeUsedForRecovery(utMiddleLevelID, adrBdlAfterAdrUpd)
	if err != nil {
		t.Error(err)
		t.Fail()
	}
	if status != RecoveryTest || token != recTokenAfterAdrUpd {
		t.Errorf("Wrong status '%s' or Token '%s': bundle of first update failed", status, token)
		t.Fail()
	}

	// try to recover
	token, foundRecoveryToken, bData, err := server.RecoveryTest(utMiddleLevelID, recTokenAfterAdrUpd, pkrAfterAdrUpd, adrBdlAfterAdrUpd)
	if err != nil {
		t.Error(err)
		t.Fail()
	}
	if token != tokenAfterPart {
		t.Errorf("wrong Token received: %s\texp:%s", token, tokenAfterPart)
		t.Fail()
	}
	if foundRecoveryToken != recTokenAfterPart {
		t.Errorf("wrong recovery Token received: %s\texp:%s", foundRecoveryToken, recTokenAfterPart)
		t.Fail()
	}
	if bData != bonusData {
		t.Error("no bonus data received")
		t.Fail()
	}
}

func TestServer_SetAddressFail(t *testing.T) {
	// address update with same address is not allowed
	client := setupClient(t)
	server := client.con.(*utConnection).server
	accountID := uint32(0)
	addressID := uint32(11)

	// simulate an access to the bonus system
	accessAdrBdl, keys, accessTokens, accessRecoveries := testExecuteAccessBonusSystem(server, accountID, addressID, t)

	// An address update
	adrIDSetAdr := uint32(12)
	tokenBeforeAdrUpd := accessTokens[utMiddleLevelID]
	recTokenBeforeAdrUpd := accessRecoveries[utMiddleLevelID]
	pkrBeforeAdrUpd, _ := client.blindRecoveryToken(recTokenBeforeAdrUpd)
	adrBdlAfterAdrUpd, tokenAfterAdrUpd, recTokenAfterAdrUpd := testExecuteSetAddress(client, server, utMiddleLevelID, tokenBeforeAdrUpd, accessAdrBdl, keys, adrIDSetAdr, pkrBeforeAdrUpd, t)
	pkrAfterAdrUpd, _ := client.blindRecoveryToken(recTokenAfterAdrUpd)

	// Execute a participation
	tokenAfterPart, recTokenAfterPart, _ := testExecuteParticipation(client, server, utMiddleLevelID, tokenAfterAdrUpd, pkrAfterAdrUpd, t)

	// second address update
	pkrBefore2ndAdrUpd, _ := client.blindRecoveryToken(recTokenAfterPart)
	_, tokenAfter2ndAdrUpd, recTokenAfter2ndAdrUpd, err := testExecuteSetAddressNoErrorHdl(client, server, utMiddleLevelID, tokenAfterPart, adrBdlAfterAdrUpd, keys, adrIDSetAdr, pkrBefore2ndAdrUpd)
	if tokenAfter2ndAdrUpd != "" || recTokenAfter2ndAdrUpd != "" || err == nil {
		t.Errorf("received tokens: normal(%s)\trecovery(%s)", tokenAfter2ndAdrUpd, recTokenAfter2ndAdrUpd)
		t.Fail()
	}
	if err != nil && err.Error() != "address is not valid. Already used" {
		t.Error(err)
		t.Fail()
	}

}

func TestServer_CanBeUsedForRecovery4(t *testing.T) {
	// Loosing Token after second address update

	client := setupClient(t)
	server := client.con.(*utConnection).server
	accountID := uint32(0)
	addressID := uint32(11)

	// simulate an access to the bonus system
	accessAdrBdl, keys, accessTokens, accessRecoveries := testExecuteAccessBonusSystem(server, accountID, addressID, t)

	// An address update
	addressIDBeforeAdrUpd := uint32(12)
	tokenBeforeAdrUpd := accessTokens[utMiddleLevelID]
	recTokenBeforeAdrUpd := accessRecoveries[utMiddleLevelID]
	pkrBeforeAdrUpd, _ := client.blindRecoveryToken(recTokenBeforeAdrUpd)
	adrBdlAfterAdrUpd, tokenAfterAdrUpd, recTokenAfterAdrUpd := testExecuteSetAddress(client, server, utMiddleLevelID, tokenBeforeAdrUpd, accessAdrBdl, keys, addressIDBeforeAdrUpd, pkrBeforeAdrUpd, t)
	pkrAfterAdrUpd, _ := client.blindRecoveryToken(recTokenAfterAdrUpd)

	// Execute a participation
	tokenAfterPart, recTokenAfterPart, bonusDataAftPart := testExecuteParticipation(client, server, utMiddleLevelID, tokenAfterAdrUpd, pkrAfterAdrUpd, t)

	// second address update
	adrIDFrom2ndAdrUpd := uint32(25)
	pkrBefore2ndAdrUpd, _ := client.blindRecoveryToken(recTokenAfterPart)
	adrBdlOf2ndUpd, tokenAfter2ndAdrUpd, recTokenAfter2ndAdrUpd := testExecuteSetAddress(client, server, utMiddleLevelID, tokenAfterPart, adrBdlAfterAdrUpd, keys, adrIDFrom2ndAdrUpd, pkrBefore2ndAdrUpd, t)

	// the server has to send the address and account id of the last address update back
	if rcvdAdr, rcvdAccount, err := server.GetLastAdrBundle(adrBdlOf2ndUpd.Seed, utMiddleLevelID); rcvdAccount != adrBdlOf2ndUpd.AccountID || rcvdAdr != adrBdlOf2ndUpd.Address || err != nil {
		t.Error("wrong address or account id received from server")
		t.Fail()
	}

	// now, the client looses his tokens
	// The initial address bundle has to fail
	status, token, err := server.CanBeUsedForRecovery(utMiddleLevelID, accessAdrBdl)
	if status != Failure || token != "" || err == nil {
		t.Errorf("Wrong status '%s' or Token '%s': access bundle did not fail", status, token)
		t.Fail()
	}

	// The address of the first update has to fail
	status, token, err = server.CanBeUsedForRecovery(utMiddleLevelID, adrBdlAfterAdrUpd)
	if status != Failure || token != "" || err == nil {
		t.Errorf("Wrong status '%s' or Token '%s': bundle of first update failed", status, token)
		t.Fail()
	}

	// The address of the second update must not fail.
	// Special case: The recovery Token of the 1st address update has to be found!
	status, token, err = server.CanBeUsedForRecovery(utMiddleLevelID, adrBdlOf2ndUpd)
	if err != nil {
		t.Error(err)
		t.Fail()
	}
	if status != RecoveryTestPenultimateAdr || token != recTokenAfterAdrUpd {
		t.Errorf("Wrong status '%s' or Token '%s': bundle of first update failed", status, token)
		t.Fail()
	}
	token, foundRecoveryToken, bData, err := server.RecoveryTest(utMiddleLevelID, recTokenAfterAdrUpd, pkrAfterAdrUpd, adrBdlOf2ndUpd)
	if err != nil {
		t.Error(err)
		t.Fail()
	}
	if token != tokenAfter2ndAdrUpd {
		t.Errorf("wrong Token received: %s\texp:%s", token, tokenAfter2ndAdrUpd)
		t.Fail()
	}
	if foundRecoveryToken != recTokenAfter2ndAdrUpd {
		t.Errorf("wrong recovery Token received: %s\texp:%s", foundRecoveryToken, recTokenAfter2ndAdrUpd)
		t.Fail()
	}
	if bData != bonusDataAftPart {
		t.Error("no bonus data received")
		t.Fail()
	}
}

func TestServer_CanBeUsedForRecovery5(t *testing.T) {
	// Loosing Token after second participation step

	client := setupClient(t)
	server := client.con.(*utConnection).server
	accountID := uint32(0)
	addressID := uint32(11)

	// simulate an access to the bonus system
	accessAdrBdl, keys, accessTokens, accessRecoveries := testExecuteAccessBonusSystem(server, accountID, addressID, t)

	// An address update
	addressIDBeforeAdrUpd := uint32(12)
	tokenBeforeAdrUpd := accessTokens[utMiddleLevelID]
	recTokenBeforeAdrUpd := accessRecoveries[utMiddleLevelID]
	pkrBeforeAdrUpd, _ := client.blindRecoveryToken(recTokenBeforeAdrUpd)
	adrBdlAfterAdrUpd, tokenAfterAdrUpd, recTokenAfterAdrUpd := testExecuteSetAddress(client, server, utMiddleLevelID, tokenBeforeAdrUpd, accessAdrBdl, keys, addressIDBeforeAdrUpd, pkrBeforeAdrUpd, t)
	pkrAfterAdrUpd, _ := client.blindRecoveryToken(recTokenAfterAdrUpd)

	// Execute a participation
	tokenAfterPart, recTokenAfterPart, _ := testExecuteParticipation(client, server, utMiddleLevelID, tokenAfterAdrUpd, pkrAfterAdrUpd, t)

	// second address update
	adrIDFrom2ndAdrUpd := uint32(25)
	pkrBefore2ndAdrUpd, _ := client.blindRecoveryToken(recTokenAfterPart)
	adrBdlOf2ndUpd, tokenAfter2ndAdrUpd, recTokenAfter2ndAdrUpd := testExecuteSetAddress(client, server, utMiddleLevelID, tokenAfterPart, adrBdlAfterAdrUpd, keys, adrIDFrom2ndAdrUpd, pkrBefore2ndAdrUpd, t)
	pkrAfter2ndAdrUpd, _ := client.blindRecoveryToken(recTokenAfter2ndAdrUpd)

	// Execute a participation
	tokenAfter2ndPart, recTokenAfter2ndPart, bDataAfter2ndPart := testExecuteParticipation(client, server, utMiddleLevelID, tokenAfter2ndAdrUpd, pkrAfter2ndAdrUpd, t)
	if tokenAfter2ndPart == "" || recTokenAfter2ndPart == "" {
		t.Error("initial tokens")
		t.Fail()
	}

	// now, the client looses his tokens
	// The initial address bundle has to fail
	status, token, err := server.CanBeUsedForRecovery(utMiddleLevelID, accessAdrBdl)
	if status != Failure || token != "" || err == nil {
		t.Errorf("Wrong status '%s' or Token '%s': access bundle did not fail", status, token)
		t.Fail()
	}

	// The address of the first update has to fail
	status, token, err = server.CanBeUsedForRecovery(utMiddleLevelID, adrBdlAfterAdrUpd)
	if status != Failure || token != "" || err == nil {
		t.Errorf("Wrong status '%s' or Token '%s': bundle of first update failed", status, token)
		t.Fail()
	}

	// The address of the second update must not fail.
	status, token, err = server.CanBeUsedForRecovery(utMiddleLevelID, adrBdlOf2ndUpd)
	if err != nil {
		t.Error(err)
		t.Fail()
	}
	if status != RecoveryTest || token != recTokenAfter2ndAdrUpd {
		t.Errorf("Wrong status '%s' or Token '%s': bundle of first update failed", status, token)
		t.Fail()
	}

	token, foundRecoveryToken, bData, err := server.RecoveryTest(utMiddleLevelID, recTokenAfterAdrUpd, pkrAfter2ndAdrUpd, adrBdlOf2ndUpd)
	if err != nil {
		t.Error(err)
		t.Fail()
	}
	if token != tokenAfter2ndPart {
		t.Errorf("wrong Token received: %s\texp:%s", token, tokenAfter2ndAdrUpd)
		t.Fail()
	}
	if foundRecoveryToken != recTokenAfter2ndPart {
		t.Errorf("wrong recovery Token received: %s\texp:%s", foundRecoveryToken, recTokenAfter2ndAdrUpd)
		t.Fail()
	}
	if bData != bDataAfter2ndPart {
		t.Error("no bonus data received")
		t.Fail()
	}

	// the server has to send the address and account id of the last address update back
	if rcvdAdr, rcvdAccount, err := server.GetLastAdrBundle(adrBdlOf2ndUpd.Seed, utMiddleLevelID); rcvdAccount != adrBdlOf2ndUpd.AccountID || rcvdAdr != adrBdlOf2ndUpd.Address || err != nil {
		t.Error("wrong address or account id received from server")
		t.Fail()
	}
}

// Access the server's bonus system for testing purposes
func testExecuteAccessBonusSystem(server *Server, accountID, addressID uint32, t *testing.T) (adrBdl *crypt.AddressBundle, keys []*hdkeychain.ExtendedKey, tokens, recoveries map[string]string) {
	codes := createValidTestCodes(t, server)

	// generate a seed and keys
	seed, keys, err := crypt.GetWalletKeys(utMnemonic, accountID, false)
	if err != nil {
		fail(t, err.Error())
	}

	// calculate the address
	address := crypt.GetAddress(keys[4], addressID)
	adrBdl = &crypt.AddressBundle{Seed: seed, AccountID: accountID, AddressID: addressID, Address: address.String()}
	tokens, recoveries, _ = server.AccessBonusSystem(codes, adrBdl)
	return
}

func testExecuteParticipation(client *Client, server *Server, bLevelID, token, pkr string, t *testing.T) (nextToken, nextRecovery, bonusData string) {
	bLevel := client.BonusLevels[bLevelID]

	////////// step 1: Get blind Token and signature for address update ////////////
	blindBundle, signature, err := client.getSignatureForToken(bLevel, token, ActionParticipate)
	if err != nil {
		t.Error(err)
		t.Fail()
	}
	// execute participate
	if nextToken, nextRecovery, bonusData, err = server.Participate(bLevelID, blindBundle.HashValue, signature, pkr); err != nil {
		t.Error(err)
		t.Fail()
	}
	return
}

func testExecuteSetAddress(client *Client, server *Server, bLevelID, token string, adrBdl *crypt.AddressBundle, keys []*hdkeychain.ExtendedKey, adrIdOfUpdate uint32, pkr string, t *testing.T) (adrBdlOfUpd *crypt.AddressBundle, tokenAfterUpd, recAfterUpd string) {
	var err error
	adrBdlOfUpd, tokenAfterUpd, recAfterUpd, err = testExecuteSetAddressNoErrorHdl(client, server, bLevelID, token, adrBdl, keys, adrIdOfUpdate, pkr)
	if err != nil {
		t.Error(err)
		t.Fail()
	}
	return
}

func testExecuteSetAddressNoErrorHdl(client *Client, server *Server, bLevelID, token string, adrBdl *crypt.AddressBundle, keys []*hdkeychain.ExtendedKey, adrIdOfUpdate uint32, pkr string) (adrBdlOfUpd *crypt.AddressBundle, tokenAfterUpd, recAfterUpd string, err error) {
	bLevel := client.BonusLevels[bLevelID]

	////////// step 1: Get blind Token and signature for address update ////////////
	var blindBundle *crypt.BlindBundle
	var signature []byte
	blindBundle, signature, err = client.getSignatureForToken(bLevel, token, ActionParticipate)
	if err != nil {
		return
	}

	////////// step 2: Update address ////////////
	// calculate the next address
	address := crypt.GetAddress(keys[4], adrIdOfUpdate)
	adrBdlOfUpd = &crypt.AddressBundle{Seed: adrBdl.Seed, AccountID: adrBdl.AccountID, AddressID: adrIdOfUpdate, Address: address.String()}
	// execute address update
	tokenAfterUpd, recAfterUpd, err = server.SetAddress(bLevelID, blindBundle.HashValue, signature, adrBdlOfUpd, ActionParticipate, pkr)
	if err != nil {
		return
	}
	return
}

func TestServer_Reset(t *testing.T) {
	s := NewServer()
	if len(s.flightMap) != 100 || len(s.Hierarchy) == 0 || len(s.BonusCodes) != 0 || len(s.BonusList) == 0 {
		t.Error("wrong setup for test")
		t.Fail()
	}
	// add a booking
	s.flightMap[1].AddBooking(1, s.BonusList["low"])
	if len(s.flightMap[1].Bookings) != 1 {
		t.Error("adding booking went wrong")
		t.Fail()
	}

	s.Reset()
	if len(s.flightMap) != 100 || len(s.Hierarchy) == 0 || len(s.BonusCodes) != 0 || len(s.BonusList) == 0 {
		t.Error("wrong length of lists/maps")
		t.Fail()
	}
	if len(s.flightMap[1].Bookings) != 0 {
		t.Error("bookings not erased")
		t.Fail()
	}
}

func TestServer_Register(t *testing.T) {
	var clientID int
	var err error

	s := NewServer()
	clientID = -1
	if clientID, err = s.Register(); err != nil {
		t.Error(err)
		t.Fail()
	}
	if clientID != 0 {
		t.Errorf("wrong clientID %d", clientID)
		t.Fail()
	}
	if clientID, err = s.Register(); err != nil {
		t.Error(err)
		t.Fail()
	}
	if clientID != 1 {
		t.Errorf("wrong clientID %d", clientID)
		t.Fail()
	}
}

func TestServer_GetStatisticSummary(t *testing.T) {
	var err error

	s := NewServer()
	if _, err = s.Register(); err != nil {
		t.Error(err)
		t.Fail()
	}
	summary := s.GetStatisticSummary()
	if summary == nil {
		t.Error("no summary found")
		t.FailNow()
	}
	if len(summary.BLevelToSummary) != 3 {
		t.Errorf("wrong number of bonus levels in summary: %d", len(summary.BLevelToSummary))
		t.Fail()
	}
}
