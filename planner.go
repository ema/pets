// Copyright (C) 2022 Emanuele Rocca

package main

import (
	"fmt"
	"os/exec"
)

// PetsCause conveys the reason behind a given action.
type PetsCause int

const (
	PKG      = iota // required package is missing
	CONTENTS        // configuration file contents differ
	OWNER           // needs chown()
	GROUP           // needs chgrp()
	MODE            // needs chmod()
)

// A PetsAction represents something to be done, namely running a certain
// Command. PetsActions exist because of some Trigger, which is a PetsFile.
type PetsAction struct {
	Cause   PetsCause
	Command *exec.Cmd
	Trigger *PetsFile
}

// Visualize prints the PetsAction to stdout
func (pa *PetsAction) Visualize() {
	if pa.Trigger != nil {
		fmt.Printf("INFO: %s triggered command: %s\n", pa.Trigger.Source, pa.Command)
	} else {
		fmt.Printf("INFO: command: %s\n", pa.Command)
	}
}

// Perform executes the Command
func (pa *PetsAction) Perform() {
	stdout, stderr, err := RunCmd(pa.Command)

	if err != nil {
		fmt.Printf("ERROR: running Perform() -> %v\n", err)
	}

	if len(stdout) > 0 {
		fmt.Printf("INFO: stdout from Perform() -> %v\n", stdout)
	}

	if len(stderr) > 0 {
		fmt.Printf("ERROR: stderr from Perform() -> %v\n", stderr)
	}
}

// PkgsToInstall returns two values, a boolean and a command. The former is
// true if there are any new packages to install, the latter is the
// distro-specific command to run to install the packages.
func PkgsToInstall(triggers []*PetsFile) (bool, *exec.Cmd) {
	installPkgs := false
	installCmd := InstallCommand()

	for _, trigger := range triggers {
		for _, pkg := range trigger.Pkgs {
			if pkg.IsInstalled() {
				fmt.Printf("DEBUG: %s already installed\n", pkg)
			} else {
				fmt.Printf("INFO: %s not installed\n", pkg)
				installCmd.Args = append(installCmd.Args, string(pkg))
				installPkgs = true
			}
		}
	}

	return installPkgs, installCmd
}

// NewPetsActions is the []PetsFile -> []PetsAction constructor.  Given a slice
// of PetsFile(s), generate a list of PetsActions to perform.
func NewPetsActions(triggers []*PetsFile) []*PetsAction {
	actions := []*PetsAction{}

	// First, install all needed packages. Build a list of all missing package
	// names first, and then install all of them in one go. This is to avoid
	// embarassing things like running in a loop apt install pkg1 ; apt install
	// pkg2 ; apt install pkg3 like some configuration management systems do.
	if installPkgs, installCmd := PkgsToInstall(triggers); installPkgs {
		actions = append(actions, &PetsAction{
			Cause:   PKG,
			Command: installCmd,
		})
	}

	for _, trigger := range triggers {
		// See if Source needs to be copied over Dest. No need to check if Dest
		// is empty, as it's a mandatory argument. Its presence is ensured at
		// parsing time.
		shaSource, err := Sha256(trigger.Source)
		if err != nil {
			fmt.Printf("ERROR: cannot determine sha256 of Source file %s: %v\n", trigger.Source, err)
		}

		shaDest, err := Sha256(trigger.Dest)
		if err != nil {
			fmt.Printf("ERROR: cannot determine sha256 of Dest file %s: %v\n", trigger.Dest, err)
		}

		if len(shaSource) > 0 && len(shaDest) > 0 && shaSource != shaDest {
			fmt.Printf("DEBUG: sha256[%s]=%s != sha256[%s]=%s\n", trigger.Source, shaSource, trigger.Dest, shaDest)

			actions = append(actions, &PetsAction{
				Cause:   CONTENTS,
				Command: NewCmd([]string{"cp", trigger.Source, trigger.Dest}),
				Trigger: trigger,
			})
		}
	}

	return actions
}
