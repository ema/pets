// Copyright (C) 2022 Emanuele Rocca

package main

import (
	"log"
	"os/exec"
	"strings"
)

// A PetsPackage represents a distribution package.
type PetsPackage string

func (pp PetsPackage) aptCachePolicy() string {
	aptCache := NewCmd([]string{"apt-cache", "policy", string(pp)})
	stdout, _, err := RunCmd(aptCache)

	if err != nil {
		log.Printf("ERROR: aptCachePolicy() command %s failed: %s\n", aptCache, err)
		return ""
	}

	return stdout
}

// IsValid returns true if the given PetsPackage is available in the distro.
func (pp PetsPackage) IsValid() bool {
	stdout := pp.aptCachePolicy()

	if strings.HasPrefix(stdout, string(pp)) {
		// Return true if the output of apt-cache policy begins with pp
		log.Printf("DEBUG: %s is a valid package name\n", pp)
		return true
	} else {
		log.Printf("ERROR: %s is not an available package\n", pp)
		return false
	}
}

// IsInstalled returns true if the given PetsPackage is installed on the
// system.
func (pp PetsPackage) IsInstalled() bool {
	stdout := pp.aptCachePolicy()
	for _, line := range strings.Split(stdout, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Installed: ") {
			version := strings.SplitN(line, ": ", 2)
			return version[1] != "(none)"
		}
	}

	log.Printf("ERROR: no 'Installed:' line in apt-cache policy %s\n", pp)
	return false
}

// InstallCommand returns the command needed to install packages on this
// system.
func InstallCommand() *exec.Cmd {
	return NewCmd([]string{"apt-get", "-y", "install"})
}
