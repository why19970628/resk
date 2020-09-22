package main

import (
	"fmt"
	"resk/infra/algo"
)

func main() {
	fmt.Printf("%v\n",
		algo.AfterShuffle(int64(10), int64(100)*100))
}
