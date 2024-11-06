package util

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"strconv"

	"github.com/mr-tron/base58"
	"golang.org/x/crypto/ripemd160"
)

const (
	CheckSumLength = 4
)

func IntToHex(n int64) []byte {
	return []byte(strconv.FormatInt(n, 16))
}

func Handle(err error) {
	if err != nil {
		panic(err)
	}
}
func BytesToInt(b []byte) int {
	n, err := strconv.Atoi(string(b))
	Handle(err)
	return n
}

func Atoi(s string) int {
	n, err := strconv.Atoi(s)
	Handle(err)
	return n
}

func Base58Encode(input []byte) []byte {
	encode := base58.Encode(input)
	return []byte(encode)
}
func Base58Decode(input []byte) []byte {
	decode, err := base58.Decode(string(input))
	Handle(err)
	return decode
}

func Serialize(data interface{}) []byte {
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	err := encoder.Encode(data)
	Handle(err)
	return buffer.Bytes()
}

func Deserialize(data []byte, to interface{}) {
	decoder := gob.NewDecoder(bytes.NewReader(data))
	err := decoder.Decode(to)
	Handle(err)
}
func CheckSum(payload []byte) []byte {
	firstHash := sha256.Sum256(payload)
	secondHash := sha256.Sum256(firstHash[:])
	return secondHash[:CheckSumLength]
}
func ValidateAddress(address string) bool {
	pubKeyHash := Base58Decode([]byte(address))
	actualCheckSum := pubKeyHash[len(pubKeyHash)-CheckSumLength:]
	version := pubKeyHash[0]
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-CheckSumLength]
	targetCheckSum := CheckSum(append([]byte{version}, pubKeyHash...))
	return bytes.Equal(actualCheckSum, targetCheckSum)
}
func HashPubKey(PublicKey []byte) []byte {
	pubHash := sha256.Sum256(PublicKey)
	hasher := ripemd160.New()
	_, err := hasher.Write(pubHash[:])
	Handle(err)
	return hasher.Sum(nil)
}

func ValidateAmount(ammount int) bool {
	if ammount < 0 {
		return false
	}
	return true
}

func StrToInt(s string) int {
	i, err := strconv.Atoi(s)
	Handle(err)
	return i
}

func Int64ToByte(n int64) []byte {
	return []byte(strconv.FormatInt(n, 10))
}
