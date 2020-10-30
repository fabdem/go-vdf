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

	firstFileName := os.Args[1]
	secondFileName := os.Args[2]
	if tokenOnlyFlg { // Shift if token diff
		firstFileName,secondFileName = secondFileName,os.Args[3]
	}

	// Let's start with checking both file's formats (does it starts with FFFE?)
	// Game files are apparently UTF16 BOM the rest is UTF8 so check that and consistency

	f1, err1 := os.Open(firstFileName)
	if err1 != nil {
		log.Fatalf("Can't open : %s - %s", firstFileName, err1)
	}
	defer f1.Close()

	fi1, err1 := f1.Stat() // Get file size
	if err1 != nil {
		log.Fatalf("Can't read file size : %s - %s", firstFileName, err1)
	}
	encoding := guessEncoding(f1)

	f2, err2 := os.Open(secondFileName)
	if err2 != nil {
		log.Fatalf("Can't open : %s - %s", secondFileName, err2)
	}
	defer f2.Close()

	fi2, err2 := f2.Stat() // Get file size
	if err2 != nil {
		log.Fatalf("Can't read file size : %s - %s", secondFileName, err2)
	}

	switch {
		case encoding == "UTF16bomle":
		case encoding == "UTF16bombe":
		case encoding == "UTF8bom":
		case encoding == "UTF8":
		default:
			log.Fatalf("Only files encoded in UTF16 BOM LE/BE or UTF8/BOM can be processed.\n")
	}

	if encoding != guessEncoding(f2) {
		log.Fatalf("Files have different encoding!\n")
	}


	// Defines the encoding of the special chars
	switch {
		case encoding == "UTF8":
			// Set special char encoding for UTF8:
			cParenth = []byte{'{'}
			cDbleQuote = []byte{'"'}
			cDbleSlash = []byte{'/', '/'}
			cTab = []byte{'\t'}
			cBackSlash = []byte{'\\'}
			cLineFeed, cCarriageRet = []byte{'\n'}, []byte{'\r'}
			cCRLF = append([]byte{'\r'}, []byte{'\n'}...)

		case encoding == "UTF8bom":
			// Set special char encoding for UTF8bom:
			cParenth = []byte{'{'}
			cDbleQuote = []byte{'"'}
			cDbleSlash = []byte{'/', '/'}
			cTab = []byte{'\t'}
			cBackSlash = []byte{'\\'}
			cLineFeed, cCarriageRet = []byte{'\n'}, []byte{'\r'}
			cCRLF = append([]byte{'\r'}, []byte{'\n'}...)
			bom = []byte{0xEF, 0xBB, 0xBF}

		case encoding == "UTF16bomle":
			// Set special char encoding for UTF16bomle:
			cParenth = []byte{'{', 0x00}
			cDbleQuote = []byte{'"', 0x00}
			cDbleSlash = []byte{'/', 0x00, '/', 0x00}
			cTab = []byte{'\t', 0x00}
			cBackSlash = []byte{'\\', 0x00}
			cLineFeed, cCarriageRet = []byte{'\n', 0x00}, []byte{'\r', 0x00}
			cCRLF = append([]byte{'\r', 0x00}, []byte{'\n', 0x00}...)
			bom = []byte{0xFF, 0xFE}

		case encoding == "UTF16bombe":
			// Set special char encoding for UTF16bomle:
			cParenth = []byte{0x00, '{'}
			cDbleQuote = []byte{0x00, '"'}
			cDbleSlash = []byte{0x00, '/', 0x00, '/'}
			cTab = []byte{0x00, '\t'}
			cBackSlash = []byte{0x00, '\\'}
			cLineFeed, cCarriageRet = []byte{0x00,'\n'}, []byte{0x00,'\r'}
			cCRLF = append([]byte{0x00,'\r'}, []byte{0x00,'\n'}...)
			bom = []byte{0xFE, 0xFF}

		default:
			// We should never end up here
			log.Fatalf("Only files encoded in UTF16 BOM LE or UTF8/bom can be processed.\n")
	}

	// Skip headers for both files
	skipHeader(f1)
	skipHeader(f2)

	/* // Debug
	fmt.Printf("%s", bom) // Add a bom if needed
	bufdebug := make([]byte, 3000000)
	if _, err := f1.Read(bufdebug); err != nil {
		log.Fatalf("skipHeader() - can't read file - %s", err)
	}
	if err := ioutil.WriteFile("./out1.txt", bufdebug, 0644) ; err != nil  {
		log.Fatalf("skipHeader() - can't write debug file - ./out1.txt", err)
	} // Debug */


	// Create a map and populate with f2 - using string as index but it's actually []byte
	fileMapped := make(map[string]bool)

	scannerf1 := bufio.NewScanner(f1)
	scannerf2 := bufio.NewScanner(f2)

	//adjust the capacity (file max characters) - the default size is 4096 bytes!!!!
	var maxCapacity int64

	if fi1.Size() >= fi2.Size() { // ...from the largest of the 2 files
		maxCapacity = fi1.Size() + 1024
	} else {
		maxCapacity = fi2.Size() + 1024
	}
	buf := make([]byte, maxCapacity)
	scannerf1.Buffer(buf, int(maxCapacity)) // Used to scan f1
	scannerf2.Buffer(buf, int(maxCapacity)) // Used to scan f2 (same buffer)

	// Set the split function
	scannerf1.Split(onDblQuotes)
	scannerf2.Split(onDblQuotes)

	// Populate map with the tokens key/value pairs
	for {
		if !scannerf2.Scan() {
			break
		}
		key := scannerf2.Bytes()

		if !scannerf2.Scan() {
			break
		}
		value := scannerf2.Bytes()

		//fmt.Printf("%s", append(key, value...))
		//fmt.Printf("%s", []byte{0x0D, 0x00, 0x0A, 0x00})
		if !tokenOnlyFlg {
			fileMapped[string(append(key, value...))] = true
		} else {
			fileMapped[string(key)] = true	 // Option -t used (compares token names only)
		}

	}

	emptyFile := true

	/* Debug:
	fmt.Printf("%s", bom) // Add a bom
	emptyFile = false    // and file is not empty anymore

	for k, _ := range fileMapped {
		fmt.Printf("%s", k)
		fmt.Printf("%s", cCRLF)
	}
	fmt.Printf("********************************************%s", cCRLF)
	fmt.Printf("********************************************%s", cCRLF)
	fmt.Printf("********************************************%s", cCRLF)
	*/

	// Compare 2nd file with the map and build the diff file
	for {
		if !scannerf1.Scan() {
			break
		}
		key := scannerf1.Bytes()
		key1 := make([]byte, len(key)) // Weird the fileMapped modifies var value - had to make a copy in order to use the var within the if
		copy(key1, key)

		if !scannerf1.Scan() {
			break
		}
		value := scannerf1.Bytes()
		value1 := make([]byte, len(value))
		copy(value1, value)

		if !tokenOnlyFlg {
			combined := append(key, value...)
			if !fileMapped[string(combined)] {
				if emptyFile {
					emptyFile = false
					fmt.Printf("%s", bom) // Add a bom if needed
				}
				// It's a diff -> print out
				fmt.Printf("%s%s%s%s", key1, cTab, value1, cCRLF)
			}

		} else { // Option -t used (compares token names only)
			if !fileMapped[string(key)] {
				if emptyFile {
					emptyFile = false
					fmt.Printf("%s", bom) // Add a bom if needed
				}
				// It's a diff -> print out
				fmt.Printf("%s%s%s%s", key1, cTab, value1, cCRLF)
			}
		}
	}

	// If file content is the same say so in UTF8
	if emptyFile {
		fmt.Printf("File contents are identical or %s is a subset of %s\n", firstFileName,secondFileName)
	}

}

