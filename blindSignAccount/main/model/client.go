package model

import (
	"blindSignAccount/main/config"
	"blindSignAccount/main/crypt"
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/btcsuite/btcutil/hdkeychain"
	"github.com/cryptoballot/rsablind"
	"io/ioutil"
	"math/rand"
	"strconv"
	"time"
)

type Client struct {
	ClientID  int
	Mnemonic  string
	Seed      []byte
	AccountID uint32
	AddressID uint32 // the last address id which was used
	Keys      []*hdkeychain.ExtendedKey
	//Bookings         []*Booking
	con              Connection
	BLevelToTokens   map[string]string
	BLevelToRecovery map[string]string // maps bonus level name to recovery Token
	BonusCodes       []*BonusCode
	BonusLevels      map[string]*BonusLevel
	flightList       map[int]*Flight
	RecoveryID       int               // used for generation of recoveryPK
	RecoveryPK       *ecdsa.PrivateKey // used for deterministic blinding of recovery tokens
}

const maxAdrID uint32 = 12
const maxAccountID uint32 = 100

func NewClient(clientID int, mnemonic string, recoveryID int) *Client {
	rand.Seed(int64(clientID))
	return NewClientWithAccountID(clientID, mnemonic, recoveryID, uint32(rand.Intn(int(maxAccountID))))
}

func NewClientWithAccountID(clientID int, mnemonic string, recoveryID int, accountID uint32) *Client {
	var err error
	client := &Client{ClientID: clientID, Mnemonic: mnemonic,
		BLevelToTokens:   map[string]string{},
		BLevelToRecovery: map[string]string{},
		BonusCodes:       []*BonusCode{},
		BonusLevels:      map[string]*BonusLevel{},
		flightList:       map[int]*Flight{},
	}

	client.AccountID = accountID
	client.AddressID = 0
	if client.Seed, client.Keys, err = crypt.GetWalletKeys(client.Mnemonic, client.AccountID, false); err != nil {
		return nil
	}

	// create the recoveryPK
	client.RecoveryID = recoveryID
	if err = client.generateRecoveryPK(); err != nil {
		return nil
	}

	return client
}

// Returns the client's account id
func (c *Client) GetAccountID() uint32 {
	return c.AccountID
}

// Sets a new connection for the given client
func (c *Client) SetConnection(connection Connection) {
	c.con = connection
}

// Checks whether a connection was set or not
func (c *Client) HasConnection() bool {
	return c.con != nil
}

func (c *Client) ReloadConfig(pathToConfig string) {
	if err := config.ReadConfigFile(pathToConfig); err != nil {
		panic(err)
	}
	ServerAddress = "http://" + config.GetConfigAddress()
}

// Registers a client at the server
// The client id will be overwritten
func (c *Client) Register() (clientID int, err error) {
	if clientID, err = c.con.Register(); err != nil {
		return
	}
	c.ClientID = clientID
	return
}

type ClientServerJson struct {
	Client *Client
	Server *Server
}

// Save the client's data to a debug file
func (c *Client) SaveDebugInfos() error {
	var server *Server
	var err error

	if server, err = c.con.GetDebugInfos(); err != nil {
		return err
	}

	// write client
	file, err := json.Marshal(c)
	if err != nil {
		return err
	}
	t := time.Now()
	err = ioutil.WriteFile(config.GetConfigResultDir()+strconv.Itoa(c.ClientID)+"_"+t.Format("2006_01_02_15_04_05")+"_client_debug.json", file, 0644)

	// write server
	file, err = json.Marshal(server)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(config.GetConfigResultDir()+strconv.Itoa(c.ClientID)+"_"+t.Format("2006_01_02_15_04_05")+"_server_debug.json", file, 0644)
	return err
}

// Fills the client's flight and bonus list with information by the server
func (c *Client) GetSystemInformation() error {
	flightList, bonusList, err := c.con.GetSystemInformation()
	if err != nil {
		return err
	}
	// fill all flights
	c.flightList = make(map[int]*Flight)
	for _, flight := range flightList {
		c.flightList[flight.ID] = flight
	}

	// fill all bonus levels
	c.BonusLevels = make(map[string]*BonusLevel)
	for _, bLevel := range bonusList {
		c.BonusLevels[bLevel.BonusID] = bLevel
	}
	return nil
}

