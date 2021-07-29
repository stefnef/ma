package model

import (
	"blindSignAccount/main/crypt"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/btcsuite/btcutil"
	"github.com/cryptoballot/rsablind"
	"io/ioutil"
	"sort"
	"strconv"
	"sync"
)

type Server struct {
	// maps bonus level ids to corresponding bonus levels
	BonusList map[string]*BonusLevel
	// an ordered list of bonus levels:
	// levels with smaller indexes correspond to levels with higher priority
	Hierarchy []*BonusLevel
	// maps code ids to corresponding bonus codes
	BonusCodes map[string]*BonusCode
	// a list of available flights
	flightMap map[int]*Flight
	// a list of known clients
	ClientIDs []int

	// sync
	Mux sync.Mutex

	// statistic
	CntReqSendBooking, CntReqGetBookingCode,
	CntReqGetSystemInformation,
	CntReqBlindSignature, CntReqSetAddress,
	CntReqAccessBonusSystem, CntReqParticipate,
	CntReqCanBesUsedForRecovery, CntReqRecoveryTest, CntReqGetLastAdrBundle,
	CntReqRegister, CntReqExit, CntReqStatistic, CntReqReset int
}

const lengthBonusCode = 64

// Creates a new server
func NewServer() *Server {
	s := &Server{BonusList: GetDefaultHBLS(),
		BonusCodes: map[string]*BonusCode{},
		flightMap:  GetDefaultFlightList(),
		ClientIDs:  []int{}}

	// check out the priorities
	priorities := map[*BonusLevel]int{}
	for _, level := range s.BonusList {
		priorities[level] = len(level.LowerLevels)
		s.Hierarchy = append(s.Hierarchy, level)
	}
	// sort the hierarchy slice by calculated priorities
	sort.Slice(s.Hierarchy, func(i, j int) bool {
		return priorities[s.Hierarchy[i]] > priorities[s.Hierarchy[j]]
	})

	return s
}

// Loads a server from a debug dump file
func NewServerFromFile(fileName string) (*Server, error) {
	var server = NewServer()
	rawVal, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, err
	}

	if err = json.Unmarshal(rawVal, server); err != nil {
		return nil, err
	}

	// correct the pointer values for bonus codes
	for _, bCode := range server.BonusCodes {
		if bCode != nil {
			bCode.ValidFor = server.BonusList[bCode.ValidFor.BonusID]
		}
	}
	// correct the pointer values for the hierarchy
	for idx, bLevel := range server.Hierarchy {
		server.Hierarchy[idx] = server.BonusList[bLevel.BonusID]
	}

	return server, nil
}

// generates and adds a new bonus code for the bonus level with given id
func (s *Server) GenerateNewBonusCode(bonusLevelID string) *BonusCode {
	code := NewBonusCode(s.BonusList[bonusLevelID])
	s.BonusCodes[code.CodeID] = code
	return code
}

// Handles a new booking of a given customer for a given bonus level.
// If successful a new Token is generated. This Token can be used for generating a new code
func (s *Server) Booking(flightID, customerID int, bonusLevelID string) (string, error) {
	// sync
	s.Mux.Lock()
	defer s.Mux.Unlock()

	// find bonus level
	bLevel := s.BonusList[bonusLevelID]
	if bLevel == nil {
		return "", errors.New("bonus level " + bonusLevelID + " does not exist")
	}
	// find flight
	flight := s.flightMap[flightID]
	if flight == nil {
		return "", errors.New("flight with id " + strconv.Itoa(flightID) + " does not exist")
	}
	// generate a code
	token := crypt.GenerateToken()
	err := bLevel.addValidToken(token, ActionBooking)
	if err != nil {
		return "", nil
	}

	// create a new booking
	flight.AddBooking(customerID, bLevel)

	return token, nil
}

