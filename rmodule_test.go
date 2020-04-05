package rainbow

import (
	"crypto"
	"fmt"
	"math"
	"math/big"
	"math/rand"
	"strings"
	"testing"
)

func TestExtract(t *testing.T) {

	bb := new(big.Int).SetInt64(62)
	n := new(big.Int).SetInt64(3)
	v := new(big.Int)
	vv := 0

	bb, vv = extract(bb, n, v)
	if vv != 1 && bb.Cmp(new(big.Int).SetInt64(20)) != 0 {
		t.FailNow()
	}

	bb, vv = extract(bb, n, v)
	if vv != 2 && bb.Cmp(new(big.Int).SetInt64(6)) != 0 {
		t.FailNow()
	}
}

func TestRBuilder1(t *testing.T) {
	min, max := 2, 5
	red := New(crypto.MD5, 10).CompileAlphabet("Aéi", min, max).buildReduce()
	p, h := []byte{}, []byte{2, 5, 12, 6, 54, 44, 55, 89, 7, 65, 46, 5, 4}
	results := make(map[string]int)
	for i := 0; i < 10_000; i++ {
		p = red(i, h, p)
		results[string(p)]++
	}
	if len(results) != 3*3+3*3*3+3*3*3*3+3*3*3*3*3 { // 360
		fmt.Printf("min %d max %d results len %d\n%+v\n", min, max, len(results), results)
		t.Fatal("unexpected result length")
	}
}

func TestRBuilder2(t *testing.T) {
	min, max := 3, 3
	red := New(crypto.MD5, 10).CompileAlphabet("aécd", min, max).buildReduce()
	p, h := []byte{}, []byte{2, 5, 12, 6, 54, 44, 55, 89, 7, 65, 46, 5, 4}
	results := make(map[string]int)
	for i := 0; i < 10_000; i++ {
		p = red(i, h, p)
		results[string(p)]++
	}
	if len(results) != 4*4*4 {
		fmt.Printf("min %d max %d results len %d\n%+v\n", min, max, len(results), results)
		t.Fatal("unexpected result length")
	}
}

func BenchmarkRBuilderAlphabet(b *testing.B) {
	min, max := 7, 12
	red := New(crypto.MD5, 10).CompileAlphabet("abcdefghijklmnopqrstuvwxyz", min, max).buildReduce()
	p, h := []byte{}, []byte{2, 5,
		12, 6, 54, 44, 55, 89, 7,
		12, 6, 54, 44, 55, 89, 7,
		12, 6, 54, 44, 55, 89, 7,
		12, 6, 54, 44, 55, 89, 7,
		65, 46, 5, 4}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		p = red(i, h, p)
	}

}

func TestExtractf(t *testing.T) {
	r := New(crypto.MD5, 10)

	buf := new(big.Int)
	var v, s, ss float64
	n := 1_000_000.
	for i := 0.; i < n; i++ {
		b := new(big.Int).Rand(r.rand, largeConstantAsBig)
		v = extractf(b, buf)
		s += v
		ss += v * v
	}
	s = s / n
	ss = ss/n - s*s
	if math.Abs(s-0.5) > 0.01 || math.Abs(ss-1./12.) > 0.001 {
		fmt.Printf("Mean  : %f\t expected : %f\nVariance : %f\t expected : %f\n", s, 0.5, ss, 1./12.)
		t.Fatal("unrealistic mean or variance")
	}
}

func BenchmarkExtractf(b *testing.B) {
	r := New(crypto.MD5, 10)

	buf := new(big.Int)
	var v float64
	bb := new(big.Int).Rand(r.rand, new(big.Int).SetInt64(10000))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v = extractf(bb, buf)
	}
	if v == 0 {
	}
}

func BenchmarkExtract(b *testing.B) {
	r := New(crypto.MD5, 10)

	buf := new(big.Int)
	var v int
	bb := new(big.Int).Rand(r.rand, new(big.Int).SetInt64(3213213131313131331))
	bbb := new(big.Int).Set(bb)
	vv := new(big.Int).Set(bb)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bb.Set(bb)
		buf.Set(bb)
		buf, v = extract(bbb, buf, vv)
	}
	if v == 0 {
	}
}

func TestVisualCompileWords(t *testing.T) {

	r := New(crypto.MD5, 10).
		CompileWordList("words_test.txt").
		CompileAlphabet("0123456789", 0, 2).
		Build()

	n := 20
	h := []byte{1, 3, 55, 6, 4, 44, 55, 88, 99, 77, 22, 33, 11, 121, 65, 4, 5, 5, 55, 4, 4}
	fmt.Println("==============================")
	fmt.Println("List of words concatened with 0 to 2 digits")
	for i := 0; i < n; i++ {
		p := r.rf(i, h, []byte{})
		fmt.Println(string(p))
	}
	fmt.Println("==============================")
}

func TestVisualCompileTransform(t *testing.T) {

	trf := func(p []byte) []byte {
		return []byte("**" + strings.ToUpper(string(p)) + "**")
	}

	r := New(crypto.MD5, 10).
		CompileWordList("words_test.txt").
		CompileTransform(trf, 1./3.).
		Build()

	n := 1000
	h := make([]byte, 16)
	rand.Read(h)

	capi := 0

	fmt.Println("List of words, 1/3rd of them capitalized ")
	for i := 0; i < n; i++ {
		p := r.rf(i, h, []byte{})
		if p[0] == byte('*') {
			capi++
		}
	}
	ratio := float64(capi) / float64(n)
	fmt.Printf("Capitalized Actual : %2.1f%%\t Target : %2.1f%%\n", 100.*ratio, 100./3.)
	if ratio < .3 || ratio > 0.35 {
		t.Fatal("unexpected ratio of capitalized letters")
	}

}