// Create a new booking for given flight and for selected bonus level
func (c *Client) Booking(flightID int, bLevelID string) error {
	bLevel := c.BonusLevels[bLevelID]
	if bLevel == nil {
		return errors.New("unknown bonus level")
	}

	// send booking and receive a new Token
	token, err := c.con.SendBooking(c.ClientID, flightID, bLevelID)
	if err != nil {
		return err
	}

	if token == "" {
		return errors.New("no code received for booking")
	}

	// create blind Token
	blindBundle, err := crypt.CreateBlindBundle(bLevel.ActionVariants[ActionBooking].PublicKey)
	if err != nil {
		return err
	}

	var blindSigHex string
	if blindSigHex, err = c.con.GetBlindSignature(bLevelID, token, blindBundle.BlindToken, ActionBooking); err != nil {
		return err
	}
	if blindBundle.BlindSig, err = base64.URLEncoding.DecodeString(blindSigHex); err != nil {
		return err
	}
	signature := rsablind.Unblind(&bLevel.ActionVariants[ActionBooking].PublicKey, []byte(blindBundle.BlindSig), blindBundle.UnBlinder)

	code, err := c.con.GetBookingCode(bLevelID, blindBundle.HashValue, signature)
	if err != nil {
		return err
	}

	c.BonusCodes = append(c.BonusCodes, NewBonusCodeWithID(code, bLevel))
	return nil
}

// Access the server's bonus system with the client's codes
func (c *Client) AccessBonusSystem() error {
	var codes []string
	var adrBundle *crypt.AddressBundle

	// create codes
	for _, code := range c.BonusCodes {
		codes = append(codes, code.CodeID)
	}

	// create a new address
	adrBundle = &crypt.AddressBundle{Seed: c.Seed,
		AccountID: c.AccountID, AddressID: c.AddressID,
		Address: crypt.GetAddress(c.Keys[4], c.AddressID).String()}

	tokenMap, recoveryMap, err := c.con.AccessBonusSystem(codes, adrBundle)
	c.AddressID++
	if err != nil {
		return err
	}
	if len(tokenMap) == 0 {
		return errors.New("no tokens received")
	}

	if len(recoveryMap) == 0 {
		return errors.New("no recovery tokens received")
	}

	// update and map all tokens to the bonus levels
	for bLevel, token := range tokenMap {
		c.BLevelToTokens[bLevel] = token
	}

	// the same for recovery tokens
	for bLevel, recoveryToken := range recoveryMap {
		c.BLevelToRecovery[bLevel] = recoveryToken
	}
	return nil
}

func (c *Client) AdrUpdate(bLevel *BonusLevel) (token, recoveryToken string, blindBundle *crypt.BlindBundle, signature []byte, err error) {
	////////// step 1: Get blind Token and signature for address update ////////////
	blindBundle, signature, err = c.getSignatureForToken(bLevel, c.BLevelToTokens[bLevel.BonusID], ActionParticipate)
	if err != nil {
		return "", "", nil, nil, err
	}

	////////// step 2: Update address ////////////
	// create a new address
	adrBundle := &crypt.AddressBundle{Seed: c.Seed, AccountID: c.AccountID, AddressID: c.AddressID,
		Address: crypt.GetAddress(c.Keys[4], c.AddressID).String()}

	pkr, err := c.blindRecoveryToken(c.BLevelToRecovery[bLevel.BonusID])
	if err != nil {
		return "", "", nil, nil, err
	}
	token, recoveryToken, err = c.con.SetAddress(bLevel.BonusID, blindBundle.HashValue, signature, adrBundle, ActionParticipate, pkr)
	if err != nil {
		return "", "", nil, nil, err
	}
	return token, recoveryToken, blindBundle, signature, nil
}

