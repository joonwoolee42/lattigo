package mkckks

import (
	"flag"
	"fmt"
	"math"
	"sort"
	"testing"

	"github.com/ldsec/lattigo/v2/ckks"
	"github.com/ldsec/lattigo/v2/ring"
	"github.com/ldsec/lattigo/v2/utils"
	"github.com/stretchr/testify/require"
)

var printPrecisionStats = flag.Bool("print-precision", false, "print precision stats")
var minPrec = 15.0

func Test_MKCKKS(t *testing.T) {

	//skip parameter 4 due to memory consumption
	for i, p := range ckks.DefaultParams {
		if i != 4 && i != 9 {
			testEncryptionEqualsDecryption(t, p)
			testAdd(t, p)
			testAddFourParticipants(t, p)
			testAddPlaintext(t, p)
			testAddPlaintextTwoParticipants(t, p)
			testSub(t, p)
			testSubPlaintext(t, p)
			testNeg(t, p)
			testSubPlaintextTwoParticipants(t, p)
			testMulPlaintext(t, p)
			testMulPlaintextTwoParticipants(t, p)
			//testMul(t, p)
			//testMulFourParticipants(t, p)
		}
	}

}

func testEncryptionEqualsDecryption(t *testing.T, params *ckks.Parameters) {

	sigma := 6.0

	participants := setupPeers(1, params, sigma)

	t.Run(testString("Test encryption equals decryption/", 1, params), func(t *testing.T) {

		// get random value
		value := newTestValue(params, complex(-1, -1), complex(1, 1))

		//encrypt
		cipher := participants[0].Encrypt(value)

		// decrypt
		partialDec := participants[0].GetPartialDecryption(cipher)
		decrypted := participants[0].Decrypt(cipher, []*ring.Poly{partialDec})

		// decode and check
		verifyTestVectors(params, value, decrypted, t)
	})

}

func testAdd(t *testing.T, params *ckks.Parameters) {

	sigma := 6.0

	participants := setupPeers(2, params, sigma)

	t.Run(testString("Test add/", 2, params), func(t *testing.T) {

		// generate new values
		values1 := newTestValue(params, complex(-1, -1), complex(1, 1))
		values2 := newTestValue(params, complex(-1, -1), complex(1, 1))

		// Encrypt
		cipher1 := participants[0].Encrypt(values1)
		cipher2 := participants[1].Encrypt(values2)

		// pad and add
		evaluator := NewMKEvaluator(params)

		out1, out2 := PadCiphers(cipher1, cipher2, params)

		resCipher := evaluator.Add(out1, out2)

		// decrypt
		partialDec1 := participants[0].GetPartialDecryption(resCipher)
		partialDec2 := participants[1].GetPartialDecryption(resCipher)

		decrypted := participants[0].Decrypt(resCipher, []*ring.Poly{partialDec1, partialDec2})

		// perform the operation in the plaintext space
		for i := 0; i < len(values1); i++ {
			values1[i] += values2[i]
		}

		// check results
		verifyTestVectors(params, values1, decrypted, t)
	})

}

func testSub(t *testing.T, params *ckks.Parameters) {

	sigma := 6.0

	participants := setupPeers(2, params, sigma)

	t.Run(testString("Test Subtraction/", 2, params), func(t *testing.T) {

		value1 := newTestValue(params, complex(-1, -1), complex(1, 1))
		value2 := newTestValue(params, complex(-1, -1), complex(1, 1))

		// encrypt
		cipher1 := participants[0].Encrypt(value1)
		cipher2 := participants[1].Encrypt(value2)

		// pad and add
		evaluator := NewMKEvaluator(params)

		out1, out2 := PadCiphers(cipher1, cipher2, params)

		resCipher := evaluator.Sub(out1, out2)

		// decrypt
		partialDec1 := participants[0].GetPartialDecryption(resCipher)
		partialDec2 := participants[1].GetPartialDecryption(resCipher)

		decrypted := participants[0].Decrypt(resCipher, []*ring.Poly{partialDec1, partialDec2})

		// perform the operation in the plaintext space
		// perform the operation in the plaintext space
		for i := 0; i < len(value1); i++ {
			value1[i] -= value2[i]
		}

		// check results
		verifyTestVectors(params, value1, decrypted, t)

	})

}

