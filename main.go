package main

import "fmt"

func main() {
	// Generate a list of PetsFiles from the given config directory.
	files, err := walkDir("sample_pet")
	if err != nil {
		fmt.Println(err)
	}

	// Config validator
	globalErrors := checkGlobalConstraints(files)

	if globalErrors != nil {
		fmt.Println(err)
		// Global validation errors mean we should stop the whole update.
		return
	}

	// Check validation errors in individual files. Get a list of valid files.
	goodPets := checkLocalConstraints(files)
	for _, pet := range goodPets {
		fmt.Println(pet)
	}
}
