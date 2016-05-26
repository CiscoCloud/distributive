package checks

import (
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"syscall"

	"github.com/CiscoCloud/distributive/chkutil"
	"github.com/CiscoCloud/distributive/errutil"
	"github.com/CiscoCloud/distributive/tabular"
	"github.com/mitchellh/go-ps"
)

/*
#### Command
Description: Does this Command exit without error?
Parameters:
- Cmd (string): Command to be executed
Example parameters:
- "cat /etc/my-config/", "/bin/my_health_check.py"
*/

type Command struct{ Command string }

func init() { 
    chkutil.Register("Command", func() chkutil.Check {
        return &Command{}
    })
    chkutil.Register("CommandOutputMatches", func() chkutil.Check {
        return &CommandOutputMatches{}
    })
    chkutil.Register("Running", func() chkutil.Check {
        return &Running{}
    })
    chkutil.Register("Temp", func() chkutil.Check {
        return &Running{}
    })
    chkutil.Register("Module", func() chkutil.Check {
        return &Module{}
    })
    chkutil.Register("KernelParameter", func() chkutil.Check {
        return &KernelParameter{}
    })
    chkutil.Register("PHPConfig", func() chkutil.Check {
        return &PHPConfig{}
    })
}

func (chk Command) New(params []string) (chkutil.Check, error) {
	// TODO further validation with LookPath - maybe in parameter-validation.go
	if len(params) != 1 {
		return chk, errutil.ParameterLengthError{1, params}
	}
	chk.Command = params[0]
	return chk, nil
}

func (chk Command) Status() (int, string, error) {
	cmd := exec.Command("bash", "-c", chk.Command)
	err := cmd.Start()
	if err != nil && strings.Contains(err.Error(), "not found in $PATH") {
		return 1, "Executable not found: " + chk.Command, nil
	} else if err != nil {
		return 1, "", err
	}
	if err = cmd.Wait(); err != nil {
		var exitCode int
		// this is convoluted, but should work on Windows & Unix
		if exiterr, ok := err.(*exec.ExitError); ok {
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				exitCode = status.ExitStatus()
			}
		}
		// dummy, in case the above failed. We know it's not zero!
		if exitCode == 0 {
			exitCode = 1
		}
		out, _ := cmd.CombinedOutput() // don't care if this fails
		exitMessage := "Command exited with non-zero exit code:"
		exitMessage += "\n\tCommand: " + chk.Command
		exitMessage += "\n\tExit code: " + fmt.Sprint(exitCode)
		exitMessage += "\n\tOutput: " + string(out)
		return 1, exitMessage, nil
	}
	return errutil.Success()
}

/*
#### CommandOutputMatches
Description: Does the combined (stdout + stderr) output of this Command match
the given regexp?
Parameters:
- Cmd (string): Command to be executed
- Regexp (regexp): Regexp to query output with
Example parameters:
- "cat /etc/my-config/", "/bin/my_health_check.py"
- "value=expected", "[rR]{1}e\we[Xx][^oiqnlkasdjc]"
*/

type CommandOutputMatches struct {
	Command string
	re      *regexp.Regexp
}

func (chk CommandOutputMatches) New(params []string) (chkutil.Check, error) {
	if len(params) != 2 {
		return chk, errutil.ParameterLengthError{2, params}
	}
	re, err := regexp.Compile(params[1])
	if err != nil {
		return chk, errutil.ParameterTypeError{params[1], "regexp"}
	}
	chk.re = re
	chk.Command = params[0]
	return chk, nil
}

func (chk CommandOutputMatches) Status() (int, string, error) {
	cmd := exec.Command("bash", "-c", chk.Command)
	out, err := cmd.CombinedOutput()
	if err != nil {
		errutil.ExecError(cmd, string(out), err)
	}
	if chk.re.Match(out) {
		return errutil.Success()
	}
	msg := "Command output did not match regexp"
	return errutil.GenericError(msg, chk.re.String(), []string{string(out)})
}

/*
#### Running
Description: Is a process by this exact name Running (excluding this process)?
Parameters:
- Name (string): Process name to look for
Example parameters:
- nginx, [kthreadd], consul-agent, haproxy-consul
Depedencies:
- `ps aux`
*/

type Running struct{ name string }

func (chk Running) New(params []string) (chkutil.Check, error) {
	if len(params) != 1 {
		return chk, errutil.ParameterLengthError{1, params}
	}
	chk.name = params[0]
	return chk, nil
}

func (chk Running) Status() (int, string, error) {
	processes, err := ps.Processes()
	if err != nil {
		return 1, "", err
	}
	var executables []string
	for _, proc := range processes {
		executables = append(executables, proc.Executable())
	}
	if tabular.StrIn(chk.name, executables) {
		return errutil.Success()
	}
	return errutil.GenericError("Process not Running", chk.name, executables)
}

/*
#### Temp
Description: Is the core Temperature under this value (in degrees Celcius)?
Parameters:
- Temp (positive int16): Maximum acceptable Temperature
Example parameters:
- 100, 110C, 98°C, 100℃
Depedencies:
- A configured lm-sensors (namely, `sensors`)
*/

type Temp struct{ max uint16 }

