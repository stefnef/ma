package model

import (
	"encoding/hex"
	"testing"
)

var utLowLevelID = "low"
var utMiddleLevelID = "middle"
var utHighLevelID = "high"

func setupBonusLevel() *BonusLevel {
	return NewBonusLevel(utLowLevelID, 5, 1)
}

func TestNewBonusLevel(t *testing.T) {
	bLevelLow := setupBonusLevel()
	if bLevelLow == nil {
		t.Error("bLevel is nil")
		t.Fail()
	}
}

func TestBonusLevel_validTokens(t *testing.T) {
	bLevelLow := setupBonusLevel()
	_ = bLevelLow.addValidToken("testToken_one", ActionParticipate)
	if err := bLevelLow.addValidToken("testToken_two", ActionParticipate); err != nil {
		t.Error(err)
		t.Fail()
	}
	if len(bLevelLow.ActionVariants[ActionParticipate].ValidTokens) != 2 {
		t.Error("wrong number of valid tokens")
		t.Fail()
	}
	// the tokens must not be valid for other action variants
	if len(bLevelLow.ActionVariants[ActionBooking].ValidTokens) != 0 {
		t.Error("Token valid for wrong action variant")
		t.Fail()
	}
	// try to add the first Token twice
	if err := bLevelLow.addValidToken("testToken_one", ActionParticipate); err == nil {
		t.Error(err)
		t.Fail()
	}
	if len(bLevelLow.ActionVariants[ActionParticipate].ValidTokens) != 2 {
		t.Error("wrong number of valid tokens")
		t.Fail()
	}
	if bLevelLow.ActionVariants[ActionParticipate].ValidTokens["testToken_one"] || bLevelLow.ActionVariants[ActionParticipate].ValidTokens["testToken_two"] {
		t.Error("tokens marked as used")
		t.Fail()
	}

	// mark the second Token as used
	bLevelLow.markTokenAsUsed("testToken_two", ActionParticipate)
	if bLevelLow.ActionVariants[ActionParticipate].ValidTokens["testToken_two"] != true {
		t.Error("Token not marked as used")
		t.Fail()
	}
	// try to add the second Token again
	if err := bLevelLow.addValidToken("testToken_two", ActionParticipate); err == nil {
		t.Error("used Token could be added")
		t.Fail()
	}
}

func TestBonusLevel_AddLowerLevel(t *testing.T) {
	lower := NewBonusLevel(utLowLevelID, 5, 1)
	middle := NewBonusLevel(utMiddleLevelID, 3, 1)
	middle.AddLowerLevel(lower)

	if len(lower.LowerLevels) != 0 || len(middle.LowerLevels) != 1 {
		t.Error("wrong hierarchical setting")
		t.Fail()
	}

	if middle.LowerLevels[0] != lower {
		t.Error("couldn't add lower level to middle")
		t.Fail()
	}
}

func TestBonusLevel_refreshMaps(t *testing.T) {
	lower := NewBonusLevel(utLowLevelID, 5, 1)
	actionVariant := lower.ActionVariants[ActionParticipate]
	action := ActionParticipate
	oldToken := "oldToken"
	token := "token1"
	oldAddress := "oldAddress"
	address := "address1"
	seed := []byte{1, 2, 3}
	oldRecovery := "oldRecovery"
	recovery := "recovery"

	// init the maps
	actionVariant.AddressToRecovery[oldAddress] = oldRecovery
	actionVariant.AddressToToken[oldAddress] = oldToken
	actionVariant.SeedToAddress[hex.EncodeToString(seed)] = oldAddress
	actionVariant.SeedToAccountID[hex.EncodeToString(seed)] = uint32(2)
	actionVariant.TokenToSeed[oldToken] = hex.EncodeToString(seed)

	lower.refreshMaps(recovery, token, address, seed, action, 4)

	if actionVariant.TokenToSeed[token] != hex.EncodeToString(seed) {
		t.Error("Token to seed")
	}
	if actionVariant.AddressToToken[address] != token {
		t.Error("address to Token")
	}
	if actionVariant.AddressToRecovery[address] != recovery {
		t.Error("address to recovery")
	}
	if _, ok := actionVariant.AddressToRecovery[oldAddress]; ok == false {
		t.Error("mapping of old address to recovery was deleted")
	}
	if actionVariant.SeedToAddress[hex.EncodeToString(seed)] != address {
		t.Error("seed to address")
	}
	if actionVariant.PenultimateAdr[hex.EncodeToString(seed)] != oldAddress {
		t.Error("seed to PenultimateAdr address")
	}
	if actionVariant.SeedToAccountID[hex.EncodeToString(seed)] != 4 {
		t.Error("seed not mapped to correct account id")
	}
}
