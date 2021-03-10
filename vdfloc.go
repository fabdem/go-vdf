package vdfloc

// Package vdfloc
//	Toolbox of functions to deal with valve vdf loc files
//	Compatible with utf8 and utf16BE encoding

import (
	// "bufio"
	// "bytes"
	// "flag"
	// "io/ioutil"
)

type VDFFile struct {
	name         string
	f            *os.File
	encoding     string
	logWriter	 io.Writer
	debug        bool
	cParenth     []byte
	cDbleQuote   []byte
	cDbleSlash   []byte
	cBackSlash   []byte
	cLineFeed    []byte
	cCarriageRet []byte
	cCRLF        []byte
	cTab         []byte
	bom          []byte
}

var	g_debug bool
var	g_logWriter io.Writer

// Create a new instance
// - lookup path to p4 command
// - Returns instance and error code
func New(fileName string) (*VDFFile, error) {

	// validate parameter
	if fileName == "" {
		return nil, errors.New(fmt.Sprintf("File name cannot be empty"))
	}

	v := &VDFFile{} // Create instance

	var err error
	v.name = fileName
	v.debug = g_debug
	v.logWriter = g_logWriter

	// Open the file for reading
	f, err := os.Open(fileName)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Unable to open file %s - %v", fileName, err))
	}

	v.f = f
	// Default encoding: utf8 no bom
	v.cParenth = []byte{'{'}
	v.cDbleQuote = []byte{'"'}
	v.cDbleSlash = []byte{'/', '/'}
	v.cTab = []byte{'\t'}
	v.cBackSlash = []byte{'\\'}
	v.cLineFeed, v.cCarriageRet = []byte{'\n'}, []byte{'\r'}
	v.cCRLF = append([]byte{'\r'}, []byte{'\n'}...)

	return v, nil
}

// Release instance
// Close the file and release the structure.
func Close(v *VDFFile) (err error) {
	err = v.f.Close()
	v = nil
	return err
}

// SetDebug()
//
// Enable or disable log for all instances created from this point
// Traces errors if it's set to true.
func SetDebug(debug bool, logWriter io.Writer) {
	g_debug = debug
	g_logWriter = logWriter
}

// Log writer
func (v *VDFFile) log(a interface{}) {
	if v.debug {
		if v.logWriter != nil {
			timestamp := time.Now().Format(time.RFC3339)
			msg := fmt.Sprintf("%v: %v", timestamp, a)
			fmt.Fprintln(v.logWriter, msg)
		} else {
			log.Println(a)
		}
	}
}