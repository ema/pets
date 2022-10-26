// Copyright (C) 2022 Emanuele Rocca

package main

import (
	"testing"
)

func TestPkgsToInstall(t *testing.T) {
	// Test with empty slice of PetsFiles
	petsFiles := []*PetsFile{}
	isTodo, _ := PkgsToInstall(petsFiles)
	assertEquals(t, isTodo, false)

	// Test with one package already installed
	pf, err := NewPetsFile("/dev/null", "coreutils", "/etc/passwd", "root", "root", "0640", "", "")
	assertNoError(t, err)

	petsFiles = append(petsFiles, pf)
	isTodo, _ = PkgsToInstall(petsFiles)
	assertEquals(t, isTodo, false)

	// Add another package to the mix, this time it's not installed
	petsFiles[0].Pkgs = append(petsFiles[0].Pkgs, PetsPackage("freedoom"))
	isTodo, _ = PkgsToInstall(petsFiles)
	assertEquals(t, isTodo, true)
}