// Generates a new code which is valid for a requested bonus level.
// The code will be generated if and only if the hash value and the signature are fitting together
func (s *Server) GetBookingCode(bLevelID string, hashValue, signature []byte) (string, error) {
	// sync
	s.Mux.Lock()
	defer s.Mux.Unlock()

	// find bonus level
	bLevel := s.BonusList[bLevelID]
	if bLevel == nil {
		return "", errors.New("bonus level '" + bLevelID + "' does not exist")
	}

	// check that the hash value fits the signature
	if len(hashValue) == 0 || len(signature) == 0 {
		return "", errors.New("hash value or signature is empty")
	}
	if err := rsablind.VerifyBlindSignature(&bLevel.ActionVariants[ActionBooking].SkKey.PublicKey, hashValue, signature); err != nil {
		return "", err
	}

	bCode := s.GenerateNewBonusCode(bLevelID)
	if bCode.CodeID == "" {
		return "", errors.New("no code generated")
	}
	return bCode.CodeID, nil
}

// Checks if codes are valid and if bonus system can be accessed
// If successful a Token is generated, returned and linked to the
// given seed.
func (s *Server) AccessBonusSystem(codes []string, adrBundle *crypt.AddressBundle) (tokens, recoveryTokens map[string]string, err error) {
	// sync
	s.Mux.Lock()
	defer s.Mux.Unlock()

	tokens = make(map[string]string, len(s.BonusCodes))
	recoveryTokens = make(map[string]string, len(s.BonusCodes))
	var calcAddress *btcutil.AddressPubKeyHash

	// check that seed and address fit together
	calcAddress, err = crypt.GetAddressFromSeed(adrBundle, false)
	if err != nil {
		return
	}
	if calcAddress.String() != adrBundle.Address {
		err = errors.New("address does not fit to given seed")
		return
	}

	validLevels := s.verifyCodes(codes)
	if len(validLevels) == 0 {
		err = errors.New("no bonus level accessible")
		return
	}

	// generate a valid Token for every valid bonus level
	for _, bLevel := range validLevels {
		bLevel.ActionVariants[ActionParticipate].Mux.Lock()
		token := crypt.GenerateToken()
		bLevel.ActionVariants[ActionParticipate].TokenToSeed[token] = hex.EncodeToString(adrBundle.Seed)
		SaveWrite(StatTokenToSeed, bLevel.ActionVariants[ActionParticipate])
		bLevel.ActionVariants[ActionParticipate].SeedToAddress[hex.EncodeToString(adrBundle.Seed)] = adrBundle.Address
		SaveWrite(StatSeedToAddress, bLevel.ActionVariants[ActionParticipate])
		bLevel.ActionVariants[ActionParticipate].SeedToAccountID[hex.EncodeToString(adrBundle.Seed)] = adrBundle.AccountID
		SaveWrite(StatSeedToAccountID, bLevel.ActionVariants[ActionParticipate])
		bLevel.ActionVariants[ActionParticipate].AddressToToken[adrBundle.Address] = token
		SaveWrite(StatAddressToToken, bLevel.ActionVariants[ActionParticipate])
		tokens[bLevel.BonusID] = token
		if err = bLevel.addValidToken(token, ActionParticipate); err != nil { //mark the Token as valid
			bLevel.ActionVariants[ActionParticipate].Mux.Unlock()
			return
		}

		// generate a recovery Token
		recoveryToken := crypt.GenerateToken()
		bLevel.ActionVariants[ActionParticipate].AddressToRecovery[adrBundle.Address] = recoveryToken
		SaveWrite(StatAddressToRecovery, bLevel.ActionVariants[ActionParticipate])
		recoveryTokens[bLevel.BonusID] = recoveryToken

		// mark the address as the one which was used for accessing
		bLevel.ActionVariants[ActionParticipate].SeedToAccessAdr[hex.EncodeToString(adrBundle.Seed)] = adrBundle.Address
		SaveWrite(StatSeedToAccessAdr, bLevel.ActionVariants[ActionParticipate])

		bLevel.ActionVariants[ActionParticipate].Mux.Unlock()
	}
	return
}

