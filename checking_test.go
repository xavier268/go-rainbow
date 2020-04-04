package rainbow

import (
	"crypto"
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
