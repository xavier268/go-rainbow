package rainbow

import (
	"crypto"

	// implementation packages
	_ "crypto/md5"
	_ "crypto/sha1"
)

// GetCryptoFunc constructs a HashFunc based on the underlying crypto.Hash.
func GetCryptoFunc(ch crypto.Hash) HashFunc {
	if !ch.Available() {
		panic("the requested crypto algorith has not been imported and is not available")
	}
	hh := ch.New()
	return func(p, h []byte) []byte {
		hh.Reset()
		hh.Write(p)
		return hh.Sum(h[:0])
	}
}

// GetMD5Func retun a MD5 hash function
func GetMD5Func() HashFunc {
	return GetCryptoFunc(crypto.MD5)
}

// GetSHA1Func retun a MD5 hash function
func GetSHA1Func() HashFunc {
	return GetCryptoFunc(crypto.SHA1)
}
