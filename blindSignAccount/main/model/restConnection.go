package model

import (
	"blindSignAccount/main/crypt"
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"time"
)

func readBody(response *http.Response, data interface{}) error {
	bodyBytes, _ := ioutil.ReadAll(response.Body)
	err := json.Unmarshal(bodyBytes, &data)
	if err != nil {
		err = errors.New(err.Error() + " (body:" + string(bodyBytes) + ")")
	}
	return err
}

type RestConnection struct {
	netClient *http.Client
}

func NewRestConnection() *RestConnection {
	return &RestConnection{netClient: &http.Client{Timeout: time.Minute * 3}}
}

func (con *RestConnection) GetSystemInformation() ([]*Flight, []*BonusLevel, error) {
	var msg MsgResponseSystemInfo
	var err error
	var resp *http.Response

	if resp, err = con.netClient.Get(ServerAddress + RoutePath(PathGetSystemInformation).String()); err != nil {
		return nil, nil, err
	}
	if err = readBody(resp, &msg); err != nil {
		return nil, nil, err
	}
	if msg.Err != "" {
		return nil, nil, errors.New(msg.Err)
	}
	return msg.Data.Flights, msg.Data.BLevels, nil
}

func (con *RestConnection) SendBooking(customerID, flightID int, bLevelID string) (string, error) {
	var msg MsgResponseSendBooking
	var err error
	var resp *http.Response

	values := map[string]interface{}{"customerID": customerID, "flightID": flightID, "bLevelID": bLevelID}
	jsonValue, _ := json.Marshal(values)
	if resp, err = con.netClient.Post(ServerAddress+RoutePath(PathSendBooking).String(),
		"application/json", bytes.NewBuffer(jsonValue)); err != nil {
		return "", err
	}

	if err = readBody(resp, &msg); err != nil {
		return "", err
	}
	if msg.Err != "" {
		return "", errors.New(msg.Err)
	}
	return msg.Data.Token, nil
}

func (con *RestConnection) GetBlindSignature(bLevelID, token string, blindToken []byte, action int) (string, error) {
	var msg MsgResponseBlindSignature
	var err error
	var resp *http.Response
	values := map[string]interface{}{"token": token, "blindToken": hex.EncodeToString(blindToken), "action": action, "bLevelID": bLevelID}
	jsonValue, _ := json.Marshal(values)
	if resp, err = con.netClient.Post(ServerAddress+RoutePath(PathBlindSignature).String(),
		"application/json", bytes.NewBuffer(jsonValue)); err != nil {
		return "", err
	}

	if err = readBody(resp, &msg); err != nil {
		return "", err
	}
	if msg.Err != "" {
		return "", errors.New(msg.Err)
	}

	return msg.Data.BlindSignature, nil
}

func (con *RestConnection) GetBookingCode(bLevelID string, hashValue, signature []byte) (string, error) {
	var msg MsgResponseGetBookingCode
	var err error
	var resp *http.Response

	values := map[string]interface{}{"hashValue": hex.EncodeToString(hashValue), "signature": hex.EncodeToString(signature), "bLevelID": bLevelID}
	jsonValue, _ := json.Marshal(values)
	if resp, err = con.netClient.Post(ServerAddress+RoutePath(PathGetBookingCode).String(),
		"application/json", bytes.NewBuffer(jsonValue)); err != nil {
		return "", err
	}

	if err = readBody(resp, &msg); err != nil {
		return "", err
	}
	if msg.Err != "" {
		return "", errors.New(msg.Err)
	}

	return msg.Data.Code, nil
}

func (con *RestConnection) AccessBonusSystem(codes []string, adrBundle *crypt.AddressBundle) (tokens, recoveries map[string]string, err error) {
	var msg MsgResponseAccessBS
	var resp *http.Response

	values := map[string]interface{}{"codes": codes, "adrBundle": encodeAdrBdl(adrBundle)}
	jsonValue, _ := json.Marshal(values)
	if resp, err = con.netClient.Post(ServerAddress+RoutePath(PathAccessBonusSystem).String(),
		"application/json", bytes.NewBuffer(jsonValue)); err != nil {
		return nil, nil, err
	}

	if err = readBody(resp, &msg); err != nil {
		return nil, nil, err
	}
	if msg.Err != "" {
		return nil, nil, errors.New(msg.Err)
	}

	return msg.Data.Tokens, msg.Data.RecoveryTokens, nil
}

func (con *RestConnection) SetAddress(bLevelID string, hashValue, signature []byte, adrBundle *crypt.AddressBundle, action int, pkr string) (token, recovery string, err error) {
	var msg MsgResponseSetAdr
	var resp *http.Response

	values := map[string]interface{}{"bLevelID": bLevelID, "hashValue": hex.EncodeToString(hashValue),
		"signature": hex.EncodeToString(signature), "adrBundle": encodeAdrBdl(adrBundle), "action": action, "pkr": pkr}
	jsonValue, _ := json.Marshal(values)
	if resp, err = con.netClient.Post(ServerAddress+RoutePath(PathSetAddress).String(),
		"application/json", bytes.NewBuffer(jsonValue)); err != nil {
		return "", "", err
	}

	if err = readBody(resp, &msg); err != nil {
		return "", "", err
	}
	if msg.Err != "" {
		return "", "", errors.New(msg.Err)
	}

	return msg.Data.Token, msg.Data.RecoveryToken, nil
}

