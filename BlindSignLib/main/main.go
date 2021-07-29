package main

import "C"
import (
	"blindSignAccount/main/model"
	"log"
	"strconv"
	"strings"
	"sync"
)

var clients map[int]*model.Client
var registerClient *model.Client
var mtx sync.Mutex

var wordList = [...]string{"air", "blind", "car", "day", "egg", "faith", "gain", "hat", "ice", "joy", "key"}

const mnemonicLength int = 24

var mnemonicDefault []string

const nrOfClients int = 50000

func init() {
	// a dummy client for register other clients is needed
	registerClient = model.NewClient(0, chooseMnemonic(0), 3)
	clients = make(map[int]*model.Client, nrOfClients)
	mnemonicDefault = make([]string, mnemonicLength)
	for i := 0; i < mnemonicLength; i++ {
		mnemonicDefault[i] = wordList[0]
	}
}

// chooses a mnemonic for given depending on given length
func chooseMnemonic(clientsLengths int) string {
	digits := giveDigits(clientsLengths)
	mnemonic := make([]string, mnemonicLength)
	copy(mnemonic, mnemonicDefault)
	digitsLength := len(digits)
	for i := digitsLength; i > 0; i-- {
		mnemonic[mnemonicLength-i] = wordList[digits[digitsLength-i]]
	}
	return strings.Join(mnemonic, " ")
}

// returns all digits of a given number
func giveDigits(clientLength int) []int {
	clientLengthStr := strconv.Itoa(clientLength)
	var digits = make([]int, 0, len(clientLengthStr))
	for _, digitChar := range clientLengthStr {
		digitInt, _ := strconv.Atoi(string(digitChar))
		digits = append(digits, digitInt)
	}
	return digits
}

//export AddClient
func AddClient(configFile string) int {
	mtx.Lock()
	defer mtx.Unlock()

	if !registerClient.HasConnection() {
		registerClient.ReloadConfig(configFile)
		registerClient.SetConnection(model.NewRestConnection())
	}
	clientID, err := registerClient.Register()
	if err != nil {
		log.Fatal(err)
	}

	client := model.NewClient(clientID, chooseMnemonic(clientID), 3)
	client.ReloadConfig(configFile)
	client.SetConnection(model.NewRestConnection())
	clients[clientID] = client
	return clientID
}

//export GetSystemInfo
func GetSystemInfo(clientID int) *C.char {
	if err := clients[clientID].GetSystemInformation(); err != nil {
		return C.CString(err.Error())
	}
	return C.CString("")
}

//export Booking
func Booking(clientID, flightID int, bLevelID string) (error *C.char) {
	if err := clients[clientID].Booking(flightID, bLevelID); err != nil {
		return C.CString(err.Error())
	}
	return C.CString("")
}

//export AccessBonusSystem
func AccessBonusSystem(clientID int) (error *C.char) {
	if err := clients[clientID].AccessBonusSystem(); err != nil {
		return C.CString(err.Error())
	}
	return C.CString("")
}

//export Participate
func Participate(clientID int, bLevelID string) (bonusData *C.char, errMsg *C.char) {
	var err error
	var bData string
	if bData, err = clients[clientID].Participate(bLevelID); err != nil {
		return C.CString(bData), C.CString(err.Error())
	}
	return C.CString(bData), C.CString("")
}

//export Restore
func Restore(clientID int) (*C.char, *C.char) {
	var bData map[string]string
	var bonusData string
	var err error

	if bData, err = clients[clientID].RestoreFromMnemonic(clientID, chooseMnemonic(clientID), 3); err != nil {
		return C.CString(""), C.CString(err.Error())
	}
	bonusData = "[ "
	for key, val := range bData {
		bonusData += key + ": " + val + " "
	}
	bonusData += "]"
	return C.CString(bonusData), C.CString("")
}

//export SaveDebugInfos
func SaveDebugInfos(clientID int) *C.char {
	var err error
	if err = clients[clientID].SaveDebugInfos(); err != nil {
		return C.CString(err.Error())
	}
	return C.CString("")
}

//export GetAccountID
func GetAccountID(clientID int) int {
	return int(clients[clientID].GetAccountID())
}

func main() {
	//go build -o blindSignLib.so -buildmode=c-shared main/main.go
}
