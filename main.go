package main

import "fmt"

func main() {
	files, err := walkDir("sample_pet")
	if err != nil {
		fmt.Println(err)
	}

	for i, f := range files {
		fmt.Printf("%d: %v\n", i, f)
	}
}
