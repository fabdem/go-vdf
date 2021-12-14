package vdfloc

// Package vdfloc
//	Toolbox of functions to deal with valve vdf loc files
//	Compatible with utf8 and utf16BE encoding

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

type VDFFile struct {
	pathAndName string // loc file path and name
	fileName    string
	f           *os.File
	encoding    string
	logWriter   io.Writer
	sourceTkn   bool // Define whether we keep the [english] tokens or not
	debug       bool
	maxKeyLen	int  // Maximum autorised char length of keys
	// cParenth     []byte
	// cDbleQuote   []byte
	// cDbleSlash   []byte
	// cBackSlash   []byte
	// cLineFeed    []byte
	// cCarriageRet []byte
	// cCRLF        []byte
	// cTab         []byte
	// bom          []byte
}

var g_debug bool
var g_logWriter io.Writer

// Create a new instance
// - In: File name and path
// - Returns instance and error code
func New(filePathAndName string) (*VDFFile, error) {

	// validate parameter
	if filePathAndName == "" {
		return nil, fmt.Errorf("File name cannot be empty")
	}

	v := &VDFFile{} // Create instance

	var err error
	v.pathAndName = filePathAndName
	v.fileName = filepath.Base(filePathAndName)
	v.debug = g_debug
	v.logWriter = g_logWriter
	v.sourceTkn = false // default behavior: we ignore tokens names including "[english]"
	v.maxKeyLen = 120    // characters - default maximum autorised length for keys

	// Open the file for reading
	f, err := os.Open(filePathAndName)
	if err != nil {
		return nil, fmt.Errorf("Unable to open file %s - %v", filePathAndName, err)
	}

	v.f = f
	// Default encoding: utf8 no bom
	// v.cParenth = []byte{'{'}
	// v.cDbleQuote = []byte{'"'}
	// v.cDbleSlash = []byte{'/', '/'}
	// v.cTab = []byte{'\t'}
	// v.cBackSlash = []byte{'\\'}
	// v.cLineFeed, v.cCarriageRet = []byte{'\n'}, []byte{'\r'}
	// v.cCRLF = append([]byte{'\r'}, []byte{'\n'}...)

	return v, nil
}

// Release instance
// Close the file and release the structure.
func Close(v *VDFFile) (err error) {
	err = v.f.Close()
	v = nil
	return err
}

// Set the flag to keep token names with [english] tag
func (v *VDFFile) SetKeepSourceTokens() {
	v.sourceTkn = true
}

// Reset the flag to filter out token names with [english] tag
func (v *VDFFile) ResetKeepSourceTokens() {
	v.sourceTkn = false
}

// Read the flag to keep (true) or filter (flase) token names with [english] tag
func (v *VDFFile) GetKeepSourceTokenFlag() bool {
	return v.sourceTkn
}

// Set max autorised char key length
func (v *VDFFile) SetMaxKeyLen(val int) {
	v.maxKeyLen = val
}

// Read max autorised char key length
func (v *VDFFile) ReadMaxKeyLen() int {
	return v.maxKeyLen
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
