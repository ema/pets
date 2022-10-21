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
	// 1) print the duplicate sources
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
