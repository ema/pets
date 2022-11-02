// Copyright (C) 2022 Emanuele Rocca

package main

import (
	"log"
	"os/exec"
	"strings"
)

// A PetsPackage represents a distribution package.
type PetsPackage string

// PackageManager available on the system. APT on Debian-based distros, YUM on
// RedHat and derivatives.
type PackageManager int

const (
	APT = iota
	YUM
	APK
)

func WhichPackageManager() PackageManager {
	var err error

	apt := NewCmd([]string{"apt", "--help"})
	_, _, err = RunCmd(apt)
	if err == nil {
		return APT
	}

	yum := NewCmd([]string{"yum", "--help"})
	_, _, err = RunCmd(yum)
	if err == nil {
		return YUM
	}

	apk := NewCmd([]string{"apk", "--version"})
	_, _, err = RunCmd(apk)
	if err == nil {
		return APK
	}

	panic("Unknown Package Manager")
}

func (pp PetsPackage) getPkgInfo() string {
	var pkgInfo *exec.Cmd

	switch WhichPackageManager() {
	case APT:
		pkgInfo = NewCmd([]string{"apt-cache", "policy", string(pp)})
	case YUM:
		pkgInfo = NewCmd([]string{"yum", "info", string(pp)})
	case APK:
		pkgInfo = NewCmd([]string{"apk", "search", "-e", string(pp)})
	}

	stdout, _, err := RunCmd(pkgInfo)

	if err != nil {
		log.Printf("[ERROR] pkgInfoPolicy() command %s failed: %s\n", pkgInfo, err)
		return ""
	}

	return stdout
}

// IsValid returns true if the given PetsPackage is available in the distro.
func (pp PetsPackage) IsValid() bool {
	stdout := pp.getPkgInfo()
	family := WhichPackageManager()

	if family == APT && strings.HasPrefix(stdout, string(pp)) {
		// Return true if the output of apt-cache policy begins with pp
		log.Printf("[DEBUG] %s is a valid package name\n", pp)
		return true
	}

	if family == YUM {
		for _, line := range strings.Split(stdout, "\n") {
			line = strings.TrimSpace(line)
			pkgName := strings.SplitN(line, ": ", 2)
			if len(pkgName) == 2 {
				if strings.TrimSpace(pkgName[0]) == "Name" {
					return pkgName[1] == string(pp)
				}
			}
		}
	}

	if family == APK && strings.HasPrefix(stdout, string(pp)) {
		// Return true if the output of apk search -e begins with pp
		log.Printf("[DEBUG] %s is a valid package name\n", pp)
		return true
	}

	log.Printf("[ERROR] %s is not an available package\n", pp)
	return false
}

// IsInstalled returns true if the given PetsPackage is installed on the
// system.
func (pp PetsPackage) IsInstalled() bool {
	family := WhichPackageManager()

	if family == APT {
		stdout := pp.getPkgInfo()
		for _, line := range strings.Split(stdout, "\n") {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "Installed: ") {
				version := strings.SplitN(line, ": ", 2)
				return version[1] != "(none)"
			}
		}

		log.Printf("[ERROR] no 'Installed:' line in apt-cache policy %s\n", pp)
		return false
	}

	if family == YUM {
		installed := NewCmd([]string{"rpm", "-qa", string(pp)})
		stdout, _, err := RunCmd(installed)
		if err != nil {
			log.Printf("[ERROR] running %s: '%s'", installed, err)
			return false
		}
		return len(stdout) > 0
	}

	if family == APK {
		installed := NewCmd([]string{"apk", "info", "-qe", string(pp)})
		_, _, err := RunCmd(installed)
		if err != nil {
			log.Printf("[ERROR] running %s: '%s'", installed, err)
			return false
		}
		return true
	}

	return false
}

// InstallCommand returns the command needed to install packages on this
// system.
func InstallCommand() *exec.Cmd {
	switch WhichPackageManager() {
	case APT:
		return NewCmd([]string{"apt-get", "-y", "install"})
	case YUM:
		return NewCmd([]string{"yum", "-y", "install"})
	case APK:
		return NewCmd([]string{"apk", "add"})
	}
  return nil
}
