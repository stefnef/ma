package handlers

import (
	"blindSignAccount/main/model"
	"bytes"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestHdlAccessBonusSystem(t *testing.T) {
	setup(t)
	var msgAccessBS model.MsgResponseAccessBS

	codes := [...]string{"code1", "code2"}
	adrBundleMap := map[string]interface{}{"Seed": hex.EncodeToString([]byte("1234567890123456")), "AccountID": uint32(2),
		"Address": "adr2", "AddressID": uint32(3)}

	values := map[string]interface{}{"codes": codes, "adrBundle": adrBundleMap}
	jsonValue, _ := json.Marshal(values)
	response := callURL("POST", model.RoutePath(model.PathAccessBonusSystem).String(), http.StatusBadRequest, bytes.NewBuffer(jsonValue), t)
	if err := json.Unmarshal([]byte(response.String()), &msgAccessBS); err != nil {
		t.Error(err)
		t.Fail()
	}
	if msgAccessBS.Err == "" {
		t.Error("no error message received")
		t.Fail()
	}
	if msgAccessBS.Err != "address does not fit to given seed" {
		t.Error(msgAccessBS.Err)
		t.Fail()
	}
	if len(msgAccessBS.Data.Tokens) != 0 {
		t.Errorf("received tokens %q", msgAccessBS.Data.Tokens)
	}
	if len(msgAccessBS.Data.RecoveryTokens) != 0 {
		t.Errorf("received recovery tokens %q", msgAccessBS.Data.RecoveryTokens)
	}

	if Server.CntReqAccessBonusSystem != 1 {
		t.Error("wrong count for request")
		t.Fail()
	}
}

func TestHdlParticipate(t *testing.T) {
	setup(t)
	var msgParticipate *model.MsgResponseParticipate

	// try to fail
	values := map[string]interface{}{}
	jsonValue, _ := json.Marshal(values)
	response := callURL("POST", model.RoutePath(model.PathParticipate).String(), http.StatusBadRequest, bytes.NewBuffer(jsonValue), t)
	if err := json.Unmarshal([]byte(response.String()), &msgParticipate); err != nil {
		t.Error(err)
		t.Fail()
	}
	if msgParticipate.Err == "" {
		t.Error("no error msg received")
		t.Fail()
	}
	if !strings.Contains(msgParticipate.Err, "missing") || !strings.Contains(msgParticipate.Err, "bLevelID") ||
		!strings.Contains(msgParticipate.Err, "hashValue") || !strings.Contains(msgParticipate.Err, "signature") ||
		!strings.Contains(msgParticipate.Err, "pkr") {
		t.Error(msgParticipate.Err)
		t.Fail()
	}

	// correct rest api call, but wrong model data
	values = map[string]interface{}{"bLevelID": "middle", "hashValue": "abc", "signature": "ABC", "pkr": "123"}
	jsonValue, _ = json.Marshal(values)
	response = callURL("POST", model.RoutePath(model.PathParticipate).String(), http.StatusBadRequest, bytes.NewBuffer(jsonValue), t)
	if err := json.Unmarshal([]byte(response.String()), &msgParticipate); err != nil {
		t.Error(err)
		t.Fail()
	}
	if msgParticipate.Err == "" {
		t.Error("no error msg received")
		t.Fail()
	}
	if msgParticipate.Err != "crypto/rsa: verification error" {
		t.Error("wrong error msg: " + msgParticipate.Err)
		t.Fail()
	}
	if msgParticipate.Data.Token != "" || msgParticipate.Data.RecoveryToken != "" {
		t.Errorf("received token (%s) or recovery token (%s)", msgParticipate.Data.Token, msgParticipate.Data.RecoveryToken)
		t.Fail()
	}
	if msgParticipate.Data.BonusData != "" {
		t.Errorf("received bonus data: %s", msgParticipate.Data.BonusData)
		t.Fail()
	}

	if Server.CntReqParticipate != 2 {
		t.Error("wrong count for request")
		t.Fail()
	}
}
