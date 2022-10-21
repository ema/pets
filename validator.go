// Copyright (C) 2022 Emanuele Rocca
//
// Pets configuration file validator. Given a list of in-memory PetsFile(s),
// see if our sanity constraints are met. For example, we do not want multiple
// files to be installed to the same destination path. Also, all validation
// commands must succeed.

package main

import (
	"fmt"
)

// checkGlobalConstraints validates assumptions that must hold across all
// configuration files
func checkGlobalConstraints(files []*PetsFile) error {
	// Keep the seen PetsFiles in a map so we can:
	// 1) identify and print duplicate sources
	// 2) avoid slices.Contains which is only in Go 1.18+ and not even bound to
	//    the Go 1 Compatibility Promise™
	seen := make(map[string]*PetsFile)

	for _, pf := range files {
		other, exist := seen[pf.Dest]
		if exist {
			return fmt.Errorf("ERROR: duplicate definition for '%s': '%s' and '%s'\n", pf.Dest, pf.Source, other.Source)
		}
		seen[pf.Dest] = pf
	}

	return nil
}

// checkLocalConstraints validates assumptions that must hold for the
// individual configuration files. An error in one file means we're gonna skip
// it but proceed with the rest. The function returns a slice of files for
// which validation passed.
func checkLocalConstraints(files []*PetsFile, pathErrorOK bool) []*PetsFile {
	var goodPets []*PetsFile

	for _, pf := range files {
		// Check if the specified package exists
		if !pf.PkgExists() {
			continue
		}

		// Check pre-update validation command
		if !pf.RunPre(pathErrorOK) {
			continue
		}

		goodPets = append(goodPets, pf)
	}

	return goodPets
}
