package model

import (
	"blindSignAccount/main/crypt"
	"crypto/rand"
	"crypto/rsa"
	"encoding/hex"
	"errors"
	"sort"
	"sync"
)

const ActionBooking = 0
const ActionParticipate = 1

type BonusLevel struct {
	// The id\name of the bonus level
	BonusID string
	// duration in days for which generated codes are valid
	ValidDuration int
	// the minimal number of codes needed to access this level
	MinNrCodes int
	// access manager store valid tokens, addresses and codes for
	// specific actions
	ActionVariants []*BonusActionVariant

	// list of lower bonus levels
	// all valid codes for this level have to be valid for lower levels also
	LowerLevels []*BonusLevel
}

type bonusDataPair struct {
	Token         string
	RecoveryToken string
	BonusData     string
}

type BonusActionVariant struct {
	VariantID int
	// The servers stores a public key for every bonus level
	// This key is used for blind signatures
	PublicKey rsa.PublicKey
	// If a bonus level was successfully accessed, then
	// the airline will give a Token which is used to verify the access
	// mapping seeds tokens and addresses is needed for reconstructing the access
	SeedToAddress   map[string]string
	AddressToToken  map[string]string
	TokenToSeed     map[string]string
	SeedToAccountID map[string]uint32
	// tokens are needed for recovery
	AddressToRecovery map[string]string
	// maps a blinded recovery Token (pkr) to the address used in address update method
	PkrToAdrUpd map[string]string
	// maps a blinded recovery Token (pkr) to the bonus data received at bonus data action
	PkrToBonusData map[string]*bonusDataPair
	// marks which tokens are valid or where used in past
	ValidTokens map[string]bool
	SkKey       *rsa.PrivateKey // private key
	// an additional map is needed which maps seeds to the address which
	// was used for accessing
	SeedToAccessAdr map[string]string
	// stores the 2nd last address that was used for address update
	PenultimateAdr map[string]string

	// statistic elements save number of reads and writes of maps
	Statistic [10]*Statistic

	// we need a mutex
	Mux            sync.Mutex
	MuxValidTokens sync.Mutex
	MuxStatistic   sync.Mutex
}

func NewBonusActionVariant(variantID int) *BonusActionVariant {
	bAV := &BonusActionVariant{
		VariantID:         variantID,
		SeedToAddress:     make(map[string]string, 0),
		SeedToAccountID:   make(map[string]uint32, 0),
		AddressToToken:    map[string]string{},
		TokenToSeed:       map[string]string{},
		ValidTokens:       map[string]bool{},
		AddressToRecovery: map[string]string{},
		SeedToAccessAdr:   map[string]string{},
		PenultimateAdr:    map[string]string{},
		PkrToAdrUpd:       map[string]string{},
		PkrToBonusData:    map[string]*bonusDataPair{},
		Statistic:         NewStatisticArray(),
	}
	bAV.SkKey, _ = rsa.GenerateKey(rand.Reader, crypt.KeyLength)
	bAV.PublicKey = bAV.SkKey.PublicKey
	return bAV
}

func NewBonusLevel(id string, duration, minNrCodes int) *BonusLevel {
	b := &BonusLevel{BonusID: id,
		ValidDuration:  duration,
		MinNrCodes:     minNrCodes,
		LowerLevels:    []*BonusLevel{},
		ActionVariants: make([]*BonusActionVariant, 2),
	}
	b.ActionVariants[ActionBooking] = NewBonusActionVariant(ActionBooking)
	b.ActionVariants[ActionParticipate] = NewBonusActionVariant(ActionParticipate)
	return b
}

func (b BonusLevel) Equals(other BonusLevel) bool {
	if b.BonusID != other.BonusID || b.MinNrCodes != other.MinNrCodes || b.ValidDuration != other.ValidDuration {
		return false
	}
	if len(b.LowerLevels) != len(other.LowerLevels) {
		return false
	}
	sort.Slice(b.LowerLevels, func(i, j int) bool {
		if b.LowerLevels[i].BonusID == b.LowerLevels[j].BonusID {
			return true
		}
		return false
	})
	sort.Slice(other.LowerLevels, func(i, j int) bool {
		if other.LowerLevels[i].BonusID == other.LowerLevels[j].BonusID {
			return true
		}
		return false
	})
	for idx, elem := range b.LowerLevels {
		if elem.BonusID != other.LowerLevels[idx].BonusID {
			return false
		}
	}
	return true
}

// Creates a copy for public use
func (b *BonusLevel) CopyPublic() *BonusLevel {
	// sync
	b.ActionVariants[ActionBooking].Mux.Lock()
	b.ActionVariants[ActionParticipate].Mux.Lock()
	defer b.ActionVariants[ActionBooking].Mux.Unlock()
	defer b.ActionVariants[ActionParticipate].Mux.Unlock()

	copyBLevel := &BonusLevel{BonusID: b.BonusID, ValidDuration: b.ValidDuration,
		MinNrCodes: b.MinNrCodes, ActionVariants: make([]*BonusActionVariant, 2)}
	copyBLevel.ActionVariants[ActionBooking] = &BonusActionVariant{PublicKey: b.ActionVariants[ActionBooking].PublicKey}
	copyBLevel.ActionVariants[ActionParticipate] = &BonusActionVariant{PublicKey: b.ActionVariants[ActionParticipate].PublicKey}
	for _, lLevel := range b.LowerLevels {
		copyBLevel.LowerLevels = append(copyBLevel.LowerLevels, lLevel.CopyPublic())
	}
	return copyBLevel
}

