package rainbow

import (
	"bytes"
	"fmt"
	"math/rand"
	"sort"
)

// Rainbow is the main type to generate tables or lookup a password.
type Rainbow struct {
	// H is the Hash function used to compute from password to hash
	H HashFunc
	// R is a reduce function, from hash to password
	// It conforms to the hash.Hash interface.
	R ReduceFunc
	// Cl is the chain length (constant)
	Cl int
	// HSize is the size of the HashFunction result in  bytes
	HSize int
	// Random generator
	Rand *rand.Rand
	// Chains
	Chains []*Chain
	// Are the chains sorted ?
	sorted bool
}

// ReduceFunc is a function that reduces a hash
// into a password, the next in the chain.
// The password value is returned as a byte slice.
// No allocations are made.
type ReduceFunc func(step int, hash []byte, password []byte) []byte

// HashFunc is a function that hashes a password into a hash
// hash is put into the provided hash slice, returning it.
// No allcations are made.
type HashFunc func(passwd []byte, hash []byte) []byte

// NewRainbow constructor
func NewRainbow(H HashFunc, R ReduceFunc, chlength int) *Rainbow {
	r := new(Rainbow)
	r.Cl = chlength
	r.H, r.R = H, R
	return r
}

// Chain in the table, contains the start and end hash values
type Chain struct {
	Start []byte
	End   []byte
}

// NewChain builds a new Chain, from random start.
// It is not immediateley added to the Chains slice in r.
// See - AddChain below.
func (r *Rainbow) NewChain() *Chain {
	c := new(Chain)
	c.Start = make([]byte, r.HSize, r.HSize)
	r.Rand.Read(c.Start)
	c.End = append([]byte{}, c.Start...)
	p := []byte{}
	for i := 0; i < r.Cl; i++ {
		p = r.R(i, c.End, p)
		c.End = r.H(p, c.End)
		fmt.Println(string(p), "-->", c.End)
	}
	return c
}

// AddChain adds c to the Rainbow Table.
// The array is not immediately sorted.
func (r *Rainbow) AddChain(c ...*Chain) {
	r.Chains = append(r.Chains, c...)
	r.sorted = false
}

// SortChains will ensure Chains are sorted for Lookup.
// Sort by increasing byte order of chain.End.
func (r *Rainbow) SortChains() {
	if r.sorted {
		return
	}
	sort.Slice(r.Chains, func(i, j int) bool {
		return bytes.Compare(r.Chains[i].End, r.Chains[j].End) < 0
	})
}

func (r *Rainbow) String() string {
	s := "\n============================" +
		"\n   DUMPING RAINBOW TABLE " +
		"\n============================"

	for i, c := range r.Chains {
		s += fmt.Sprintf("\n%d\nStart:\t%v\nEnd : \t%v\n", i, c.Start, c.End)
	}
	return s + "\n"
}

// Lookup finds the password p that generated the hash h,
// if it exists. Found indicates if found.
func (r *Rainbow) Lookup(h []byte) (p []byte, found bool) {

	buf := make([]byte, r.HSize, r.HSize)
	for depth := 0; depth < r.Cl; depth++ {
		buf = append(buf[:0], h...)

		// compute the chain ending to look for ...
		for i := r.Cl - depth; i < r.Cl; i++ {
			p = r.R(i, buf, p)
			buf = r.H(p, buf)
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
	buf := append([]byte{}, c.Start...)
	p = []byte{}
	for i := 0; i < r.Cl; i++ {
		p = r.R(i, buf, p)
		buf = r.H(p, buf)
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
	for _, c := range r.Chains {
		switch bytes.Compare(endHash, c.End) {
		case 0:
			return c, true
		case 1: // loop ...
		case -1: // too far ...
			return nil, false
		}
	}
	return nil, false
}
