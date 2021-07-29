package model

import (
	"testing"
	"time"
)

func setupBonusCode() *BonusCode {
	bLevel := setupBonusLevel()
	return NewBonusCode(bLevel)
}

func TestNewBonusCode(t *testing.T) {
	bLevel := setupBonusLevel()
	bCode := NewBonusCode(bLevel)
	if bCode.CodeID == "" {
		t.Error("couldn't create correct ID")
		t.Fail()
	}
	if bCode.ValidFor != bLevel {
		t.Error("couldn't set the bonus level for new code")
		t.Fail()
	}
}

func TestBonusCode_compare(t *testing.T) {
	bCode := setupBonusCode()
	// compare with 'future' (1 day in future)
	isAfter := bCode.After(time.Now().AddDate(0, 0, 1))
	if isAfter {
		t.Error("wrong time comparison with future")
		t.Fail()
	}

	// compare with 'present'
	isAfter = bCode.After(bCode.CreatedAt)
	if isAfter {
		t.Error("wrong time comparison with present")
		t.Fail()
	}

	// compare with 'past' (3 days before)
	isAfter = bCode.After(time.Now().AddDate(0, 0, -3))
	if !isAfter {
		t.Error("wrong time comparison with past")
		t.Fail()
	}
}

func TestBonusCode_isExpiredForBonusLevel(t *testing.T) {
	bLevel := setupBonusLevel()
	bCode := BonusCode{ValidFor: bLevel}

	// create a valid date
	bCode.CreatedAt = time.Now().AddDate(0, 0, 1-bCode.ValidFor.ValidDuration)
	if !bCode.isValidForBonusLevel(bLevel) {
		t.Error("valid code not detected")
		t.Fail()
	}

	// create an invalid date
	bCode.CreatedAt = time.Now().AddDate(0, 0, -bCode.ValidFor.ValidDuration-1)
	if bCode.isValidForBonusLevel(bLevel) {
		t.Error("expired code not detected")
		t.Fail()
	}
}

func TestBonusCode_GetValidBonusLevels(t *testing.T) {
	lower := NewBonusLevel(utLowLevelID, 5, 1)
	middle := NewBonusLevel(utMiddleLevelID, 3, 1)
	middle.AddLowerLevel(lower)
	bCode := BonusCode{ValidFor: middle}

	// create a valid date
	bCode.CreatedAt = time.Now().AddDate(0, 0, 1-bCode.ValidFor.ValidDuration)
	validLevels := bCode.GetValidBonusLevels()
	if len(validLevels) != 2 {
		t.Errorf("Wrong length of valid levels: %v", len(validLevels))
		t.Fail()
	}
	if validLevels[0] != middle || validLevels[1] != lower {
		t.Errorf("Wrong list of valid levels: %v, %v", validLevels[0], validLevels[1])
		t.Fail()
	}

	// create am invalid date
	bCode.CreatedAt = time.Now().AddDate(0, 0, -bCode.ValidFor.ValidDuration-1)
	validLevels = bCode.GetValidBonusLevels()
	if len(validLevels) != 0 {
		t.Errorf("Wrong length of valid levels: %v", len(validLevels))
		t.Fail()
	}
}
