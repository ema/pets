// Copyright (C) 2022 Emanuele Rocca

package main

import (
	"fmt"
	"os/exec"
)

// A PetsAction represents something to be done, namely running a certain
// Command. PetsActions exist because of some Trigger, which is a PetsFile.
type PetsAction struct {
	Command *exec.Cmd
	Trigger *PetsFile
}

// Visualize prints the PetsAction to stdout
func (pa *PetsAction) Visualize() {
	fmt.Printf("INFO: %s triggered command: %s\n", pa.Trigger.Source, pa.Command)
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

// NewPetsActions is the PetsFile -> []PetsAction constructor.
// Given a PetsFile, generate a list of PetsActions to perform.
func NewPetsActions(trigger *PetsFile) []*PetsAction {
	actions := []*PetsAction{}

	// First, install the package
	if trigger.Pkg != "" {
		actions = append(actions, &PetsAction{
			Command: trigger.Pkg.InstallCommand(),
			Trigger: trigger,
		})
	}

	// Then, see if Source needs to be copied over Dest. No need to check if
	// Dest is empty, as it's a mandatory argument. Its presence is ensured at
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
			Command: NewCmd([]string{"cp", trigger.Source, trigger.Dest}),
			Trigger: trigger,
		})
	}

	return actions
}
