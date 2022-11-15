// Copyright (C) 2022 Emanuele Rocca

package main

import (
	"os"
	"testing"
)

func coreutils_bin() string {

	switch os.Getenv("os") {
	case "alpine":
		coreutils_bin := "/bin"
	case "debian":
		coreutils_bin := "/bin"
	case "ubuntu":
		coreutils_bin := "/bin"
	default:
		coreutils_bin := "/usr/bin"
	}
	return coreutils_bin
}

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

func TestChmod(t *testing.T) {
	// Expect Chmod() to return nil if the 'mode' directive is missing.
	pf := &PetsFile{
		Source: "/dev/null",
		Dest:   "/dev/null",
	}

	pa := Chmod(pf)
	if pa != nil {
		t.Errorf("Expecting nil, got %v instead", pa)
	}

	pf.AddMode("0644")

	pa = Chmod(pf)

	assertEquals(t, pa.Cause.String(), "CHMOD")
	assertEquals(t, pa.Command.String(), fmt.Sprintf("%s/chmod 0644 /dev/null", coreutils_bin()))

	pf.Dest = "/etc/passwd"
	pa = Chmod(pf)
	if pa != nil {
		t.Errorf("Expecting nil, got %v instead", pa)
	}
}

func TestChown(t *testing.T) {
	pf := &PetsFile{
		Source: "/dev/null",
		Dest:   "/etc/passwd",
	}

	// If no 'user' or 'group' directives are specified
	pa := Chown(pf)
	if pa != nil {
		t.Errorf("Expecting nil, got %v instead", pa)
	}

	// File owned by 'root:root' already
	pf.AddUser("root")
	pf.AddGroup("root")
	pa = Chown(pf)
	if pa != nil {
		t.Errorf("Expecting nil, got %v instead", pa)
	}

	pf.AddUser("nobody")
	pa = Chown(pf)
	if pa == nil {
		t.Errorf("Expecting some action, got nil instead")
	}

	assertEquals(t, pa.Cause.String(), "OWNER")
	assertEquals(t, pa.Command.String(), fmt.Sprintf("%s/chown nobody:root /etc/passwd", coreutils_bin()))
}
