package handlers

import (
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
)

func SendBooking(c *gin.Context) {
	var status = http.StatusBadRequest
	var err error
	var token = make(map[string]string, 0)
	var customerID, flightID int
	var bLevelID string

	Server.CntReqSendBooking++

	err = errors.New("unknown error")
	elements := map[string]interface{}{"customerID": customerID, "flightID": flightID, "bLevelID": bLevelID}

	defer render(c, gin.H{"payload": &token}, &status, &err)

	if err = parseBody(c, &elements); err != nil {
		return
	}
	customerID = elements["customerID"].(int)
	flightID = elements["flightID"].(int)
	bLevelID = elements["bLevelID"].(string)

	if token["token"], err = Server.Booking(flightID, customerID, bLevelID); err != nil {
		return
	}

	status = http.StatusAccepted
}

func HdlGetBookingCode(c *gin.Context) {
	var status = http.StatusBadRequest
	var err error
	var data = make(map[string]string, 0)
	var hashValue, signature []byte
	var bLevelID string

	Server.CntReqGetBookingCode++

	err = errors.New("unknown error")
	elements := map[string]interface{}{"hashValue": hashValue, "signature": signature, "bLevelID": bLevelID}
	defer render(c, gin.H{"payload": &data}, &status, &err)

	if err = parseBody(c, &elements); err != nil {
		return
	}

	bLevelID = elements["bLevelID"].(string)
	hashValue = elements["hashValue"].([]byte)
	signature = elements["signature"].([]byte)

	if data["code"], err = Server.GetBookingCode(bLevelID, hashValue, signature); err != nil {
		return
	}
	status = http.StatusAccepted
}