func (c *Client) Participate(bLevelID string) (bonusData string, err error) {
	var participateToken string

	bLevel := c.BonusLevels[bLevelID]
	if bLevel == nil {
		return "", errors.New("bonus level does not exist")
	}

	//////////// step 1 & step 2: Get blind Token and signature for address update and update address ////////////
	var blindBundle *crypt.BlindBundle
	var signature []byte
	var pkr string
	participateToken, c.BLevelToRecovery[bLevelID], blindBundle, signature, err = c.AdrUpdate(bLevel)
	c.calcNextAddressID(c.AddressID)
	if err != nil {
		return "", err
	}

	////////// step 3: Get blind Token and signature for participation data ////////////
	blindBundle, signature, err = c.getSignatureForToken(bLevel, participateToken, ActionParticipate)
	if err != nil {
		return "", err
	}

	////////// step 4: get participation data ////////////
	if pkr, err = c.blindRecoveryToken(c.BLevelToRecovery[bLevelID]); err != nil {
		return "", err
	}
	c.BLevelToTokens[bLevelID], c.BLevelToRecovery[bLevelID], bonusData, err = c.con.Participate(bLevelID, blindBundle.HashValue, signature, pkr)
	if err != nil {
		return "", err
	}
	return bonusData, nil
}

func (c *Client) calcNextAddressID(currAddressID uint32) {
	c.AddressID++
	if c.AddressID >= maxAdrID {
		c.AccountID = uint32(rand.Intn(int(maxAccountID)))
		c.AddressID = 0
		_, c.Keys, _ = crypt.GetWalletKeys(c.Mnemonic, c.AccountID, false)
		c.generateRecoveryPK()
	}
}

// Requests a signature from the server for a given bonus level.
// The given Token is used for authorisation.
func (c *Client) getSignatureForToken(bLevel *BonusLevel, token string, action int) (blindBundle *crypt.BlindBundle, signature []byte, err error) {
	var blindSigHex string

	blindBundle, err = crypt.CreateBlindBundle(bLevel.ActionVariants[action].PublicKey)
	if err != nil {
		return nil, nil, err
	}
	if blindSigHex, err = c.con.GetBlindSignature(bLevel.BonusID, token, blindBundle.BlindToken, action); err != nil {
		return nil, nil, err
	}
	if blindBundle.BlindSig, err = base64.URLEncoding.DecodeString(blindSigHex); err != nil {
		return nil, nil, err
	}
	signature = rsablind.Unblind(&bLevel.ActionVariants[action].PublicKey, []byte(blindBundle.BlindSig), blindBundle.UnBlinder)
	return blindBundle, signature, nil
}

// Blinds a given recovery Token
func (c *Client) blindRecoveryToken(recoveryToken string) (pkr string, err error) {
	hash := sha256.Sum256([]byte(recoveryToken))
	r, _, err := ecdsa.Sign(crypt.NewReader(c.RecoveryID), c.RecoveryPK, hash[:])
	if err != nil {
		return "", err
	}

	return r.String(), nil
}

// Generate a private key which is used for blinding of
// recovery tokens
func (c *Client) generateRecoveryPK() error {
	if len(c.Keys) != 5 {
		return errors.New("wrong key setup")
	}

	c.RecoveryPK = crypt.GetPrivateKey(c.Keys[4], uint32(c.RecoveryID)).ToECDSA()
	return nil
}

// Tries to restore access data for given bonus level id
// True is returned if successful, false otherwise
func (c *Client) Restore(bLevelID string) (string, error) {
	var currAddressID, curAccountID uint32 = 0, 0
	var accessed bool
	var err error
	var bData string
	var lastAdr string

	// ask the server for the last used address
	lastAdr, curAccountID, err = c.con.GetLastAdrBdl(c.Seed, bLevelID)
	if err != nil {
		return "", err
	}
	// update the keys under the new account node
	c.AccountID = curAccountID
	_, c.Keys, _ = crypt.GetWalletKeys(c.Mnemonic, c.AccountID, false)
	if err = c.generateRecoveryPK(); err != nil {
		return "", err
	}

	// try to search the correct address id, which was recently used
	for currAddressID = c.getAdrIDFromString(lastAdr); currAddressID < maxAdrID; currAddressID++ {
		if accessed, bData, err = c.RestoreWithAddress(bLevelID, currAddressID); err != nil {
			return "", err
		}
		if accessed {
			// increase the address id since it must point to the next address id
			c.AddressID = currAddressID + 1
			return bData, nil
		}
	}
	return "", errors.New("no restore possible")
}