func testAddPlaintext(t *testing.T, params *ckks.Parameters) {

	sigma := 6.0

	participants := setupPeers(1, params, sigma)

	t.Run(testString("Test Plaintext Addition/", 1, params), func(t *testing.T) {

		value1 := newTestValue(params, complex(-1, -1), complex(1, 1))

		// encrypt
		cipher := participants[0].Encrypt(value1)

		// add plaintext to ciphertext
		evaluator := NewMKEvaluator(params)
		value2 := newTestValue(params, complex(-1, -1), complex(1, 1))

		pt := evaluator.NewPlaintextFromValue(value2)
		resCipher := evaluator.AddPlaintext(pt, cipher)

		// decrypt
		partialDec1 := participants[0].GetPartialDecryption(resCipher)

		decrypted := participants[0].Decrypt(resCipher, []*ring.Poly{partialDec1})

		// perform the operation in the plaintext space
		for i := 0; i < len(value1); i++ {
			value1[i] += value2[i]
		}

		// check results
		verifyTestVectors(params, value1, decrypted, t)
	})

}

func testAddPlaintextTwoParticipants(t *testing.T, params *ckks.Parameters) {

	sigma := 6.0

	participants := setupPeers(2, params, sigma)

	t.Run(testString("Test Plaintext Addition/", 2, params), func(t *testing.T) {

		value1 := newTestValue(params, complex(-1, -1), complex(1, 1))
		value2 := newTestValue(params, complex(-1, -1), complex(1, 1))

		// encrypt
		cipher1 := participants[0].Encrypt(value1)
		cipher2 := participants[1].Encrypt(value2)

		// add plaintext to one of the ciphertext then add both ciphertexts
		evaluator := NewMKEvaluator(params)
		value3 := newTestValue(params, complex(-1, -1), complex(1, 1))

		pt := evaluator.NewPlaintextFromValue(value3)
		resCipher1 := evaluator.AddPlaintext(pt, cipher1)

		out1, out2 := PadCiphers(resCipher1, cipher2, params)

		resCipher := evaluator.Add(out1, out2)

		// decrypt
		partialDec1 := participants[0].GetPartialDecryption(resCipher)
		partialDec2 := participants[1].GetPartialDecryption(resCipher)

		decrypted := participants[0].Decrypt(resCipher, []*ring.Poly{partialDec1, partialDec2})

		// perform the operation in the plaintext space
		for i := 0; i < len(value1); i++ {
			value1[i] += value2[i] + value3[i]
		}

		// check results
		verifyTestVectors(params, value1, decrypted, t)

	})

}

func testSubPlaintext(t *testing.T, params *ckks.Parameters) {

	sigma := 6.0

	participants := setupPeers(1, params, sigma)

	t.Run(testString("Test Plaintext Subtraction/", 1, params), func(t *testing.T) {

		value1 := newTestValue(params, complex(-1, -1), complex(1, 1))

		// encrypt
		cipher := participants[0].Encrypt(value1)

		// sub plaintext to ciphertext
		evaluator := NewMKEvaluator(params)
		value2 := newTestValue(params, complex(-1, -1), complex(1, 1))

		pt := evaluator.NewPlaintextFromValue(value2)
		resCipher := evaluator.SubPlaintext(pt, cipher)

		// decrypt
		partialDec1 := participants[0].GetPartialDecryption(resCipher)

		decrypted := participants[0].Decrypt(resCipher, []*ring.Poly{partialDec1})

		// perform the operation in the plaintext space
		for i := 0; i < len(value1); i++ {
			value1[i] -= value2[i]
		}

		// check results
		verifyTestVectors(params, value1, decrypted, t)

	})

}

