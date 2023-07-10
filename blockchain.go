// simple blockchain implementation

package main

import (
	"fmt"
	"log"

	"github.com/boltdb/bolt"
)

const dbFile = "blockchain.db"
const blocksBucket = "blocks"

type Blockchain struct {
	tip []byte // latest block's hash
	db *bolt.DB // database BoltDB
}

// iterate over keys in bucket
type BlockchainIterator struct {
	currentHash []byte
	db *bolt.DB
}

func (bc *Blockchain) AddBlock(data string) {
	var lastHash []byte

	// start read-only transaction ('View') on BoltDB database
	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket)) // retrieves the bucket 'blocksBucket' in transaction 
		lastHash = b.Get([]byte("l")) // value associated with 'l' is fetched from bucket and assigned to lastHash

		return nil // return nil for successful execution
	})

	if err != nil {
		log.Panic(err)
	}

	newBlock := NewBlock(data, lastHash)

	err = bc.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		err := b.Put(newBlock.Hash, newBlock.Serialize())
		if err != nil {
			log.Panic(err)
		}

		err = b.Put([]byte("l"), newBlock.Hash) // update the last block's hash to hash of newly added block
		if err != nil {
			log.Panic(err)
		}

		bc.tip = newBlock.Hash // update tip

		return nil
	})
}

// creates new instance of BlockchainIterator and initializes with the hash of the latest block and BoltDB database associated with it
func (bc *Blockchain) Iterator() *BlockchainIterator {
	bci := &BlockchainIterator{bc.tip, bc.db} // return pointer to new BlockchainIterator

	return bci
}


// retrieve the next block from the blockchain during iteration
func (i *BlockchainIterator) Next() *Block {
	var block *Block

	err := i.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		encodedBlock := b.Get(i.currentHash) // fetches encoded block associated with current hash from bucket
		block = DeserializeBlock(encodedBlock) // deserializes

		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	i.currentHash = block.PrevBlockHash // updates the currentHash to the PrevBlockHash to ensure iterator will move to prevBlock

	return block // return retrieved block
}


func NewBlockchain() *Blockchain {
	var tip []byte
	db, err := bolt.Open(dbFile, 0600, nil) // open BoltDB file, won't return error if DNE
	if err != nil {
		log.Panic(err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket)) // obtain bucket storing blocks

		if b == nil { // if DNE, generate genesis block
			fmt.Println("No existing blockchain found. Creating a new one...")
			genesis := NewGenesisBlock()

			// create bucket
			b, err := tx.CreateBucket([]byte(blocksBucket))
			if err != nil {
				log.Panic(err)
			}
			// save block into bucket
			err = b.Put(genesis.Hash, genesis.Serialize())
			if err != nil {
				log.Panic(err)
			}
			// update 'l' key storing the last block hash of chain
			err = b.Put([]byte("l"), genesis.Hash)
			if err != nil {
				log.Panic(err)
			}
			// update the tip
			tip = genesis.Hash
		} else {
			// else, read 'l' key
			tip = b.Get([]byte("l"))
		}

		return nil
	})

	if err != nil {
		log.Panic(err)
	}
	//store tip and DB connection in blockchain struct
	bc := Blockchain{tip, db}

	return &bc
}