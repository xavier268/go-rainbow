package rainbow

// GetAlphaReduceFunc creates a ReduceFunc that generates
// alphabetic loawercases passwords of the exact specified length.
func GetAlphaReduceFunc(nbOfChar int) ReduceFunc {
	if nbOfChar > 16 {
		panic("GetAlphaReduceFunc only accepts up to 16 letters length passwords")
	}
	return func(step int, h, p []byte) (pp []byte) {
		pp = p[:0]
		for i := 0; i < nbOfChar; i++ {
			pp = append(pp, 'a'+(byte(step*11)+h[i])%26)
		}
		return pp
	}
}
