package workers

import (
	"fmt"
	"github.com/CiscoCloud/distributive/tabular"
	"github.com/CiscoCloud/distributive/wrkutils"
	"io/ioutil"
	"log"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// RegisterMisc registers these checks so they can be used.
func RegisterMisc() {
	wrkutils.RegisterCheck("command", command, 1)
	wrkutils.RegisterCheck("commandoutputmatches", commandOutputMatches, 2)
	wrkutils.RegisterCheck("running", running, 1)
	wrkutils.RegisterCheck("phpconfig", phpConfig, 2)
	wrkutils.RegisterCheck("diskusage", diskUsage, 2)
	wrkutils.RegisterCheck("memoryusage", memoryUsage, 1)
	wrkutils.RegisterCheck("swapusage", swapUsage, 1)
	wrkutils.RegisterCheck("cpuusage", cpuUsage, 1)
	wrkutils.RegisterCheck("temp", temp, 1)
	wrkutils.RegisterCheck("module", module, 1)
	wrkutils.RegisterCheck("kernelparameter", kernelParameter, 1)
}

// command runs a shell command, and collapses its error code to 0 or 1.
// It outputs stderr and stdout if the command has error code != 0.
func command(parameters []string) (exitCode int, exitMessage string) {
	toExec := parameters[0]
	cmd := exec.Command("bash", "-c", toExec)
	err := cmd.Start()
	if err != nil && strings.Contains(err.Error(), "not found in $PATH") {
		return 1, "Executable not found: " + toExec
	} else if err != nil {
		wrkutils.ExecError(cmd, "", err)
	}
	if err = cmd.Wait(); err != nil {
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
		exitMessage += "Command exited with non-zero exit code:"
		exitMessage += "\n\tCommand: " + toExec
		exitMessage += "\n\tExit code: " + fmt.Sprint(exitCode)
		exitMessage += "\n\tOutput: " + string(out)
		return 1, exitMessage
	}
	return 0, ""
}

// commandOutputMatches checks to see if a command's combined output matches a
// given regexp
func commandOutputMatches(parameters []string) (exitCode int, exitMessage string) {
	toExec := parameters[0]
	re := wrkutils.ParseUserRegex(parameters[1])
	cmd := exec.Command("bash", "-c", toExec)
	out, err := cmd.CombinedOutput()
	if err != nil {
		wrkutils.ExecError(cmd, string(out), err)
	}
	if re.Match(out) {
		return 0, ""
	}
	msg := "Command output did not match regexp"
	return wrkutils.GenericError(msg, re.String(), []string{string(out)})
}

// running checks if a process is running using `ps aux`, and searching for the
// process name, excluding this process (in case the process name is in the JSON
// file name)
func running(parameters []string) (exitCode int, exitMessage string) {
	// getRunningCommands returns the entries in the "COMMAND" column of `ps aux`
	getRunningCommands := func() (commands []string) {
		cmd := exec.Command("ps", "aux")
		return wrkutils.CommandColumnNoHeader(10, cmd)
	}
	proc := parameters[0]
	// remove this process from consideration
	commands := getRunningCommands()
	var filtered []string
	for _, cmd := range commands {
		if !strings.Contains(cmd, "distributive") {
			filtered = append(filtered, cmd)
		}
	}
	if tabular.StrIn(proc, filtered) {
		return 0, ""
	}
	return wrkutils.GenericError("Process not running", proc, filtered)
}

// temp parses the output of lm_sensors and determines if Core 0 (all cores) are
// over a certain threshold as specified in the JSON.
func temp(parameters []string) (exitCode int, exitMessage string) {
	// getCoreTemp returns an integer temperature for a certain core
	getCoreTemp := func(core int) (temp int) {
		out, err := exec.Command("sensors").Output()
		if err != nil {
			log.Fatal("Error while executing `sensors`:\n\t" + err.Error())
		}
		// get all-core line up to paren
		lineRegex := regexp.MustCompile("Core " + fmt.Sprint(core) + ":?(.*)\\(")
		line := lineRegex.Find(out)
		// get temp from that line
		tempRegex := regexp.MustCompile("\\d+\\.\\d*")
		tempString := string(tempRegex.Find(line))
		tempFloat, err := strconv.ParseFloat(tempString, 64)
		if err != nil {
			msg := "Error while parsing output from `sensors`:\n\t"
			log.Fatal(msg + err.Error())
		}
		return int(tempFloat)

	}
	max := wrkutils.ParseMyInt(parameters[0])
	temp := getCoreTemp(0)
	if temp < max {
		return 0, ""
	}
	msg := "Core temp exceeds defined maximum"
	return wrkutils.GenericError(msg, fmt.Sprint(max), []string{fmt.Sprint(temp)})
}

// module checks to see if a kernel module is installed
func module(parameters []string) (exitCode int, exitMessage string) {
	// kernelModules returns a list of all modules that are currently loaded
	kernelModules := func() (modules []string) {
		cmd := exec.Command("/sbin/lsmod")
		return wrkutils.CommandColumnNoHeader(0, cmd)
	}
	name := parameters[0]
	modules := kernelModules()
	if tabular.StrIn(name, modules) {
		return 0, ""
	}
	return wrkutils.GenericError("Module is not loaded", name, modules)
}

// kernelParameter checks to see if a kernel parameter was set
func kernelParameter(parameters []string) (exitCode int, exitMessage string) {
	// parameterValue returns the value of a kernel parameter
	parameterSet := func(name string) bool {
		_, err := exec.Command("/sbin/sysctl", "-q", "-n", name).Output()
		// failed on incorrect module name
		if err != nil && strings.Contains(err.Error(), "255") {
			return false
		} else if err != nil {
			log.Fatal("Error while executing /sbin/systctl:\n\tError: " + err.Error())
		}
		return true
	}
	name := parameters[0]
	if parameterSet(name) {
		return 0, ""
	}
	return 1, "Kernel parameter not set: " + name
}

// phpConfig checks the value of a PHP configuration variable
func phpConfig(parameters []string) (exitCode int, exitMessage string) {
	// getPHPVariable returns the value of a PHP configuration value as a string
	// or just "" if it doesn't exist
	getPHPVariable := func(name string) (val string) {
		quote := func(str string) string {
			return "\"" + str + "\""
		}
		// php -r 'echo get_cfg_var("default_mimetype");
		echo := fmt.Sprintf("echo get_cfg_var(%s);", quote(name))
		cmd := exec.Command("php", "-r", echo)
		out, err := cmd.CombinedOutput()
		if err != nil {
			wrkutils.ExecError(cmd, string(out), err)
		}
		return string(out)
	}
	name := parameters[0]
	value := parameters[1]
	actualValue := getPHPVariable(name)
	if actualValue == value {
		return 0, ""
	} else if actualValue == "" {
		msg := "PHP configuration variable not set"
		return wrkutils.GenericError(msg, value, []string{actualValue})
	}
	msg := "PHP variable did not match expected value"
	return wrkutils.GenericError(msg, value, []string{actualValue})
}

// getSwapOrMemory returns output from `free`, it is an abstraction of
// getSwap and getMemory.  inputs: status: free | used | total
// swapOrMem: memory | swap, units: b | kb | mb | gb | tb
func getSwapOrMemory(status string, swapOrMem string, units string) int {
	statusToColumn := map[string]int{
		"total": 1,
		"used":  2,
		"free":  3,
	}
	unitsToFlag := map[string]string{
		"b":  "--bytes",
		"kb": "--kilo",
		"mb": "--mega",
		"gb": "--giga",
		"tb": "--terra",
	}
	typeToRow := map[string]int{
		"memory": 0,
		"swap":   1,
	}
	// check to see that our keys are really in our dict
	if _, ok := statusToColumn[status]; !ok {
		msg := "Invalid status passed to getSwapOrMemory"
		_, e := wrkutils.GenericError(msg, status, []string{"total", "used", "free"})
		log.Fatal(e)
	} else if _, ok := unitsToFlag[units]; !ok {
		msg := "Invalid units passed to getSwapOrMemory"
		_, e := wrkutils.GenericError(msg, status, []string{"b", "kb", "mb", "gb", "tb"})
		log.Fatal(e)
	} else if _, ok := typeToRow[swapOrMem]; !ok {
		msg := "Invalid swapOrMem passed to getSwapOrMemory"
		_, e := wrkutils.GenericError(msg, status, []string{"memory", "swap"})
		log.Fatal(e)
	}
	// execute free and return the appropriate output
	cmd := exec.Command("free", unitsToFlag[units])
	entireColumn := wrkutils.CommandColumnNoHeader(statusToColumn[status], cmd)
	// TODO parse and catch here, for better error reporting
	return wrkutils.ParseMyInt(entireColumn[typeToRow[swapOrMem]])
}

// getSwap returns KiB of swap with a certain status: free | used | total
// and with a certain given unit: b | kb | mb | gb | tb
func getSwap(status string, units string) int {
	return getSwapOrMemory(status, "swap", units)
}

// getMemory returns KiB of swap with a certain status: free | used | total
// and with a certain given unit: b | kb | mb | gb | tb
func getMemory(status string, units string) int {
	return getSwapOrMemory(status, "memory", units)
}

// getUsedPercent returns the % of physical memory or swap currently in use
func getUsedPercent(swapOrMem string) float32 {
	used := getSwapOrMemory("used", swapOrMem, "b")
	total := getSwapOrMemory("total", swapOrMem, "b")
	return (float32(used) / float32(total)) * 100
}

// memoryUsage checks to see whether or not the system has a memory usage
// percentage below a certain threshold
func memoryUsage(parameters []string) (exitCode int, exitMessage string) {
	maxPercentUsed := wrkutils.ParseMyInt(parameters[0])
	actualPercentUsed := getUsedPercent("memory")
	if actualPercentUsed < float32(maxPercentUsed) {
		return 0, ""
	}
	msg := "Memory usage above defined maximum"
	slc := []string{fmt.Sprint(actualPercentUsed)}
	return wrkutils.GenericError(msg, fmt.Sprint(maxPercentUsed), slc)
}

// memoryUsage checks to see whether or not the system has a memory usage
// percentage below a certain threshold
func swapUsage(parameters []string) (exitCode int, exitMessage string) {
	maxPercentUsed := wrkutils.ParseMyInt(parameters[0])
	actualPercentUsed := getUsedPercent("swap")
	if actualPercentUsed < float32(maxPercentUsed) {
		return 0, ""
	}
	msg := "Swap usage above defined maximum"
	slc := []string{fmt.Sprint(actualPercentUsed)}
	return wrkutils.GenericError(msg, fmt.Sprint(maxPercentUsed), slc)
}

// getCPUSample helps cpuUsage do its thing. Taken from a stackoverflow:
// http://stackoverflow.com/questions/11356330/getting-cpu-usage-with-golang
func getCPUSample() (idle, total uint64) {
	contents, err := ioutil.ReadFile("/proc/stat")
	if err != nil {
		return
	}
	lines := strings.Split(string(contents), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if fields[0] == "cpu" {
			numFields := len(fields)
			for i := 1; i < numFields; i++ {
				val, err := strconv.ParseUint(fields[i], 10, 64)
				if err != nil {
					fmt.Println("Error: ", i, fields[i], err)
				}
				total += val // tally up all the numbers to get total ticks
				if i == 4 {  // idle is the 5th field in the cpu line
					idle = val
				}
			}
			return
		}
	}
	return
}

// cpuUsage checks to see whether or not CPU usage is below a certain %.
func cpuUsage(parameters []string) (exitCode int, exitMessage string) {
	idle0, total0 := getCPUSample()
	time.Sleep(3 * time.Second)
	idle1, total1 := getCPUSample()
	idleTicks := float32(idle1 - idle0)
	totalTicks := float32(total1 - total0)
	actualPercentUsed := 100 * (totalTicks - idleTicks) / totalTicks
	maxPercentUsed := wrkutils.ParseMyInt(parameters[0])
	if actualPercentUsed < float32(maxPercentUsed) {
		return 0, ""
	}
	msg := "CPU usage above defined maximum"
	slc := []string{fmt.Sprint(actualPercentUsed)}
	return wrkutils.GenericError(msg, fmt.Sprint(maxPercentUsed), slc)
}
