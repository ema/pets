// Copyright (C) 2022 Emanuele Rocca

package main

import (
	"testing"
)

func TestBadUser(t *testing.T) {
	f, err := NewPetsFile("", "", "", "never-did-this-user-exist", "", "", "", "")

	assertError(t, err)

	if f != nil {
		t.Errorf("Expecting f to be nil, got %v instead", f)
	}
}

func TestBadGroup(t *testing.T) {
	f, err := NewPetsFile("", "", "", "root", "never-did-this-user-exist", "", "", "")

	assertError(t, err)

	if f != nil {
		t.Errorf("Expecting f to be nil, got %v instead", f)
	}
}

func TestShortModes(t *testing.T) {
	f, err := NewPetsFile("", "", "", "root", "root", "600", "", "")

	assertNoError(t, err)

	assertEquals(t, f.Mode, "600")

	f, err = NewPetsFile("", "", "", "root", "root", "755", "", "")

	assertNoError(t, err)

	assertEquals(t, f.Mode, "755")
}

func TestOK(t *testing.T) {
	f, err := NewPetsFile("syntax on\n", "vim", "/tmp/vimrc", "root", "root", "0600", "cat -n /etc/motd /etc/passwd", "df")
	assertNoError(t, err)

	assertEquals(t, f.Pkgs[0], PetsPackage("vim"))
	assertEquals(t, f.Dest, "/tmp/vimrc")
	assertEquals(t, f.Mode, "0600")
}

func TestFileIsValidTrue(t *testing.T) {
	// Everything correct
	f, err := NewPetsFile("/dev/null", "gvim", "/dev/null", "root", "root", "0600", "/bin/true", "")
	assertNoError(t, err)

	assertEquals(t, f.IsValid(false), true)
}

func TestFileIsValidBadPackage(t *testing.T) {
	// Bad package name
	f, err := NewPetsFile("/dev/null", "not-an-actual-package", "/dev/null", "root", "root", "0600", "/bin/true", "")
	assertNoError(t, err)

	assertEquals(t, f.IsValid(false), false)
}

func TestFileIsValidPrePathError(t *testing.T) {
	// Path error in validation command
	f, err := NewPetsFile("README.adoc", "gvim", "/etc/motd", "root", "root", "0600", "/bin/whatever-but-not-a-valid-path", "")
	assertNoError(t, err)
	assertEquals(t, f.IsValid(true), true)
}

func TestFileIsValidPathError(t *testing.T) {
	f, err := NewPetsFile("README.adoc", "gvim", "/etc/motd", "root", "root", "0600", "/bin/whatever-but-not-a-valid-path", "")
	assertNoError(t, err)

	// Passing pathErrorOK=true to IsValid
	assertEquals(t, f.IsValid(true), true)

	// Passing pathErrorOK=false to IsValid
	assertEquals(t, f.IsValid(false), false)
}
