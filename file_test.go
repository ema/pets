// Copyright (C) 2022 Emanuele Rocca

package main

import (
	"testing"
)

func TestBadUser(t *testing.T) {
	f, err := NewTestFile("", "", "", "never-did-this-user-exist", "", "", "", "")

	assertError(t, err)

	if f != nil {
		t.Errorf("Expecting f to be nil, got %v instead", f)
	}
}

func TestBadGroup(t *testing.T) {
	f, err := NewTestFile("", "", "", "root", "never-did-this-user-exist", "", "", "")

	assertError(t, err)

	if f != nil {
		t.Errorf("Expecting f to be nil, got %v instead", f)
	}
}

func TestShortModes(t *testing.T) {
	f, err := NewTestFile("", "", "", "root", "root", "600", "", "")

	assertNoError(t, err)

	assertEquals(t, f.Mode, "600")

	f, err = NewTestFile("", "", "", "root", "root", "755", "", "")

	assertNoError(t, err)

	assertEquals(t, f.Mode, "755")
}

func TestOK(t *testing.T) {
	f, err := NewTestFile("syntax on\n", "vim", "/tmp/vimrc", "root", "root", "0600", "cat -n /etc/motd /etc/passwd", "df")
	assertNoError(t, err)

	assertEquals(t, f.Pkgs[0], PetsPackage("vim"))
	assertEquals(t, f.Dest, "/tmp/vimrc")
	assertEquals(t, f.Mode, "0600")
}

func TestFileIsValidTrue(t *testing.T) {
	// Everything correct
	f, err := NewTestFile("/dev/null", "gvim", "/dev/null", "root", "root", "0600", "/bin/true", "")
	assertNoError(t, err)

	assertEquals(t, f.IsValid(false), true)
}

func TestFileIsValidBadPackage(t *testing.T) {
	// Bad package name
	f, err := NewTestFile("/dev/null", "not-an-actual-package", "/dev/null", "root", "root", "0600", "/bin/true", "")
	assertNoError(t, err)

	assertEquals(t, f.IsValid(false), false)
}

func TestFileIsValidPrePathError(t *testing.T) {
	// Path error in validation command
	f, err := NewTestFile("README.adoc", "gvim", "/etc/motd", "root", "root", "0600", "/bin/whatever-but-not-a-valid-path", "")
	assertNoError(t, err)
	assertEquals(t, f.IsValid(true), true)
}

func TestFileIsValidPathError(t *testing.T) {
	f, err := NewTestFile("README.adoc", "gvim", "/etc/motd", "root", "root", "0600", "/bin/whatever-but-not-a-valid-path", "")
	assertNoError(t, err)

	// Passing pathErrorOK=true to IsValid
	assertEquals(t, f.IsValid(true), true)

	// Passing pathErrorOK=false to IsValid
	assertEquals(t, f.IsValid(false), false)
}

func TestNeedsCopyNoSource(t *testing.T) {
	f := NewPetsFile()
	f.Source = ""
	assertEquals(t, int(f.NeedsCopy()), int(NONE))
}

func TestNeedsCopySourceNotThere(t *testing.T) {
	f := NewPetsFile()
	f.Source = "something-very-funny.lol"
	assertEquals(t, int(f.NeedsCopy()), int(NONE))
}

func TestNeedsLinkNoDest(t *testing.T) {
	f := NewPetsFile()
	f.Source = "sample_pet/vimrc"
	assertEquals(t, int(f.NeedsLink()), int(NONE))
}

func TestNeedsLinkHappyPathLINK(t *testing.T) {
	f := NewPetsFile()
	f.Source = "sample_pet/vimrc"
	f.AddLink("/tmp/this_does_not_exist_yet.vimrc")
	assertEquals(t, int(f.NeedsLink()), int(LINK))
}

func TestNeedsLinkHappyPathNONE(t *testing.T) {
	f := NewPetsFile()
	f.Source = "sample_pet/README"
	f.AddLink("sample_pet/README.txt")
	assertEquals(t, int(f.NeedsLink()), int(NONE))
}

func TestNeedsLinkDestExists(t *testing.T) {
	f := NewPetsFile()
	f.Source = "sample_pet/vimrc"
	f.AddLink("/etc/passwd")
	assertEquals(t, int(f.NeedsLink()), int(NONE))
}

func TestNeedsLinkDestIsSymlink(t *testing.T) {
	f := NewPetsFile()
	f.Source = "sample_pet/vimrc"
	f.AddLink("/etc/mtab")
	assertEquals(t, int(f.NeedsLink()), int(NONE))
}

func TestNeedsDirNoDirectory(t *testing.T) {
	f := NewPetsFile()
	assertEquals(t, int(f.NeedsDir()), int(NONE))
}

func TestNeedsDirHappyPathDIR(t *testing.T) {
	f := NewPetsFile()
	f.Directory = "/etc/does/not/exist"
	assertEquals(t, int(f.NeedsDir()), int(DIR))
}

func TestNeedsDirHappyPathNONE(t *testing.T) {
	f := NewPetsFile()
	f.Directory = "/etc"
	assertEquals(t, int(f.NeedsDir()), int(NONE))
}

func TestNeedsDirDestIsFile(t *testing.T) {
	f := NewPetsFile()
	f.Directory = "/etc/passwd"
	assertEquals(t, int(f.NeedsDir()), int(NONE))
}
