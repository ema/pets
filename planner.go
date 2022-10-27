// Copyright (C) 2022 Emanuele Rocca

package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"syscall"
)

// PetsCause conveys the reason behind a given action.
type PetsCause int

const (
	PKG    = iota // required package is missing
	CREATE        // configuration file is missing and needs to be created
	UPDATE        // configuration file differs and needs to be updated
	OWNER         // needs chown()
	MODE          // needs chmod()
	POST          // post-update command
)

func (pc PetsCause) String() string {
	return map[PetsCause]string{
		PKG:    "PACKAGE_INSTALL",
		CREATE: "FILE_CREATE",
		UPDATE: "FILE_UPDATE",
		OWNER:  "OWNER",
		MODE:   "CHMOD",
		POST:   "POST_UPDATE",
	}[pc]
}

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
		fmt.Printf("INFO: [%s] %s triggered command: %s\n", pa.Cause, pa.Trigger.Source, pa.Command)
	} else {
		fmt.Printf("INFO: [%s] triggered command: %s\n", pa.Cause, pa.Command)
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

// FileToCopy figures out if the given trigger represents a file that needs to
// be updated, and returns the corresponding PetsAction.
func FileToCopy(trigger *PetsFile) *PetsAction {
	// See if Source needs to be copied over Dest. No need to check if Dest
	// is empty, as it's a mandatory argument. Its presence is ensured at
	// parsing time.
	shaSource, err := Sha256(trigger.Source)
	if err != nil {
		fmt.Printf("ERROR: cannot determine sha256 of Source file %s: %v\n", trigger.Source, err)
		return nil
	}

	shaDest, err := Sha256(trigger.Dest)
	if os.IsNotExist(err) {
		return &PetsAction{
			Cause:   CREATE,
			Command: NewCmd([]string{"cp", trigger.Source, trigger.Dest}),
			Trigger: trigger,
		}
	} else if err != nil {
		fmt.Printf("ERROR: cannot determine sha256 of Dest file %s: %v\n", trigger.Dest, err)
		return nil
	}

	if shaSource == shaDest {
		fmt.Printf("DEBUG: same sha256 for %s and %s: %s\n", trigger.Source, trigger.Dest, shaSource)
		return nil
	}

	fmt.Printf("INFO: sha256[%s]=%s != sha256[%s]=%s\n", trigger.Source, shaSource, trigger.Dest, shaDest)

	return &PetsAction{
		Cause:   UPDATE,
		Command: NewCmd([]string{"cp", trigger.Source, trigger.Dest}),
		Trigger: trigger,
	}
}

// Chown returns a chown PetsAction or nil if none is needed.
func Chown(trigger *PetsFile) *PetsAction {
	// Build arg (eg: 'root:staff', 'root', ':staff')
	arg := ""
	var wantUid, wantGid int
	var err error

	if trigger.User != nil {
		arg = trigger.User.Username

		// get the requested uid as integer
		wantUid, err = strconv.Atoi(trigger.User.Uid)
		if err != nil {
			// This should really never ever happen, unless we're
			// running on Windows. :)
			panic(err)
		}
	}

	if trigger.Group != nil {
		arg = fmt.Sprintf("%s:%s", arg, trigger.Group.Name)

		// get the requested gid as integer
		wantGid, err = strconv.Atoi(trigger.Group.Gid)
		if err != nil {
			panic(err)
		}
	}

	if arg == "" {
		// Return immediately if the file had no 'owner' / 'group' directives
		return nil
	}

	// The action to (possibly) perform is a chown of the file.
	action := &PetsAction{
		Cause:   OWNER,
		Command: NewCmd([]string{"chown", arg, trigger.Dest}),
		Trigger: trigger,
	}

	// stat(2) the destination file to see if a chown is needed
	fileInfo, err := os.Stat(trigger.Dest)
	if os.IsNotExist(err) {
		// If the destination file is not there yet, prepare a chown
		// for later on.
		return action
	}

	stat, _ := fileInfo.Sys().(*syscall.Stat_t)

	if trigger.User != nil && int(stat.Uid) != wantUid {
		fmt.Printf("INFO: %s is owned by uid %d instead of %s\n", trigger.Dest, stat.Uid, trigger.User.Username)
		return action
	}

	if trigger.Group != nil && int(stat.Gid) != wantGid {
		fmt.Printf("INFO: %s is owned by gid %d instead of %s\n", trigger.Dest, stat.Gid, trigger.Group.Name)
		return action
	}

	fmt.Printf("DEBUG: %s is owned by %d:%d already\n", trigger.Dest, stat.Uid, stat.Gid)
	return nil
}

// Chmod returns a chmod PetsAction or nil if none is needed.
func Chmod(trigger *PetsFile) *PetsAction {
	if trigger.Mode == "" {
		// Return immediately if the 'mode' directive was not specified.
		return nil
	}

	// The action to (possibly) perform is a chmod of the file.
	action := &PetsAction{
		Cause:   MODE,
		Command: NewCmd([]string{"chmod", trigger.Mode, trigger.Dest}),
		Trigger: trigger,
	}

	// stat(2) the destination file to see if a chmod is needed
	fileInfo, err := os.Stat(trigger.Dest)
	if os.IsNotExist(err) {
		// If the destination file is not there yet, prepare a mod
		// for later on.
		return action
	}

	// See if the desired mode and reality differ.
	newMode, err := StringToFileMode(trigger.Mode)
	if err != nil {
		fmt.Println("ERROR: unexpected error in Chmod()", err)
		return nil
	}

	oldMode := fileInfo.Mode()

	if oldMode != newMode {
		fmt.Printf("INFO: %s is %s instead of %s\n", trigger.Dest, oldMode, newMode)
		return action
	}

	fmt.Printf("DEBUG: %s is %s already\n", trigger.Dest, newMode)
	return nil
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
		// Then, figure out which files need to be modified/created.
		if fileAction := FileToCopy(trigger); fileAction != nil {
			actions = append(actions, fileAction)
		}

		// Any owner changes needed
		if chown := Chown(trigger); chown != nil {
			actions = append(actions, chown)
		}

		// Any mode changes needed
		if chmod := Chmod(trigger); chmod != nil {
			actions = append(actions, chmod)
		}

		// Finally, post-update commands
		if trigger.Post != nil {
			actions = append(actions, &PetsAction{
				Cause:   POST,
				Command: trigger.Post,
			})
		}
	}

	return actions
}
