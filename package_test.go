// Copyright (C) 2022 Emanuele Rocca

package main

import (
	"testing"
)

func TestPkgIsValid(t *testing.T) {
	pkg := PetsPackage("coreutils")
	assertEquals(t, pkg.IsValid(), true)
}

func TestPkgIsNotValid(t *testing.T) {
	pkg := PetsPackage("obviously-this-cannot-be valid ?")
	assertEquals(t, pkg.IsValid(), false)
}

func TestIsInstalled(t *testing.T) {
	pkg := PetsPackage("binutils")
	assertEquals(t, pkg.IsInstalled(), true)
}

func TestIsNotInstalled(t *testing.T) {
	pkg := PetsPackage("abiword")
	assertEquals(t, pkg.IsInstalled(), false)

	pkg = PetsPackage("this is getting ridiculous")
	assertEquals(t, pkg.IsInstalled(), false)
}
