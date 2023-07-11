package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"math"
	"math/big"
)

var (
	maxNonce = math.MaxInt64
)

const targetBits = 24 // defines difficulty of mining, using constant instead of target adjusting algorithm

type ProofOfWork struct {
	block *Block
	target *big.Int
}

func NewProofOfWork(b *Block) *ProofOfWork {
	target := big.NewInt(1) // initialize big.Int
	target.Lsh(target, uint(256-targetBits)) //shift by 256 - targetBits

	pow := &ProofOfWork{b, target}
	return pow
}

// prepare data to hash
// nonce =  counter to introduce randomness in hash generation
func (pow *ProofOfWork) prepareData(nonce int) []byte {
	data := bytes.Join(
		[][]byte{ // merge block fields with the target and nonce
			pow.block.PrevBlockHash,
			pow.block.HashTransactions(),
			IntToHex(pow.block.Timestamp),
			IntToHex(int64(targetBits)),
			IntToHex(int64(nonce)),
		},
		[]byte{},
	)
	return data
}

func (pow *ProofOfWork) Run() (int, []byte) {
	var hashInt big.Int // int representation of hash
	var hash [32]byte
	nonce := 0 // counter

	fmt.Printf("Mining a new block")
	for nonce < maxNonce { // "infinite" loop limited by maxNonce to avoid possible overflow of nonce
		data := pow.prepareData(nonce) // prepare data

		hash = sha256.Sum256(data) //hash data
		fmt.Printf("\r%x", hash)
		hashInt.SetBytes(hash[:]) //convert int to bigInt

		// compare integer with target
		if hashInt.Cmp(pow.target) == -1 { // if less, break
			break 
		} else {
			nonce ++ // increment nonce by 1
		}

	}
	fmt.Printf("\n\n")
	return nonce, hash[:]
}

func (pow *ProofOfWork) Validate() bool {
	var hashInt big.Int

	data := pow.prepareData((pow.block.Nonce))
	hash := sha256.Sum256(data)
	hashInt.SetBytes(hash[:])

	isValid := hashInt.Cmp(pow.target) == -1
	return isValid
}