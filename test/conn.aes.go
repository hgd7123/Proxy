package main

import (
	"crypto/aes"
	"crypto/cipher"
	"errors"
)

type ConnAes struct {
	key      string
	keyBlock cipher.Block
	iv       string
}

func NewConnAes(key string, iv string) (ca ConnAes, err error) {
	if len(iv) != aes.BlockSize {
		return ca, errors.New("error iv size not 16")
	}
	ca.key = key
	ca.iv = iv
	ca.keyBlock, err = aes.NewCipher([]byte(key))
	return ca, err
}

func (ca *ConnAes) Encrypt(src []byte) ([]byte, error) {
	paddinglen := aes.BlockSize - (len(src) % aes.BlockSize)
	for i := 0; i < paddinglen; i++ {
		src = append(src, byte(paddinglen))
	}
	enbuf := make([]byte, len(src))
	cbce := cipher.NewCBCEncrypter(ca.keyBlock, []byte(iv))
	cbce.CryptBlocks(enbuf, src)
	return enbuf, nil
}

func (ca *ConnAes) Decrypt(src []byte) ([]byte, error) {

	if (len(src) < aes.BlockSize) || (len(src)%aes.BlockSize != 0) {
		return nil, errors.New("error encrypt data size")
	}

	debuf := make([]byte, len(src))
	cbcd := cipher.NewCBCDecrypter(ca.keyBlock, []byte(iv))
	cbcd.CryptBlocks(debuf, src)
	paddinglen := int(debuf[len(src)-1])
	if paddinglen > 16 {
		return nil, errors.New("error encrypt data size")
	}
	return debuf[:len(src)-paddinglen], nil
}
