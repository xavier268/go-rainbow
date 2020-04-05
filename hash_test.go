package rainbow

import (
	"bytes"
	"crypto"
	"testing"
)

func TestMD5(t *testing.T) {
	md := getCryptoFunc(crypto.MD5)
	h1 := md([]byte("hello world"), []byte{})
	h2 := md([]byte("hello world"), []byte{})
	// two different slices
	if bytes.Compare(h1, h2) != 0 {
		t.Log("h1 h2", h1, h2)
		t.Fatal("md5 hash is not consistent")
	}
	h3 := md([]byte("hello world"), h2)
	// h2 was rewritten into h3, same slices
	if bytes.Compare(h2, h3) != 0 {
		t.Log("h2 h3", h2, h3)
		t.Fatal("md5 hash is not consistent")
	}
	// h2 rewritten as h4
	h4 := md([]byte("hello wArld"), h2)
	if bytes.Compare(h4, h1) == 0 {
		t.Log("h4 h3", h4, h3)
		t.Fatal("unexpected md5 collision")
	}
	if bytes.Compare(h4, h2) != 0 {
		t.Log("h4 h2", h4, h2)
		t.Fatal("rewrite of slice did not happen as expected")
	}
}

func BenchmarkHashFunctionMD5(b *testing.B) {
	md := getCryptoFunc(crypto.MD5)
	h := make([]byte, 16, 16)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h = md(h, h)
	}
}
