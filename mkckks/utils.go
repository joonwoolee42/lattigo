package mkckks

import (
	"sort"

	"github.com/ldsec/lattigo/v2/ckks"
	"github.com/ldsec/lattigo/v2/ring"
	"github.com/ldsec/lattigo/v2/utils"
)

// Dot computes the dot product of two decomposed polynomials in ring^d and store the result in res
func Dot(p1 *MKDecomposedPoly, p2 *MKDecomposedPoly, res *ring.Poly, r *ring.Ring) {
	if len(p1.poly) != len(p2.poly) {
		panic("Cannot compute dot product on vectors of different size !")
	}

	for l := uint64(0); l < uint64(len(p1.poly)); l++ {
		r.MulCoeffsMontgomeryAndAdd(p1.poly[l], p2.poly[l], res)
	}

}

// DotLvl computes the dot product of two decomposed polynomials in ringQ^d up to q_level and store the result in res
func DotLvl(level uint64, p1 *MKDecomposedPoly, p2 *MKDecomposedPoly, res *ring.Poly, r *ring.Ring) {
	if len(p1.poly) != len(p2.poly) {
		panic("Cannot compute dot product on vectors of different size !")
	}

	for l := uint64(0); l < uint64(len(p1.poly)); l++ {
		r.MulCoeffsMontgomeryAndAddLvl(level, p1.poly[l], p2.poly[l], res)
	}

}

// MergeSlices merges two slices of uint64 and places the result in s3
// the resulting slice is sorted in ascending order
func MergeSlices(s1, s2 []uint64) []uint64 {

	s3 := make([]uint64, len(s1))

	copy(s3, s1)

	for _, el := range s2 {

		if Contains(s3, el) < 0 {
			s3 = append(s3, el)
		}
	}

	sort.Slice(s3, func(i, j int) bool { return s3[i] < s3[j] })

	return s3
}

// Contains return the element's index if the element is in the slice. -1 otherwise
func Contains(s []uint64, e uint64) int {

	for i, el := range s {
		if el == e {
			return i
		}
	}
	return -1
}

// GetRandomPoly samples a polynomial with a uniform distribution in the given ring
func GetRandomPoly(params *ckks.Parameters, r *ring.Ring) *ring.Poly {

	prng, err := utils.NewPRNG()
	if err != nil {
		panic(err)
	}

	return GetUniformSampler(params, r, prng).ReadNew()
}