// Checks if given codes are valid and receive list of bonus levels for which
// they are valid
func (s *Server) verifyCodes(codes []string) (accessible []*BonusLevel) {
	// initialize a map of valid levels
	validLevels := map[*BonusLevel]int{}
	for _, level := range s.BonusList {
		validLevels[level] = 0
	}

	// run threw all codes and check for which levels they are valid
	for _, code := range codes {
		if bCode := s.BonusCodes[code]; bCode != nil {
			validForCode := bCode.GetValidBonusLevels()
			if len(validForCode) == 0 {
				continue
			}
			for _, level := range validForCode {
				validLevels[level]++
			}
		}
	}

	// check that the minimal number of codes per level was reached
	// The levels are ordered by hierarchy priority
	for _, level := range s.Hierarchy {
		if validLevels[level] >= level.MinNrCodes {
			// a valid level was found and can be accessed
			// all lower levels can be accessed as well
			accessible = append(level.LowerLevels, level)
			// mark all codes as used
			for _, code := range codes {
				s.BonusCodes[code] = nil
			}
			return accessible
		}
	}

	return accessible
}

// Calculates a blind signature for a given blind Token.
// The signature is calculated if an other given Token is valid
func (s *Server) GetBlindSignature(bLevelID, token string, blindToken []byte, action int) (string, error) {
	// sync
	s.Mux.Lock()
	defer s.Mux.Unlock()

	bLevel := s.getBonusLevel(bLevelID)
	if bLevel == nil {
		return "", errors.New("no level known with given id")
	}
	// check that the Token is valid
	isValid := bLevel.isTokenValid(token, action)
	if !isValid {
		return "", errors.New("Token is not valid")
	}
	blindSig, err := rsablind.BlindSign(bLevel.ActionVariants[action].SkKey, blindToken)
	bLevel.markTokenAsUsed(token, action)
	return base64.URLEncoding.EncodeToString(blindSig), err
}

// sets a new address if signature of given the hash value is valid to the given signature
// pkr - blinded recovery Token
func (s *Server) SetAddress(bLevelID string, hashed, sig []byte, adrBundle *crypt.AddressBundle, action int, pkr string) (token, recoveryToken string, err error) {
	// sync
	s.Mux.Lock()
	defer s.Mux.Unlock()

	var calcAddress *btcutil.AddressPubKeyHash

	bLevel := s.getBonusLevel(bLevelID)
	if bLevel == nil {
		return "", "", errors.New("no level known with given id")
	}
	// check that the seed is known
	_, ok := bLevel.ActionVariants[action].SeedToAddress[hex.EncodeToString(adrBundle.Seed)]
	SaveRead(StatSeedToAddress, bLevel.ActionVariants[action])
	if !ok {
		return "", "", errors.New("seed unknown")
	}

	if len(hashed) == 0 || len(sig) == 0 {
		return "", "", errors.New("hash value or signature is empty")
	}

	// check that the hash value fits the signature
	if err = rsablind.VerifyBlindSignature(&bLevel.ActionVariants[action].SkKey.PublicKey, hashed, sig); err != nil {
		return "", "", err
	}

	// check that the address fits to the seed
	calcAddress, err = crypt.GetAddressFromSeed(adrBundle, false)
	if err != nil {
		return "", "", err
	}
	if calcAddress.String() != adrBundle.Address {
		return "", "", errors.New("address does not fit to given seed")
	}

	// check that the address was not used before
	SaveRead(StatAddressToToken, bLevel.ActionVariants[action])
	if _, found := bLevel.ActionVariants[action].AddressToToken[adrBundle.Address]; found {
		return "", "", errors.New("address is not valid. Already used")
	}

	token = crypt.GenerateToken()
	recoveryToken = crypt.GenerateToken()

	// refresh maps
	bLevel.refreshMaps(recoveryToken, token, adrBundle.Address, adrBundle.Seed, action, adrBundle.AccountID)
	if err = bLevel.addValidToken(token, action); err != nil {
		return "", "", err
	}
	// the pkr has to be mapped to the address: Needed for recovery test
	SaveWrite(StatPkrToAdrUpd, bLevel.ActionVariants[action])
	bLevel.ActionVariants[action].PkrToAdrUpd[pkr] = adrBundle.Address
	return token, recoveryToken, nil
}

