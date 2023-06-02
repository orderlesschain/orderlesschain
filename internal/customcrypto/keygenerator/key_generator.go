package keygenerator

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"gitlab.lrz.de/orderless/orderlesschain/internal/logger"
	"os"
	"path/filepath"
)

const privateKeyPath = "./configs/privateKey.pem"
const publicKeyPath = "./configs/publicKey.pem"

type RSAKey struct {
	PublicKey       *rsa.PublicKey
	PublicKeyString string
	PrivateKey      *rsa.PrivateKey
}

func NewRSAKeys() (RSAKey, error) {
	var k RSAKey
	privateKey, err := rsa.GenerateKey(rand.Reader, 512)
	if err != nil {
		return k, err
	}
	k.PublicKey = &privateKey.PublicKey
	k.PrivateKey = privateKey
	privateKeyFilePath, _ := filepath.Abs(privateKeyPath)
	privateKeyFile, err := os.Create(privateKeyFilePath)
	if err != nil {
		logger.FatalLogger.Fatalln(err)
	}
	privateBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(k.PrivateKey),
	}
	err = pem.Encode(privateKeyFile, privateBlock)
	if err != nil {
		logger.FatalLogger.Fatalln(err)
	}
	publicKeyFilePath, _ := filepath.Abs(publicKeyPath)
	publicKeyFile, err := os.Create(publicKeyFilePath)
	if err != nil {
		logger.FatalLogger.Fatalln(err)
	}
	publicBlock := &pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: x509.MarshalPKCS1PublicKey(k.PublicKey),
	}
	err = pem.Encode(publicKeyFile, publicBlock)
	if err != nil {
		logger.FatalLogger.Fatalln(err)
	}
	return k, nil
}

func (k RSAKey) PrivateKeyToPemString() string {
	return string(
		pem.EncodeToMemory(
			&pem.Block{
				Type:  "RSA PRIVATE KEY",
				Bytes: x509.MarshalPKCS1PrivateKey(k.PrivateKey),
			},
		),
	)
}

func (k RSAKey) PublicKeyToPemString() string {
	return string(
		pem.EncodeToMemory(
			&pem.Block{
				Type:  "RSA PUBLIC KEY",
				Bytes: x509.MarshalPKCS1PublicKey(k.PublicKey),
			},
		),
	)
}

func LoadPublicPrivateKeyFromFile() *RSAKey {
	key := &RSAKey{}
	key.PrivateKey = ReadPrivateKeyFromFile()
	key.PublicKey = ReadPublicKeyFromFile()
	key.PublicKeyString = string(
		pem.EncodeToMemory(
			&pem.Block{
				Type:  "RSA PUBLIC KEY",
				Bytes: x509.MarshalPKCS1PublicKey(key.PublicKey),
			},
		),
	)
	return key
}

func ReadPrivateKeyFromFile() *rsa.PrivateKey {
	privateKeyFilePath, _ := filepath.Abs(privateKeyPath)
	privateKeyFile, err := os.Open(privateKeyFilePath)
	if err != nil {
		logger.FatalLogger.Fatalln(err)
	}
	defer func(privateKeyFile *os.File) {
		err = privateKeyFile.Close()
		if err != nil {
			logger.ErrorLogger.Println(err)
		}
	}(privateKeyFile)
	privateKeyInfo, err := privateKeyFile.Stat()
	if err != nil {
		logger.FatalLogger.Fatalln(err)
	}
	privateKeyBuffer := make([]byte, privateKeyInfo.Size())
	_, err = privateKeyFile.Read(privateKeyBuffer)
	if err != nil {
		logger.FatalLogger.Fatalln(err)
	}
	block, _ := pem.Decode(privateKeyBuffer)
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		logger.FatalLogger.Fatalln(err)
	}
	return privateKey
}

func ReadPublicKeyFromFile() *rsa.PublicKey {
	publicKeyFilePath, _ := filepath.Abs(publicKeyPath)
	publicKeyFile, err := os.Open(publicKeyFilePath)
	if err != nil {
		logger.FatalLogger.Fatalln(err)
	}
	defer func(publicKeyFile *os.File) {
		err = publicKeyFile.Close()
		if err != nil {
			logger.ErrorLogger.Println(err)
		}
	}(publicKeyFile)
	publicKeyInfo, err := publicKeyFile.Stat()
	if err != nil {
		logger.FatalLogger.Fatalln(err)
	}
	publicKeyBuffer := make([]byte, publicKeyInfo.Size())
	_, err = publicKeyFile.Read(publicKeyBuffer)
	if err != nil {
		logger.FatalLogger.Fatalln(err)
	}
	block, _ := pem.Decode(publicKeyBuffer)
	publicKey, err := x509.ParsePKCS1PublicKey(block.Bytes)
	if err != nil {
		logger.FatalLogger.Fatalln(err)
	}
	return publicKey
}

func ConvertStringToPublicKey(publicKeyString string) *rsa.PublicKey {
	block, _ := pem.Decode([]byte(publicKeyString))
	publicKey, err := x509.ParsePKCS1PublicKey(block.Bytes)
	if err != nil {
		logger.FatalLogger.Fatalln(err)
	}
	return publicKey
}
