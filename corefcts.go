package vdfloc

import (
	// "bufio"
	// "bytes"
	// "flag"
	"fmt"
	"io"
	"errors"
	"io/ioutil"
	// "log"
	"os"
)

/*
// Skip the file BOM and determine encoding
//
// Very basic check - detection utf8 could be improved
//
// Returns:
//  - update structure with encoding details ("UTF16bomle" or "UTF16bombe" or "UTF8bom" or "UTF8")
//  - move the file pointer past the bom
//  - err if any
//
//
func (v *VDFFile) SkipBOM() error {
	v.log("SkipBOM()")

	v.f.Seek(0, 0) // rewind the file - BOM is at the begining

	const BUFF_SIZE = 4
	var UTF16bomle	= []byte{0xFF, 0xFE}
	var UTF16bombe	= []byte{0xFE, 0xFF}
	var UTF8bom		= []byte{0xEF, 0xBB, 0xBF}

	firstFewBytes := make([]byte, BUFF_SIZE)
	if _, err := v.f.Read(firstFewBytes); err != nil {
		return errors.New(fmt.Sprintf("Unable to read file - %v", err))
	}

	switch {
	case bytes.Equal(firstFewBytes[0:len(UTF16bomle)], UTF16bomle):
		v.f.Seek(int64(len(UTF16bomle)-BUFF_SIZE), 1) // remove the BOM (relative to position)
		v.encoding = "UTF16bomle"
		v.cParenth = []byte{'{', 0x00}
		v.cDbleQuote = []byte{'"', 0x00}
		v.cDbleSlash = []byte{'/', 0x00, '/', 0x00}
		v.cTab = []byte{'\t', 0x00}
		v.cBackSlash = []byte{'\\', 0x00}
		v.cLineFeed, cCarriageRet = []byte{'\n', 0x00}, []byte{'\r', 0x00}
		v.cCRLF = append([]byte{'\r', 0x00}, []byte{'\n', 0x00}...)
		v.bom = []byte{0xFF, 0xFE}

	case bytes.Equal(firstFewBytes[0:len(UTF16bombe)], UTF16bombe):
		v.f.Seek(int64(len(UTF16bombe)-BUFF_SIZE), 1) // remove the BOM (relative to position)
		v.encoding = "UTF16bombe"
		v.cParenth = []byte{0x00, '{'}
		v.cDbleQuote = []byte{0x00, '"'}
		v.cDbleSlash = []byte{0x00, '/', 0x00, '/'}
		v.cTab = []byte{0x00, '\t'}
		v.cBackSlash = []byte{0x00, '\\'}
		v.cLineFeed, cCarriageRet = []byte{0x00, '\n'}, []byte{0x00, '\r'}
		v.cCRLF = append([]byte{0x00, '\r'}, []byte{0x00, '\n'}...)
		v.bom = []byte{0xFE, 0xFF}

	case bytes.Equal(firstFewBytes[0:len(UTF8bom)], UTF8bom):
		v.f.Seek(int64(len(UTF8bom)-BUFF_SIZE), 1) // remove the BOM (relative to position)
		v.encoding = "UTF8bom"
		v.bom = []byte{0xEF, 0xBB, 0xBF}

	default:
		v.f.Seek(int64(-BUFF_SIZE), 1) // rewind file
		v.encoding = "UTF8"
	}
	return nil
}
*/


// ReadSource() Read entire source in a buffer.
//
// Process utf8 with or without bom and utf16 be/le
// Determine the encoding.
// Store the file in a slice for further procesing.
//
func (v *VDFFile) ReadSource() (buf []byte, err error) {
	v.log("ReadSource() - %s", v.name)

    // Open file
	f, err := os.Open(v.name)
    if err != nil {
        return nil, errors.New(fmt.Sprintf("ReadSource() - Can't open file %s - %v", v.name, err))
    }

    // Make a Reader
    unicodeReader, v.encoding, err := Utf8Reader(f, "")
    if err != nil {
        return nil, errors.New(fmt.Sprintf("ReadSource() - %v", err))
    }
	
    // Read, decode (if needed) and store file content in a slice (utf8 no bom)
    buf, err := ioutil.ReadAll(unicodeReader)
    if err != nil {
        return nil, errors.New(fmt.Sprintf("ReadSource() - Fail to read file %v", err))
    }

	f.Close()
	
	return buf, err
}


