package handlers

import (
	"blindSignAccount/main/crypt"
	"blindSignAccount/main/model"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"reflect"
	"strconv"
)

var Server *model.Server

type Msg interface{}

type msg struct {
	Data interface{} `json:"data"`
	Err  string      `json:"err"`
}

func init() {
	Server = model.NewServer()
	//var err error
	//var fileName = ""
	//if Server, err = model.NewServerFromFile(fileName); err != nil {
	//	panic(err)
	//}
}

func New(data interface{}, err error) Msg {
	var msg msg
	msg.Data = data
	if err != nil {
		msg.Err = err.Error()
	}
	return msg
}

func render(c *gin.Context, data gin.H, statusCode *int, err *error) {
	msg := New(data["payload"], *err)

	switch c.Request.Header.Get("Accept") {
	case "application/xml":
		// Respond with XML
		c.XML(*statusCode, msg)
	default: //"application/json"
		// Respond with JSON
		c.JSON(*statusCode, msg)
	}
}

func readBody(c *gin.Context) (map[string]interface{}, error) {
	if c.Request.Body == nil {
		return make(map[string]interface{}, 0), nil
	}
	bodyBytes, _ := ioutil.ReadAll(c.Request.Body)
	var values = make(map[string]interface{})
	err := json.Unmarshal(bodyBytes, &values)
	return values, err
}

// Check if the body contains all needed elements
func checkMissingElements(body *map[string]interface{}, elements *map[string]interface{}) error {
	var missingElements string

	for elemName := range *elements {
		if (*body)[elemName] == nil {
			if missingElements != "" {
				missingElements += ", "
			}
			missingElements += elemName
		}
	}
	if missingElements != "" {
		return errors.New("missing body element(s): " + missingElements)
	}
	return nil
}

func parseElements(elements *map[string]interface{}, body *map[string]interface{}) error {
	var err error

	var misTypes string
	for elemName, elem := range *elements {
		parameter := (*body)[elemName]
		switch elem.(type) {
		case int:
			switch parameter.(type) {
			case float64:
				(*elements)[elemName] = int(parameter.(float64))
			case int:
				(*elements)[elemName] = parameter.(int)
			case string:
				(*elements)[elemName], err = strconv.Atoi(parameter.(string))
			default:
				err = errors.New("wrong type (" + reflect.TypeOf(parameter).String() + ")")
			}
			if err != nil {
				if misTypes != "" {
					misTypes += ", "
				}
				misTypes += elemName
				err = nil
			}
		case string:
			switch parameter.(type) {
			case string:
				(*elements)[elemName] = parameter
			default:
				misTypes += elemName + " "
			}
		case []byte:
			switch parameter.(type) {
			case string:
				paramToString, _ := hex.DecodeString(parameter.(string))
				(*elements)[elemName] = paramToString
			default:
				misTypes += elemName + " "
			}
		case []string:
			switch parameter.(type) {
			case []string:
				(*elements)[elemName] = parameter
			case []interface{}:
				(*elements)[elemName], err = parseToStringSlice(parameter.([]interface{}))
				if err != nil {
					misTypes += elemName + "( " + err.Error() + ") "
					err = nil
				}
			default:
				fmt.Println(reflect.TypeOf(parameter))
				misTypes += " " + elemName
			}
		case *crypt.AddressBundle:
			switch parameter.(type) {
			case crypt.AddressBundle:
				(*elements)[elemName] = &parameter
			case map[string]interface{}:
				paramMap := parameter.(map[string]interface{})
				(*elements)[elemName], err = parseAdrBundle(paramMap)
				if err != nil {
					misTypes += "adrBdl (" + err.Error() + ")"
					err = nil
				}
			default:
				misTypes += elemName + " "
			}
		default:
			err = errors.New("type " + reflect.TypeOf(elem).String() + " not yet implemented")
			return err
		}
	}

	if misTypes != "" {
		return errors.New("wrong parameter types : " + misTypes)
	}

	return nil
}

func parseAdrBundle(paramMap map[string]interface{}) (*crypt.AddressBundle, error) {
	var adrBdl = &crypt.AddressBundle{}
	// check if the map contains all elements
	if _, found := paramMap["Seed"]; !found || paramMap["Seed"] == nil || reflect.TypeOf(paramMap["Seed"]).String() != "string" {
		return nil, errors.New("missing seed")
	}
	if _, found := paramMap["AccountID"]; !found || reflect.TypeOf(paramMap["AccountID"]).String() != "float64" {
		return nil, errors.New("missing AccountID")
	}
	if _, found := paramMap["AddressID"]; !found || reflect.TypeOf(paramMap["AddressID"]).String() != "float64" {
		return nil, errors.New("missing AddressID")
	}
	if _, found := paramMap["Address"]; !found || reflect.TypeOf(paramMap["Address"]).String() != "string" {
		return nil, errors.New("missing AddressID")
	}
	decodedSeed, _ := hex.DecodeString(paramMap["Seed"].(string))
	adrBdl.Seed = decodedSeed
	adrBdl.AccountID = uint32(paramMap["AccountID"].(float64))
	adrBdl.AddressID = uint32(paramMap["AddressID"].(float64))
	adrBdl.Address = paramMap["Address"].(string)
	return adrBdl, nil
}

func parseToStringSlice(elements []interface{}) ([]string, error) {
	var out = make([]string, len(elements))

	for idx, elem := range elements {
		if reflect.TypeOf(elem).String() != "string" {
			return nil, errors.New("non-string found")
		}
		out[idx] = elem.(string)
	}
	return out, nil
}

func parseBody(c *gin.Context, elements *map[string]interface{}) error {
	var err error
	var body map[string]interface{}

	if body, err = readBody(c); err != nil {
		return err
	}

	if err = checkMissingElements(&body, elements); err != nil {
		return err
	}

	if err = parseElements(elements, &body); err != nil {
		return err
	}

	return nil
}
