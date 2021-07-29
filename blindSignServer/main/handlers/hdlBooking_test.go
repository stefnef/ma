package handlers

import (
	"blindSignAccount/main/model"
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestSendBooking(t *testing.T) {
	setup(t)
	var msgBooking *model.MsgResponseSendBooking

	// try to fail
	msgBooking = nil
	values := map[string]interface{}{}
	jsonValue, _ := json.Marshal(values)
	response := callURL("POST", model.RoutePath(model.PathSendBooking).String(), http.StatusBadRequest, bytes.NewBuffer(jsonValue), t)
	if err := json.Unmarshal([]byte(response.String()), &msgBooking); err != nil {
		t.Error(err)
		t.Fail()
	}
	if msgBooking.Err == "" {
		t.Error("no error msg received")
		t.Fail()
	}
	if !strings.Contains(msgBooking.Err, "missing") || !strings.Contains(msgBooking.Err, "bLevelID") ||
		!strings.Contains(msgBooking.Err, "customerID") || !strings.Contains(msgBooking.Err, "flightID") {
		t.Error(msgBooking.Err)
		t.Fail()
	}
	if msgBooking.Data.Token != "" {
		t.Error("token received: " + msgBooking.Data.Token)
		t.Fail()
	}

	// must not fail
	msgBooking = nil
	values = map[string]interface{}{"customerID": 1, "flightID": 2, "bLevelID": "middle"}
	jsonValue, _ = json.Marshal(values)
	response = callURL("POST", model.RoutePath(model.PathSendBooking).String(), http.StatusAccepted, bytes.NewBuffer(jsonValue), t)
	if err := json.Unmarshal([]byte(response.String()), &msgBooking); err != nil {
		t.Error(err)
		t.Fail()
	}
	if msgBooking.Err != "" {
		t.Error(msgBooking.Err)
		t.Fail()
	}
	if msgBooking.Data.Token == "" {
		t.Error("no token received")
		t.Fail()
	}

	if Server.CntReqSendBooking != 2 {
		t.Error("wrong count for request")
		t.Fail()
	}
}

func TestHdlGetBookingCode(t *testing.T) {
	var msgCode *model.MsgResponseGetBookingCode

	// try to fail
	values := map[string]interface{}{"hashValue": "abc", "signature": "ABC", "bLevelID": "low"}
	jsonValue, _ := json.Marshal(values)
	response := callURL("POST", model.RoutePath(model.PathGetBookingCode).String(), http.StatusBadRequest, bytes.NewBuffer(jsonValue), t)
	if err := json.Unmarshal([]byte(response.String()), &msgCode); err != nil {
		t.Error(err)
		t.Fail()
	}
	if msgCode.Err == "" {
		t.Error("no error message received")
		t.Fail()
	}
	if !strings.Contains(msgCode.Err, "verification error") {
		t.Error(msgCode.Err)
		t.Fail()
	}
	if msgCode.Data.Code != "" {
		t.Error("received code '" + msgCode.Data.Code + "'")
		t.Fail()
	}
	if Server.CntReqGetBookingCode != 1 {
		t.Error("wrong count for request")
		t.Fail()
	}
}

func TestBookingBlindSignature(t *testing.T) {
	setup(t)
	var msgBooking *model.MsgResponseSendBooking
	var msgBlindSign *model.MsgResponseBlindSignature

	values := map[string]interface{}{"customerID": 1, "flightID": 2, "bLevelID": "middle"}
	jsonValue, _ := json.Marshal(values)
	response := callURL("POST", model.RoutePath(model.PathSendBooking).String(), http.StatusAccepted, bytes.NewBuffer(jsonValue), t)
	if err := json.Unmarshal([]byte(response.String()), &msgBooking); err != nil {
		t.Error(err)
		t.Fail()
	}

	values = map[string]interface{}{"bLevelID": "middle", "token": msgBooking.Data.Token, "action": model.ActionBooking,
		"blindToken": "123455BlindToken"}
	jsonValue, _ = json.Marshal(values)
	response = callURL("POST", model.RoutePath(model.PathBlindSignature).String(), http.StatusAccepted, bytes.NewBuffer(jsonValue), t)
	if err := json.Unmarshal([]byte(response.String()), &msgBlindSign); err != nil {
		t.Error(err)
		t.Fail()
	}
	if msgBlindSign.Err != "" {
		t.Error(msgBlindSign.Err)
		t.Fail()
	}
	if msgBlindSign.Data.BlindSignature == "" {
		t.Error("no blind signature received")
		t.Fail()
	}

	if Server.CntReqSendBooking != 1 || Server.CntReqBlindSignature != 1 {
		t.Error("wrong count for request")
		t.Fail()
	}
}
