package handlers

import (
	"blindSignAccount/main/crypt"
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
)

func HdlAccessBonusSystem(c *gin.Context) {
	var status = http.StatusBadRequest
	var codes []string
	var tokens, recoveryTokens map[string]string
	var adrBundle *crypt.AddressBundle
	var err error
	var data = make(map[string]interface{}, 0)

	Server.CntReqAccessBonusSystem++

	elements := map[string]interface{}{"codes": codes, "adrBundle": adrBundle}
	err = errors.New("unknown error")
	defer render(c, gin.H{"payload": &data}, &status, &err)

	if err = parseBody(c, &elements); err != nil {
		return
	}

	codes = elements["codes"].([]string)
	adrBundle = elements["adrBundle"].(*crypt.AddressBundle)

	if tokens, recoveryTokens, err = Server.AccessBonusSystem(codes, adrBundle); err != nil {
		return
	}

	data["tokens"] = tokens
	data["recoveryTokens"] = recoveryTokens

	status = http.StatusAccepted
}

func HdlParticipate(c *gin.Context) {
	var status = http.StatusBadRequest
	var bLevelID, pkr, token, recoveryToken, bonusData string
	var hashValue, signature []byte
	var err error
	var data = make(map[string]interface{}, 0)

	Server.CntReqParticipate++

	elements := map[string]interface{}{"bLevelID": bLevelID, "hashValue": hashValue, "signature": signature, "pkr": pkr}
	err = errors.New("unknown error")
	defer render(c, gin.H{"payload": &data}, &status, &err)

	if err = parseBody(c, &elements); err != nil {
		return
	}

	bLevelID = elements["bLevelID"].(string)
	hashValue = elements["hashValue"].([]byte)
	signature = elements["signature"].([]byte)
	pkr = elements["pkr"].(string)

	if token, recoveryToken, bonusData, err = Server.Participate(bLevelID, hashValue, signature, pkr); err != nil {
		return
	}

	data["token"] = token
	data["recoveryToken"] = recoveryToken
	data["bonusData"] = bonusData

	status = http.StatusAccepted
}
