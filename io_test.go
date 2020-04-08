package rainbow

import (
	"crypto"
	"fmt"
	"os"
	"testing"
)

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
	fmt.Println(r.signature)

	err := r.Save(b)
	if err != nil {
		t.Fatal(err)
	}
	b.Close()
	fmt.Println("Saved ", r.signature)

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

	fmt.Println("rr loaded : ", rr.signature, " with ", len(rr.chains), "chains now")

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
	fmt.Println("rr loaded twice : ", rr.signature, " with ", len(rr.chains), "chains now")
	if len(rr.chains) != len(r.chains) {
		t.Fatalf("after dedup, length do not match %d saved, but %d loaded", len(r.chains), len(rr.chains))
	}
	for i := range r.chains {
		if !r.chains[i].Equal(r.chains[i]) {
			t.Fatalf("chains # %d differ", i)
		}
	}
}
