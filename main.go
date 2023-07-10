package main


func main() {
	bc := NewBlockchain()
	defer bc.db.Close() // Close DB when main function finishes execution

	cli := CLI{bc} // initialize CLI struct
	cli.Run()
	
}