// Checks if participation action is legal.
// pkr - The blinded recovery Token
// A new Token is generated in case of success.
func (s *Server) Participate(bLevelID string, hashed, sig []byte, pkr string) (token, recoveryToken, bonusData string, err error) {
	// sync
	s.Mux.Lock()
	defer s.Mux.Unlock()

	bLevel := s.getBonusLevel(bLevelID)
	if bLevel == nil {
		return "", "", "", errors.New("no level known with given id")
	}

	// check that the hash value fits the signature
	if len(hashed) == 0 || len(sig) == 0 {
		return "", "", "", errors.New("hash value or signature is empty")
	}
	if err = rsablind.VerifyBlindSignature(&bLevel.ActionVariants[ActionParticipate].SkKey.PublicKey, hashed, sig); err != nil {
		return "", "", "", err
	}

	token = crypt.GenerateToken()
	if err = bLevel.addValidToken(token, ActionParticipate); err != nil {
		return "", "", "", err
	}

	// generate a new recovery Token
	recoveryToken = crypt.GenerateToken()

	// generate bonus data
	bonusData = bLevel.GetBonusData()

	// map the pkr to the bonus data
	bonusDataPair := &bonusDataPair{Token: token, RecoveryToken: recoveryToken, BonusData: bonusData}
	bLevel.ActionVariants[ActionParticipate].PkrToBonusData[pkr] = bonusDataPair
	SaveWrite(StatPkrToBonusData, bLevel.ActionVariants[ActionParticipate])

	return token, recoveryToken, bonusData, nil
}

// Returns a pointer to the requested bonus level object
func (s *Server) getBonusLevel(levelID string) (bLevel *BonusLevel) {
	bLevel, _ = s.BonusList[levelID]
	return bLevel
}

// Returns all information about the server's available flights and
// bonus levels
func (s *Server) GetSystemInformation() ([]*Flight, []*BonusLevel, error) {
	// sync
	s.Mux.Lock()
	defer s.Mux.Unlock()

	flights := make([]*Flight, 0)
	bLevels := make([]*BonusLevel, 0)
	var err error

	// get flight information
	for fID := range s.flightMap {
		// only IDs are needed
		flights = append(flights, &Flight{ID: fID, Bookings: []*Booking{}})
	}
	sort.Slice(flights, func(i, j int) bool { return flights[i].ID < flights[j].ID })

	// get bonus level information
	for _, bLevel := range s.BonusList {
		bLevels = append(bLevels, bLevel.CopyPublic())
	}
	if len(flights) == 0 || len(bLevels) == 0 {
		err = errors.New("unloaded server")
	}
	return flights, bLevels, err
}

