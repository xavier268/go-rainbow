package rainbow

import (
	"bytes"
	"crypto"
	"fmt"
	"math/rand"
	"sort"
	"time"
)

// Version of the package
func Version() (major, minor, sub int) {
	return 0, 6, 2
}

// VersionString for human consumption
func VersionString() string {
	M, m, s := Version()
	return fmt.Sprintf("version_%d.%d.%d", M, m, s)
}

// Rainbow is the main type to generate tables or lookup a password.
type Rainbow struct {
	// a human readable signature of the configuration
	signature string
	// hashing algorithm
	halgo crypto.Hash
	// hf is the Hash function used to compute from password to hash
	hf HashFunction
	// hsize is the size of the HashFunction result in  bytes
	hsize int

	// rf is a reduce function, from hash to password
	// It conforms to the hash.Hash interface.
	rf ReduceFunction
	// number of bytes of entropy consumed by the reduce function
	used int

	// cl is the chain length (constant)
	cl int

	// Random generator
	rand *rand.Rand
	// chains
	chains []*Chain
	// Are the chains sorted ?
	sorted bool

	// The modules used to build the reduce name space
	rms []*rmodule
	// flag : you can build only once.
	built bool
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

// Build finish compiling the Rainbow table "reduce" function.
func (r *Rainbow) Build() *Rainbow {

	if r.rf != nil {
		panic("a reduce function was already defined, you cannot redefine it")
	}

	r.signature = fmt.Sprintf("go-rainbow %s\nchain length %d\nhash algorithm : %d\n",
		VersionString(), r.cl, r.halgo)

	// update the reduce function
	r.rf = r.buildReduce()

	r.signature += fmt.Sprintf("used bytes %d", r.used)

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
		// Do we know of such  chains ?
		from, to, found := r.findChain(buf)
		if found {
			// loop onpotential candidates ...
			for i := from; i < to; i++ {
				c := r.chains[i]
				// Now, c might contain the solution ?
				p, found = r.walkChain(c, h)
				if found {
					// got it !
					return p, found
				}
				// false positive,
				// check other matching chains ...
			}
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

// findChain look for the chains given its ending.
// return the index of the matching chain, from (included) to (excluded)
func (r *Rainbow) findChain(endHash []byte) (from, to int, found bool) {
	if !r.sorted {
		r.SortChains()
	}

	from = -1
	to = -1
	for i, c := range r.chains {
		// look for lower bound ...
		if from < 0 {
			switch bytes.Compare(endHash, c.end) {
			case 0:
				from = i
				to = from + 1
			case 1: // not yet, search again
			case -1: // too far, nothing found
				return from, to, false
			}
		} else {
			switch bytes.Compare(endHash, c.end) {
			case 0:
				to = i + 1
			default:
				return from, to, true
			}
		}
	}
	return from, to, from >= 0 && to >= 0
}

// DedupChains deduplicate identical chains.
// Heavy operation, needs to be triggered manually.
func (r *Rainbow) DedupChains() {

	// dedup will potentially change order
	r.sorted = false

	fmt.Print("Dedup from ", len(r.chains))

	// remove duplicates one by one
	for i := 0; i < len(r.chains); i++ {
		for j := i + 1; j < len(r.chains); j++ {
			if r.chains[i].Equal(r.chains[j]) {
				r.chains[i] = r.chains[len(r.chains)-1]
				r.chains = r.chains[:len(r.chains)-1]
				j--
				i--
				break
			}
		}
	}

	fmt.Println(" to ", len(r.chains))

	// sort back
	r.SortChains()

}
