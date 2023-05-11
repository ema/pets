// Copyright (C) 2022 Emanuele Rocca

package main

import (
	"log"
	"os"
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
	YAY
	PACMAN
)

var WhichPackageManager = whichPackageManager()

// whichPackageManager is available on the system
func whichPackageManager() PackageManager {
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

	// Yay has to be first because yay wraps pacman
	yay := NewCmd([]string{"yay", "--version"})
	if _, _, err = RunCmd(yay); err == nil {
		return YAY
	}

	pacman := NewCmd([]string{"pacman", "--version"})
	if _, _, err = RunCmd(pacman); err == nil {
		return PACMAN
	}

	panic("Unknown Package Manager")
}

func (pp PetsPackage) getPkgInfo() string {
	var pkgInfo *exec.Cmd

	switch WhichPackageManager {
	case APT:
		pkgInfo = NewCmd([]string{"apt-cache", "policy", string(pp)})
	case YUM:
		pkgInfo = NewCmd([]string{"yum", "info", string(pp)})
	case APK:
		pkgInfo = NewCmd([]string{"apk", "search", "-e", string(pp)})
	case PACMAN:
		pkgInfo = NewCmd([]string{"pacman", "-Si", string(pp)})
	case YAY:
		pkgInfo = NewCmd([]string{"yay", "-Si", string(pp)})
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
	if WhichPackageManager == APT && strings.HasPrefix(stdout, string(pp)) {
		// Return true if the output of apt-cache policy begins with pp
		log.Printf("[DEBUG] %s is a valid package name\n", pp)
		return true
	}

	if WhichPackageManager == YUM {
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

	if WhichPackageManager == APK && strings.HasPrefix(stdout, string(pp)) {
		// Return true if the output of apk search -e begins with pp
		log.Printf("[DEBUG] %s is a valid package name\n", pp)
		return true
	}

	if (WhichPackageManager == PACMAN || WhichPackageManager == YAY) && !strings.HasPrefix(stdout, "error:") {
		// Return true if the output of pacman -Si doesnt begins with error
		log.Printf("[DEBUG] %s is a valid package name\n", pp)
		return true
	}

	log.Printf("[ERROR] %s is not an available package\n", pp)
	return false
}

// IsInstalled returns true if the given PetsPackage is installed on the
// system.
func (pp PetsPackage) IsInstalled() bool {
	if WhichPackageManager == APT {
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

	if WhichPackageManager == YUM {
		installed := NewCmd([]string{"rpm", "-qa", string(pp)})
		stdout, _, err := RunCmd(installed)
		if err != nil {
			log.Printf("[ERROR] running %s: '%s'", installed, err)
			return false
		}
		return len(stdout) > 0
	}

	if WhichPackageManager == APK {
		installed := NewCmd([]string{"apk", "info", "-e", string(pp)})
		stdout, _, err := RunCmd(installed)
		if err != nil {
			log.Printf("[ERROR] running %s: '%s'\n", installed, err)
			return false
		}

		// apk info -e $pkg prints the package name to stdout if the package is
		// installed, nothing otherwise
		return strings.TrimSpace(stdout) == string(pp)
	}

	if WhichPackageManager == PACMAN || WhichPackageManager == YAY {
		installed := NewCmd([]string{"pacman", "-Q", string(pp)})
		if WhichPackageManager == YAY {
			installed = NewCmd([]string{"yay", "-Q", string(pp)})
		}
		// pacman and yay will return 0 if the package is installed 1 if not
		if _, _, err := RunCmd(installed); err != nil {
			if exitError, ok := err.(*exec.ExitError); ok {
				return exitError.ExitCode() == 0
			}
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
	switch WhichPackageManager {
	case APT:
		cmd := NewCmd([]string{"apt-get", "-y", "install"})
		cmd.Env = os.Environ()
		cmd.Env = append(cmd.Env, "DEBIAN_FRONTEND=noninteractive")
		return cmd
	case YUM:
		return NewCmd([]string{"yum", "-y", "install"})
	case APK:
		return NewCmd([]string{"apk", "add"})
	case PACMAN:
		return NewCmd([]string{"pacman", "-S", "--noconfirm"})
	case YAY:
		return NewCmd([]string{"yay", "-S", "--noconfirm"})
	}
	return nil
}
