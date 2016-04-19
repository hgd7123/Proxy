package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"
	cfg "goconfig"
	"io"
	"net"
	"os"
)

var (
	aesTable string = "ywlSRb80TaCQ4b7bZXCQzxs9"
	AESADDR  string = ":5001"
	//LOCALAESADDR string = ":5001"
	Saddr    string = ":5000"
	Haddr    string = ":5002"
	G_Config *cfg.ConfigFile
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

func HttpProxy() {
	listener, err := net.Listen("tcp", Haddr)
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}

	for {
		conn, liserr := listener.Accept()
		if liserr != nil {
			fmt.Println(liserr)
		}
		comt := NewConnMngr(conn)
		go comt.DealHttp()
	}
}

func main() {

	if len(os.Args) > 1 {
		if os.Args[1] == "-h" || os.Args[1] == "--help" {
			fmt.Println("args 1 is local socks addr Defaut ':5000'")
			fmt.Println("args 2 is remote socks addr Defaut ':5001'")
			fmt.Println("args 3 is local http addr Default ':5002'")
			os.Exit(1)
		}
		Saddr = os.Args[1]
		fmt.Println(os.Args[1])
	}
	if len(os.Args) > 2 {
		AESADDR = os.Args[2]
		fmt.Println(os.Args[2])
	}

	if len(os.Args) > 3 {
		Haddr = os.Args[3]
		fmt.Println(os.Args[3])
	}

	go HttpProxy()

	listener, err := net.Listen("tcp", Saddr)
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}
	for {
		conn, liserr := listener.Accept()
		if liserr != nil {
			fmt.Println(liserr)
		}
		comt := NewConnMngr(conn)
		go comt.DealSock()
	}
}