// Checks if a given address was set for the last address update. If it was used, then the recovery Token will be
// returned as well.
func (s *Server) CanBeUsedForRecovery(bLevelID string, adrBdl *crypt.AddressBundle) (status RecoveryStatus, token string, err error) {
	bLevelParticipate := s.getBonusLevel(bLevelID).ActionVariants[ActionParticipate]

	bLevelParticipate.Mux.Lock()
	defer bLevelParticipate.Mux.Unlock()
	adr := bLevelParticipate.SeedToAddress[hex.EncodeToString(adrBdl.Seed)]
	SaveRead(StatSeedToAddress, bLevelParticipate)
	//bLevelParticipate.Mux.Unlock()
	if adr == "" || adr != adrBdl.Address {
		return Failure, "", errors.New("address invalid")
	}

	// the address was used
	// check if the Token was used for a blind signature
	//bLevelParticipate.Mux.Lock()
	token = bLevelParticipate.AddressToToken[adr]
	SaveRead(StatAddressToToken, bLevelParticipate)
	//bLevelParticipate.Mux.Unlock()
	if s.getBonusLevel(bLevelID).isTokenValid(token, ActionParticipate) {
		// the Token is still valid and was not used for participation
		// => Special case: 1st address update
		//bLevelParticipate.Mux.Lock()
		//defer bLevelParticipate.Mux.Unlock()
		SaveRead(StatSeedToAccessAdr, bLevelParticipate)
		if bLevelParticipate.SeedToAccessAdr[hex.EncodeToString(adrBdl.Seed)] == adr {
			return RecoveryTestAfterAccess, token, nil
		}
		// => Use the 2nd last address for address update
		SaveRead(StatPenultimateAdr, bLevelParticipate)
		penultimateAdr := bLevelParticipate.PenultimateAdr[hex.EncodeToString(adrBdl.Seed)]
		if penultimateAdr == "" {
			return Failure, "", errors.New("no penultimate address found")
		}
		token = bLevelParticipate.AddressToRecovery[penultimateAdr]
		SaveRead(StatAddressToRecovery, bLevelParticipate)
		// check if the penultimate address is equal to the address used for access
		SaveRead(StatSeedToAccessAdr, bLevelParticipate)
		if penultimateAdr == bLevelParticipate.SeedToAccessAdr[hex.EncodeToString(adrBdl.Seed)] {
			return RecoveryTestAfterFirstAdrUpd, token, nil
		} else {
			// A connection has to be proven between that 2nd last address and the corresponding participation
			return RecoveryTestPenultimateAdr, token, nil
		}

	} else {
		// the Token was used for a participation step
		// => search the recovery Token and send it back
		//bLevelParticipate.Mux.Lock()
		token = bLevelParticipate.AddressToRecovery[adr]
		SaveRead(StatAddressToRecovery, bLevelParticipate)
		//bLevelParticipate.Mux.Unlock()
		return RecoveryTest, token, nil
	}
}

// Checks if the given address can be used for address update and if the RecoveryToken and the pkr are valid
func (s *Server) RecoveryTest(bLevelID, recoveryToken, pkr string, adrBdl *crypt.AddressBundle) (token, foundRecoveryToken, bonusData string, err error) {
	// sync
	s.Mux.Lock()
	defer s.Mux.Unlock()

	var adr string
	var found bool
	var bData *bonusDataPair
	var status RecoveryStatus

	bAction := s.BonusList[bLevelID].ActionVariants[ActionParticipate]

	// the check has to be redone
	status, _, err = s.CanBeUsedForRecovery(bLevelID, adrBdl)
	if err != nil {
		return
	}
	switch status {
	case RecoveryTestAfterAccess:
		foundRecoveryToken = bAction.AddressToRecovery[adrBdl.Address]
		SaveRead(StatAddressToRecovery, bAction)
	case Failure:
		err = errors.New("address bundle cannot be used for recovery")
	case RecoveryTestAfterFirstAdrUpd:
		// check the connection <access,recToken> --> <pkr, adrUpdate>
		SaveRead(StatPkrToAdrUpd, bAction)
		if adr, found = bAction.PkrToAdrUpd[pkr]; found != true {
			err = errors.New("pkr does not match to any address update step")
			return
		}
		if adr != adrBdl.Address {
			err = errors.New("the pkr does not match with the given address")
			return
		}
		token = bAction.AddressToToken[adrBdl.Address]
		SaveRead(StatAddressToToken, bAction)
		foundRecoveryToken = bAction.AddressToRecovery[string(adrBdl.Address)]
		SaveRead(StatAddressToRecovery, bAction)
	case RecoveryTest:
		SaveRead(StatPkrToBonusData, bAction)
		if bData, found = bAction.PkrToBonusData[pkr]; found != true {
			err = errors.New("pkr does not match to any participation step")
			return
		}
		token = bData.Token
		foundRecoveryToken = bData.RecoveryToken
		bonusData = bData.BonusData
	case RecoveryTestPenultimateAdr:
		// check that there exists a participation step with given pkr
		SaveRead(StatPkrToBonusData, bAction)
		if bData, found = bAction.PkrToBonusData[pkr]; found != true {
			err = errors.New("pkr does not match to any participation step")
			return
		}
		// step was found => search Token and recovery Token of address update step
		// that was executed after the participation step
		token = bAction.AddressToToken[adrBdl.Address]
		SaveRead(StatAddressToToken, bAction)
		foundRecoveryToken = bAction.AddressToRecovery[adrBdl.Address]
		SaveRead(StatAddressToRecovery, bAction)
		// show bonus data of the participation step
		bonusData = bData.BonusData
	default:
		err = errors.New("unknown status")
	}

	return
}

