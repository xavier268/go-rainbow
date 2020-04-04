package rainbow

import (
	"math/big"
)

// a rmodule accptets a big.Int derived from the hash and the previous
// rmodules. It will extract from it (DivMod) a decision, act on the decision,
// thus modifying both the previous hash and the byte array.
// It wil return both values, trying not to allocate,
// and therefore potentially modifying passed content.
type rmodule func(b *big.Int, p []byte) (bb *big.Int, pp []byte)

// RBuilder is a structure that can be compiled and generate a ReduceFunc
// for a specific name space.
type RBuilder struct {
	rms []rmodule
	// flag if already built, you can build only once.
	built bool
	// cumulative size of the big.Int that will be used
	used *big.Int
}

// NewRBuilder constructs and initialize an RBuilder
func NewRBuilder() *RBuilder {
	rm := new(RBuilder)
	rm.used = new(big.Int).SetInt64(1)
	return rm
}

// Build builds a new ReduceFunction from the RMBuilder.
func (rm *RBuilder) Build() ReduceFunction {
	if rm.built {
		panic("build was already called")
	}
	if len(rm.rms) == 0 {
		panic("cannot build : not rmodules were compiled yet")
	}
	rm.built = true

	// allocate initial capacity outside ReduceFunction
	var buf [64]byte
	bi := (&big.Int{}).SetBytes(buf[:])
	st := (&big.Int{})

	return func(step int, h, p []byte) []byte {

		// prepare bi, handling the step
		st.SetInt64(int64(step))
		bi.SetBytes(h)
		bi.Add(bi, st)

		// reset password, keeping capacity
		p = p[:0]

		// apply the various rmodule
		for _, f := range rm.rms {
			bi, p = f(bi, p)
		}
		// return the last password generated
		return p
	}
}

// CompileAlphabet will compile an alpbet of runes (a string).
// It will append to the password, ensuring lenghth
// is between min(included) and max(included) runes.
func (rm *RBuilder) CompileAlphabet(alphabet string, min, max int) *RBuilder {

	if len(alphabet) == 0 || max < min || max <= 0 || min < 0 {
		panic("invalid input parameters")
	}

	// prepocess alphabet
	alp := make([][]byte, 0, len(alphabet))
	for _, r := range alphabet {
		// ranging rune by rune ...
		alp = append(alp, []byte(string(r)))
	}

	// allocate memory and "constants"
	buf := new(big.Int).SetInt64(10000)
	ns := new(big.Int).SetInt64(int64(max - min + 1)) // size choice
	n := new(big.Int).SetInt64(int64(len(alp)))       // letter choice

	// update used capacity
	rm.used.Mul(rm.used, ns)
	for i := 0; i < max; i++ {
		rm.used.Mul(rm.used, n)
	}

	//fmt.Println("Used capacity so far : ", rm.used)

	// append the rmodule
	rm.rms = append(rm.rms,
		func(b *big.Int, p []byte) (*big.Int, []byte) {
			var v, s int

			// decide on the size
			b, v = extract(b, ns, buf)
			// TODO - adjust calculation to ensure a better distribution.
			// Currently, too many smaller since probability is the same
			// for the full size group.
			// Smaller size should be selected less often ?
			// A distribution fonction should be provided ?
			s = v + min
			// ensure minimum values are set
			for i := 0; i < s; i++ {
				b, v = extract(b, n, buf)
				p = append(p, alp[v]...)
			}
			return b, p
		})
	// return the updated RBuilder
	return rm
}

// extract from a big int a value from 0 to (n-1),
// returning the new big.Int and the extracted value.
// v is a big.Int passed to hold v, and avoid allocation.
// It will be overwritten.
// No control is made that the big.Int is big enough
// and that the extraction is significant.
func extract(b *big.Int, n *big.Int, v *big.Int) (bb *big.Int, vv int) {
	b, v = b.DivMod(b, n, v)
	return b, int(v.Int64())
}
