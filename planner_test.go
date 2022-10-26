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

func TestFileToCopy(t *testing.T) {
	pf, err := NewPetsFile("sample_pet/ssh/sshd_config", "ssh", "sample_pet/ssh/sshd_config", "root", "root", "0640", "", "")
	assertNoError(t, err)

	pa := FileToCopy(pf)
	if pa != nil {
		t.Errorf("Expecting nil, got %v instead", pa)
	}

	pf, err = NewPetsFile("sample_pet/ssh/sshd_config", "ssh", "/tmp/polpette", "root", "root", "0640", "", "")
	assertNoError(t, err)

	pa = FileToCopy(pf)
	if pa == nil {
		t.Errorf("Expecting a PetsAction, got nil instead")
	}

	assertEquals(t, pa.Cause.String(), "FILE_CREATE")

	pf, err = NewPetsFile("sample_pet/ssh/sshd_config", "ssh", "sample_pet/ssh/user_ssh_config", "root", "root", "0640", "", "")
	assertNoError(t, err)

	pa = FileToCopy(pf)
	if pa == nil {
		t.Errorf("Expecting a PetsAction, got nil instead")
	}

	assertEquals(t, pa.Cause.String(), "FILE_UPDATE")
}
