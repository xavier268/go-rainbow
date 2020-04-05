package rainbow

import (
	"crypto"
	"fmt"
	"math"
	"math/big"
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
		b := new(big.Int).Rand(r.rand, BIGDIV)
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
