package signer

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"errors"
	"gitlab.lrz.de/orderless/orderlesschain/internal/customcrypto/hasher"
	"gitlab.lrz.de/orderless/orderlesschain/internal/customcrypto/keygenerator"
	"gitlab.lrz.de/orderless/orderlesschain/internal/logger"
	"sync"
)

type Signer struct {
	publicKeys map[string]*rsa.PublicKey
	privateKey *rsa.PrivateKey
	lock       *sync.RWMutex
}

func NewSigner(key *keygenerator.RSAKey) *Signer {
	privateKey := key.PrivateKey
	tempSigner := &Signer{
		publicKeys: map[string]*rsa.PublicKey{},
		privateKey: privateKey,
		lock:       &sync.RWMutex{},
	}
	return tempSigner
}

func (s *Signer) AddPublicKey(id, publicKey string) {
	s.lock.Lock()
	s.publicKeys[id] = keygenerator.ConvertStringToPublicKey(publicKey)
	s.lock.Unlock()
}

func (s *Signer) Sign(data []byte) []byte {
	signature, err := rsa.SignPKCS1v15(rand.Reader, s.privateKey, hasher.GlobalHashType, hasher.Hash(data))
	if err != nil {
		logger.ErrorLogger.Println(err)
	}
	return signature
}

func (s *Signer) Verify(publicKeyOwner string, data []byte, signature []byte) error {
	s.lock.RLock()
	publicKey := s.publicKeys[publicKeyOwner]
	s.lock.RUnlock()
	if publicKey == nil || len(data) == 0 || len(signature) == 0 {
		return errors.New("empty public key or data or signature")
	}
	return rsa.VerifyPKCS1v15(publicKey, hasher.GlobalHashType, hasher.Hash(data), signature)
}

func (s *Signer) SignAndEncode(data []byte) string {
	return base64.StdEncoding.EncodeToString(s.Sign(data))
}

func (s *Signer) DecodeAndVerify(publicKeyOwner string, data []byte, signature string) error {
	decodedSignature, _ := base64.StdEncoding.DecodeString(signature)
	return s.Verify(publicKeyOwner, data, decodedSignature)
}
