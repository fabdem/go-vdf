package main

//	F.Demurger 2021-12
//  	args:
//			vdfloccheck <options> <filename>
//			e.g. vdfloccheck csgo_latam.txt
//
//      Option:
//				-v version
//				-d debug/log file
//
//	Perform a sanity check on vdf localization files.
//		- check key/value/[cond statements]
//		- check that there are no isolated (with no key/value) conditional statements.
//		- check unicity of keys
//
//	Err codes:
//		0: All good
//		1:
//		2: key unicity error 
//
//
//	cross compilation AMD64:  env GOOS=windows GOARCH=amd64 go build vdfloccheck.go

import (
	// "bytes"
	// "errors"
	"flag"
	"fmt"
	"go-vdfloc"
	"io"
	"log"
	"os"
	"time"
	"strings"
	// "golang.org/x/text/encoding"
	// "golang.org/x/text/encoding/unicode"
)

var g_logFile *os.File // log

func main() {

	var versionFlg bool
	var debug string
	var err error

	const usageVersion = "Display Version"
	const usageDebug = "Specify a debug file"

	// Have to create a specific set, the default one is poluted by some test stuff from another lib (?!)
	checkFlags := flag.NewFlagSet("check", flag.ExitOnError)

	checkFlags.BoolVar(&versionFlg, "version", false, usageVersion)
	checkFlags.BoolVar(&versionFlg, "v", false, usageVersion+" (shorthand)")
	checkFlags.StringVar(&debug, "debug", "", usageDebug)
	checkFlags.StringVar(&debug, "d", "", usageDebug+" (shorthand)")
	checkFlags.Usage = func() {
		fmt.Printf("Usage: %s <options> <filename>\n", os.Args[0])
		fmt.Println(" Check a vdf loc file to:\n")
		fmt.Println(" - make sure keys are unique\n")
		fmt.Println(" - make sure there is no isolated conditional statement.\n")
		fmt.Println(" Returns an error any of these conditions is not met\n")
		fmt.Println(" plus print out the list of non unique keys or isolated \n")
		fmt.Println(" conditional statements.\n\n")
		checkFlags.PrintDefaults()
	}

	// Check parameters
	checkFlags.Parse(os.Args[1:])

	if versionFlg {
		fmt.Printf("Version %s\n", "2021-12  v1.0.6")
		os.Exit(0)
	}

	// Parse the command parameters
	index := len(os.Args)
	if index < 2 {
		fmt.Printf("Not enough parameters defined\n")
		fmt.Printf("Check usage with option -help\n")
		os.Exit(1)
	}

	filename := os.Args[index-1]

	// Set a log file if one is defined
	if len(debug) > 0 {
		// append to a log file if it exists otherwise create it
		g_logFile, err = os.OpenFile(debug, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil { // Report but ignore the error. Logs will go to std err
			log.Println(fmt.Sprintf("%s - Can't open or create debug file %s", os.Args[0], debug))
		}
		SetDebug(true, g_logFile)
		vdfloc.SetDebug(true, g_logFile)		
	}

	logThis(fmt.Sprintf("%s Params - file to processed: %s ", os.Args[0], filename))

	// Get encoding and lines with conditional statements
	v, err := vdfloc.New(filename)
	if_err_fatal(err, fmt.Sprintf("Error accessing file %s", filename))
	

	// Filter in tokens with a conditional statement
	// slice, err := v.GetStringsWithConditionalStatement()
	// if_err_fatal(err, fmt.Sprintf("Error parsing file %s", filename))

	// fileEncoding := v.GetEncoding()

	globalErr := false
	
	// Parse tokens
	buf, err := v.ReadSource()
	if_err_fatal(err, fmt.Sprintf("Error accessing file %s", filename))

	res, err := v.SkipHeader(buf)
	if_err_fatal(err, fmt.Sprintf("Error reading vdf header of %s", filename))

	tokens, err := v.ParseInSlice(res)
	if_err_fatal(err, fmt.Sprintf("Error parsing vdf header of %s", filename))


	// Check unicity of keys
	listBrokenUnicity, err := v.CheckKeyUnicity(tokens)
	if err != nil {
		fmt.Printf("Error checking unicity %s \nOffending tokens: %s\n", err, listBrokenUnicity)
		globalErr = true
	}

	// Look for isolated conditional statements
	listIsolatedCond,err := v.CheckIsolatedConditionalStatements(buf)
	if err != nil {
		fmt.Printf("Error isolated conditional statements %s \nOffending tokens: %s\n", err, listIsolatedCond)
		globalErr = true
	}

	// Check key validity to detect corrupted vdf
	listinvalidKeys,err := v.CheckKeyValidity(tokens)
	if err != nil {
		fmt.Printf("Invalid keys %s \nOffending tokens: %s\n", err, listinvalidKeys)
		globalErr = true
	}

	// Check values shouldn't be "[empty string]" (not really a vdf requirement though)
	var listWrongVal []string
	err_flag := false
	for _, tkn := range tokens {
		// fmt.Printf("%s\n",strings.ToUpper(tkn[2]))
		if strings.ToUpper(tkn[2]) == "[EMPTY STRING]" {
			listWrongVal = append(listWrongVal, tkn[1] + " " + tkn[2])
			err_flag = true
		}
	}
	if err_flag {
		fmt.Printf("Invalid value(s) found: %s\n", listWrongVal)
		globalErr = true
	}

	// Exiting
	vdfloc.Close(v)
	g_logFile.Close()
	if globalErr == false {
		fmt.Println("All good")
		os.Exit(0)
	} else {
		os.Exit(1)
	}
}

// fileExists()
//
// Checks if a file exists and is not a directory.
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// if_err_fatal()
//
// Catch all fatal error cases
//	If there is an error:
//  - string is displayed on stdout
//	- then we quit with error code 1 + log
func if_err_fatal(err error, msg string) {
	if err != nil {
		fmt.Printf("\n%s\n%s\n", msg, err)
		logThis(fmt.Sprintf("Fatal:%s - %v", msg, err))
		g_logFile.Close()
		os.Exit(1)
	}
}

// Log stuff:
//
// Enable or disable log
// define optional log writer
// Log on the std error if no writer defined.
// etc.

type Mylog struct {
	debug     bool
	logWriter io.Writer
}

var l Mylog

// SetDebug()
//
// Enable or disable log
func SetDebug(debug bool, logWriter io.Writer) {
	l.debug = debug
	l.logWriter = logWriter
}

// GetDebugWriter()
//
// Return log writer
func GetDebugWriter() (logWriter io.Writer) {
	return l.logWriter
}

// logThis()
//
// Actual logger
func logThis(a interface{}) {
	if l.debug {
		if l.logWriter != nil {
			timestamp := time.Now().Format(time.RFC3339)
			msg := fmt.Sprintf("%v: %v", timestamp, a)
			fmt.Fprintln(l.logWriter, msg)
		} else {
			log.Println(a)
		}
	}
}
