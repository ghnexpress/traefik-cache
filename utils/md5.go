package utils

import (
	"crypto/md5"
	"encoding/hex"
)

func GetMD5Hash(d []byte) string {
	hash := md5.Sum(d)
	return hex.EncodeToString(hash[:])
}
