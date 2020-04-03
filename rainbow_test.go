package rainbow

import (
	"bytes"
	"crypto"
	"fmt"
	"math/rand"
	"testing"
)

func TestBasicRainbow(t *testing.T) {
	r := getTestRainbow()
	c1 := r.NewChain()
	c2 := r.NewChain()
	if bytes.Compare(c1.Start, c2.Start) == 0 ||
		bytes.Compare(c1.End, c2.End) == 0 ||
		bytes.Compare(c1.End, c1.Start) == 0 ||
		bytes.Compare(c2.End, c2.Start) == 0 {
		fmt.Printf("%+v\n", c1)
		fmt.Printf("%+v\n", c2)
		t.Fatal("unexpected chain collisions")
	}
}

func TestChainsBasic(t *testing.T) {
	r := getTestRainbow()
	for i := 0; i < 10; i++ {
		c := r.NewChain()
		r.AddChain(c)
	}
	fmt.Println(r)
	r.SortChains()
	fmt.Println(r)

	h := r.Chains[5].End
	c, found := r.findChain(h)
	if !found || bytes.Compare(c.End, h) != 0 {
		t.Log(h)
		t.Log(r.Chains[5])
		t.Fatal("should have found chain # 5")
	}

	// Check retrieving a known password, in the right or wrong chain
	p := []byte{}
	h = []byte{202, 215, 41, 34, 70, 162, 213, 246, 40, 111, 16, 218, 90, 118, 92, 243}
	p, found = r.walkChain(r.Chains[1], h)
	if !found || string(p) != "phzzoozt" {
		t.Log(string(p), " --> ", h)
		t.Fatal("was not able to find predefined password")
	}
	_, found = r.walkChain(r.Chains[2], h)
	if found {
		t.Fatal("found a non exiting hash in chain #2 ")
	}

	// use the Lookup to do the full cycle
	p, found = r.Lookup(h)
	if !found || string(p) != "phzzoozt" {
		t.Fatal("lookup failed,  retrieving ", string(p), "instead of phzzoozt")
	}

	hh := append([]byte{}, h...)
	hh[0]++ // slight change should prevent retrieving the password ...
	p, found = r.Lookup(hh)
	if found || string(p) == "phzzoozt" {
		t.Fatal("lookup should have failed, but did not ! ")
	}

}

// ============================= utilities ==============================

func getTestRainbow() *Rainbow {
	return &Rainbow{
		H:     GetMD5Func(),
		R:     GetAlphaReduceFunc(8),
		Cl:    10,
		HSize: crypto.MD5.Size(),
		Rand:  rand.New(rand.NewSource(42)),
	}
}
