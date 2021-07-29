package handlers

import (
	"blindSignAccount/main/crypt"
	"blindSignAccount/main/model"
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestHdlCanBeUsedForRecovery(t *testing.T) {
	setup(t)
	var msgRecovery *model.MsgResponseRecStatus

	// try to fail
	values := map[string]interface{}{}
	jsonValue, _ := json.Marshal(values)
	response := callURL("POST", model.RoutePath(model.PathCanBesUsedForRecovery).String(), http.StatusBadRequest, bytes.NewBuffer(jsonValue), t)
	if err := json.Unmarshal([]byte(response.String()), &msgRecovery); err != nil {
		t.Error(err)
		t.Fail()
	}
	if msgRecovery.Err == "" {
		t.Error("no error msgRecovery received")
		t.Fail()
	}
	if !strings.Contains(msgRecovery.Err, "missing") || !strings.Contains(msgRecovery.Err, "bLevelID") ||
		!strings.Contains(msgRecovery.Err, "adrBundle") {
		t.Error(msgRecovery.Err)
		t.Fail()
	}

	// correct rest api call, but wrong model data
	values = map[string]interface{}{"bLevelID": "middle", "adrBundle": crypt.AddressBundle{Seed: []byte{1}}}
	jsonValue, _ = json.Marshal(values)
	response = callURL("POST", model.RoutePath(model.PathCanBesUsedForRecovery).String(), http.StatusAccepted, bytes.NewBuffer(jsonValue), t)
	if err := json.Unmarshal([]byte(response.String()), &msgRecovery); err != nil {
		t.Error(err)
		t.Fail()
	}
	if msgRecovery.Err == "" {
		t.Error("received no error msg")
		t.Fail()
	}
	if msgRecovery.Data.RecoveryStatus != model.Failure {
		t.Errorf("received wrong recovery status: %s", msgRecovery.Data.RecoveryStatus)
		t.Fail()
	}
	if msgRecovery.Data.Token != "" {
		t.Errorf("received non-empty token: %s", msgRecovery.Data.Token)
		t.Fail()
	}

	if Server.CntReqCanBesUsedForRecovery != 2 {
		t.Error("wrong count for request")
		t.Fail()
	}
}

func TestHdlRecoveryTest(t *testing.T) {
	setup(t)
	var msgRecovery *model.MsgResponseRecoveryTest

	// try to fail
	values := map[string]interface{}{}
	jsonValue, _ := json.Marshal(values)
	response := callURL("POST", model.RoutePath(model.PathRecoveryTest).String(), http.StatusBadRequest, bytes.NewBuffer(jsonValue), t)
	if err := json.Unmarshal([]byte(response.String()), &msgRecovery); err != nil {
		t.Error(err)
		t.Fail()
	}
	if msgRecovery.Err == "" {
		t.Error("no error msgRecovery received")
		t.Fail()
	}
	if !strings.Contains(msgRecovery.Err, "missing") || !strings.Contains(msgRecovery.Err, "bLevelID") ||
		!strings.Contains(msgRecovery.Err, "recoveryToken") || !strings.Contains(msgRecovery.Err, "pkr") ||
		!strings.Contains(msgRecovery.Err, "adrBundle") {
		t.Error(msgRecovery.Err)
		t.Fail()
	}
	// correct rest api call, but wrong model data
	values = map[string]interface{}{"bLevelID": "middle", "adrBundle": crypt.AddressBundle{Seed: []byte{1}},
		"recoveryToken": "recoveryToken", "pkr": "pkr"}
	jsonValue, _ = json.Marshal(values)
	response = callURL("POST", model.RoutePath(model.PathRecoveryTest).String(), http.StatusBadRequest, bytes.NewBuffer(jsonValue), t)
	if err := json.Unmarshal([]byte(response.String()), &msgRecovery); err != nil {
		t.Error(err)
		t.Fail()
	}
	if msgRecovery.Err == "" {
		t.Error("no error msg received")
		t.Fail()
	}
	if msgRecovery.Data.Token != "" || msgRecovery.Data.FoundRecoveryToken != "" || msgRecovery.Data.BonusData != "" {
		t.Errorf("wrong value(s): token (%s), recovery token (%s), bonus data (%s)", msgRecovery.Data.Token,
			msgRecovery.Data.FoundRecoveryToken, msgRecovery.Data.BonusData)
	}

	if Server.CntReqRecoveryTest != 2 {
		t.Error("wrong count for request")
		t.Fail()
	}
}

func TestHdlGetLastAdrBundle(t *testing.T) {
	setup(t)
	var msgLastAdrBdl *model.MsgResponseLastAdrBdl

	// try to fail
	values := map[string]interface{}{}
	jsonValue, _ := json.Marshal(values)
	response := callURL("POST", model.RoutePath(model.PathLastAdrBdl).String(), http.StatusBadRequest, bytes.NewBuffer(jsonValue), t)
	if err := json.Unmarshal([]byte(response.String()), &msgLastAdrBdl); err != nil {
		t.Error(err)
		t.Fail()
	}
	if msgLastAdrBdl.Err == "" {
		t.Error("no error msgLastAdrBdl received")
		t.Fail()
	}
	if !strings.Contains(msgLastAdrBdl.Err, "missing") || !strings.Contains(msgLastAdrBdl.Err, "bLevelID") ||
		!strings.Contains(msgLastAdrBdl.Err, "seed") {
		t.Error(msgLastAdrBdl.Err)
		t.Fail()
	}
	// correct rest api call, but wrong model data
	values = map[string]interface{}{"bLevelID": "middle", "seed": []byte{1}}
	jsonValue, _ = json.Marshal(values)
	response = callURL("POST", model.RoutePath(model.PathLastAdrBdl).String(), http.StatusNotFound, bytes.NewBuffer(jsonValue), t)
	if err := json.Unmarshal([]byte(response.String()), &msgLastAdrBdl); err != nil {
		t.Error(err)
		t.Fail()
	}
	if msgLastAdrBdl.Err == "" {
		t.Error("no error msg received")
		t.Fail()
	}
	if msgLastAdrBdl.Data.Address != "" || msgLastAdrBdl.Data.AccountID != uint32(0) {
		t.Errorf("wrong value(s): adr (%s), accountID (%d)", msgLastAdrBdl.Data.Address,
			msgLastAdrBdl.Data.AccountID)
	}

	if Server.CntReqGetLastAdrBundle != 2 {
		t.Error("wrong count for request")
		t.Fail()
	}
}
