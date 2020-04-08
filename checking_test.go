package rainbow

import (
	"crypto"
	"math/big"
	"testing"
	"unsafe"
)

func TestChecks(t *testing.T) {

	// integer arithmetic
	if unsafe.Sizeof(int(0)) != unsafe.Sizeof(int64(0)) {
		t.Fatal("expected 64 bits default ints")
	}

	if i := 1024 - 1; byte(i) != 255 {
		t.Fatal("int to byte conversion error")
	}

	i := 256*17 + 5
	if byte(5) != byte(i) {
		t.Fatal("byte truncating not behaving as expected")
	}

	// crypto engines

	if !crypto.MD5.Available() {
		t.Fatal("MD5 is not available")
	}
	if !crypto.SHA1.Available() {
		t.Fatal("SHA1 is not available")
	}

	// string, char and rune
	s := "a⌘é"
	if len(s) != 6 {
		t.Fatal("wrong string byte length with utf8 : ", len(s))
	}
	if len([]rune(s)) != 3 {
		t.Fatal("wrong string rune length with utf8 : ", len([]rune(s)))
	}

}

func BenchmarkBigIntDivMod(b *testing.B) {
	var i, j, k = (&big.Int{}).SetInt64(111),
		(&big.Int{}).SetInt64(117),
		(&big.Int{}).SetInt64(213)
		// make them bigger ...
	for t := 0; t < 10; t++ {
		i.Mul(i, i)
		j.Mul(j, j)
		k.Mul(k, k)
	}
	// start benchmark
	b.ResetTimer()
	for t := 0; t < b.N; t++ {
		_, _ = i.DivMod(i, j, k)
	}
}
