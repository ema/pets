// Copyright (C) 2022 Emanuele Rocca

package main

import (
	"fmt"
	"os/exec"
	"strings"
)

// A PetsPackage represents a distribution package.
type PetsPackage string

// IsValid returns true if the given PetsPackage is available in the distro.
func (pp PetsPackage) IsValid() bool {
	aptCache := NewCmd([]string{"apt-cache", "policy", string(pp)})
	stdout, _, err := RunCmd(aptCache)

	if err != nil {
		fmt.Printf("ERROR: PetsPackage.IsValid() command %s failed: %s\n", aptCache, err)
		return false
	}

	if strings.HasPrefix(stdout, string(pp)) {
		// Return true if the output of apt-cache policy begins with pp
		fmt.Printf("DEBUG: %s is a valid package name\n", pp)
		return true
	} else {
		fmt.Printf("ERROR: %s is not an available package\n", pp)
		return false
	}
}

// InstallCommand returns the command needed to install the given PetsPackage.
func (pp PetsPackage) InstallCommand() *exec.Cmd {
	return NewCmd([]string{"apt-get", "-y", "install", string(pp)})
}
