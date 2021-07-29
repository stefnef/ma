package crypt

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"testing"
)

func Test_GetWalletKeys(t *testing.T) {
	mnemonic := "coil early bronze maze battle any core sweet burger busy cotton impact evoke oven jeans glance clock final eight crowd tool okay mushroom shrimp"
	realAddress := "1FFRUe1Lp4psmZHrq6NKrarN9Q9ntx2w3m"
	accountID := uint32(0)
	addressID := uint32(2)
	seed, keys, err := GetWalletKeys(mnemonic, accountID, false)
	if err != nil {
		t.Error(err)
		t.Fail()
	}

	// get the 3rd address
	address := GetAddress(keys[4], addressID)
	if address == nil {
		t.Error("could not generate address")
		t.FailNow()
	}
	if address.String() != realAddress {
		t.Errorf("wrong address generated: %s", address.String())
		t.Fail()
	}

	// calculate the same keys from the seed
	recalculatedKeys, err := getKeysFromSeed(seed, accountID, false)
	if err != nil {
		t.Error(err)
		t.Fail()
	}
	address = GetAddress(recalculatedKeys[4], addressID)
	if address == nil {
		t.Error("could not generate address")
		t.FailNow()
	}
	if address.String() != realAddress {
		t.Error("wrong address generated")
		t.Fail()
	}

	// calculate the same address from the seed
	adrBundle := &AddressBundle{Seed: seed, AccountID: accountID, AddressID: addressID}
	address, err = GetAddressFromSeed(adrBundle, false)
	if address == nil {
		t.Error(err)
		t.FailNow()
	}
	if address.String() != realAddress {
		t.Error("wrong address from seed")
		t.Fail()
	}
}

func TestGetPKey(t *testing.T) {
	mnemonic := "coil early bronze maze battle any core sweet burger busy cotton impact evoke oven jeans glance clock final eight crowd tool okay mushroom shrimp"
	accountID := uint32(0)
	_, keys, _ := GetWalletKeys(mnemonic, accountID, false)
	pk := GetPrivateKey(keys[4], 0)
	publicKey := GetPublicKey(keys[4], 0)
	if pk == nil || publicKey == nil {
		t.Error("could not calculate keys")
		t.FailNow()
	}

	msg := "hello, world"
	hash := sha256.Sum256([]byte(msg))

	reader := Reader{randNR: 7}
	r, s, err := ecdsa.Sign(reader, pk.ToECDSA(), hash[:])
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	if !ecdsa.Verify(&pk.PublicKey, hash[:], r, s) {
		t.Error("not valid")
		t.Fail()
	}
}