func testSubPlaintextTwoParticipants(t *testing.T, params *ckks.Parameters) {

	sigma := 6.0

	participants := setupPeers(2, params, sigma)

	t.Run(testString("Test Plaintext Subtraction/", 2, params), func(t *testing.T) {

		value1 := newTestValue(params, complex(-1, -1), complex(1, 1))
		value2 := newTestValue(params, complex(-1, -1), complex(1, 1))

		// encrypt
		cipher1 := participants[0].Encrypt(value1)
		cipher2 := participants[1].Encrypt(value2)

		// sub plaintext to one of the ciphertext then sub both ciphertexts
		evaluator := NewMKEvaluator(params)
		value3 := newTestValue(params, complex(-1, -1), complex(1, 1))

		pt := evaluator.NewPlaintextFromValue(value3)
		resCipher1 := evaluator.SubPlaintext(pt, cipher1)

		out1, out2 := PadCiphers(resCipher1, cipher2, params)

		resCipher := evaluator.Sub(out1, out2)

		// decrypt
		partialDec1 := participants[0].GetPartialDecryption(resCipher)
		partialDec2 := participants[1].GetPartialDecryption(resCipher)

		decrypted := participants[0].Decrypt(resCipher, []*ring.Poly{partialDec1, partialDec2})

		// perform the operation in the plaintext space
		for i := 0; i < len(value1); i++ {
			value1[i] -= value2[i] + value3[i]
		}

		// check results
		verifyTestVectors(params, value1, decrypted, t)

	})

}

func testNeg(t *testing.T, params *ckks.Parameters) {

	sigma := 6.0

	participants := setupPeers(1, params, sigma)

	t.Run(testString("Test Negation/", 1, params), func(t *testing.T) {

		value1 := newTestValue(params, complex(-1, -1), complex(1, 1))

		// encrypt
		cipher := participants[0].Encrypt(value1)

		// add with negated ciphertext
		evaluator := NewMKEvaluator(params)

		resCipher := evaluator.Add(evaluator.Neg(cipher), cipher)

		// decrypt
		partialDec1 := participants[0].GetPartialDecryption(resCipher)

		decrypted := participants[0].Decrypt(resCipher, []*ring.Poly{partialDec1})

		// perform the operation in the plaintext space
		for i := 0; i < len(value1); i++ {
			value1[i] -= value1[i]
		}

		// check results
		verifyTestVectors(params, value1, decrypted, t)

	})
}

func testAddFourParticipants(t *testing.T, params *ckks.Parameters) {

	sigma := 6.0

	participants := setupPeers(4, params, sigma)

	t.Run(testString("Test add/", 4, params), func(t *testing.T) {

		// generate test values
		values1 := newTestValue(params, complex(-1, -1), complex(1, 1))
		values2 := newTestValue(params, complex(-1, -1), complex(1, 1))
		values3 := newTestValue(params, complex(-1, -1), complex(1, 1))
		values4 := newTestValue(params, complex(-1, -1), complex(1, 1))

		// Encrypt
		cipher1 := participants[0].Encrypt(values1)
		cipher2 := participants[1].Encrypt(values2)
		cipher3 := participants[2].Encrypt(values3)
		cipher4 := participants[3].Encrypt(values4)

		// pad and add in 2 steps
		evaluator := NewMKEvaluator(params)

		out1, out2 := PadCiphers(cipher1, cipher2, params)
		out3, out4 := PadCiphers(cipher3, cipher4, params)

		resCipher1 := evaluator.Add(out1, out2)
		resCipher2 := evaluator.Add(out3, out4)

		finalOut1, finalOut2 := PadCiphers(resCipher1, resCipher2, params)

		resCipher := evaluator.Add(finalOut1, finalOut2)

		// decrypt
		partialDec1 := participants[0].GetPartialDecryption(resCipher)
		partialDec2 := participants[1].GetPartialDecryption(resCipher)
		partialDec3 := participants[2].GetPartialDecryption(resCipher)
		partialDec4 := participants[3].GetPartialDecryption(resCipher)

		decrypted := participants[0].Decrypt(resCipher, []*ring.Poly{partialDec1, partialDec2, partialDec3, partialDec4})

		// perform the operation in the plaintext space
		for i := 0; i < len(values1); i++ {
			values1[i] += values2[i] + values3[i] + values4[i]
		}

		// check results
		verifyTestVectors(params, values1, decrypted, t)
	})

}

