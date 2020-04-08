package rainbow

import (
	"bufio"
	"math/bits"
	"os"
)

// a rmodule define the intermediate step for the reduce function.
type rmodule struct {
	// a rmodule is provided a uint64 derived from the hash and the previous
	// rmodules. It will extract from its decisions,
	// thus modifying the byte array.
	// It will try not to allocate,
	// and therefore potentially modifying passed content.
	run func(entropy uint64, p []byte) (pp []byte)
	// number of bits needed by this module
	bits int
}

// buildReduce builds a new ReduceFunction from the RMBuilder.
// Will be called by the Rainbow Build, not to be called directly.
func (r *Rainbow) buildReduce() ReduceFunction {
	if r.built {
		panic("build was already called")
	}
	if len(r.rms) == 0 {
		panic("cannot build : no rmodules were compiled yet")
	}

	ttlbits := 0
	for _, m := range r.rms {
		ttlbits += m.bits
	}
	if ttlbits >= 8*r.hsize {
		panic("too many entropy required versus available hash size")
	}
	r.usedBits = ttlbits

	r.built = true

	return func(step int, h, p []byte) []byte {

		// merge the step into the hash
		for i := range h {
			switch i & 3 {
			case 0:
				h[i] ^= byte(step)
			case 1:
				h[i] ^= byte(19 * step)
			case 2:
				h[i] ^= byte(571 * step)
			case 3:
				h[i] ^= byte(1093 * step)
			default:
				panic("arithmetic internal error")
			}
		}

		// reset password, keeping capacity
		p = p[:0]
		// bit index
		bi := 0

		// apply the various rmodule
		for _, m := range r.rms {

			// extract needed entropy from current pointer
			ent := extractEntropy(h, bi, bi+m.bits)

			// apply module
			p = m.run(ent, p)

			// refresh bit index
			bi += m.bits
		}
		// return the last password generated
		return p
	}
}

// extract 'nb' bits starting at the 'from' position,
// returning the corresponding uint64
func extractEntropy(h []byte, from, nb int) uint64 {
	if nb > 63 {
		panic("trying to extract more than would fit in a uint64")
	}
	if from+nb >= 8*len(h) {
		panic("attempting to extract bits beyond input h length")
	}
	var acc, mask uint64
	// first non aligned bits
	if from%8 != 0 {
		mask = 2<<(from%8) - 1
		acc = uint64(h[from/8]) & mask
	}
	// aligned bits
	for i := 1 + from/8; i < (from+nb)/8; i++ {
		acc = acc<<8 + uint64(h[i])
	}
	if (from+nb)%8 != 0 {
		// last non aligned bits
		mask = 2<<((nb+from)%8) - 1
		acc = acc<<8 + uint64(h[(nb+from)/8])&mask
	}
	return acc
}

// CompileAlphabet will compile an alpbet of runes (a string).
// It will append to the password, ensuring lenghth
// is between min(included) and max(included) runes.
func (r *Rainbow) CompileAlphabet(alphabet string, min, max int) *Rainbow {

	if len(alphabet) == 0 || max < min || max <= 0 || min < 0 {
		panic("invalid input parameters")
	}

	// preprocess alphabet
	alp := make([][]byte, 0, len(alphabet))
	for _, r := range alphabet {
		// ranging rune by rune ...
		alp = append(alp, []byte(string(r)))
	}
	alpl := uint64(len(alp))
	// create rmodule
	mod := new(rmodule)
	bb := uint64(1)
	for i := 0; i < max; i++ {
		bb *= alpl
	}
	mod.bits = bits.Len64(bb) + bits.Len64(uint64(max-min))

	mod.run = func(ent uint64, p []byte) []byte {
		var s int
		var v uint64

		// decide on the size, s
		if max > min {
			s = min + int(ent)%(max-min)
			v = v >> bits.Len64(uint64(max-min))
		} else {
			s = min
		}

		// append values to p
		for i := 0; i < s; i++ {
			v = ent % alpl
			p = append(p, alp[v]...)
			v = v / alpl
		}
		return p
	}

	// append the rmodule
	r.rms = append(r.rms, mod)

	return r
}

// CompileTransform compile the password transfarmation, with provided probability.
// 0.0 - means never, and 1.0 means always.
func (r *Rainbow) CompileTransform(trf func(p []byte) []byte, probability float64) *Rainbow {
	if probability < 0 || probability > 1 {
		panic("probability needs to be between 0.0 and 1.0")
	}
	if trf == nil {
		panic("transformation fonction must be provided, cannot be nul")
	}

	var pp []byte
	var prob = probability

	mod := new(rmodule)

	mod.bits = 1 // shift only by 1 bit for each choice
	mod.run = func(ent uint64, p []byte) []byte {
		// decide on probability
		v := float64(ent%1000) / 1000. // precision is 0.1% probability

		if v < prob {
			pp = trf(p)
			return pp
		}
		// else, do nothing
		return p
	}

	// add the module, that would call trf
	r.rms = append(r.rms, mod)

	return r

}

// CompileWordList will load the word list from file,
// store it in memory, and select one on every reduce operation,
// that will be appended to the current password.
// The words are listed in the provided file, space separated,
// as per the ScanWords SplitFunction.
func (r *Rainbow) CompileWordList(fName string) *Rainbow {

	// Open file
	f, err := os.Open(fName)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	// scan and store words in memory
	var words [][]byte
	scanner := bufio.NewScanner(f)
	scanner.Split(bufio.ScanWords)
	for scanner.Scan() {
		words = append(words, scanner.Bytes())
	}

	mod := new(rmodule)
	mod.bits = bits.Len64(uint64(len(words)))

	mod.run = func(ent uint64, p []byte) []byte {

		// Select the word
		v := ent % uint64(len(words))

		// append selected word
		p = append(p, words[v]...)

		// return
		return p
	}

	// append the module
	r.rms = append(r.rms, mod)

	return r
}

/* DEPRECATED
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

// largeConstantAsBig is a large number constant used to generate a float with extractf
var largeConstantAsBig = new(big.Int).SetInt64(largeConstant)
var largePrimeAsBig = new(big.Int).SetInt64(largePrime)

const largeConstant = 1_000_000_000
const largePrime = 2_147_483_647

// return a float uniformely distributed between 0 and 1
func extractf(b *big.Int, buf *big.Int) float64 {
	buf = buf.Mul(b, largePrimeAsBig)
	buf = buf.Mod(buf, largeConstantAsBig)
	return float64(buf.Int64()) / float64(largeConstant)
}

*/
