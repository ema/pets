package main

import "fmt"

func main() {
	err := walkDir("sample_pet")
	if err != nil {
		fmt.Println(err)
	}
}
