package handlers

import (
	"blindSignAccount/main/crypt"
	"blindSignAccount/main/model"
	"bytes"
	"encoding/json"
	"net/http"
	"reflect"
	"strings"
	"testing"
)

func TestGetSystemInformation(t *testing.T) {
	setup(t)
	var msgSystemInfo model.MsgResponseSystemInfo

	response := callURL("GET", model.RoutePath(model.PathGetSystemInformation).String(), http.StatusOK, nil, t)

	if err := json.Unmarshal([]byte(response.String()), &msgSystemInfo); err != nil {
		t.Error(err)
		t.Fail()
	}
	if msgSystemInfo.Err != "" {
		t.Error(msgSystemInfo.Err)
	}
	if len(msgSystemInfo.Data.Flights) == 0 {
		t.Error("wrong number of flights")
		t.Fail()
	}

	for _, sBLevel := range model.GetDefaultHBLS() {
		if !contains(sBLevel, msgSystemInfo.Data.BLevels) {
			t.Error("bonus level " + sBLevel.BonusID + " does not exist")
			t.Fail()
		}
	}
	if len(msgSystemInfo.Data.BLevels) != 3 {
		t.Error("wrong number of bonus levels")
		t.Fail()
	}

	if Server.CntReqGetSystemInformation != 1 {
		t.Error("wrong count for request")
		t.Fail()
	}
}

func contains(elem *model.BonusLevel, collection []*model.BonusLevel) bool {
	for idx := range collection {
		if reflect.TypeOf(elem) == reflect.TypeOf(collection[idx]) {
			if collection[idx].Equals(*elem) {
				return true
			}
		}
	}
	return false
}

func TestGetSystemRegister(t *testing.T) {
	setup(t)
	var msgSystemRegister *model.MsgResponseRegister

	response := callURL("GET", model.RoutePath(model.PathRegister).String(), http.StatusOK, nil, t)

	if err := json.Unmarshal([]byte(response.String()), &msgSystemRegister); err != nil {
		t.Error(err)
		t.Fail()
	}
	if msgSystemRegister.Err != "" {
		t.Error(msgSystemRegister.Err)
		t.Fail()
	}
	if msgSystemRegister.Data.ClientID != 0 {
		t.Error("wrong ID of client")
		t.Fail()
	}

	// 2nd time: New ID is expected
	response = callURL("GET", model.RoutePath(model.PathRegister).String(), http.StatusOK, nil, t)

	if err := json.Unmarshal([]byte(response.String()), &msgSystemRegister); err != nil {
		t.Error(err)
		t.Fail()
	}
	if msgSystemRegister.Err != "" {
		t.Error(msgSystemRegister.Err)
		t.Fail()
	}
	if msgSystemRegister.Data.ClientID != 1 {
		t.Error("wrong ID of client")
		t.Fail()
	}

	if Server.CntReqRegister != 2 {
		t.Error("wrong count for request")
		t.Fail()
	}
}

func TestPostBlindSignature(t *testing.T) {
	setup(t)
	var msgBlindSign *model.MsgResponseBlindSignature

	// try to fail
	values := map[string]interface{}{}
	jsonValue, _ := json.Marshal(values)
	response := callURL("POST", model.RoutePath(model.PathBlindSignature).String(), http.StatusBadRequest, bytes.NewBuffer(jsonValue), t)
	if err := json.Unmarshal([]byte(response.String()), &msgBlindSign); err != nil {
		t.Error(err)
		t.Fail()
	}
	if msgBlindSign.Err == "" {
		t.Error("no error msg received")
		t.Fail()
	}
	if !strings.Contains(msgBlindSign.Err, "missing") || !strings.Contains(msgBlindSign.Err, "bLevelID") ||
		!strings.Contains(msgBlindSign.Err, "token") || !strings.Contains(msgBlindSign.Err, "blindToken") ||
		!strings.Contains(msgBlindSign.Err, "action") {
		t.Error(msgBlindSign.Err)
		t.Fail()
	}
	if msgBlindSign.Data.BlindSignature != "" {
		t.Error("token received: " + msgBlindSign.Data.BlindSignature)
		t.Fail()
	}

	if Server.CntReqBlindSignature != 1 {
		t.Error("wrong count for request")
		t.Fail()
	}
}

