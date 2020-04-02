// Package bfv implements a RNS-accelerated Fan-Vercauteren version of Brakerski's scale invariant homomorphic encryption scheme. It provides modular arithmetic over the integers.
package bfv

import (
	"github.com/ldsec/lattigo/ring"
)

// GaloisGen is an integer of order N/2 modulo M and that spans Z_M with the integer -1. The j-th ring automorphism takes the root zeta to zeta^(5j).
const GaloisGen uint64 = 5

// bfvContext is a struct which contains all the elements required to instantiate the BFV Scheme. This includes the parameters (polynomial degree, plaintext modulus, ciphertext modulus,
// Gaussian sampler, polynomial contexts and other parameters required for the homomorphic operations).
type bfvContext struct {
	params *Parameters

	// Polynomial degree
	n uint64

	gaussianSampler *ring.KYSampler

	// Polynomial contexts
	contextT    *ring.Context // Plaintext modulus
	contextQ    *ring.Context // Ciphertext modulus
	contextQMul *ring.Context // Ciphertext extended modulus (for multiplication)
	contextP    *ring.Context // Keys additional modulus
	contextQP   *ring.Context // Keys modulus

	galElRotRow      uint64   // Rows rotation generator
	galElRotColLeft  []uint64 // Columns right rotations generators
	galElRotColRight []uint64 // Columsn left rotations generators
}

func newBFVContext(params *Parameters) (context *bfvContext) {

	if !params.isValid {
		panic("cannot newBFVContext: params not valid (check if they were generated properly)")
	}

	context = new(bfvContext)
	var err error

	LogN := params.LogN
	N := uint64(1 << LogN)

	context.n = N

	if context.contextT, err = ring.NewContextWithParams(N, []uint64{params.T}); err != nil {
		panic(err)
	}

	if context.contextQ, err = ring.NewContextWithParams(N, params.Qi); err != nil {
		panic(err)
	}

	if context.contextQMul, err = ring.NewContextWithParams(N, params.QiMul); err != nil {
		panic(err)
	}

	if len(params.Pi) != 0 {
		if context.contextP, err = ring.NewContextWithParams(N, params.Pi); err != nil {
			panic(err)
		}
	}

	if context.contextQP, err = ring.NewContextWithParams(N, append(params.Qi, params.Pi...)); err != nil {
		panic(err)
	}

	context.gaussianSampler = context.contextQP.NewKYSampler(params.Sigma, int(6*params.Sigma))

	context.galElRotColLeft = ring.GenGaloisParams(N, GaloisGen)
	context.galElRotColRight = ring.GenGaloisParams(N, ring.ModExp(GaloisGen, 2*N-1, 2*N))
	context.galElRotRow = 2*N - 1
	return
}
