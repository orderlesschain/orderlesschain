package hasher

import (
	"crypto"
	"crypto/md5"
	"crypto/sha256"
)

const LimitHashedBlock = 30
const Limit = 10

// GlobalHashType possible types: SHA256 or MD5
const GlobalHashType = crypto.SHA256

func HashMD5(data []byte) []byte {
	hashedValue := md5.Sum(data)
	return hashedValue[:]
}

func HashSHA256(data []byte) []byte {
	hashSha256 := sha256.New()
	hashSha256.Write(data)
	return hashSha256.Sum(nil)
}

func Hash(data []byte) []byte {
	return HashSHA256(data)
}
