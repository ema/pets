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
	pf, err := NewTestFile("/dev/null", "binutils", "/etc/passwd", "root", "root", "0640", "", "")
	assertNoError(t, err)

	petsFiles = append(petsFiles, pf)
	isTodo, _ = PkgsToInstall(petsFiles)
	assertEquals(t, isTodo, false)

	// Add another package to the mix, this time it's not installed
	petsFiles[0].Pkgs = append(petsFiles[0].Pkgs, PetsPackage("abiword"))
	isTodo, _ = PkgsToInstall(petsFiles)
	assertEquals(t, isTodo, true)
}

func TestFileToCopy(t *testing.T) {
	pf, err := NewTestFile("sample_pet/ssh/sshd_config", "ssh", "sample_pet/ssh/sshd_config", "root", "root", "0640", "", "")
	assertNoError(t, err)

	pa := FileToCopy(pf)
	if pa != nil {
		t.Errorf("Expecting nil, got %v instead", pa)
	}

	pf, err = NewTestFile("sample_pet/ssh/sshd_config", "ssh", "/tmp/polpette", "root", "root", "0640", "", "")
	assertNoError(t, err)

	pa = FileToCopy(pf)
	if pa == nil {
		t.Errorf("Expecting a PetsAction, got nil instead")
	}

	assertEquals(t, pa.Cause.String(), "FILE_CREATE")

	pf, err = NewTestFile("sample_pet/ssh/sshd_config", "ssh", "sample_pet/ssh/user_ssh_config", "root", "root", "0640", "", "")
	assertNoError(t, err)

	pa = FileToCopy(pf)
	if pa == nil {
		t.Errorf("Expecting a PetsAction, got nil instead")
	}

	assertEquals(t, pa.Cause.String(), "FILE_UPDATE")
}

func TestChmod(t *testing.T) {
	// Expect Chmod() to return nil if the 'mode' directive is missing.
	pf := NewPetsFile()
	pf.Source = "/dev/null"
	pf.Dest = "/dev/null"

	pa := Chmod(pf)
	if pa != nil {
		t.Errorf("Expecting nil, got %v instead", pa)
	}

	pf.AddMode("0644")

	pa = Chmod(pf)

	assertEquals(t, pa.Cause.String(), "CHMOD")
	assertEquals(t, pa.Command.String(), "/bin/chmod 0644 /dev/null")

	pf.Dest = "/etc/passwd"
	pa = Chmod(pf)
	if pa != nil {
		t.Errorf("Expecting nil, got %v instead", pa)
	}
}

func TestChown(t *testing.T) {
	pf := NewPetsFile()
	pf.Source = "/dev/null"
	pf.Dest = "/etc/passwd"

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
	assertEquals(t, pa.Command.String(), "/bin/chown nobody:root /etc/passwd")
}

func TestLn(t *testing.T) {
	pf := NewPetsFile()
	pf.Source = "sample_pet/vimrc"

	// Link attribute and Dest not set
	pa := LinkToCreate(pf)
	if pa != nil {
		t.Errorf("Expecting nil, got %v instead", pa)
	}

	// Destination already exists
	pf.AddLink("/etc/passwd")

	pa = LinkToCreate(pf)
	if pa != nil {
		t.Errorf("Expecting nil, got %v instead", pa)
	}

	// Happy path, destination does not exist yet
	pf.AddLink("/tmp/vimrc")

	pa = LinkToCreate(pf)
	if pa == nil {
		t.Errorf("Expecting some action, got nil instead")
	}

	assertEquals(t, pa.Cause.String(), "LINK_CREATE")
	assertEquals(t, pa.Command.String(), "/bin/ln -s sample_pet/vimrc /tmp/vimrc")
}

func TestMkdir(t *testing.T) {
	pf := NewPetsFile()

	pa := DirToCreate(pf)
	if pa != nil {
		t.Errorf("Expecting nil, got %v instead", pa)
	}

	pf.Directory = "/etc"

	pa = DirToCreate(pf)
	if pa != nil {
		t.Errorf("Expecting nil, got %v instead", pa)
	}

	pf.Directory = "/etc/polpette/al/sugo"

	pa = DirToCreate(pf)
	if pa == nil {
		t.Errorf("Expecting some action, got nil instead")
	}

	assertEquals(t, pa.Cause.String(), "DIR_CREATE")
	assertEquals(t, pa.Command.String(), "/bin/mkdir -p /etc/polpette/al/sugo")
}
