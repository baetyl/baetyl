package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/baetyl/baetyl"
	_ "github.com/baetyl/baetyl/docker"
	_ "github.com/mattn/go-sqlite3"
)

// Brand of this application
var Brand string = "baetyl"

// Run baetyl from command line
func Run(args []string, env []string) error {
	fs := flag.NewFlagSet(Brand, flag.ContinueOnError)

	var help = fs.Bool("h", false, "this help")
	var showVersion = fs.Bool("v", false, "show version and exit")
	var showVersionAndOptions = fs.Bool("V", false, "show version and configure options then exit")
	var prefix = fs.String("p", baetyl.DefaultPrefix, fmt.Sprintf("set prefix path (default: %s)", baetyl.DefaultPrefix))
	var configPath = fs.String("c", baetyl.DefaultConfigPath, fmt.Sprintf("set configuration file (default: %s)", baetyl.DefaultConfigPath))
	//var directives = flag.String("g", "", "set global directives out of configuration file")
	var debug = fs.Bool("debug", false, "run in debug mode")

	err := fs.Parse(args)
	if err != nil {
		return err
	}

	if *help {
		fmt.Printf("%s version: baetyl/%s\n", Brand, baetyl.Version)
		fmt.Printf("Usage: %s [-hvV] [-c filename] [-p prefix] [-g directives] [--debug])\n", Brand)
		fmt.Println("")
		fmt.Println("Options:")
		fmt.Printf("  -h      : %s\n", fs.Lookup("h").Usage)
		fmt.Printf("  -v      : %s\n", fs.Lookup("v").Usage)
		fmt.Printf("  -V      : %s\n", fs.Lookup("V").Usage)
		fmt.Printf("  -p      : %s\n", fs.Lookup("p").Usage)
		fmt.Printf("  -c      : %s\n", fs.Lookup("c").Usage)
		fmt.Printf("  --debug : %s\n", fs.Lookup("debug").Usage)
		return nil
	}

	if *showVersion {
		fmt.Printf("%s version: baetyl/%s\n", Brand, baetyl.Version)
		return nil
	}

	if *showVersionAndOptions {
		fmt.Printf("%s version: baetyl/%s\n", Brand, baetyl.Version)
		fmt.Println("Configure options:")
		fmt.Printf("  PREFIX=%s\n", baetyl.DefaultPrefix)
		fmt.Printf("  CONF_PATH=%s\n", baetyl.DefaultConfigPath)
		fmt.Printf("  LOG_PATH=%s\n", baetyl.DefaultLoggerPath)
		fmt.Printf("  DATA_PATH=%s\n", baetyl.DefaultDataPath)
		fmt.Printf("  PID_PATH=%s\n", baetyl.DefaultPidPath)
		fmt.Printf("  API_ADDR=%s\n", baetyl.DefaultAPIAddress)
		return nil
	}

	if *debug {
		return errors.New("not implement debug flag yet")
		//cfg.Daemon = false
		//cfg.Logger.Level = "debug"
		//cfg.OTALog.Level = "debug"
	}

	return baetyl.Run(*prefix, *configPath)
}

func main() {
	if err := Run(os.Args[1:], os.Environ()); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err.Error())
		os.Exit(-1)
	}
}
