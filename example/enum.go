package main

import "fmt"

type Status int

const (
	StatusOk Status = 200
	StatusFailed Status = 400
	StatusTimeOut Status = 500
)

func main() {
	var s Status
	s = StatusFailed

	fmt.Println(s)
}
