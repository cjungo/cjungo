package ext

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
)

type HashResult []byte

func Md5(v string) HashResult {
	hash := md5.New()
	hash.Write([]byte(v))
	return hash.Sum(nil)
}

func Sha256(v string) HashResult {
	hash := sha256.New()
	hash.Write([]byte(v))
	return hash.Sum(nil)
}

func (result HashResult) Hex() string {
	return hex.EncodeToString(result)
}

func (result HashResult) Base64() string {
	return base64.StdEncoding.EncodeToString(result)
}