// Marks a new Token as valid
func (b *BonusLevel) addValidToken(token string, action int) error {
	// sync
	//b.ActionVariants[action].Mux.Lock()
	//defer b.ActionVariants[action].Mux.Unlock()
	b.ActionVariants[action].MuxValidTokens.Lock()
	defer b.ActionVariants[action].MuxValidTokens.Unlock()

	used, contained := b.ActionVariants[action].ValidTokens[token]
	SaveRead(StatValidTokens, b.ActionVariants[action])
	if used == true {
		return errors.New("token was already used")
	}
	if contained == true {
		return errors.New("token is already valid")
	}
	b.ActionVariants[action].ValidTokens[token] = false
	SaveWrite(StatValidTokens, b.ActionVariants[action])
	return nil
}

func (b *BonusLevel) isTokenValid(token string, action int) (valid bool) {
	// sync
	//b.ActionVariants[action].Mux.Lock()
	//defer b.ActionVariants[action].Mux.Unlock()
	b.ActionVariants[action].MuxValidTokens.Lock()
	defer b.ActionVariants[action].MuxValidTokens.Unlock()

	SaveRead(StatValidTokens, b.ActionVariants[action])
	used, contained := b.ActionVariants[action].ValidTokens[token]
	if contained == false || used == true {
		return false
	}
	return true
}

// Deletes a given Token from the list of valid tokens
func (b *BonusLevel) markTokenAsUsed(token string, action int) {
	// sync
	b.ActionVariants[action].Mux.Lock()
	defer b.ActionVariants[action].Mux.Unlock()
	b.ActionVariants[action].MuxValidTokens.Lock()
	defer b.ActionVariants[action].MuxValidTokens.Unlock()
	b.ActionVariants[action].ValidTokens[token] = true
	SaveWrite(StatValidTokens, b.ActionVariants[action])
}

// Remove old values which were related to the seed or address
func (b *BonusLevel) refreshMaps(recoveryToken, token, address string, seed []byte, action int, acntID uint32) {
	// sync
	b.ActionVariants[action].Mux.Lock()
	defer b.ActionVariants[action].Mux.Unlock()

	// get the old values
	oldAddress := b.ActionVariants[action].SeedToAddress[hex.EncodeToString(seed)]
	SaveRead(StatSeedToAddress, b.ActionVariants[action])
	oldToken := b.ActionVariants[action].AddressToToken[oldAddress]
	SaveRead(StatAddressToToken, b.ActionVariants[action])

	// remove the old values and add the new ones
	b.ActionVariants[action].TokenToSeed[oldToken] = ""
	SaveWrite(StatTokenToSeed, b.ActionVariants[action])
	b.ActionVariants[action].TokenToSeed[token] = hex.EncodeToString(seed)
	SaveWrite(StatTokenToSeed, b.ActionVariants[action])
	b.ActionVariants[action].AddressToToken[oldAddress] = ""
	SaveWrite(StatAddressToToken, b.ActionVariants[action])
	b.ActionVariants[action].AddressToToken[address] = token
	SaveWrite(StatAddressToToken, b.ActionVariants[action])
	b.ActionVariants[action].SeedToAddress[hex.EncodeToString(seed)] = address
	SaveWrite(StatSeedToAddress, b.ActionVariants[action])
	b.ActionVariants[action].SeedToAccountID[hex.EncodeToString(seed)] = acntID
	SaveWrite(StatSeedToAccountID, b.ActionVariants[action])
	// the last address is now the 2nd last address
	b.ActionVariants[action].PenultimateAdr[hex.EncodeToString(seed)] = oldAddress
	SaveWrite(StatPenultimateAdr, b.ActionVariants[action])
	b.ActionVariants[action].AddressToRecovery[address] = recoveryToken
	SaveWrite(StatAddressToRecovery, b.ActionVariants[action])
	//no delete!!! delete(b.ActionVariants[action].AddressToRecovery, oldAddress)

}

// Adds a new lower level. Codes for this level are also valid for hierarchically
// lower levels.
func (b *BonusLevel) AddLowerLevel(lower *BonusLevel) {
	b.LowerLevels = append(b.LowerLevels, lower)
}

// Generates a default hierarchical bonus level system with three levels
// low - middle - high
func GetDefaultHBLS() (hbls map[string]*BonusLevel) {
	// create 3 levels
	hbls = map[string]*BonusLevel{
		"low":    NewBonusLevel("low", 50, 5),
		"middle": NewBonusLevel("middle", 30, 3),
		"high":   NewBonusLevel("high", 10, 1),
	}

	// create hierarchies
	hbls["high"].LowerLevels = []*BonusLevel{hbls["middle"], hbls["low"]}
	hbls["middle"].LowerLevels = []*BonusLevel{hbls["low"]}

	return hbls
}

func (b *BonusLevel) GetBonusData() (bonusData string) {
	const bonusDataLength = 16
	values := make([]byte, bonusDataLength)
	if _, err := rand.Read(values); err != nil {
		return ""
	}
	bonusData = b.BonusID + hex.EncodeToString(values)
	return
}

func (v *BonusActionVariant) GetName() string {
	switch v.VariantID {
	case ActionBooking:
		return "ActionBooking"
	case ActionParticipate:
		return "ActionParticipate"
	default:
		return "unknown action"
	}
}
