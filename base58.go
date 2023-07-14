package main

import (
	"bytes"
	"math/big"
)

var b58Alphabet = []byte("123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz") // prepares lookup table for Base58 encoding

// Base58Encode encodes a byte array to Base58
func Base58Encode(input []byte) []byte {
	var result []byte

	x := big.NewInt(0).SetBytes(input) // creates big int 'x' by converting input byte slice to big int

	base := big.NewInt(int64(len(b58Alphabet))) // create big int 'base' represeting the length of the Base58 Alphabet
	zero := big.NewInt(0) // big int 0
	mod := &big.Int{} //creates pointer to big int 'mod' used to store the remainder during division

	for x.Cmp(zero) != 0 { // repeatedly divide 'x' by 'base'
		x.DivMod(x, base, mod)
		result = append(result, b58Alphabet[mod.Int64()]) // remainder is used to index the corresponding character in Base58 alphabet, which is appended to result
	}

	ReverseBytes(result) // reverse order of bytes
	for b := range input {
		if b == 0x00 { // check for leading 0 bytes in the input, if encountered, prepend
			result = append([]byte{b58Alphabet[0]}, result...)
		} else {
			break // terminate when non-zero byte is encountered
		}
	}

	return result
}

// Base58Decode decodes Base58-encoded data
func Base58Decode(input []byte) []byte {
	result := big.NewInt(0) // bigInt 'result' initialized with 0 stores decoded numerical value
	zeroBytes := 0 // counter to keep track of number of leading 0 bytes in input

	for b := range input { // counts number of leading 0 bytes
		if b == 0x00 {
			zeroBytes++
		}
	}

	payload := input[zeroBytes:] // extracts the portion of the input slice after the leading 0 bytes (actual payload to be decoded)
	for _, b := range payload {
		charIndex := bytes.IndexByte(b58Alphabet, b) // finds index of byte in b58Alphabet
		result.Mul(result, big.NewInt(58)) // multiply 58 by bigInt
		result.Add(result, big.NewInt(int64(charIndex))) // add index
	}

	decoded := result.Bytes() // retrieves byte representation of the 'result' bigInt
	decoded = append(bytes.Repeat([]byte{byte(0x00)}, zeroBytes), decoded...) // appends leading 0 bytes to 'decoded' slice corresponding to num 0 bytes found in input

	return decoded
}