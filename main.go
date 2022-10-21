package main

import "fmt"

func main() {
	// *** Config parser + watcher ***

	// Generate a list of PetsFiles from the given config directory.
	files, err := walkDir("sample_pet")
	if err != nil {
		fmt.Println(err)
	}

	// *** Config validator ***
	globalErrors := checkGlobalConstraints(files)

	if globalErrors != nil {
		fmt.Println(err)
		// Global validation errors mean we should stop the whole update.
		return
	}

	// Check validation errors in individual files. Get a list of valid files.
	// TODO: see if the specified packages are available (eg: apt-cache policy)
	goodPets := checkLocalConstraints(files)
	for _, pet := range goodPets {
		fmt.Println(pet)
	}

	// *** Update visualizer ***
	// packages to install
	// pre-update command output
	// files created/modified
	// content diff
	// permissions/owner changes
	// show which post-update commands will be executed

	// *** Update executor ***
	// Install missing packages
	// Create missing directories
	// Update files
	// Change permissions/owners
	// Run post-update commands
}
