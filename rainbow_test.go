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
	if bytes.Compare(c1.start, c2.start) == 0 ||
		bytes.Compare(c1.end, c2.end) == 0 ||
		bytes.Compare(c1.end, c1.start) == 0 ||
		bytes.Compare(c2.end, c2.start) == 0 {
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

	h := r.chains[5].end
	from, to, found := r.findChain(h)
	t.Logf("%d matching chain(s) : from %d(included) to %d(excluded)", to-from, from, to)
	c := r.chains[from]
	if !found || bytes.Compare(c.end, h) != 0 {
		t.Log(h)
		t.Log(r.chains[5])
		t.Fatal("should have found chain # 5")
	}

	// Check retrieving a known password, in the right or wrong chain

	psswdTest, hashTest := r.getPHSample(r.chains[0], 5)
	p := []byte{}
	p, found = r.walkChain(r.chains[0], hashTest) // correct chain
	if !found || string(p) != string(psswdTest) {
		t.Log(string(p), " --> ", hashTest)
		t.Fatal("was not able to find predefined password, should have been ther")
	}
	_, found = r.walkChain(r.chains[1], hashTest) // wrong chain
	if found {
		t.Fatal("found a non existing hash in chain #2 ")
	}

	// use the Lookup to do the full cycle
	p, found = r.Lookup(hashTest)
	if !found || string(p) != string(psswdTest) {
		t.Fatal("lookup failed,  retrieving ", string(p), "instead of ", psswdTest)
	}

	// verify ?
	if bytes.Compare(r.hf(p, []byte{}), hashTest) != 0 {
		t.Fatal("password returned did not match the requested hash")
	}

	hh := append([]byte{}, hashTest...)
	hh[0]++ // slight change should prevent retrieving the password ...
	p, found = r.Lookup(hh)
	if found || string(p) == string(psswdTest) {
		t.Fatal("lookup should have failed, but did not ! ")
	}

}

// ============================= benchmarks =============================

func BenchmarkAddChainWithAlphabet(b *testing.B) {
	r := New(crypto.MD5, 100).
		CompileAlphabet("abcdefghijklmnopqrstuvwxyz", 0, 16).
		Build()
	b.ResetTimer()
	r.benchmarkAddChain(b)
}

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

/*
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
*/

func BenchmarkLookup500x2_000(b *testing.B) {
	r := getTestRainbow(500)    // chain length
	r.benchmarkLookup(2_000, b) // nb of chains
}

func BenchmarkLookup500x200(b *testing.B) {
	r := getTestRainbow(500)  // chain length
	r.benchmarkLookup(200, b) // nb of chains
}

// ----------------------------------------------------------

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
	r := New(crypto.MD5, chainLength).CompileAlphabet("abcdefghijklmnopqrstuvwxyz", 2, 3).Build()
	r.rand = rand.New(rand.NewSource(42))
	return r
}

// Get a sample hash with coresponding password from the specified chain.
func (r *Rainbow) getPHSample(c *Chain, level int) (p, h []byte) {
	h, p = []byte{}, []byte{}
	h = append(h, c.start...)
	for i := 0; i < level; i++ {
		p = r.rf(i, h, p)
		h = r.hf(p, h)
	}
	return p, h

}

func TestDedup1(t *testing.T) {

	// create a rainbow table
	r := New(crypto.MD5, 20)
	r.CompileAlphabet("abcdefgh", 2, 2).Build()
	for i := 0; i < 5_000; i++ {
		r.AddChain(r.NewChain())
	}
	// ensure duplication
	r.chains = append(r.chains, r.chains[200])

	r.DedupChains()
	if len(r.chains) != 5000 {
		fmt.Println(r.signature)
		fmt.Println("Nb of chains : ", len(r.chains))
		t.Fatal("insufficinet deduplication")
	}
}
func TestDedup100(t *testing.T) {

	// create a rainbow table
	r := New(crypto.MD5, 20)
	r.CompileAlphabet("abcdefgh", 2, 2).Build()
	for i := 0; i < 5_000; i++ {
		r.AddChain(r.NewChain())
	}
	// ensure duplication
	r.chains = append(r.chains, r.chains[200:300]...)

	r.DedupChains()
	if len(r.chains) != 5000 {
		fmt.Println(r.signature)
		fmt.Println("Nb of chains : ", len(r.chains))
		t.Fatal("insufficinet deduplication")
	}

}

func TestDedup150(t *testing.T) {

	// create a rainbow table
	r := New(crypto.MD5, 20)
	r.CompileAlphabet("abcdefgh", 2, 2).Build()
	for i := 0; i < 5_000; i++ {
		r.AddChain(r.NewChain())
	}
	// ensure duplication and shuffle
	r.chains = append(r.chains, r.chains[200:300]...)
	r.chains = append(r.chains, r.chains[250:300]...)
	r.chains[0], r.chains[250] = r.chains[250], r.chains[0]
	r.chains[10], r.chains[1250] = r.chains[1250], r.chains[10]

	r.DedupChains()
	if len(r.chains) != 5000 {
		fmt.Println(r.signature)
		fmt.Println("Nb of chains : ", len(r.chains))
		t.Fatal("insufficinet deduplication")
	}
}

func TestFindChain(t *testing.T) {

	// create a rainbow table
	r := New(crypto.MD5, 20)
	r.CompileAlphabet("abcdefg", 3, 5).Build()
	for i := 0; i < 500; i++ {
		r.AddChain(r.NewChain())
	}

	r.SortChains()

	testH := r.chains[25].end

	from, to, found := r.findChain(testH)
	if found {
		t.Logf("Found %d chains, from %d(incl.) to %d(excl.)", to-from, from, to)
		if from > 25 || to <= 25 {
			t.Fatal("selected chain (25) in not within the retuned interval")
		}
	} else {
		t.Logf("from %d, to %d", from, to)
		t.Fatal("should have found at least a chain")
	}
}
