package main

import (
	"github.com/CiscoCloud/distributive/errutil"
	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"net/url"
	"os"
)

// validateFlags ensures that all options passed via the command line are valid
func validateFlags(file string, URL string, directory string) {
	// validatePath ensures that something is at a given path
	validatePath := func(path string) {
		if _, err := os.Stat(path); err != nil {
			errutil.CouldntReadError(path, err)
		}
	} // validateURL ensures that the given URL is valid, or logs an error
	validateURL := func(urlstr string) {
		if _, err := url.Parse(urlstr); err != nil {
			log.WithFields(log.Fields{
				"url":   urlstr,
				"error": err.Error(),
			}).Fatal("Couldn't parse URL")
		}
	}
	if URL != "" {
		validateURL(URL)
	}
	if directory != "" {
		validatePath(directory)
	}
	if file != "" {
		validatePath(file)
	}
}

// initializeLogrus sets the logrus log level according to the specified
// verbosity level, both for packages main and chkutils
func initializeLogrus(verbosity string) {
	lvls := "info | debug | fatal | panic | warn"
	var logLevel log.Level
	logLevel = 0
	switch verbosity {
	case "info":
		logLevel = log.InfoLevel
	case "debug":
		logLevel = log.DebugLevel
	case "fatal":
		logLevel = log.FatalLevel
	case "error":
		logLevel = log.ErrorLevel
	case "panic":
		logLevel = log.PanicLevel
	case "warn":
		logLevel = log.WarnLevel
	default:
		log.WithFields(log.Fields{
			"given":    verbosity,
			"expected": lvls,
		}).Fatal("Invalid verbosity option")
	}
	log.SetLevel(logLevel)
	log.WithFields(log.Fields{
		"verbosity": verbosity,
	}).Debug("Verbosity level specified")
}

// getFlags validates and returns command line options
func getFlags() (f string, u string, d string, s bool) {
	lvls := "info | debug | fatal | error | panic | warn"
	defaultVerbosity := "warn"

	app := cli.NewApp()
	app.Name = "Distributive"
	app.Usage = "Perform distributed health tests"
	app.Version = Version
	app.Author = "Langston Barrett"
	/* For a newer version of cli:
	app.Authors = []cli.Author{
		cli.Author{
			Name:  "Langston Barrett",
			Email: "langston@aster.is",
		},
	}
	*/
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "verbosity",
			Value: defaultVerbosity,
			Usage: lvls,
		},
		cli.StringFlag{
			Name:  "file, f",
			Value: "",
			Usage: "Read a checklist from a file",
		},
		cli.StringFlag{
			Name:  "url, u",
			Value: "",
			Usage: "Read a checklist from a URL",
		},
		cli.StringFlag{
			Name:  "directory, d",
			Value: "",
			Usage: "Read all of the checklists in this directory",
		},
		cli.BoolFlag{
			Name:  "stdin, s",
			Usage: "Read data piped from stdin as a checklist",
		},
		cli.BoolFlag{
			Name:  "no-cache",
			Usage: "Don't use a cached version of a remote check, fetch it.",
		},
	}
	var verbosity string
	var file string
	var URL string
	var directory string
	var stdin bool
	app.Action = func(c *cli.Context) {
		version := c.Bool("version")
		if version {
			os.Exit(0)
		}
		verbosity = c.String("verbosity")
		file = c.String("file")
		URL = c.String("url")
		directory = c.String("directory")
		stdin = c.Bool("stdin")

		if file == "" && URL == "" && stdin == false && directory == "" {
			// use default directory if no other options specified
			directory = "/etc/distributive.d/"
		}
		log.WithFields(log.Fields{
			"file":      file,
			"URL":       URL,
			"directory": directory,
			"stdin":     stdin,
		}).Debug("Command line options")
		useCache = !c.Bool("no-cache")
	}
	if verbosity == "" {
		verbosity = "warn"
	}
	app.Run(os.Args)            // parse the arguments, execute app.Action
	initializeLogrus(verbosity) // set logLevel appropriately for chkutils
	return file, URL, directory, stdin
}
