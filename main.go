// Copyright (C) 2022 Emanuele Rocca

package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/hashicorp/logutils"
)

// ParseFlags parses the CLI flags and returns: the configuration directory as
// string, a bool for debugging output, and another bool for dryRun.
func ParseFlags() (string, bool, bool) {
	var confDir string
	defaultConfDir := filepath.Join(os.Getenv("HOME"), "pets")
	flag.StringVar(&confDir, "conf-dir", defaultConfDir, "Pets configuration directory")
	debug := flag.Bool("debug", false, "Show debugging output")
	dryRun := flag.Bool("dry-run", false, "Only show changes without applying them")
	flag.Parse()
	return confDir, *debug, *dryRun
}

// GetLogFilter returns a LevelFilter suitable for log.SetOutput().
func GetLogFilter(debug bool) *logutils.LevelFilter {
	minLogLevel := "INFO"
	if debug {
		minLogLevel = "DEBUG"
	}

	return &logutils.LevelFilter{
		Levels:   []logutils.LogLevel{"DEBUG", "INFO", "ERROR"},
		MinLevel: logutils.LogLevel(minLogLevel),
		Writer:   os.Stdout,
	}
}

func main() {
	startTime := time.Now()

	confDir, debug, dryRun := ParseFlags()

	log.SetOutput(GetLogFilter(debug))

	// Print distro family
	if WhichPackageManager == APT {
		log.Println("[DEBUG] Running on a Debian-like system")
	} else if WhichPackageManager == YUM {
		log.Println("[DEBUG] Running on a RedHat-like system")
	}
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

	if dryRun {
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
	exitStatus := 0
	for _, action := range actions {
		log.Printf("[INFO] running '%s'\n", action.Command)

		err = action.Perform()
		if err != nil {
			log.Printf("[ERROR] performing action %s: %s\n", action, err)
			exitStatus = 1
			break
		}
	}

	log.Printf("[INFO] pets run took %v\n", time.Since(startTime).Round(time.Millisecond))

	os.Exit(exitStatus)
}
