// Copyright (C) 2022 Emanuele Rocca
//
// Pets configuration file validator. Given a list of in-memory PetsFile(s),
// see if our sanity constraints are met. For example, we do not want multiple
// files to be installed to the same destination path. Also, all validation
// commands must succeed.

package main

import (
	"fmt"
	"io/fs"
	"log"
)

// CheckGlobalConstraints validates assumptions that must hold across all
// configuration files.
func CheckGlobalConstraints(files []*PetsFile) error {
	// Keep the seen PetsFiles in a map so we can:
	// 1) identify and print duplicate sources
	// 2) avoid slices.Contains which is only in Go 1.18+ and not even bound to
	//    the Go 1 Compatibility Promise™
	seen := make(map[string]*PetsFile)

	for _, pf := range files {
		other, exist := seen[pf.Dest]
		if exist {
			return fmt.Errorf("[ERROR] duplicate definition for '%s': '%s' and '%s'\n", pf.Dest, pf.Source, other.Source)
		}
		seen[pf.Dest] = pf
	}

	return nil
}

// runPre returns true if the pre-update validation command passes, or if it
// was not specificed at all. The boolean argument pathErrorOK controls whether
// or not we want to fail if the validation command is not around.
func runPre(pf *PetsFile, pathErrorOK bool) bool {
	if pf.Pre == nil {
		return true
	}

	// Some optimism.
	toReturn := true

	// Run 'pre' validation command, append Source filename to
	// arguments.
	// eg: /usr/sbin/sshd -t -f sample_pet/ssh/sshd_config
	pf.Pre.Args = append(pf.Pre.Args, pf.Source)

	stdout, stderr, err := RunCmd(pf.Pre)

	_, pathError := err.(*fs.PathError)

	if err == nil {
		log.Printf("[INFO] pre-update command %s successful\n", pf.Pre.Args)
	} else if pathError && pathErrorOK {
		// The command has failed because the validation command itself is
		// missing. This could be a chicken-and-egg problem: at this stage
		// configuration is not validated yet, hence any "package" directives
		// have not been applied.  Do not consider this as a failure, for now.
		log.Printf("[INFO] pre-update command %s failed due to PathError. Ignoring for now\n", pf.Pre.Args)
	} else {
		log.Printf("[ERROR] pre-update command %s: %s\n", pf.Pre.Args, err)
		toReturn = false
	}

	if len(stdout) > 0 {
		log.Printf("[INFO] stdout: %s", stdout)
	}

	if len(stderr) > 0 {
		log.Printf("[ERROR] stderr: %s", stderr)
	}

	return toReturn
}

// CheckLocalConstraints validates assumptions that must hold for the
// individual configuration files. An error in one file means we're gonna skip
// it but proceed with the rest. The function returns a slice of files for
// which validation passed.
func CheckLocalConstraints(files []*PetsFile, pathErrorOK bool) []*PetsFile {
	var goodPets []*PetsFile

	for _, pf := range files {
		log.Printf("[DEBUG] validating %s\n", pf.Source)

		if pf.IsValid(pathErrorOK) {
			log.Printf("[DEBUG] valid configuration file: %s\n", pf.Source)
			goodPets = append(goodPets, pf)
		} else {
			log.Printf("[ERROR] invalid configuration file: %s\n", pf.Source)
		}
	}

	return goodPets
}
