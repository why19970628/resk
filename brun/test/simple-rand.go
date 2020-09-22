package main

import (
	"fmt"
	"resk/infra/algo"
)

func main()  {
	count, amount := int64(10), int64(100)
	for i := int64(0); i < count; i++ {
		x := algo.SimpleRand(count, amount*100)
		fmt.Printf("x:%0.2f\n", float64(x)/100)
	}

}
