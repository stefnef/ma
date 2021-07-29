package model

import (
	"blindSignAccount/main/crypt"
)

type RecoveryStatus int

const (
	RecoveryTestAfterAccess = iota
	RecoveryTestAfterFirstAdrUpd
	RecoveryTestPenultimateAdr
	RecoveryTest
	Failure
)

// String returns the name of the Recovery Status
func (status RecoveryStatus) String() string {
	names := [...]string{"RecoveryTestAfterAccess", "RecoveryTest After 1st Adr Upd", "RecoveryTest Penultimate Adr",
		"RecoveryTest", "Failure"}

	// handle out-of-range
	if status < RecoveryTestAfterAccess || status > Failure {
		return "unknown status"
	}

	return names[status]
}

type Connection interface {

	// Sends a booking of given customer, flight and bonus id to the server.
	SendBooking(customerID, flightID int, bLevelID string) (string, error)
	// Receives information about flights and the bonus system of the server
	GetSystemInformation() ([]*Flight, []*BonusLevel, error)
	// Sends a blind signature request to the server
	GetBlindSignature(bLevelID, token string, blindToken []byte, action int) (string, error)
	// Gets a code from the server
	GetBookingCode(bLevelID string, hashValue, signature []byte) (string, error)
	// Sends a request to the server for accessing the server's bonus system
	AccessBonusSystem(codes []string, adrBundle *crypt.AddressBundle) (tokens, recoveries map[string]string, err error)
	// Sends an address update to the server
	SetAddress(bLevelID string, hashValue, signature []byte, adrBundle *crypt.AddressBundle, action int, pkr string) (token, recovery string, err error)
	// Requests participation data from the server
	Participate(bLevelID string, hashValue, signature []byte, pkr string) (token, recoveryToken, bonusData string, err error)
	// Checks if a given address was set for the last address update.
	CanBeUsedForRecovery(bLevelID string, adrBdl *crypt.AddressBundle) (status RecoveryStatus, token string, err error)
	// A recovery test for receiving 'normal' Token or recovery Token from server
	RecoveryTest(bLevelID, recoveryToken, pkr string, adrBdl *crypt.AddressBundle) (token, foundRecoveryToken, bonusData string, err error)
	// Register a new user
	Register() (clientID int, err error)
	// PathReset needed for tests
	Reset() error
	// Gets additional information about the server's state
	GetDebugInfos() (s *Server, err error)
	// Get the last address and account id of an address update step
	GetLastAdrBdl(seed []byte, bLevelID string) (adr string, accountID uint32, err error)
}

type utConnection struct {
	server *Server
}

func newUtConnection() *utConnection {
	return &utConnection{server: NewServer()}
}

func (con *utConnection) GetSystemInformation() ([]*Flight, []*BonusLevel, error) {
	return con.server.GetSystemInformation()
}

func (con *utConnection) SendBooking(customerID, flightID int, bLevelID string) (string, error) {
	return con.server.Booking(flightID, customerID, bLevelID)
}

func (con *utConnection) GetBlindSignature(bLevelID, token string, blindToken []byte, action int) (string, error) {
	return con.server.GetBlindSignature(bLevelID, token, blindToken, action)
}

func (con *utConnection) GetBookingCode(bLevelID string, hashValue, signature []byte) (string, error) {
	return con.server.GetBookingCode(bLevelID, hashValue, signature)
}

func (con *utConnection) AccessBonusSystem(codes []string, adrBundle *crypt.AddressBundle) (tokens, recoveries map[string]string, err error) {
	return con.server.AccessBonusSystem(codes, adrBundle)
}

func (con *utConnection) SetAddress(bLevelID string, hashValue, signature []byte, adrBundle *crypt.AddressBundle, action int, pkr string) (token, recoveryToken string, err error) {
	return con.server.SetAddress(bLevelID, hashValue, signature, adrBundle, action, pkr)
}

func (con *utConnection) Participate(bLevelID string, hashValue, signature []byte, pkr string) (token, recoveryToken, bonusData string, err error) {
	return con.server.Participate(bLevelID, hashValue, signature, pkr)
}

func (con *utConnection) CanBeUsedForRecovery(bLevelID string, adrBdl *crypt.AddressBundle) (status RecoveryStatus, token string, err error) {
	return con.server.CanBeUsedForRecovery(bLevelID, adrBdl)
}

func (con *utConnection) RecoveryTest(bLevelID, recoveryToken, pkr string, adrBdl *crypt.AddressBundle) (token, foundRecoveryToken, bonusData string, err error) {
	return con.server.RecoveryTest(bLevelID, recoveryToken, pkr, adrBdl)
}

func (con *utConnection) Register() (clientID int, err error) {
	return con.server.Register()
}

func (con *utConnection) Reset() error {
	con.server.Reset()
	return nil
}

func (con *utConnection) GetDebugInfos() (s *Server, err error) {
	return con.server, nil
}

func (con *utConnection) GetLastAdrBdl(seed []byte, bLevelID string) (adr string, accountID uint32, err error) {
	return con.server.GetLastAdrBundle(seed, bLevelID)
}