func TestHdlSetAddress(t *testing.T) {
	setup(t)
	var msgSetAddress *model.MsgResponseSetAdr

	// try to fail
	values := map[string]interface{}{}
	jsonValue, _ := json.Marshal(values)
	response := callURL("POST", model.RoutePath(model.PathSetAddress).String(), http.StatusBadRequest, bytes.NewBuffer(jsonValue), t)
	if err := json.Unmarshal([]byte(response.String()), &msgSetAddress); err != nil {
		t.Error(err)
		t.Fail()
	}
	if msgSetAddress.Err == "" {
		t.Error("no error msg received")
		t.Fail()
	}
	if !strings.Contains(msgSetAddress.Err, "missing") || !strings.Contains(msgSetAddress.Err, "bLevelID") ||
		!strings.Contains(msgSetAddress.Err, "hashValue") || !strings.Contains(msgSetAddress.Err, "signature") ||
		!strings.Contains(msgSetAddress.Err, "action") || !strings.Contains(msgSetAddress.Err, "pkr") {
		t.Error(msgSetAddress.Err)
		t.Fail()
	}

	// correct rest api call, but wrong model data
	values = map[string]interface{}{"bLevelID": "middle", "hashValue": []byte{1, 2, 3, 4}, "signature": []byte{5, 6, 7, 8},
		"action": model.ActionBooking, "pkr": "123", "adrBundle": crypt.AddressBundle{Seed: []byte{1}}}
	jsonValue, _ = json.Marshal(values)
	response = callURL("POST", model.RoutePath(model.PathSetAddress).String(), http.StatusBadRequest, bytes.NewBuffer(jsonValue), t)
	if err := json.Unmarshal([]byte(response.String()), &msgSetAddress); err != nil {
		t.Error(err)
		t.Fail()
	}
	if msgSetAddress.Err == "" {
		t.Error("no error msg received")
		t.Fail()
	}
	if strings.Contains(msgSetAddress.Err, "missing") {
		t.Error("wrong error msg: " + msgSetAddress.Err)
		t.Fail()
	}
	if msgSetAddress.Data.Token != "" || msgSetAddress.Data.RecoveryToken != "" {
		t.Errorf("received token (%s) or recovery token (%s)", msgSetAddress.Data.Token, msgSetAddress.Data.RecoveryToken)
		t.Fail()
	}

	if Server.CntReqSetAddress != 2 {
		t.Error("wrong count for request")
		t.Fail()
	}
}

func TestStatistic(t *testing.T) {
	setup(t)
	var msgStatistic *model.MsgResponseStatistic
	var actionBooking = 0
	var actionParticipate = 1

	_, _ = Server.Booking(1, 1, "low")
	response := callURL("GET", model.RoutePath(model.PathStatistic).String(), http.StatusOK, nil, t)

	if err := json.Unmarshal([]byte(response.String()), &msgStatistic); err != nil {
		t.Error(err)
		t.Fail()
	}
	if msgStatistic.Err != "" {
		t.Error(msgStatistic.Err)
	}
	if len(msgStatistic.Data.BLevelToSummary) != 3 {
		t.Error("wrong number of bonus levels")
		t.Fail()
	}

	// find the correct slice index of action booking and action participate
	if msgStatistic.Data.BLevelToSummary["low"][actionBooking].BonusActionVariant != "ActionBooking" {
		actionBooking = 1
		actionParticipate = 0
	}

	// check number of reads, writes and length for action booking
	if msgStatistic.Data.BLevelToSummary["low"][actionBooking].Statistic[model.StatValidTokens].Length != 1 {
		t.Error("wrong number of valid tokens")
		t.Fail()
	}
	if msgStatistic.Data.BLevelToSummary["low"][actionBooking].Statistic[model.StatValidTokens].NrWrites != 1 {
		t.Error("wrong number of WRITES of valid tokens")
		t.Fail()
	}
	if msgStatistic.Data.BLevelToSummary["low"][actionBooking].Statistic[model.StatValidTokens].NrReads != 1 {
		t.Error("wrong number of READS of valid tokens")
		t.Fail()
	}

	// check number of reads, writes and length for action participate
	if msgStatistic.Data.BLevelToSummary["low"][actionParticipate].Statistic[model.StatValidTokens].Length != 0 {
		t.Error("wrong number of valid tokens")
		t.Fail()
	}
	if msgStatistic.Data.BLevelToSummary["low"][actionParticipate].Statistic[model.StatValidTokens].NrWrites != 0 {
		t.Error("wrong number of WRITES of valid tokens")
		t.Fail()
	}
	if msgStatistic.Data.BLevelToSummary["low"][actionParticipate].Statistic[model.StatValidTokens].NrReads != 0 {
		t.Error("wrong number of WRITES of valid tokens")
		t.Fail()
	}

	if Server.CntReqStatistic != 1 {
		t.Error("wrong count for request")
		t.Fail()
	}
}

func TestGetDebugInformation(t *testing.T) {
	setup(t)
	var msgSystemInfo model.MsgResponseDebugInfo

	response := callURL("GET", model.RoutePath(model.PathDebugInfos).String(), http.StatusOK, nil, t)

	if err := json.Unmarshal([]byte(response.String()), &msgSystemInfo); err != nil {
		t.Error(err)
		t.Fail()
	}
	if msgSystemInfo.Err != "" {
		t.Error(msgSystemInfo.Err)
	}
	if msgSystemInfo.Data.Server == nil {
		t.Error("server is nil")
		t.Fail()
	}
}