func (s *Server) Reset() {
	// sync
	s.Mux.Lock()
	defer s.Mux.Unlock()

	sReset := NewServer()
	s.BonusList = sReset.BonusList
	s.BonusCodes = sReset.BonusCodes
	s.flightMap = sReset.flightMap
	s.Hierarchy = sReset.Hierarchy
	s.ClientIDs = sReset.ClientIDs

	// reset the statistic also
	s.CntReqSendBooking = 0
	s.CntReqGetBookingCode = 0
	s.CntReqGetSystemInformation = 0
	s.CntReqBlindSignature = 0
	s.CntReqSetAddress = 0
	s.CntReqAccessBonusSystem = 0
	s.CntReqParticipate = 0
	s.CntReqCanBesUsedForRecovery = 0
	s.CntReqRecoveryTest = 0
	s.CntReqRegister = 0
	s.CntReqExit = 0
	s.CntReqStatistic = 0
	s.CntReqReset = 0
}

func (s *Server) Register() (clientID int, err error) {
	s.Mux.Lock()
	defer s.Mux.Unlock()

	clientID = len(s.ClientIDs)
	s.ClientIDs = append(s.ClientIDs, clientID)
	return
}

// Returns a statistical summary for all action variants of
// the server's bonus levels
func (s *Server) GetStatisticSummary() *StatisticSummary {
	var stat StatisticSummary
	stat.BLevelToSummary = make(map[string][]StatisticSummaryTuple)

	for bLevelName, bLevel := range s.BonusList {
		for _, variant := range bLevel.ActionVariants {
			var statTuple StatisticSummaryTuple
			statTuple.BonusActionVariant = variant.GetName()
			statTuple.Statistic = variant.Statistic
			stat.BLevelToSummary[bLevelName] = append(stat.BLevelToSummary[bLevelName], statTuple)
		}
	}

	stat.CntReqSendBooking = s.CntReqSendBooking
	stat.CntReqGetBookingCode = s.CntReqGetBookingCode
	stat.CntReqGetSystemInformation = s.CntReqGetSystemInformation
	stat.CntReqBlindSignature = s.CntReqBlindSignature
	stat.CntReqSetAddress = s.CntReqSetAddress
	stat.CntReqAccessBonusSystem = s.CntReqAccessBonusSystem
	stat.CntReqParticipate = s.CntReqParticipate
	stat.CntReqCanBesUsedForRecovery = s.CntReqCanBesUsedForRecovery
	stat.CntReqRecoveryTest = s.CntReqRecoveryTest
	stat.CntReqRegister = s.CntReqRegister
	stat.CntReqExit = s.CntReqExit
	stat.CntReqStatistic = s.CntReqStatistic
	stat.CntReqReset = s.CntReqReset
	stat.CntReqGetLastAdrBundle = s.CntReqGetLastAdrBundle

	return &stat
}

func (s *Server) GetLastAdrBundle(seed []byte, bLevelID string) (adr string, accountID uint32, err error) {
	var found bool
	bLevelParticipate := s.getBonusLevel(bLevelID).ActionVariants[ActionParticipate]

	bLevelParticipate.Mux.Lock()
	defer bLevelParticipate.Mux.Unlock()
	if adr, found = bLevelParticipate.SeedToAddress[hex.EncodeToString(seed)]; !found {
		err = errors.New("not found")
		return
	}
	if accountID, found = bLevelParticipate.SeedToAccountID[hex.EncodeToString(seed)]; !found {
		err = errors.New("unknown error. Account id not found")
		return
	}
	SaveRead(StatSeedToAddress, bLevelParticipate)
	SaveRead(StatSeedToAccountID, bLevelParticipate)
	return adr, accountID, nil
}
