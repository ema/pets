// Copyright (C) 2022 Emanuele Rocca

package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"

	"github.com/hashicorp/logutils"
)

func defaultConfDir() string {
	home := os.Getenv("HOME")
	return filepath.Join(home, "pets")
}

func main() {
	// Parse CLI flags
	var confDir string
	flag.StringVar(&confDir, "conf-dir", defaultConfDir(), "Pets configuration directory")
	debug := flag.Bool("debug", false, "Show debugging output")
	dryRun := flag.Bool("dry-run", false, "Only show changes without applying them")
	flag.Parse()

	// Setup logger
	minLogLevel := "INFO"
	if *debug {
		minLogLevel = "DEBUG"
	}

	filter := &logutils.LevelFilter{
		Levels:   []logutils.LogLevel{"DEBUG", "INFO", "ERROR"},
		MinLevel: logutils.LogLevel(minLogLevel),
		Writer:   os.Stderr,
	}
	log.SetOutput(filter)

	// *** Config parser ***

	// Generate a list of PetsFiles from the given config directory.
	log.Println("[DEBUG] * configuration parsing starts *")

	files, err := ParseFiles(confDir)
	if err != nil {
		log.Println(err)
	}

	log.Printf("[INFO] Found %d pets configuration files", len(files))

	log.Println("[DEBUG] * configuration parsing ends *")

	// *** Config validator ***
	log.Println("[DEBUG] * configuration validation starts *")
	globalErrors := CheckGlobalConstraints(files)

	if globalErrors != nil {
		log.Println(err)
		// Global validation errors mean we should stop the whole update.
		return
	}

	// Check validation errors in individual files. At this stage, the
	// command in the "pre" validation directive may not be installed yet.
	// Ignore PathErrors for now. Get a list of valid files.
	goodPets := CheckLocalConstraints(files, true)

	log.Println("[DEBUG] * configuration validation ends *")

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
		log.Println("[INFO]", action)
	}

	if *dryRun {
		log.Println("[INFO] user requested dry-run mode, not applying any changes")
		return
	}

	// *** Update executor ***
	// Install missing packages
	// Create missing directories
	// Run pre-update command and stop the update if it fails
	// Update files
	// Change permissions/owners
	// Run post-update commands
	for _, action := range actions {
		log.Printf("[INFO] running '%s'\n", action.Command)

		err = action.Perform()
		if err != nil {
			log.Printf("[ERROR] performing action %s: %s\n", action, err)
			break
		}
	}
}
