// Copyright (C) 2022 Emanuele Rocca

package main

import (
	"testing"
)

func TestReadModelinesFileNotFound(t *testing.T) {
	modelines, err := ReadModelines("very-unlikely-to-find-this.txt")

	assertError(t, err)

	if modelines != nil {
		t.Errorf("Expecting nil modelines, got %v instead", modelines)
	}
}

func TestReadModelinesZero(t *testing.T) {
	modelines, err := ReadModelines("README.adoc")
	assertNoError(t, err)
	assertEquals(t, len(modelines), 0)
}

func TestReadModelinesNonZero(t *testing.T) {
	modelines, err := ReadModelines("sample_pet/ssh/user_ssh_config")
	assertNoError(t, err)
	assertEquals(t, len(modelines), 2)
}

func TestParseModelineErr(t *testing.T) {
	var pf PetsFile
	err := ParseModeline("", &pf)
	assertError(t, err)
}

func TestParseModelineBadKeyword(t *testing.T) {
	var pf PetsFile
	err := ParseModeline("# pets: something=funny", &pf)
	assertError(t, err)
}

func TestParseModelineOKDestfile(t *testing.T) {
	var pf PetsFile
	err := ParseModeline("# pets: destfile=/etc/ssh/sshd_config, owner=root, group=root, mode=0644", &pf)
	assertNoError(t, err)

	assertEquals(t, pf.Dest, "/etc/ssh/sshd_config")
	assertEquals(t, pf.User.Uid, "0")
	assertEquals(t, pf.Group.Gid, "0")
	assertEquals(t, pf.Mode, "0644")
	assertEquals(t, pf.Link, false)
}

func TestParseModelineOKSymlink(t *testing.T) {
	var pf PetsFile
	err := ParseModeline("# pets: symlink=/etc/ssh/sshd_config", &pf)
	assertNoError(t, err)

	assertEquals(t, pf.Dest, "/etc/ssh/sshd_config")
	assertEquals(t, pf.Link, true)
}

func TestParseModelineOKPackage(t *testing.T) {
	var pf PetsFile
	err := ParseModeline("# pets: package=vim", &pf)
	assertNoError(t, err)

	assertEquals(t, pf.Dest, "")
	assertEquals(t, string(pf.Pkgs[0]), "vim")
}
