package rainbow

import (
	"bytes"
	"crypto"
	"fmt"
	"math/big"
	"math/rand"
	"sort"
	"time"
)

// Version of the package
func Version() (major, minor int) {
	return 0, 4
}

// Rainbow is the main type to generate tables or lookup a password.
type Rainbow struct {
	// hashing algorithm
	halgo crypto.Hash
	// hf is the Hash function used to compute from password to hash
	hf HashFunction
	// hsize is the size of the HashFunction result in  bytes
	hsize int

	// rf is a reduce function, from hash to password
	// It conforms to the hash.Hash interface.
	rf ReduceFunction

	// cl is the chain length (constant)
	cl int

	// Random generator
	rand *rand.Rand
	// chains
	chains []*Chain
	// Are the chains sorted ?
	sorted bool

	// The modules used to build the reduce name space
	rms []rmodule
	// flag : you can build only once.
	built bool
	// cumulative size of the big.Int that will be used
	used *big.Int
}

// New constructs a new, empty rainbow table,
// expecting fixed length chains of the specified length.
func New(hashAlgo crypto.Hash, chainLength int) *Rainbow {
	r := new(Rainbow)
	// define hash
	r.hf = getCryptoFunc(hashAlgo)
	r.hsize = hashAlgo.Size()
	r.halgo = hashAlgo
	// set chain length
	r.cl = chainLength
	// set random generator
	r.rand = rand.New(rand.NewSource(time.Now().UnixNano()))

	// Used so far by the modules
	r.used = new(big.Int).SetInt64(1)

	return r
}

// ReduceFunction is a function that reduces a hash
// into a password, the next in the chain.
// The password value is returned as a byte slice.
// We try to use the provided password slice, to avoid allocation.
// It might be modified - or not, and is not garanteed to equal the result.
type ReduceFunction func(step int, hash []byte, password []byte) []byte

// HashFunction is a function that hashes a password into a hash
// hash is put into the provided hash slice, returning it.
// We try to avoid allocation, using the passed hash slice.
// Since hash is usually don't change, no alloc are normally made.
// The passed hash may be modified - or not.
type HashFunction func(passwd []byte, hash []byte) []byte

// Chain in the table, contains the start and end hash values
type Chain struct {
	start []byte
	end   []byte
}

// Equal compare chains
func (c *Chain) Equal(cc *Chain) bool {
	if cc == nil {
		return false
	}
	return bytes.Compare(c.start, cc.start) == 0 &&
		bytes.Compare(c.end, cc.end) == 0
}

// NewChain builds a new Chain, from random start.
// It is not immediateley added to the Chains slice in r.
// See - AddChain below.
func (r *Rainbow) NewChain() *Chain {
	c := new(Chain)
	c.start = make([]byte, r.hsize, r.hsize)
	r.rand.Read(c.start)
	c.end = append([]byte{}, c.start...)
	p := []byte{}
	for i := 0; i < r.cl; i++ {
		p = r.rf(i, c.end, p)
		c.end = r.hf(p, c.end)
		//fmt.Println(string(p), "-->", c.End)
	}
	return c
}

// BitLen provides the number of bits needed to encode the namespace.
func (r *Rainbow) BitLen() int {
	return r.used.BitLen()
}

// Build finish compiling the Rainbow table "reduce" function.
func (r *Rainbow) Build() *Rainbow {

	if r.rf != nil {
		panic("a reduce function was already defined, you cannot redefine it")
	}

	// check reduce name space cardinality
	if r.used.BitLen() > 8*r.hsize {
		panic("the full reduced name space is larger than the hash space")
	}

	// update the reduce function
	r.rf = r.buildReduce()

	return r
}

// AddChain adds c to the Rainbow Table.
// The array is not immediately sorted.
func (r *Rainbow) AddChain(c ...*Chain) {
	r.chains = append(r.chains, c...)
	r.sorted = false
}

// SortChains will ensure Chains are sorted for Lookup.
// Sort by increasing byte order of chain.End.
// It also dedup chains.
func (r *Rainbow) SortChains() {
	if r.sorted {
		return
	}
	fmt.Println("Sorting chains ...")
	sort.Slice(r.chains, func(i, j int) bool {
		return bytes.Compare(r.chains[i].end, r.chains[j].end) < 0
	})
	r.sorted = true
	// dedup - TODO - not working efficiently enough yet
	// r.dedupChains()
}

func (r *Rainbow) String() string {
	s := "\n============================" +
		"\n   DUMPING RAINBOW TABLE " +
		"\n============================"

	for i, c := range r.chains {
		s += fmt.Sprintf("\n%d\nStart:\t% X\nEnd : \t% X\n", i, c.start, c.end)
	}
	return s + "\n"
}

// Lookup finds the password p that generated the hash h,
// if it exists. Found indicates if found.
func (r *Rainbow) Lookup(h []byte) (p []byte, found bool) {

	var buf []byte
	for depth := 0; depth < r.cl; depth++ {
		buf = append(buf[0:0], h...)

		// compute the chain ending to look for ...
		for i := r.cl - depth; i < r.cl; i++ {
			p = r.rf(i, buf, p)
			buf = r.hf(p, buf)
		}
		// Do we know of such a chain ?
		c, found := r.findChain(buf)
		if found {
			// Now, c might contain the solution ...
			p, found = r.walkChain(c, h)
			if found {
				// got it !
				return p, found
			}
			// too bad, false positive ...
		}
	}
	return nil, false
}

// walkChain walks the chain looking for the password
// that led to the provided hash h.
func (r *Rainbow) walkChain(c *Chain, h []byte) (p []byte, found bool) {
	found = false
	buf := append([]byte{}, c.start...)
	p = make([]byte, 0, r.hsize)
	for i := 0; i < r.cl; i++ {
		p = r.rf(i, buf, p)
		buf = r.hf(p, buf)
		if bytes.Compare(buf, h) == 0 {
			return p, true
		}
	}
	return nil, false
}

// findChain look for the chain given its ending.
func (r *Rainbow) findChain(endHash []byte) (c *Chain, found bool) {
	if !r.sorted {
		r.SortChains()
	}
	for _, c := range r.chains {
		switch bytes.Compare(endHash, c.end) {
		case 0:
			return c, true
		case 1: // loop ...
		case -1: // too far ...
			return nil, false
		}
	}
	return nil, false
}

// dedupChains deduplicate identical chains.
func (r *Rainbow) dedupChains() {

	panic("deprecated and not tested enough")

	/*
		// only dedup when sorted
		if !r.sorted {
			r.SortChains()
			return
		}



		// dedup, maintaining order
		for i := 0; i < len(r.chains); i++ {
			for j := i + 1; j < len(r.chains); j++ {
				if r.chains[i].Equal(r.chains[j]) {
					copy(r.chains[i:], r.chains[i+1:])
					r.chains = r.chains[:len(r.chains)-1]
					j--
					break
				}
			}
		}
	*/
}