func (con *RestConnection) Participate(bLevelID string, hashValue, signature []byte, pkr string) (token, recoveryToken, bonusData string, err error) {
	var msg MsgResponseParticipate
	var resp *http.Response

	values := map[string]interface{}{"bLevelID": bLevelID, "hashValue": hex.EncodeToString(hashValue), "signature": hex.EncodeToString(signature), "pkr": pkr}
	jsonValue, _ := json.Marshal(values)
	if resp, err = con.netClient.Post(ServerAddress+RoutePath(PathParticipate).String(),
		"application/json", bytes.NewBuffer(jsonValue)); err != nil {
		return "", "", "", err
	}

	if err = readBody(resp, &msg); err != nil {
		return "", "", "", err
	}
	if msg.Err != "" {
		return "", "", "", errors.New(msg.Err)
	}

	return msg.Data.Token, msg.Data.RecoveryToken, msg.Data.BonusData, nil
}

func (con *RestConnection) CanBeUsedForRecovery(bLevelID string, adrBdl *crypt.AddressBundle) (status RecoveryStatus, token string, err error) {
	var msg MsgResponseRecStatus
	var resp *http.Response

	values := map[string]interface{}{"bLevelID": bLevelID, "adrBundle": encodeAdrBdl(adrBdl)}
	jsonValue, _ := json.Marshal(values)
	if resp, err = con.netClient.Post(ServerAddress+RoutePath(PathCanBesUsedForRecovery).String(),
		"application/json", bytes.NewBuffer(jsonValue)); err != nil {
		return Failure, "", err
	}

	if err = readBody(resp, &msg); err != nil {
		return Failure, "", err
	}
	if msg.Err != "" {
		return Failure, "", errors.New(msg.Err)
	}

	return msg.Data.RecoveryStatus, msg.Data.Token, nil
}

func (con *RestConnection) RecoveryTest(bLevelID, recoveryToken, pkr string, adrBdl *crypt.AddressBundle) (token, foundRecoveryToken, bonusData string, err error) {
	var msg MsgResponseRecoveryTest
	var resp *http.Response

	values := map[string]interface{}{"bLevelID": bLevelID, "recoveryToken": recoveryToken, "pkr": pkr, "adrBundle": encodeAdrBdl(adrBdl)}
	jsonValue, _ := json.Marshal(values)
	if resp, err = con.netClient.Post(ServerAddress+RoutePath(PathRecoveryTest).String(),
		"application/json", bytes.NewBuffer(jsonValue)); err != nil {
		return "", "", "", err
	}

	if err = readBody(resp, &msg); err != nil {
		return "", "", "", err
	}
	if msg.Err != "" {
		return "", "", "", errors.New(msg.Err)
	}

	return msg.Data.Token, msg.Data.FoundRecoveryToken, msg.Data.BonusData, nil
}

func (con *RestConnection) Register() (clientID int, err error) {
	var msg MsgResponseRegister
	var resp *http.Response

	if resp, err = con.netClient.Get(ServerAddress + RoutePath(PathRegister).String()); err != nil {
		return -1, err
	}
	if err := readBody(resp, &msg); err != nil {
		return -1, err
	}
	if msg.Err != "" {
		return -1, errors.New(msg.Err)
	}
	return msg.Data.ClientID, nil
}

func (con *RestConnection) Reset() error {
	if _, err := con.netClient.Get(ServerAddress + RoutePath(PathReset).String()); err != nil {
		return err
	}
	return nil
}

func encodeAdrBdl(adrBundle *crypt.AddressBundle) *MsgRequestAdrBundle {
	return &MsgRequestAdrBundle{Seed: hex.EncodeToString(adrBundle.Seed), AddressID: adrBundle.AddressID,
		AccountID: adrBundle.AccountID, Address: adrBundle.Address}
}

func (con *RestConnection) GetDebugInfos() (s *Server, err error) {
	var msg MsgResponseDebugInfo
	var resp *http.Response
	if resp, err = con.netClient.Get(ServerAddress + RoutePath(PathDebugInfos).String()); err != nil {
		return nil, err
	}
	if err = readBody(resp, &msg); err != nil {
		return nil, err
	}
	if msg.Err != "" {
		return nil, errors.New(msg.Err)
	}
	return msg.Data.Server, nil
}

func (con *RestConnection) GetLastAdrBdl(seed []byte, bLevelID string) (adr string, accountID uint32, err error) {
	var msg MsgResponseLastAdrBdl
	var resp *http.Response

	values := map[string]interface{}{"bLevelID": bLevelID, "seed": hex.EncodeToString(seed)}
	jsonValue, _ := json.Marshal(values)
	if resp, err = con.netClient.Post(ServerAddress+RoutePath(PathLastAdrBdl).String(),
		"application/json", bytes.NewBuffer(jsonValue)); err != nil {
		return "", uint32(0), err
	}

	if err = readBody(resp, &msg); err != nil {
		return "", uint32(0), err
	}
	if msg.Err != "" {
		return "", uint32(0), errors.New(msg.Err)
	}

	return msg.Data.Address, msg.Data.AccountID, nil
}
