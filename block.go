// simple block

package main

import (
	"bytes"
	"encoding/gob"
	"log"
	"time"
)

type Block struct {
	Timestamp		int64
	Data			[]byte // slice of byte
	PrevBlockHash	[]byte
	Hash 			[]byte
	Nonce			int
}

func (b *Block) Serialize() []byte {
	var result bytes.Buffer // buffer that will store serialized data
	encoder := gob.NewEncoder(&result) // instance of gob.Encoder that is used to encode Go values into binary

	err := encoder.Encode(b) // attempts to encode
	if err != nil { // if error, log error and terminate
		log.Panic(err)
	}

	return result.Bytes() // return serialized data as byte array
}


func NewBlock(data string, prevBlockHash []byte) *Block {
	block := &Block{time.Now().Unix(), []byte(data), prevBlockHash, []byte{}, 0}
	pow := NewProofOfWork(block)
	nonce, hash := pow.Run()

	block.Hash = hash[:]
	block.Nonce = nonce
	return block
}

func NewGenesisBlock() *Block {
	return NewBlock("Genesis Block", []byte{})
}

// Deserializes block
func DeserializeBlock(d []byte) *Block {
	var block Block

	decoder := gob.NewDecoder(bytes.NewReader(d))
	err := decoder.Decode(&block)
	if err != nil {
		log.Panic(err)
	}

	return &block
}