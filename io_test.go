package rainbow

import (
	"bytes"
	"crypto"
	"fmt"
	"testing"
)

func TestHeader(t *testing.T) {
	r := New(crypto.SHA1, 568)
	hd := r.getHeader()
	fmt.Println(hd)

	bb := hd.toBytes()
	hd2 := new(header)
	hd2.fromBytes(bb)
	if !hd.Equal(hd2) {
		t.Log(hd)
		t.Log(bb)
		t.Log(hd2)
		t.Fatal("error header-byte-header conversion")
	}

	if e := r.checkHeader(hd); e != nil {
		t.Log(e)
		t.Log(hd)
		t.Fatal("header does not pass self check")
	}
	if e := r.checkHeader(hd2); e != nil {
		t.Log(e)
		t.Log(hd2)
		t.Fatal("header2 does not pass self check")
	}
}

func TestSaveLoad(t *testing.T) {

	b := new(bytes.Buffer)

	r := New(crypto.MD5, 2_000)
	r.CompileAlphabet("abcdefgh", 2, 2).Build()
	for i := 0; i < 5_000; i++ {
		r.AddChain(r.NewChain())
	}
	fmt.Println("r unsorted : ", r.getHeader())
	r.SortChains()
	fmt.Println("r sorted : ", r.getHeader())

	err := r.Save(b)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("Saved ", r.getHeader())

	rr := New(crypto.MD5, 2_000)
	rr.CompileAlphabet("abcdefgh", 2, 2).Build()

	// clone buffer and load
	bb := bytes.NewBuffer(b.Bytes())
	err = rr.Load(bb)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("rr loaded : ", rr.getHeader())

	// load original back
	err = r.Load(b)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("original r (re)loaded : ", r.getHeader())

}