func testMulPlaintext(t *testing.T, params *ckks.Parameters) {

	sigma := 6.0

	participants := setupPeers(1, params, sigma)

	t.Run(testString("Test Plaintext Multiplication/", 1, params), func(t *testing.T) {

		value1 := newTestValue(params, complex(-1, -1), complex(1, 1))

		// encrypt
		cipher := participants[0].Encrypt(value1)

		// multiply plaintext and ciphertext
		evaluator := NewMKEvaluator(params)
		value2 := newTestValue(params, complex(-1, -1), complex(1, 1))

		pt := evaluator.NewPlaintextFromValue(value2)
		resCipher := evaluator.MultPlaintext(pt, cipher)

		// decrypt
		partialDec1 := participants[0].GetPartialDecryption(resCipher)

		decrypted := participants[0].Decrypt(resCipher, []*ring.Poly{partialDec1})

		// perform the operation in the plaintext space
		for i := 0; i < len(value1); i++ {
			value1[i] *= value2[i]
		}

		// check results
		verifyTestVectors(params, value1, decrypted, t)

	})

}

func testMulPlaintextTwoParticipants(t *testing.T, params *ckks.Parameters) {

	sigma := 6.0

	participants := setupPeers(2, params, sigma)

	t.Run(testString("Test Plaintext Multiplication/", 2, params), func(t *testing.T) {

		value1 := newTestValue(params, complex(-1, -1), complex(1, 1))
		value2 := newTestValue(params, complex(-1, -1), complex(1, 1))

		// encrypt
		cipher1 := participants[0].Encrypt(value1)
		cipher2 := participants[1].Encrypt(value2)

		// add both ciphertexts then multiply by plaintext
		evaluator := NewMKEvaluator(params)
		value3 := newTestValue(params, complex(-1, -1), complex(1, 1))

		out1, out2 := PadCiphers(cipher1, cipher2, params)

		resCipherTMP := evaluator.Add(out1, out2)

		pt := evaluator.NewPlaintextFromValue(value3)
		resCipher := evaluator.MultPlaintext(pt, resCipherTMP)

		// decrypt
		partialDec1 := participants[0].GetPartialDecryption(resCipher)
		partialDec2 := participants[1].GetPartialDecryption(resCipher)

		decrypted := participants[0].Decrypt(resCipher, []*ring.Poly{partialDec1, partialDec2})

		// perform the operation in the plaintext space
		for i := 0; i < len(value1); i++ {
			value1[i] = (value1[i] + value2[i]) * value3[i]
		}

		// check results
		verifyTestVectors(params, value1, decrypted, t)

	})

}

func testMul(t *testing.T, params *ckks.Parameters) {

	sigma := 6.0

	participants := setupPeers(2, params, sigma)

	t.Run(testString("Test multiplication/", 2, params), func(t *testing.T) {

		// generate test values
		values1 := newTestValue(params, complex(-1, -1), complex(1, 1))
		values2 := newTestValue(params, complex(-1, -1), complex(1, 1))

		// Encrypt
		cipher1 := participants[0].Encrypt(values1)
		cipher2 := participants[1].Encrypt(values2)

		// pad and add in 2 steps
		evaluator := NewMKEvaluator(params)
		evalKeys := []*MKEvaluationKey{participants[0].GetEvaluationKey(), participants[1].GetEvaluationKey()}
		publicKeys := []*MKPublicKey{participants[0].GetPublicKey(), participants[1].GetPublicKey()}

		out1, out2 := PadCiphers(cipher1, cipher2, params)

		resCipher := evaluator.MultRelinDynamic(out1, out2, evalKeys, publicKeys)

		// decrypt
		partialDec1 := participants[0].GetPartialDecryption(resCipher)
		partialDec2 := participants[1].GetPartialDecryption(resCipher)

		decrypted := participants[0].Decrypt(resCipher, []*ring.Poly{partialDec1, partialDec2})

		// perform the operation in the plaintext space
		for i := 0; i < len(values1); i++ {
			values1[i] *= values2[i]
		}

		// check results
		verifyTestVectors(params, values1, decrypted, t)
	})

}

