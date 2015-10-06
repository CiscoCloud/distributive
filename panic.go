// This file covers the panic logging mechanism for distributive
package main

import (
	"fmt"
	"io/ioutil"
	"os"
)

var (
	recoveryFile    = "distributive.panic.log"
	issueURL        = "github.com/CiscoCloud/distributive/issues"
	recoveryMessage = `######## DISTRIBUTIVE CRASH ########
Distributive has recovered from a panic but had to shut down ungracefully. This
is always a bug! If you would, please file the output placed in the local
directory at "` + recoveryFile + `" as an issue at ` + issueURL + `, along with
what you were trying to do at the time of the crash. Thanks, and sorry for the
hassle!

version: ` + Version
)

func panicHandler(output string) {
	// first print the recovery message and status to the screen, just in case
	// dumping fails.
	fmt.Println(output)
	fmt.Println("\n" + recoveryMessage)

	// now write to the disk! Ignoring the error intentionally because we've
	// already written to the screen.
	_ = ioutil.WriteFile(recoveryFile, []byte(output), 0644)
	os.Exit(2) // different exit status than just failing checks
}
