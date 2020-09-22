package main

import (
	"encoding/json"
	"fmt"
	"reflect"
)

type UserInfo struct {
	Name string
}

func main() {
	u := UserInfo{
		Name: "wang",
	}

	data , err := json.Marshal(&u)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(data)
	fmt.Println(string(data) ,reflect.TypeOf(data))


}