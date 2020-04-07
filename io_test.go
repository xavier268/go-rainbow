package rainbow

import (
	"crypto"
	"fmt"
	"os"
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
	testNbChains := 20_000
	fname := "testTable.rbw"
	b, e := os.Create(fname)
	if e != nil {
		t.Fatal(e)
	}

	// SAVING ============================
	r := New(crypto.MD5, 20)
	r.CompileAlphabet("abcdefgh", 2, 2).Build()
	for i := 0; i < testNbChains; i++ {
		r.AddChain(r.NewChain())
	}
	fmt.Println(r.getHeader())

	err := r.Save(b)
	if err != nil {
		t.Fatal(err)
	}
	b.Close()
	fmt.Println("Saved ", r.getHeader())

	// LOADING ===========================
	rr := New(crypto.MD5, 20)
	rr.CompileAlphabet("abcdefgh", 2, 2).Build()

	b, e = os.Open(fname)
	if e != nil {
		t.Fatal(e)
	}

	err = rr.Load(b)
	b.Close()
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println("rr loaded : ", rr.getHeader(), " with ", len(rr.chains), "chains now")

	// Detailled comparisons ...
	if len(rr.chains) != len(r.chains) {
		t.Fatalf("length do not match %d saved, but %d loaded", len(r.chains), len(rr.chains))
	}
	for i := range r.chains {
		if !r.chains[i].Equal(r.chains[i]) {
			t.Fatalf("chains # %d differ", i)
		}
	}

	// Load rr a second time
	b, e = os.Open(fname)
	if e != nil {
		t.Fatal(e)
	}

	err = rr.Load(b)
	b.Close()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("rr loaded twice : ", rr.getHeader(), " with ", len(rr.chains), "chains now")
	if len(rr.chains) != len(r.chains) {
		t.Fatalf("after dedup, length do not match %d saved, but %d loaded", len(r.chains), len(rr.chains))
	}
	for i := range r.chains {
		if !r.chains[i].Equal(r.chains[i]) {
			t.Fatalf("chains # %d differ", i)
		}
	}
}
