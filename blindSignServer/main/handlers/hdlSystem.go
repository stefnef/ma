package handlers

import (
	"blindSignAccount/main/crypt"
	"blindSignAccount/main/model"
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
	"syscall"
	"time"
)

func GetReset(c *gin.Context) {
	var status = http.StatusBadRequest
	var err error
	var data = make(map[string]interface{}, 0)

	Server.CntReqReset++

	err = errors.New("unknown error")
	defer render(c, gin.H{"payload": &data}, &status, &err)

	Server.Reset()
	status = http.StatusOK
	err = nil
}

func GetSystemInformation(c *gin.Context) {
	var status = http.StatusBadRequest
	var err error
	var data = make(map[string]interface{}, 0)
	var flights []*model.Flight
	var bLevels []*model.BonusLevel

	Server.CntReqGetSystemInformation++

	err = errors.New("unknown error")
	defer render(c, gin.H{"payload": &data}, &status, &err)

	if flights, bLevels, err = Server.GetSystemInformation(); err != nil {
		return
	}
	data["flights"] = flights
	data["bLevels"] = bLevels

	status = http.StatusOK
}

func GetDebugInformation(c *gin.Context) {
	var status = http.StatusOK
	var err error
	data := map[string]interface{}{"server": Server}
	defer Server.Mux.Unlock()
	Server.Mux.Lock()
	render(c, gin.H{"payload": &data}, &status, &err)
}

func GetSystemStatistic(c *gin.Context) {
	var status = http.StatusBadRequest
	var err error
	var data *model.StatisticSummary

	Server.CntReqStatistic++

	err = errors.New("unknown error")
	defer render(c, gin.H{"payload": &data}, &status, &err)

	err = nil
	data = Server.GetStatisticSummary()
	status = http.StatusOK
}

func PostSystemExit(c *gin.Context) {
	var data string
	var status = http.StatusBadRequest
	var err error
	var exitStatus int

	Server.CntReqExit++

	elements := map[string]interface{}{"exitStatus": exitStatus}
	err = errors.New("unknown error")
	defer render(c, gin.H{"payload": &data}, &status, &err)

	if err = parseBody(c, &elements); err != nil {
		return
	}

	exitStatus = elements["exitStatus"].(int)
	if exitStatus != 27 {
		return
	}

	go func() {
		time.Sleep(60 * time.Second)
		syscall.Exit(0)
	}()
}

func GetSystemRegister(c *gin.Context) {
	var status = http.StatusBadRequest
	var err error
	var clientID int
	var data = make(map[string]interface{}, 1)

	Server.CntReqRegister++

	err = errors.New("unknown error")
	defer render(c, gin.H{"payload": &data}, &status, &err)

	if clientID, err = Server.Register(); err != nil {
		return
	}

	data["clientID"] = clientID
	status = http.StatusOK
}

func PostBlindSignature(c *gin.Context) {
	var status = http.StatusBadRequest
	var bLevelID, token, blindSignature string
	var blindToken []byte
	var action int
	var err error
	var data = make(map[string]interface{}, 0)

	Server.CntReqBlindSignature++

	elements := map[string]interface{}{"bLevelID": bLevelID, "token": token, "blindToken": blindToken, "action": action}
	err = errors.New("unknown error")
	defer render(c, gin.H{"payload": &data}, &status, &err)

	if err = parseBody(c, &elements); err != nil {
		return
	}

	bLevelID = elements["bLevelID"].(string)
	token = elements["token"].(string)
	blindToken = elements["blindToken"].([]byte)
	action = elements["action"].(int)
	if action != model.ActionBooking && action != model.ActionParticipate {
		err = errors.New("unknown action")
		return
	}

	if blindSignature, err = Server.GetBlindSignature(bLevelID, token, blindToken, action); err != nil {
		return
	}

	data["blindSignature"] = blindSignature

	status = http.StatusAccepted
}

func HdlSetAddress(c *gin.Context) {
	var status = http.StatusBadRequest
	var bLevelID, pkr string
	var hashValue, signature []byte
	var adrBundle *crypt.AddressBundle
	var action int
	var err error
	var data = make(map[string]interface{}, 0)

	Server.CntReqSetAddress++

	elements := map[string]interface{}{"bLevelID": bLevelID, "hashValue": hashValue, "signature": signature,
		"adrBundle": adrBundle, "action": action, "pkr": pkr}
	err = errors.New("unknown error")
	defer render(c, gin.H{"payload": &data}, &status, &err)

	if err = parseBody(c, &elements); err != nil {
		return
	}

	bLevelID = elements["bLevelID"].(string)
	hashValue = elements["hashValue"].([]byte)
	signature = elements["signature"].([]byte)
	adrBundle = elements["adrBundle"].(*crypt.AddressBundle)
	action = elements["action"].(int)
	pkr = elements["pkr"].(string)
	if action != model.ActionBooking && action != model.ActionParticipate {
		err = errors.New("unknown action")
		return
	}
	if data["token"], data["recoveryToken"], err = Server.SetAddress(bLevelID, hashValue, signature, adrBundle, action, pkr); err != nil {
		return
	}

	status = http.StatusAccepted
}
