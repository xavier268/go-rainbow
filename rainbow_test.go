package rainbow

import (
	"bytes"
	"crypto"
	"fmt"
	"math/rand"
	"testing"
)

func TestBasicRainbow10(t *testing.T) {
	r := getTestRainbow(10)
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

func TestChainsBasic10(t *testing.T) {
	r := getTestRainbow(10)
	for i := 0; i < 10; i++ {
		c := r.NewChain()
		r.AddChain(c)
	}
	fmt.Println("unsorted table", r)
	r.SortChains()
	fmt.Println("sorted table", r)

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

	// verify ?
	if bytes.Compare(r.H(p, []byte{}), h) != 0 {
		t.Fatal("password returned did not match the requested hash")
	}

	hh := append([]byte{}, h...)
	hh[0]++ // slight change should prevent retrieving the password ...
	p, found = r.Lookup(hh)
	if found || string(p) == "phzzoozt" {
		t.Fatal("lookup should have failed, but did not ! ")
	}

}

// ============================= benchmarks =============================
func BenchmarkAddChain1_000(b *testing.B) {
	r := getTestRainbow(1_000)
	b.ResetTimer()
	r.benchmarkAddChain(b)
}
func BenchmarkAddChain10_000(b *testing.B) {
	r := getTestRainbow(10_000)
	r.benchmarkAddChain(b)
}

func BenchmarkAddChain100_000(b *testing.B) {
	r := getTestRainbow(100_000)
	r.benchmarkAddChain(b)
}
func BenchmarkAddChain1_000_000(b *testing.B) {
	r := getTestRainbow(1_000_000)
	r.benchmarkAddChain(b)
}

func BenchmarkLookup2_000x5_000(b *testing.B) {
	r := getTestRainbow(2_000)  // chain length
	r.benchmarkLookup(5_000, b) // nb of chains
}
func BenchmarkLookup5_000x2_000(b *testing.B) {
	r := getTestRainbow(5_000)  // chain length
	r.benchmarkLookup(2_000, b) // nb of chains
}
func BenchmarkLookup500x20_000(b *testing.B) {
	r := getTestRainbow(500)     // chain length
	r.benchmarkLookup(20_000, b) // nb of chains
}

func BenchmarkLookup500x2_000(b *testing.B) {
	r := getTestRainbow(500)    // chain length
	r.benchmarkLookup(2_000, b) // nb of chains
}

func BenchmarkLookup500x200(b *testing.B) {
	r := getTestRainbow(500)  // chain length
	r.benchmarkLookup(200, b) // nb of chains
}

// run a benchmark, adding chains to the  Rainbow object.
func (r *Rainbow) benchmarkAddChain(b *testing.B) {
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		c := r.NewChain()
		r.AddChain(c)
	}
}

// run an (unsuccessfull) lookup on the Rainbow object
func (r *Rainbow) benchmarkLookup(nbChains int, b *testing.B) {
	var c *Chain
	for i := 0; i < nbChains; i++ {
		c = r.NewChain()
		r.AddChain(c)
	}
	b.ResetTimer()
	h := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	var found bool
	var p []byte
	for n := 0; n < b.N; n++ {
		p, found = r.Lookup(h)
		if found {
			b.Log("Found ", p, "-->", h)
		}
	}

}

// ============================= utilities ==============================

func getTestRainbow(chainLength int) *Rainbow {
	return &Rainbow{
		H:     GetMD5Func(),
		R:     GetAlphaReduceFunc(8),
		Cl:    chainLength,
		HSize: crypto.MD5.Size(),
		Rand:  rand.New(rand.NewSource(42)),
	}
}
