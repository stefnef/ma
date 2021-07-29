package model

type MsgDataRecStatus struct {
	Token          string
	RecoveryStatus RecoveryStatus
}

type MsgResponseRecStatus struct {
	Data MsgDataRecStatus
	Err  string
}

type MsgDataRecoveryTest struct {
	Token              string
	FoundRecoveryToken string
	BonusData          string
}

type MsgResponseRecoveryTest struct {
	Data MsgDataRecoveryTest
	Err  string
}

type MsgDataAccessBS struct {
	Tokens         map[string]string
	RecoveryTokens map[string]string
}

type MsgResponseAccessBS struct {
	Data MsgDataAccessBS
	Err  string
}

type MsgDataParticipate struct {
	Token         string
	RecoveryToken string
	BonusData     string
}

type MsgResponseParticipate struct {
	Data MsgDataParticipate
	Err  string
}

type MsgDataSystemInfo struct {
	Flights []*Flight
	BLevels []*BonusLevel
}

type MsgResponseStatistic struct {
	Data *StatisticSummary
	Err  string
}

type MsgResponseSystemInfo struct {
	Data MsgDataSystemInfo
	Err  string
}

type MsgDataBlindSignature struct {
	BlindSignature string
}

type MsgResponseBlindSignature struct {
	Data MsgDataBlindSignature
	Err  string
}

type MsgDataSetAdr struct {
	Token         string
	RecoveryToken string
}

type MsgResponseSetAdr struct {
	Data MsgDataSetAdr
	Err  string
}

type MsgDataSendBooking struct {
	Token string
}

type MsgResponseSendBooking struct {
	Data MsgDataSendBooking
	Err  string
}

type MsgDataGetBookingCode struct {
	Code string
}

type MsgResponseGetBookingCode struct {
	Data MsgDataGetBookingCode
	Err  string
}

type MsgRequestAdrBundle struct {
	Seed      string
	AccountID uint32
	AddressID uint32
	Address   string
}

type MsgDataRegister struct {
	ClientID int
}

type MsgResponseRegister struct {
	Data MsgDataRegister
	Err  string
}

type MsgDataDebugInfo struct {
	Server *Server
}

type MsgResponseDebugInfo struct {
	Data MsgDataDebugInfo
	Err  string
}

type MsgDataLastAdrBdl struct {
	Address   string
	AccountID uint32
}

type MsgResponseLastAdrBdl struct {
	Data MsgDataLastAdrBdl
	Err  string
}