// Define a split function that separates strings (anything between valid double quotes).
func onDblQuotes(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if err != nil {
		log.Fatalf("Error while scanning: %s ", err)
	}
	var beg int
	var i int

	// Seek the opening double quote or '//'
	for {
		// Is this a double quote?
		if bytes.Equal(data[i:i+len(cDbleQuote)], cDbleQuote) {
			beg = i
			break
		}

		// Is this a '//'?
		if (i <= len(data)-len(cDbleSlash)) && bytes.Equal(data[i:i+len(cDbleSlash)], cDbleSlash) {
			// Skip everything until we get to a line feed
			i += len(cDbleSlash)
			for {
				if i >= len(data)-1 {
					return 0, data[0:1], bufio.ErrFinalToken // last line: empty
				}
				if bytes.Equal(data[i:i+len(cLineFeed)], cLineFeed) {
					beg = i + len(cLineFeed)
					break
				}
				i++
			}
		}

		i++
		if i >= len(data)-1 {
			return 0, data[0:1], bufio.ErrFinalToken // last line: empty
		}
	}

	// Seek the closing (non escaped) double quote
	for i = beg + len(cDbleQuote); i < len(data)-1; i++ {
		if bytes.Equal(data[i:i+len(cDbleQuote)], cDbleQuote) && !bytes.Equal(data[i-len(cBackSlash):i], cBackSlash) {
			return i + len(cDbleQuote), data[beg : i+len(cDbleQuote)], nil // Fine this is the end of the token
		}
	}
	// There is one final line to be delivered, which is an empty string.
	// Returning bufio.ErrFinalToken here tells Scan there are no more tokens after this
	// but does not trigger an error to be returned from Scan itself.
	return 0, data[0:1], bufio.ErrFinalToken
}