func (chk Temp) New(params []string) (chkutil.Check, error) {
	if len(params) != 1 {
		return chk, errutil.ParameterLengthError{1, params}
	}
	maxStr := params[0]
	// list includes: C, c, U+00B0, U+2103
	for _, char := range []string{"C", "c", "°", "℃"} {
		maxStr = strings.Replace(maxStr, char, "", -1)
	}
	maxInt, err := strconv.ParseInt(maxStr, 10, 16)
	if err != nil || maxInt < 0 {
		return chk, errutil.ParameterTypeError{params[0], "uint16"}
	}
	chk.max = uint16(maxInt)
	return chk, nil
}

// parse the output of `sensors` to get temperatures
func parseSensorsOutput(out string) (Temps []int) {
	re := regexp.MustCompile(`Core\s\d+:\s+[\+\-](?P<Temp>\d+)\.*\d*(°|\s)C`)
	for _, line := range regexp.MustCompile(`\n+`).Split(out, -1) {
		if re.MatchString(line) {
			// submatch captures only the integer part of the Temperature
			matchDict := chkutil.SubmatchMap(re, line)
			if _, ok := matchDict["Temp"]; !ok {
				return Temps
			}
			TempInt64, err := strconv.ParseInt(matchDict["Temp"], 10, 64)
			if err != nil {
				return Temps
			}
			Temps = append(Temps, int(TempInt64))
		}
	}
	return Temps
}

func (chk Temp) Status() (int, string, error) {
	cmd := exec.Command("sensors")
	out, err := cmd.CombinedOutput()
	if err != nil {
		errutil.ExecError(cmd, string(out), err)
	}
	temps := parseSensorsOutput(string(out))
	if len(temps) <= 1 {
		return 1, "", errors.New("Couldn't parse the output of lm-sensors")
	}
	if temps[0] < int(chk.max) {
		return errutil.Success()
	}
	msg := "Core Temp exceeds defined maximum"
	return errutil.GenericError(msg, chk.max, []string{fmt.Sprint(temps[0])})
}

/*
#### Module
Description: Is this kernel module installed?
Parameters:
- Name (string): Module name
Example parameters:
- hid, drm, rfkill
Depedencies:
- `/sbin/lsmod`
*/

type Module struct{ name string }

func (chk Module) New(params []string) (chkutil.Check, error) {
	if len(params) != 1 {
		return chk, errutil.ParameterLengthError{1, params}
	}
	chk.name = params[0]
	return chk, nil
}

func (chk Module) Status() (int, string, error) {
	// kernelModules returns a list of all Modules that are currently loaded
	// TODO just read from /proc/modules
	kernelModules := func() (Modules []string) {
		cmd := exec.Command("/sbin/lsmod")
		return chkutil.CommandColumnNoHeader(0, cmd)
	}
	modules := kernelModules()
	if tabular.StrIn(chk.name, modules) {
		return errutil.Success()
	}
	return errutil.GenericError("Module is not loaded", chk.name, modules)
}

/*
#### KernelParameter
Description: Is this kernel parameter set?
Parameters:
- Name (string): Kernel parameter to check
Example parameters:
- "net.ipv6.route.gc_interval", "fs.file-max"
Depedencies:
- `/sbin/sysctl`
*/

type KernelParameter struct{ name string }

func (chk KernelParameter) New(params []string) (chkutil.Check, error) {
	if len(params) != 1 {
		return chk, errutil.ParameterLengthError{1, params}
	}
	chk.name = params[0]
	return chk, nil
}

func (chk KernelParameter) Status() (int, string, error) {
	// parameterValue returns the value of a kernel parameter
	parameterSet := func(name string) bool {
		cmd := exec.Command("/sbin/sysctl", "-q", "-n", name)
		out, err := cmd.CombinedOutput()
		// failed on incorrect module name
		if err != nil && strings.Contains(err.Error(), "255") {
			return false
		} else if err != nil {
			errutil.ExecError(cmd, string(out), err)
		}
		return true
	}
	if parameterSet(chk.name) {
		return errutil.Success()
	}
	return 1, "Kernel parameter not set: " + chk.name, nil
}

/*
#### PHPConfig
Description: Does this PHP configuration variable have this value?
Parameters:
- Variable (string): PHP variable to check
- Value (string): Expected value
Example parameters:
- "default_mimetype"
Depedencies:
- php
*/

type PHPConfig struct{ variable, value string }

func (chk PHPConfig) New(params []string) (chkutil.Check, error) {
	if len(params) != 2 {
		return chk, errutil.ParameterLengthError{2, params}
	}
	chk.variable = params[0]
	chk.value = params[1]
	return chk, nil
}

func (chk PHPConfig) Status() (int, string, error) {
	// getPHPVariable returns the value of a PHP configuration value as a string
	// or just "" if it doesn't exist
	getPHPVariable := func(name string) (val string) {
		quote := func(str string) string { return "\"" + str + "\"" }
		// php -r 'echo get_cfg_var("variable_name");'
		echo := fmt.Sprintf("echo get_cfg_var(%s);", quote(name))
		cmd := exec.Command("php", "-r", echo)
		out, err := cmd.CombinedOutput()
		if err != nil {
			errutil.ExecError(cmd, string(out), err)
		}
		return string(out)
	}
	actualValue := getPHPVariable(chk.variable)
	if actualValue == chk.value {
		return errutil.Success()
	} else if actualValue == "" {
		msg := "PHP configuration variable not set"
		return errutil.GenericError(msg, chk.value, []string{actualValue})
	}
	msg := "PHP variable did not match expected value"
	return errutil.GenericError(msg, chk.value, []string{actualValue})
}
