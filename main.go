package main

import "fmt"

func main() {
	// *** Config parser + watcher ***

	// Generate a list of PetsFiles from the given config directory.
	fmt.Println("DEBUG: * configuration parsing starts *")

	files, err := walkDir("sample_pet")
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("DEBUG: * configuration parsing ends *")

	// *** Config validator ***
	fmt.Println("DEBUG: * configuration validation starts *")
	globalErrors := checkGlobalConstraints(files)

	if globalErrors != nil {
		fmt.Println(err)
		// Global validation errors mean we should stop the whole update.
		return
	}

	// Check validation errors in individual files. At this stage, the
	// command in the "pre" validation directive may not be installed yet.
	// Ignore PathErrors for now. Get a list of valid files.
	goodPets := checkLocalConstraints(files, true)
	for _, pet := range goodPets {
		fmt.Printf("DEBUG: valid configuration file: %s\n", pet.Source)
	}

	fmt.Println("DEBUG: * configuration validation ends *")

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