func testMulFourParticipants(t *testing.T, params *ckks.Parameters) {

	sigma := 6.0

	participants := setupPeers(4, params, sigma)

	t.Run(testString("Test multiplication/", 4, params), func(t *testing.T) {

		// generate test values
		values1 := newTestValue(params, complex(-1, -1), complex(1, 1))
		values2 := newTestValue(params, complex(-1, -1), complex(1, 1))
		values3 := newTestValue(params, complex(-1, -1), complex(1, 1))
		values4 := newTestValue(params, complex(-1, -1), complex(1, 1))

		// Encrypt
		cipher1 := participants[0].Encrypt(values1)
		cipher2 := participants[1].Encrypt(values2)
		cipher3 := participants[2].Encrypt(values3)
		cipher4 := participants[3].Encrypt(values4)

		// pad and add in 2 steps
		evaluator := NewMKEvaluator(params)
		evalKeys := []*MKEvaluationKey{participants[0].GetEvaluationKey(), participants[1].GetEvaluationKey(), participants[2].GetEvaluationKey(), participants[3].GetEvaluationKey()}
		publicKeys := []*MKPublicKey{participants[0].GetPublicKey(), participants[1].GetPublicKey(), participants[2].GetPublicKey(), participants[3].GetPublicKey()}

		out1, out2 := PadCiphers(cipher1, cipher2, params)
		out3, out4 := PadCiphers(cipher3, cipher4, params)

		resCipher1 := evaluator.MultRelinDynamic(out1, out2, evalKeys[:2], publicKeys[:2])
		resCipher2 := evaluator.MultRelinDynamic(out3, out4, evalKeys[2:], publicKeys[2:])

		finalOut1, finalOut2 := PadCiphers(resCipher1, resCipher2, params)

		resCipher := evaluator.MultRelinDynamic(finalOut1, finalOut2, evalKeys, publicKeys)

		// decrypt
		partialDec1 := participants[0].GetPartialDecryption(resCipher)
		partialDec2 := participants[1].GetPartialDecryption(resCipher)
		partialDec3 := participants[2].GetPartialDecryption(resCipher)
		partialDec4 := participants[3].GetPartialDecryption(resCipher)

		decrypted := participants[0].Decrypt(resCipher, []*ring.Poly{partialDec1, partialDec2, partialDec3, partialDec4})

		// perform the operation in the plaintext space
		for i := 0; i < len(values1); i++ {
			values1[i] *= values2[i] * values3[i] * values4[i]
		}

		// check results
		verifyTestVectors(params, values1, decrypted, t)
	})

}
func Test_Utils(t *testing.T) {

	s1 := []uint64{0, 2, 1}

	s2 := []uint64{3, 7, 12, 1, 0}

	expected := []uint64{0, 1, 2, 3, 7, 12}

	res := MergeSlices(s1, s2)

	if !equalsSlice(expected, res) {
		t.Errorf("MergeSlices method failed test")
	}

}

// equalsSlice returns true if both slices are equal
func equalsSlice(s1, s2 []uint64) bool {

	if len(s1) != len(s2) {
		return false
	}

	for i, e := range s1 {
		if e != s2[i] {
			fmt.Printf("%d not equal to %d . Error at index %d", e, s2[i], i)
			return false
		}
	}

	return true
}

// equalsPoly returns true if both polynomials are equal
func equalsPoly(p1 *ring.Poly, p2 *ring.Poly) bool {

	if len(p1.Coeffs) != len(p2.Coeffs) {
		return false
	}

	for i, e := range p1.Coeffs {

		if !equalsSlice(e, p2.Coeffs[i]) {
			return false
		}
	}

	return true
}

// Generates keys for a set of peers identified by their peerID using a certain ckks parameter index
// returns the slice of keys with the ckks parameters
func setupPeers(peerNbr uint64, params *ckks.Parameters, sigmaSmudging float64) []MKParticipant {

	res := make([]MKParticipant, peerNbr)

	prng, err := utils.NewKeyedPRNG([]byte{'l', 'a', 't', 't', 'i', 'g', 'o'})

	if err != nil {
		panic(err)
	}

	a := GenCommonPublicParam(params, prng)

	for i := 0; i < int(peerNbr); i++ {

		res[i] = NewParticipant(params, sigmaSmudging, a)
	}

	return res
}

func newTestValue(params *ckks.Parameters, a, b complex128) []complex128 {

	logSlots := params.LogSlots()

	values := make([]complex128, 1<<logSlots)

	for i := uint64(0); i < 1<<logSlots; i++ {
		values[i] = complex(utils.RandFloat64(real(a), real(b)), utils.RandFloat64(imag(a), imag(b)))
	}

	values[0] = complex(0.607538, 0)

	return values
}

