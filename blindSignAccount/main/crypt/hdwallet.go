package crypt

import (
	"fmt"
	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil"
	"github.com/btcsuite/btcutil/hdkeychain"
	"github.com/tyler-smith/go-bip39"
)

const coinType = uint32(626)

type AddressBundle struct {
	Seed      []byte
	AccountID uint32
	AddressID uint32
	Address   string
}

type Reader struct {
	randNR int
}

func NewReader(randNR int) *Reader {
	return &Reader{randNR: randNR}
}

func (r Reader) Read(p []byte) (n int, err error) {
	return r.randNR, nil
}

func GetWalletKeys(mnemonic string, accountID uint32, protocol bool) (seed []byte, keys []*hdkeychain.ExtendedKey, err error) {
	// Generate a Bip32 HD wallet for the mnemonic and a user supplied password
	seed = bip39.NewSeed(mnemonic, "")
	master, err := hdkeychain.NewMaster(seed, &chaincfg.MainNetParams)
	if err != nil {
		return nil, nil, err
	}

	purpose, _ := master.Child(hdkeychain.HardenedKeyStart + 44)
	coin, _ := purpose.Child(hdkeychain.HardenedKeyStart + coinType)
	account, _ := coin.Child(hdkeychain.HardenedKeyStart + 0)
	external, _ := account.Child(accountID)

	if protocol {
		fmt.Println("bip32 root key", master.String())
		fmt.Println("purpose", purpose.String())
		fmt.Println("coin", coin.String())
		fmt.Println("account", account)
		fmt.Println("external", external)
	}

	keys = []*hdkeychain.ExtendedKey{master, purpose, coin, account, external}
	return seed, keys, nil
}

func GetAddress(key *hdkeychain.ExtendedKey, i uint32) *btcutil.AddressPubKeyHash {
	add0, _ := key.Child(i)
	address, _ := add0.Address(&chaincfg.MainNetParams)
	return address
}

func GetPrivateKey(key *hdkeychain.ExtendedKey, i uint32) *btcec.PrivateKey {
	add0, _ := key.Child(i)
	privateKey, _ := add0.ECPrivKey()
	return privateKey
}

func GetPublicKey(key *hdkeychain.ExtendedKey, i uint32) *btcec.PublicKey {
	add0, _ := key.Child(i)
	publicKey, _ := add0.ECPubKey()
	return publicKey
}

func getKeysFromSeed(seed []byte, accountID uint32, protocol bool) ([]*hdkeychain.ExtendedKey, error) {
	master, err := hdkeychain.NewMaster(seed, &chaincfg.MainNetParams)
	if err != nil {
		return nil, err
	}

	purpose, _ := master.Child(hdkeychain.HardenedKeyStart + 44)
	coin, _ := purpose.Child(hdkeychain.HardenedKeyStart + coinType)
	account, _ := coin.Child(hdkeychain.HardenedKeyStart + 0)
	external, _ := account.Child(accountID)

	if protocol {
		fmt.Println("bip32 root key", master.String())
		fmt.Println("purpose", purpose.String())
		fmt.Println("coin", coin.String())
		fmt.Println("account", account)
		fmt.Println("external", external)
	}

	keys := []*hdkeychain.ExtendedKey{master, purpose, coin, account, external}
	return keys, nil
}

// Calculates the address from the seed of the given address bundle.
func GetAddressFromSeed(adrBundle *AddressBundle, protocol bool) (*btcutil.AddressPubKeyHash, error) {
	keys, err := getKeysFromSeed(adrBundle.Seed, adrBundle.AccountID, protocol)
	if err != nil {
		return nil, err
	}
	return GetAddress(keys[4], adrBundle.AddressID), nil
}
