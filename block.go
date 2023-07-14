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
	Transactions	[]*Transaction
	PrevBlockHash	[]byte
	Hash 			[]byte
	Nonce			int
}



func NewBlock(transactions []*Transaction, prevBlockHash []byte) *Block {
	block := &Block{time.Now().Unix(), transactions, prevBlockHash, []byte{}, 0}
	pow := NewProofOfWork(block)
	nonce, hash := pow.Run()

	block.Hash = hash[:]
	block.Nonce = nonce
	return block
}

func NewGenesisBlock(coinbase *Transaction) *Block { // takes in initial transaction that creates the first block and awards cryptocurrency to miner
	return NewBlock([]*Transaction{coinbase}, []byte{})
}

// HashTransactions returns a hash of the transactions in the block
func (b *Block) HashTransactions() []byte {
	var transactions [][]byte

	for _, tx := range b.Transactions {
		transactions = append(transactions, tx.Serialize())
	}
	mTree := NewMerkleTree(transactions)

	return mTree.RootNode.Data
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