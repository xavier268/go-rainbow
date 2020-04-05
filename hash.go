package rainbow

import (
	"crypto"

	// implementation packages
	_ "crypto/md5"
	_ "crypto/sha1"
)

// getCryptoFunc constructs a HashFunc based on the underlying crypto.Hash.
func getCryptoFunc(ch crypto.Hash) HashFunction {
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
