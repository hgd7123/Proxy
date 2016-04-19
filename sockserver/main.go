package main

import (
	"crypto/aes"
	"fmt"
	"net"

	"crypto/cipher"

	"crypto/rand"

	"errors"

	"io"
)

const Saddr = ":5001"

const (
	aesTable = "ywlSRb80TaCQ4b7bZXCQzxs9"
)

var (
	aesBlock cipher.Block

	ErrAESTextSize = errors.New("ciphertext is not a multiple of the block size")

	ErrAESPadding = errors.New("cipher padding size error")
)

func init() {

	var err error

	aesBlock, err = aes.NewCipher([]byte(aesTable))

	if err != nil {

		panic(err)

	}

}

// AES解密

func aesDecrypt(src []byte) ([]byte, error) {

	// 长度不能小于aes.Blocksize

	if len(src) < aes.BlockSize*2 || len(src)%aes.BlockSize != 0 {

		return nil, ErrAESTextSize

	}

	srcLen := len(src) - aes.BlockSize

	decryptText := make([]byte, srcLen)

	iv := src[srcLen:]

	mode := cipher.NewCBCDecrypter(aesBlock, iv)

	mode.CryptBlocks(decryptText, src[:srcLen])

	paddingLen := int(decryptText[srcLen-1])

	if paddingLen > 16 {

		return nil, ErrAESPadding

	}

	return decryptText[:srcLen-paddingLen], nil

}

// AES加密

func aesEncrypt(src []byte) ([]byte, error) {

	padLen := aes.BlockSize - (len(src) % aes.BlockSize)

	for i := 0; i < padLen; i++ {

		src = append(src, byte(padLen))

	}

	srcLen := len(src)

	encryptText := make([]byte, srcLen+aes.BlockSize)

	iv := encryptText[srcLen:]

	if _, err := io.ReadFull(rand.Reader, iv); err != nil {

		return nil, err

	}

	mode := cipher.NewCBCEncrypter(aesBlock, iv)

	mode.CryptBlocks(encryptText[:srcLen], src)

	return encryptText, nil

}

func main() {

	listener, err := net.Listen("tcp", Saddr)
	if err != nil {
		fmt.Println(err)
	}
	for {
		conn, liserr := listener.Accept()
		if liserr != nil {
			fmt.Println(liserr)
		}
		comt := NewConnMngr(conn)
		go comt.Deal()
	}
}
