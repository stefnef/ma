package model

type StatName int

const (
	StatSeedToAddress = iota
	StatSeedToAccountID
	StatAddressToToken
	StatTokenToSeed
	StatValidTokens
	StatAddressToRecovery
	StatSeedToAccessAdr
	StatPenultimateAdr
	StatPkrToAdrUpd
	StatPkrToBonusData
)

func (name StatName) String() string {
	names := [10]string{
		"SeedToAddress", "SeedToAccountID", "AddressToToken", "TokenToSeed",
		"ValidTokens", "AddressToRecovery", "SeedToAccessAdr",
		"PenultimateAdr", "PkrToAdrUpd", "PkrToBonusData",
	}
	if !name.IsValid() {
		return "unknown statistic name"
	}
	return names[name]
}

func (name StatName) IsValid() bool {
	if name < StatSeedToAddress || name > StatPkrToBonusData {
		return false
	}
	return true
}

func (name StatName) GetSize(variant *BonusActionVariant) int {
	switch name {
	case StatSeedToAddress:
		return len(variant.SeedToAddress)
	case StatSeedToAccountID:
		return len(variant.SeedToAccountID)
	case StatAddressToToken:
		return len(variant.AddressToToken)
	case StatTokenToSeed:
		return len(variant.TokenToSeed)
	case StatValidTokens:
		return len(variant.ValidTokens)
	case StatAddressToRecovery:
		return len(variant.AddressToRecovery)
	case StatSeedToAccessAdr:
		return len(variant.SeedToAccessAdr)
	case StatPenultimateAdr:
		return len(variant.PenultimateAdr)
	case StatPkrToAdrUpd:
		return len(variant.PkrToAdrUpd)
	case StatPkrToBonusData:
		return len(variant.PkrToBonusData)
	}

	return 0
}

type Statistic struct {
	Name     string `json:"Name"`
	NrReads  int    `json:"NrReads"`
	NrWrites int    `json:"NrWrites"`
	Length   int    `json:"Length"`
}

type StatisticSummary struct {
	BLevelToSummary             map[string][]StatisticSummaryTuple `json:"BLevelToSummary"`
	CntReqGetLastAdrBundle      int                                `json:"CntReqGetLastAdrBundle"`
	CntReqSendBooking           int                                `json:"CntReqSendBooking"`
	CntReqGetBookingCode        int                                `json:"CntReqGetBookingCode"`
	CntReqGetSystemInformation  int                                `json:"CntReqGetSystemInformation"`
	CntReqBlindSignature        int                                `json:"CntReqBlindSignature"`
	CntReqSetAddress            int                                `json:"CntReqSetAddress"`
	CntReqAccessBonusSystem     int                                `json:"CntReqAccessBonusSystem"`
	CntReqParticipate           int                                `json:"CntReqParticipate"`
	CntReqCanBesUsedForRecovery int                                `json:"CntReqCanBesUsedForRecovery"`
	CntReqRecoveryTest          int                                `json:"CntReqRecoveryTest"`
	CntReqRegister              int                                `json:"CntReqRegister"`
	CntReqExit                  int                                `json:"CntReqExit"`
	CntReqStatistic             int                                `json:"CntReqStatistic"`
	CntReqReset                 int                                `json:"CntReqReset"`
}

type StatisticSummaryTuple struct {
	BonusActionVariant string         `json:"BonusActionVariant"`
	Statistic          [10]*Statistic `json:"Statistic"`
}

func NewStatisticArray() [10]*Statistic {
	statistic := [10]*Statistic{
		{Name: StatName(StatSeedToAddress).String(), NrReads: 0, NrWrites: 0, Length: 0},
		{Name: StatName(StatSeedToAccountID).String(), NrReads: 0, NrWrites: 0, Length: 0},
		{Name: StatName(StatAddressToToken).String(), NrReads: 0, NrWrites: 0, Length: 0},
		{Name: StatName(StatTokenToSeed).String(), NrReads: 0, NrWrites: 0, Length: 0},
		{Name: StatName(StatValidTokens).String(), NrReads: 0, NrWrites: 0, Length: 0},
		{Name: StatName(StatAddressToRecovery).String(), NrReads: 0, NrWrites: 0, Length: 0},
		{Name: StatName(StatSeedToAccessAdr).String(), NrReads: 0, NrWrites: 0, Length: 0},
		{Name: StatName(StatPenultimateAdr).String(), NrReads: 0, NrWrites: 0, Length: 0},
		{Name: StatName(StatPkrToAdrUpd).String(), NrReads: 0, NrWrites: 0, Length: 0},
		{Name: StatName(StatPkrToBonusData).String(), NrReads: 0, NrWrites: 0, Length: 0},
	}
	return statistic
}

func SaveWrite(name StatName, variant *BonusActionVariant) {
	if !name.IsValid() {
		return
	}
	variant.MuxStatistic.Lock()
	defer variant.MuxStatistic.Unlock()
	variant.Statistic[name].NrWrites++
	variant.Statistic[name].Length = name.GetSize(variant)
}

func SaveRead(name StatName, variant *BonusActionVariant) {
	if !name.IsValid() {
		return
	}
	variant.MuxStatistic.Lock()
	defer variant.MuxStatistic.Unlock()
	variant.Statistic[name].NrReads++
}