// SkipHeader() Skip header by moving the file beginning past the header
//
// Search for '{' outside "" strings
// and move the begining of the file to the last occurence
// Could be improved: inc by 1 char
// Weak algo, can be fooled by escaped double quotes in strings.
//
func (v *VDFFile) SkipHeader() (err error) {
	v.log("SkipHeader()")

	const BUFF_SIZE = 1024 // Read 1KB of data. That should contain the entire header
	var sizeRead int

	buf := make([]byte, BUFF_SIZE+4) // lazy (add few bytes to avoid out-of-range issues)
	if sizeRead, err = theFile.Read(buf); err != nil {
		return errors.New(fmt.Sprintf("skipHeader() - can't read file - %v", err))
	}

	relPosition := sizeRead // Reader offset

	// Scan all the data read
	for i := 0; i < sizeRead; {
		//fmt.Printf("%s",string(buf[i]))
		opening := true

		// Skip strings between "
		for {
			// Is this a double quote?
			if bytes.Equal(buf[i:i+len(v.cDbleQuote)], v.cDbleQuote) {
				if !opening {
					i += len(v.cDbleQuote)
					break
				} else {
					opening = false
				}
			}
			i++
			if i >= BUFF_SIZE { // When we reach the end of the buffer we're done
				return
			}
		}

		// Look for a '{'
		for {
			// Is this a double quote?
			if bytes.Equal(buf[i:i+len(v.cDbleQuote)], v.cDbleQuote) {
				break
			}
			// Is this a '{'?
			if bytes.Equal(buf[i:i+len(v.cParenth)], v.cParenth) {
				theFile.Seek(int64(i+len(v.cParenth)-relPosition), 1) // skip the header
				// fmt.Printf("\nSeek relative=%d and i=%d\n",int64(i + len(cParenth) - relPosition),i)
				relPosition = i + len(v.cParenth) // update reader offset
				// fmt.Printf("relPosition=%d\n", relPosition)
				// fmt.Printf("\nbuf=%s\n",buf[i-5:i+5])
			}
			i++
			if i >= BUFF_SIZE {
				return // When we reach the end of the buffer we're done
			}
		}
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

// Parse the file into a map
//
//
func (v *VDFFile) ParseFile() (fileMapped map[string]string, err error) {
	v.log("ParseFile(%s), v.fileName")

	fs, err := v.f.Stat() // Get file size
	if err != nil {
		return fileMapped, errors.New(fmt.Sprintf("Can't read file size : %v", err))
	}

	err = SkipBom()
	if err != nil {
		return fileMapped, err
	}

	// Skip headers for both files
	err = skipHeader()
	if err != nil {
		return fileMapped, err
	}

	// Create a map and populate with file content - using string as index but it's actually []byte
	fileMapped = make(map[string]string)

	scannerf := bufio.NewScanner(v.f)

	//adjust the capacity (file max characters) - the default size is 4096 bytes!!!!
	maxCapacity := fi1.Size() + 1024
	buf := make([]byte, maxCapacity)
	scannerf.Buffer(buf, int(maxCapacity))

	// Set the split function
	scannerf.Split(v.onDblQuotes)

	// Populate map with the tokens key/value pairs
	for {
		if !scannerf.Scan() {
			break
		}
		key := scannerf.Bytes()

		if !scannerf.Scan() {
			break
		}
		value := scannerf.Bytes()

		fileMapped[string(key)] = value
	}

	emptyFile := true

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
		fmt.Printf("File contents are identical or %s is a subset of %s\n", firstFileName, secondFileName)
	}

}