// Tries to restore access data for given bonus level id and with given address id.
// True is returned if successful, false otherwise
func (c *Client) RestoreWithAddress(bLevelID string, addressID uint32) (bool, string, error) {
	adr := crypt.GetAddress(c.Keys[4], addressID).String()
	adrBdl := &crypt.AddressBundle{Seed: c.Seed, Address: adr, AddressID: addressID, AccountID: c.AccountID}
	status, token, _ := c.con.CanBeUsedForRecovery(bLevelID, adrBdl) // error handling is not needed here
	switch status {
	case RecoveryTestAfterAccess:
		// a recoveryTest has to be executed for receiving the RecoveryToken
		_, foundRecovery, bData, err := c.con.RecoveryTest(bLevelID, "", "", adrBdl)
		if err != nil {
			return false, "", err
		}
		c.BLevelToTokens[bLevelID] = token
		c.BLevelToRecovery[bLevelID] = foundRecovery
		return true, bData, nil
	case Failure:
		return false, "", nil
	default:
		// the received Token is a recovery Token => calculate pkr
		pkr, err := c.blindRecoveryToken(token)
		if err != nil {
			return false, "", err
		}
		// execute a recovery test
		foundToken, foundRecovery, bData, err := c.con.RecoveryTest(bLevelID, token, pkr, adrBdl)
		if err != nil {
			return false, "", err
		}
		c.BLevelToTokens[bLevelID] = foundToken
		c.BLevelToRecovery[bLevelID] = foundRecovery
		return true, bData, nil
	}
}

func (c *Client) RestoreFromMnemonic(clientID int, mnemonic string, recoveryID int) (map[string]string, error) {
	var err error
	var bData = map[string]string{}

	// ask for server information
	if err = c.GetSystemInformation(); err != nil {
		return bData, err
	}

	// search with account id within interval [accountID,maxAccountID)
	bData, err = c.restoreWithAccountIDBounds(clientID, recoveryID, mnemonic)
	return bData, err

}

// Restores a client completely from the mnemonic, the recovery id and the client id
// The connection has to be set up before calling this function!
func (c *Client) restoreWithAccountIDBounds(clientID, recoveryID int, mnemonic string) (map[string]string, error) {
	var err error
	var success bool
	var highestAdrID uint32
	var bData = map[string]string{}

	// init the account id to 0
	//for c.AccountID = lBound; c.AccountID < uBound; c.AccountID++ {
	// create a new helper client
	c.AccountID = 0
	newClient := NewClientWithAccountID(clientID, mnemonic, recoveryID, c.AccountID)
	// assign recalculated values
	c.RecoveryID = newClient.RecoveryID
	c.Mnemonic = newClient.Mnemonic
	c.Seed = newClient.Seed
	c.Keys = newClient.Keys
	c.RecoveryPK = newClient.RecoveryPK
	// init the address to 0
	c.AddressID = 0
	for bLevel := range c.BonusLevels {
		if bData[bLevel], err = c.Restore(bLevel); err == nil {
			// restoration succeeded for one bonus level
			success = true
			if highestAdrID < c.AddressID {
				highestAdrID = c.AddressID
			}
		}
	}
	if success {
		c.AddressID = highestAdrID
		return bData, nil
	}
	//}
	return bData, err
}

// Finds the address id for a string representation of an address
func (c *Client) getAdrIDFromString(adr string) (addressID uint32) {
	for addressID = uint32(0); addressID < maxAdrID; addressID++ {
		nextAdr := crypt.GetAddress(c.Keys[4], addressID).String()
		if nextAdr == adr {
			return
		}
	}
	return
}
