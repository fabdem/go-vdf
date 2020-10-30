package vdfloc

// Package vdfloc
//	Toolbox of functions to deal with valve vdf loc files
//	Compatible with utf8 and utf16BE encoding

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	// "io/ioutil"
	"log"
	"os"
)


type VDFFile struct {
	name						string
	f								*os.File
	encoding				string
	logWriter       io.Writer
	debug           bool
	cParenth				[]byte
	cDbleQuote			[]byte
	cDbleSlash			[]byte
	cBackSlash			[]byte
	cLineFeed				[]byte
	cCarriageRet		[]byte
	cCRLF						[]byte
	cTab						[]byte
	bom							[]byte
}

// Create a new instance
// - lookup path to p4 command
// - Returns instance and error code
func New(fileName string) (*VDFFile, error) {
	v := &VDFFile{} // Create instance

	var err error
	v.name = fileName
	v.debug = false // default

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
	v.cLineFeed, cCarriageRet = []byte{'\n'}, []byte{'\r'}
	v.cCRLF = append([]byte{'\r'}, []byte{'\n'}...)

	return v, nil
}



func main() {

	var versionFlg bool
	var tokenOnlyFlg bool
	const usageVersion   = "Display Version"
	const diffTokensOnly   = "Diff tokens only"

	flag.BoolVar(&versionFlg, "version", false, usageVersion)
	flag.BoolVar(&versionFlg, "v", false, usageVersion + " (shorthand)")
	flag.BoolVar(&tokenOnlyFlg, "tokens", false, diffTokensOnly)
	flag.BoolVar(&tokenOnlyFlg, "t", false, diffTokensOnly + " (shorthand)")
	flag.Usage = usageIs  // Display app usage

	flag.Parse()

	if versionFlg {
		fmt.Printf("Version %s\n", "2019-02-04  v1.2.5")
		os.Exit(0)
	}

	if len(os.Args) < 2 {
		usageIs()  // Display usage
		log.Fatalf("Missing parameters\n")
	}
