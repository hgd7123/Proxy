package main

import (
	"crypto/cipher"
	"encoding/json"
	"errors"
	"fmt"
)

const (
	aesTable = "ywlSRb80TaCQ4b7bZXCQzxs9"
	iv       = "0xf7u8ilp5gthyjn"
)

var (
	aesBlock cipher.Block

	ErrAESTextSize = errors.New("ciphertext is not a multiple of the block size")

	ErrAESPadding = errors.New("cipher padding size error")
)

//func init() {
//	var err error
//	aesBlock, err = aes.NewCipher([]byte(aesTable))
//	if err != nil {
//		panic(err)
//	}

//}

//// AES解密

//func aesDecrypt(src []byte) ([]byte, error) {

//	// 长度不能小于aes.Blocksize
//	if len(src) < aes.BlockSize*2 || len(src)%aes.BlockSize != 0 {
//		return nil, ErrAESTextSize
//	}
//	srcLen := len(src)
//	decryptText := make([]byte, srcLen)

//	fmt.Println(iv)
//	mode := cipher.NewCBCDecrypter(aesBlock, iv)
//	mode.CryptBlocks(decryptText, src)
//	paddingLen := int(decryptText[srcLen-1])
//	if paddingLen > 16 {
//		return nil, ErrAESPadding
//	}
//	return decryptText[:srcLen-paddingLen], nil

//}

//// AES加密

//func aesEncrypt(src []byte) ([]byte, error) {

//	padLen := aes.BlockSize - (len(src) % aes.BlockSize)
//	fmt.Println("panding len", padLen)

//	for i := 0; i < padLen; i++ {
//		src = append(src, byte(padLen))
//	}

//	srcLen := len(src)
//	encryptText := make([]byte, srcLen)

//	fmt.Println(iv)
//	mode := cipher.NewCBCEncrypter(aesBlock, iv)
//	mode.CryptBlocks(encryptText, src)
//	return encryptText, nil

//}
type IatResultArgs struct {
	Ls bool `json:"ls,omitempty"`
	Sn int  `json:"sn,omitempty"`
}

func main() {
	args := "{\"sn\":1,\"ls\":true,\"sub\":\"iat\"}"
	fmt.Println(args[0:2], args[2:5])
	var iatarg IatResultArgs
	err := json.Unmarshal([]byte(args), &iatarg)

	fmt.Print(err)
	fmt.Print(iatarg.Ls, iatarg.Sn)
	ca, err := NewConnAes(aesTable, iv)
	if err != nil {
		fmt.Println(err)
	}
	enbuf, err := ca.Encrypt([]byte("11111111111111111"))

	fmt.Println(enbuf, err, len(enbuf))
	debuf, err := ca.Decrypt(enbuf)
	fmt.Println(string(debuf), err)

	//	en, err := aesEncrypt([]byte(""))
	//	fmt.Println(len(en))
	//	if err != nil {
	//		panic(err)
	//	}

	//	de, err := aesDecrypt(en)
	//	if err != nil {
	//		panic(err)
	//	}
	//	fmt.Println(string(de))

}
