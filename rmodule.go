package rainbow

import (
	"bufio"
	"fmt"
	"os"
)

// a rmodule define the intermediate step for the reduce function.
type rmodule struct {
	// a rmodule is provided a subslice from the hash and the previous
	// rmodules. It will use it to make its decisions,
	// thus modifying the password p byte array.
	// It will try not to allocate,
	// and therefore potentially modifying passed content.
	run func(entropy, p []byte) (pp []byte)
	// number of bytes of entropy expected by this module
	bytes int
	// signature of the rmodule
	signature string
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

	r.used = 0
	for _, m := range r.rms {
		r.used += m.bytes
		r.signature = m.signature + "\n"
	}

	r.built = true

	return func(step int, h, p []byte) []byte {

		// autoextend hash size
		for r.used > len(h) {
			h = append(h, h...)
		}

		// merge the step into the hash
		for i := range h {
			h[i] = byte(int(h[i]) + step*(i+1))
		}

		// reset password, keeping capacity
		p = p[:0]
		// byte index
		bi := 0

		// apply the various rmodule
		for _, m := range r.rms {

			// extract needed entropy from current pointer
			ent := extractEntropy(h, bi, bi+m.bytes)

			// apply module
			p = m.run(ent, p)

			//fmt.Printf("Entropy before module %d= %d->%s\n", i, ent, string(p))

			// refresh bit index
			bi += m.bytes
		}
		// return the last password generated
		return p
	}
}

// extract 'nb' bytes starting at the 'from' position,
// returning the corresponding uint64
func extractEntropy(h []byte, from, nb int) []byte {
	if from+nb > len(h) {
		// should not happen (auto extension of h)
		panic("attempting to extract beyond h length")
	}
	return h[from : from+nb]
}

// CompileAlphabet will compile an alpbabet of runes (a string).
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

	if len(alp) >= 255 {
		panic("alphabet should not exceed 256 signs")
	}

	// create rmodule
	mod := new(rmodule)
	mod.signature = fmt.Sprintf("CompileAlphabet with alphabet : %v, min:%d, max%d\n", alphabet, min, max)
	mod.bytes = (max + 1) // 1 for the size and one per letter
	mod.run = func(ent, p []byte) []byte {
		var s, v int

		// decide on the size, s
		if max > min {
			s = min + int(ent[0])%(max-min+1)
		} else {
			s = min
		}

		// append values to p
		for i := 0; i < s; i++ {
			v = int(ent[i+1]) % len(alp)
			p = append(p, alp[v]...)
		}
		return p
	}

	// append the rmodule
	r.rms = append(r.rms, mod)

	return r
}

// CompileTransform compile the password transfarmation, selecting one among all transformation.
// One or more alternative can be nil.
func (r *Rainbow) CompileTransform(trf ...func(p []byte) []byte) *Rainbow {

	if len(trf) > 255 {
		panic("too many alternative transformation provided - max 255")
	}

	if len(trf) == 0 {
		panic("there should be at least one transformation, possibly nil")
	}

	mod := new(rmodule)
	mod.signature = fmt.Sprintf("CompileTransform with %d transformations", len(trf))

	mod.bytes = 1
	mod.run = func(ent, p []byte) []byte {
		// decide on transformation to use
		v := int(ent[0]) % len(trf)

		// transform if selection is not nil
		if trf[v] != nil {
			p = trf[v](p)
		}
		// else, do nothing
		return p
	}

	// add the module
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
	mod.signature = fmt.Sprintf("CompileWordList with %d words", len(words))
	sz := 1
	mod.bytes = 1
	for {
		mod.bytes++
		sz *= 256
		if sz > len(words) {
			break
		}
	}

	mod.run = func(ent, p []byte) []byte {

		// Select the word
		var v uint64
		for _, e := range ent {
			v = 256*v + uint64(e)
		}
		v = v % uint64(len(words))

		// append selected word
		p = append(p, words[v]...)

		// return
		return p
	}

	// append the module
	r.rms = append(r.rms, mod)

	return r
}
