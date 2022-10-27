// Copyright (C) 2022 Emanuele Rocca

package main

import "fmt"

func main() {
	// *** Config parser + watcher ***

	// Generate a list of PetsFiles from the given config directory.
	fmt.Println("DEBUG: * configuration parsing starts *")

	files, err := ParseFiles("sample_pet")
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("DEBUG: * configuration parsing ends *")

	// *** Config validator ***
	fmt.Println("DEBUG: * configuration validation starts *")
	globalErrors := CheckGlobalConstraints(files)

	if globalErrors != nil {
		fmt.Println(err)
		// Global validation errors mean we should stop the whole update.
		return
	}

	// Check validation errors in individual files. At this stage, the
	// command in the "pre" validation directive may not be installed yet.
	// Ignore PathErrors for now. Get a list of valid files.
	goodPets := CheckLocalConstraints(files, true)

	fmt.Println("DEBUG: * configuration validation ends *")

	// Generate the list of actions to perform.
	actions := NewPetsActions(goodPets)

	// *** Update visualizer ***
	// Display:
	// - packages to install
	// - files created/modified
	// - content diff (maybe?)
	// - owner changes
	// - permissions changes
	// - which post-update commands will be executed
	for _, action := range actions {
		fmt.Println("INFO:", action)
	}

	// *** Update executor ***
	// Install missing packages
	// Create missing directories
	// Run pre-update command and stop the update if it fails
	// Update files
	// Change permissions/owners
	// Run post-update commands
	for _, action := range actions {
		fmt.Printf("INFO: running '%s'\n", action.Command)

		err = action.Perform()
		if err != nil {
			fmt.Printf("ERROR: performing action %s: %s\n", action, err)
			break
		}
	}
}
