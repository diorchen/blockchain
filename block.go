// simple block

package main

import (
	"bytes"
	"crypto/sha256"
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

func (b *Block) Serialize() []byte {
	var result bytes.Buffer // buffer that will store serialized data
	encoder := gob.NewEncoder(&result) // instance of gob.Encoder that is used to encode Go values into binary

	err := encoder.Encode(b) // attempts to encode
	if err != nil { // if error, log error and terminate
		log.Panic(err)
	}

	return result.Bytes() // return serialized data as byte array
}

func (b *Block) HashTransactions() []byte {
	var txHashes [][]byte // stores the hashes of each transaction in the block
	var txHash [32]byte // stores the final hash of all the transaction hashes combined

	for _, tx := range b.Transactions { // iterates over all transactions
		txHashes = append(txHashes, tx.ID)
	}
	txHash = sha256.Sum256(bytes.Join(txHashes, []byte{})) // concatenante the transaction hashes into single byte slice

	return txHash[:]
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