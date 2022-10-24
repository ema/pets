// Copyright (C) 2022 Emanuele Rocca

package main

import (
	"fmt"
	"strings"
)

// A PetsPackage represents a distribution package.
type PetsPackage struct {
	Name string
}

// IsValid returns true if the given PetsPackage is available in the distro.
func (pp *PetsPackage) IsValid() bool {
	aptCache := NewCmd([]string{"apt-cache", "policy", pp.Name})
	stdout, _, err := RunCmd(aptCache)

	if err != nil {
		fmt.Printf("ERROR: PetsPackage.IsValid() command %s failed: %s\n", aptCache, err)
		return false
	}

	if strings.HasPrefix(stdout, pp.Name) {
		// Return true if the output of apt-cache policy begins with Name
		fmt.Printf("DEBUG: %s is a valid package name\n", pp.Name)
		return true
	} else {
		fmt.Printf("ERROR: %s is not an available package\n", pp.Name)
		return false
	}
}