func verifyTestVectors(params *ckks.Parameters, valuesWant, valuesTest []complex128, t *testing.T) {

	precStats := GetPrecisionStats(params, valuesWant, valuesTest)

	if *printPrecisionStats {
		t.Log(precStats.String())
	}

	require.GreaterOrEqual(t, real(precStats.MeanPrecision), minPrec)
	require.GreaterOrEqual(t, imag(precStats.MeanPrecision), minPrec)
}

// GetPrecisionStats generates a PrecisionStats struct from the reference values and the decrypted values
func GetPrecisionStats(params *ckks.Parameters, valuesWant, valuesTest []complex128) (prec ckks.PrecisionStats) {

	logSlots := params.LogSlots()
	slots := uint64(1 << logSlots)

	var deltaReal, deltaImag float64

	var delta complex128

	diff := make([]complex128, slots)

	prec.MaxDelta = complex(0, 0)
	prec.MinDelta = complex(1, 1)

	prec.MeanDelta = complex(0, 0)

	cdfResol := 500

	prec.RealDist = make([]struct {
		Prec  float64
		Count int
	}, cdfResol)
	prec.ImagDist = make([]struct {
		Prec  float64
		Count int
	}, cdfResol)

	precReal := make([]float64, len(valuesWant))
	precImag := make([]float64, len(valuesWant))

	for i := range valuesWant {

		delta = valuesTest[i] - valuesWant[i]
		deltaReal = math.Abs(real(delta))
		deltaImag = math.Abs(imag(delta))
		precReal[i] = math.Log2(1 / deltaReal)
		precImag[i] = math.Log2(1 / deltaImag)

		diff[i] += complex(deltaReal, deltaImag)

		prec.MeanDelta += diff[i]

		if deltaReal > real(prec.MaxDelta) {
			prec.MaxDelta = complex(deltaReal, imag(prec.MaxDelta))
		}

		if deltaImag > imag(prec.MaxDelta) {
			prec.MaxDelta = complex(real(prec.MaxDelta), deltaImag)
		}

		if deltaReal < real(prec.MinDelta) {
			prec.MinDelta = complex(deltaReal, imag(prec.MinDelta))
		}

		if deltaImag < imag(prec.MinDelta) {
			prec.MinDelta = complex(real(prec.MinDelta), deltaImag)
		}
	}

	calcCDF(precReal, cdfResol, prec.RealDist)
	calcCDF(precImag, cdfResol, prec.ImagDist)

	prec.MinPrecision = deltaToPrecision(prec.MaxDelta)
	prec.MaxPrecision = deltaToPrecision(prec.MinDelta)
	prec.MeanDelta /= complex(float64(slots), 0)
	prec.MeanPrecision = deltaToPrecision(prec.MeanDelta)
	prec.MedianDelta = calcmedian(diff)
	prec.MedianPrecision = deltaToPrecision(prec.MedianDelta)
	return prec
}

func deltaToPrecision(c complex128) complex128 {
	return complex(math.Log2(1/real(c)), math.Log2(1/imag(c)))
}

func calcCDF(precs []float64, cdfResol int, res []struct {
	Prec  float64
	Count int
}) {
	sortedPrecs := make([]float64, len(precs))
	copy(sortedPrecs, precs)
	sort.Float64s(sortedPrecs)
	minPrec := sortedPrecs[0]
	maxPrec := sortedPrecs[len(sortedPrecs)-1]
	for i := 0; i < cdfResol; i++ {
		curPrec := minPrec + float64(i)*(maxPrec-minPrec)/float64(cdfResol)
		for countSmaller, p := range sortedPrecs {
			if p >= curPrec {
				res[i].Prec = curPrec
				res[i].Count = countSmaller
				break
			}
		}
	}
}

func calcmedian(values []complex128) (median complex128) {

	tmp := make([]float64, len(values))

	for i := range values {
		tmp[i] = real(values[i])
	}

	sort.Float64s(tmp)

	for i := range values {
		values[i] = complex(tmp[i], imag(values[i]))
	}

	for i := range values {
		tmp[i] = imag(values[i])
	}

	sort.Float64s(tmp)

	for i := range values {
		values[i] = complex(real(values[i]), tmp[i])
	}

	index := len(values) / 2

	if len(values)&1 == 1 {
		return values[index]
	}

	if index+1 == len(values) {
		return values[index]
	}

	return (values[index] + values[index+1]) / 2
}
