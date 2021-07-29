package crypt

import (
	"crypto"
	"crypto/rsa"
	"github.com/cryptoballot/fdh"
	"github.com/cryptoballot/rsablind"
)

const KeyLength = 1024

type BlindBundle struct {
	Token      string
	BlindToken []byte
	HashValue  []byte
	UnBlinder  []byte
	BlindSig   []byte
}

func CreateBlindBundle(key rsa.PublicKey) (*BlindBundle, error) {
	token := GenerateToken()
	hashValue := fdh.Sum(crypto.SHA256, 768, []byte(token))
	blindToken, unBlinder, err := rsablind.Blind(&key, hashValue)
	if err != nil {
		return nil, err
	}
	return &BlindBundle{Token: token, HashValue: hashValue, BlindToken: blindToken, UnBlinder: unBlinder}, nil
}

func GetBlindSignatureTestData(token string, key *rsa.PrivateKey) (blindToken, blindSig, hashValue, sig []byte, err error) {
	var unBlind []byte
	// hash it
	hashValue = fdh.Sum(crypto.SHA256, 768, []byte(token))
	// blind and unBlind
	blindToken, unBlind, err = rsablind.Blind(&key.PublicKey, hashValue)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	if blindSig, err = rsablind.BlindSign(key, blindToken); err != nil {
		return nil, nil, nil, nil, err
	}
	sig = rsablind.Unblind(&key.PublicKey, blindSig, unBlind)
	return
}
