// Copyright (C) 2022 Emanuele Rocca

package main

import (
	"os"

	"testing"
)

func assertEquals(t *testing.T, a, b interface{}) {
	if a != b {
		t.Errorf("%v != %v", a, b)
	}
}

func TestBadUser(t *testing.T) {
	f, err := NewPetsFile("", "", "", "never-did-this-user-exist", "", "", "", "")

	if err == nil {
		t.Errorf("Expecting an error, got nil instead")
	}

	if f != nil {
		t.Errorf("Expecting nil error, got %v instead", f)
	}
}

func TestShortModes(t *testing.T) {
	f, err := NewPetsFile("", "", "", "root", "root", "600", "", "")
	if err != nil {
		t.Errorf("Expecting err to be nil, got %v instead", err)
	}

	assertEquals(t, f.Mode, os.FileMode(int(0600)))

	f, err = NewPetsFile("", "", "", "root", "root", "755", "", "")
	if err != nil {
		t.Errorf("Expecting err to be nil, got %v instead", err)
	}

	assertEquals(t, f.Mode, os.FileMode(int(0755)))
}

func TestOK(t *testing.T) {
	f, err := NewPetsFile("syntax on\n", "vim", "/tmp/vimrc", "root", "root", "0600", "cat -n /etc/motd /etc/passwd", "df")
	if err != nil {
		t.Errorf("Expecting err to be nil, got %v instead", err)
	}

	assertEquals(t, f.Pkg, "vim")
	assertEquals(t, f.Dest, "/tmp/vimrc")
	assertEquals(t, f.Mode, os.FileMode(int(0600)))
}
