package rainbow

import "math/big"

// getAlphaReduceFunc creates a ReduceFunction that generates
// alphabetic lowercases passwords of the exact specified length,
// within the limit of the number of bytes from the hash.
func getAlphaReduceFunc(nbOfChar int) ReduceFunction {
	return GetStringReduceFunc(nbOfChar, "abcdefghijklmnopqrstuvwxyz", true)
}

// GetStringReduceFunc will generate a reduce function that makes words
// of the specified nbRune size (exactsize or maximimum  size,
// depending on the boolean).
// IMPORTANT : we specify a number of rune, not of bytes !
func GetStringReduceFunc(nbRunes int, alphabet string, exactsize bool) ReduceFunction {
	runes := []rune(alphabet)
	if len(alphabet) >= 255 {
		panic("alphabet byte length should not equal or exceed 255")
	}
	mod := (&big.Int{}).SetInt64(int64(len(runes)))
	if !exactsize { // one more, pseudo char, to signal to ignore
		mod.Add(mod, (&big.Int{}).SetInt64(1))
	}
	remain := &big.Int{}
	z := &big.Int{}
	return func(step int, h, p []byte) (pp []byte) {
		pp = p[:0] // reusing memory ...
		var s int
		z.SetBytes(h)
		for i := 0; i < nbRunes; i++ {
			if i%64 == 0 {
				s = int(step)
			}
			z.DivMod(z, mod, remain)
			rr := int(remain.Int64())
			if exactsize {
				rr = (s + rr) % len(runes)
			} else {
				rr = (s + rr) % (len(runes) + 1)
			}
			if rr < len(runes) {
				// here, we append a new rune, pseudo char is ignored
				// s is used to rate the runes within the alphabet.
				pp = append(pp, []byte(string(runes[rr]))...)
			}
			// slower shifting, the modulo can be very small
			s = s >> 1
		}
		return pp
	}
}

// GetBaseReduceFunc returns a ReduceFunction that has the same value space
// as the hash function. It cannot just be the identity, because it needs
// to vary with the step, to avoid cycles.
//
// Note on the step computation applied :
// The computation garantees that the period w.r.t. the step is
// at least the size of the hash space and at max the maxint value.
// It also ensure bytes beyond the first are also touched, to maximize
// entropy.
func GetBaseReduceFunc() ReduceFunction {
	return func(step int, h, p []byte) (pp []byte) {
		var s uint64
		pp = p[:0]
		for i := 0; i < len(h); i++ {
			if i%8 == 0 {
				s = uint64(step)
			}
			pp = append(pp, byte(s)+h[i])
			s = s >> 8
		}
		return pp
	}
}
