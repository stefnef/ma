package handlers

import (
	"blindSignAccount/main/crypt"
	"blindSignAccount/main/model"
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
)

func HdlCanBeUsedForRecovery(c *gin.Context) {
	var status = http.StatusBadRequest
	var bLevelID, token string
	var recoveryStatus model.RecoveryStatus
	var adrBundle *crypt.AddressBundle
	var err error
	var data = make(map[string]interface{}, 0)

	Server.CntReqCanBesUsedForRecovery++

	elements := map[string]interface{}{"bLevelID": bLevelID, "adrBundle": adrBundle}
	err = errors.New("unknown error")
	defer render(c, gin.H{"payload": &data}, &status, &err)

	if err = parseBody(c, &elements); err != nil {
		return
	}

	bLevelID = elements["bLevelID"].(string)
	adrBundle = elements["adrBundle"].(*crypt.AddressBundle)

	recoveryStatus, token, err = Server.CanBeUsedForRecovery(bLevelID, adrBundle)

	data["token"] = token
	data["recoveryStatus"] = recoveryStatus

	status = http.StatusAccepted
}

func HdlRecoveryTest(c *gin.Context) {
	var status = http.StatusBadRequest
	var bLevelID, token, recoveryToken, pkr, foundRecovery, bonusData string
	var adrBundle *crypt.AddressBundle
	var err error
	var data = make(map[string]interface{}, 0)

	Server.CntReqRecoveryTest++

	elements := map[string]interface{}{"bLevelID": bLevelID, "recoveryToken": recoveryToken, "pkr": pkr, "adrBundle": adrBundle}
	err = errors.New("unknown error")
	defer render(c, gin.H{"payload": &data}, &status, &err)

	if err = parseBody(c, &elements); err != nil {
		return
	}

	bLevelID = elements["bLevelID"].(string)
	recoveryToken = elements["recoveryToken"].(string)
	pkr = elements["pkr"].(string)
	adrBundle = elements["adrBundle"].(*crypt.AddressBundle)

	if token, foundRecovery, bonusData, err = Server.RecoveryTest(bLevelID, recoveryToken, pkr, adrBundle); err != nil {
		return
	}

	data["token"] = token
	data["foundRecoveryToken"] = foundRecovery
	data["bonusData"] = bonusData
	status = http.StatusAccepted
}

func HdlGetLastAdrBundle(c *gin.Context) {
	var status = http.StatusBadRequest
	var seed []byte
	var adr, bLevelID string
	var accountID uint32
	var err error
	var data = make(map[string]interface{}, 0)

	Server.CntReqGetLastAdrBundle++

	elements := map[string]interface{}{"bLevelID": bLevelID, "seed": seed}
	err = errors.New("unknown error")
	defer render(c, gin.H{"payload": &data}, &status, &err)

	if err = parseBody(c, &elements); err != nil {
		return
	}

	bLevelID = elements["bLevelID"].(string)
	seed = elements["seed"].([]byte)

	adr, accountID, err = Server.GetLastAdrBundle(seed, bLevelID)

	data["address"] = adr
	data["accountID"] = accountID

	if err == nil {
		status = http.StatusAccepted
	} else {
		status = http.StatusNotFound
	}
}
