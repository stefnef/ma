package model

import (
	"crypto/rand"
	"encoding/base64"
	"time"
)

type BonusCode struct {
	CodeID    string
	CreatedAt time.Time
	ValidFor  *BonusLevel
}

func NewBonusCodeWithID(codeID string, validFor *BonusLevel) *BonusCode {
	return &BonusCode{CodeID: codeID, ValidFor: validFor}
}

func NewBonusCode(validFor *BonusLevel) *BonusCode {
	codeID, _ := randomString(lengthBonusCode)
	return &BonusCode{CodeID: codeID, ValidFor: validFor, CreatedAt: time.Now()}
}

// Reports whether bc.createdAt is after t
func (bc *BonusCode) After(t time.Time) bool {
	return bc.CreatedAt.After(t)
}

// Generates a random string (base64 encoded)
func randomString(length int) (string, error) {
	var err error
	b := make([]byte, length)
	if _, err = rand.Read(b); err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(b), err
}

// Checks if the bonus code is expired for a given bonus level.
// A code is expired if the creation time was too long ago due to the
// bonus levels' validDuration.
func (bc *BonusCode) isValidForBonusLevel(bLevel *BonusLevel) bool {
	// get the expiration date of the bonus level
	expireAt := bc.CreatedAt.AddDate(0, 0, bLevel.ValidDuration)
	if time.Now().After(expireAt) {
		return false
	}
	return true
}

// Returns a list of valid bonus levels for the bonus code
func (bc *BonusCode) GetValidBonusLevels() []*BonusLevel {
	var validLevels []*BonusLevel
	visited := make(map[*BonusLevel]bool)
	if bc.isValidForBonusLevel(bc.ValidFor) {
		bc.searchValidBonusLevels(&validLevels, bc.ValidFor, &visited)
	}

	return validLevels
}

// Search recursively all bonus levels for which the bonus code is valid
func (bc *BonusCode) searchValidBonusLevels(validList *[]*BonusLevel, currBLevel *BonusLevel, visited *map[*BonusLevel]bool) {

	// mark the current level as visited
	(*visited)[currBLevel] = true

	// check the expiry date
	if bc.isValidForBonusLevel(currBLevel) {
		*validList = append(*validList, currBLevel)
	}

	// check all other 'lower' bonus level
	for _, lowerLevel := range currBLevel.LowerLevels {
		// if the lower level was not already visited => check it
		if (*visited)[lowerLevel] == false {
			bc.searchValidBonusLevels(validList, lowerLevel, visited)
		}
	}
}
