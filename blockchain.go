// simple blockchain implementation

package main

import (
	"encoding/hex"
	"fmt"
	"log"
	"os"

	"github.com/boltdb/bolt"
)

const dbFile = "blockchain.db"
const blocksBucket = "blocks"
const genesisCoinbaseData = "The Times 03/Jan/2009 Chancellor on brink of second bailout for banks"

type Blockchain struct {
	tip []byte // latest block's hash
	db *bolt.DB // database BoltDB
}

// iterate over keys in bucket
type BlockchainIterator struct {
	currentHash []byte
	db *bolt.DB
}

func (bc *Blockchain) MineBlock(transactions []*Transaction) {
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

	newBlock := NewBlock(transactions, lastHash)

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

func (bc *Blockchain) FindUnspentTransactions(address string) []Transaction { // takes address, returns slice of unspent transactions associated with it
	var unspentTXs []Transaction
	spentTXOs := make(map[string][]int)
	bci := bc.Iterator()

	for {
		block := bci.Next()

		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID) // convert transaction ID to hexadecimal string

		Outputs:
			for outIdx, out := range tx.Vout {
				// Was the output spent?
				if spentTXOs[txID] != nil {
					for _, spentOut := range spentTXOs[txID] { // for all transaction outputs
						if spentOut == outIdx { // if spent, continue
							continue Outputs
						}
					}
				}

				if out.CanBeUnlockedWith(address) { // if unlockable with adress, address is unspent, so add to slice
					unspentTXs = append(unspentTXs, *tx)
				}
			}

			if tx.IsCoinbase() == false { // if not coinbase transaction(not the reward for mining)
				for _, in := range tx.Vin { // iterate over input and check if any input can unlock an ouput
					if in.CanUnlockOutputWith(address) {
						inTxID := hex.EncodeToString(in.Txid) // add transaction id and output index to 'spentTXOs' map to mark as spent
						spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Vout)
					}
				}
			}
		}

		if len(block.PrevBlockHash) == 0 { // continue until genesis block reached (empty prevBlockHash)
			break
		}
	}

	return unspentTXs
}

// FindUTXO finds and returns all unspent transaction outputs
func (bc *Blockchain) FindUTXO(address string) []TXOutput {
	var UTXOs []TXOutput
	unspentTransactions := bc.FindUnspentTransactions(address) // retrieve all unspent transactions associated with aress

	for _, tx := range unspentTransactions { // iterate through each unspent transaction
		for _, out := range tx.Vout { // iterate through each transaction output
			if out.CanBeUnlockedWith(address) { // if can be unlocked, add to unspent TXOs slice
				UTXOs = append(UTXOs, out)
			}
		}
	}

	return UTXOs
}

// FindSpendableOutputs finds and returns unspent outputs to reference in inputs
func (bc *Blockchain) FindSpendableOutputs(address string, amount int) (int, map[string][]int) {
	unspentOutputs := make(map[string][]int)
	unspentTXs := bc.FindUnspentTransactions(address)
	accumulated := 0

Work:
	for _, tx := range unspentTXs { // for each transaction in UTXOs, convert tx.ID to hex format
		txID := hex.EncodeToString(tx.ID)

		for outIdx, out := range tx.Vout { // for each output, check if can be unlocked and if the accumulate amount is less than desired "amount"
			if out.CanBeUnlockedWith(address) && accumulated < amount {
				accumulated += out.Value // add value to accumulated
				unspentOutputs[txID] = append(unspentOutputs[txID], outIdx) // output index mapped to tx.ID

				if accumulated >= amount {
					break Work
				}
			}
		}
	}

	return accumulated, unspentOutputs
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

func dbExists() bool {
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return false
	}

	return true
}

func NewBlockchain(address string) *Blockchain {
	if dbExists() == false {
		fmt.Println("No existing blockchain found. Create one first.")
		os.Exit(1)
	}

	var tip []byte
	db, err := bolt.Open(dbFile, 0600, nil) // open BoltDB file, won't return error if DNE
	if err != nil {
		log.Panic(err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket)) // obtain bucket storing blocks
		tip = b.Get([]byte("l"))

		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	//store tip and DB connection in blockchain struct
	bc := Blockchain{tip, db}

	return &bc
}