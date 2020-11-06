package vdfloc

import (
	// "bufio"
	// "bytes"
	// "flag"
	"errors"
	"fmt"
	// "io"
	"io/ioutil"
	// "log"
	"os"
	"regexp"
)

// ReadSource() Read entire source in a buffer.
//
// Process utf8 with or without bom and utf16 be/le
// Determine the encoding.
// Store the file in a slice for further procesing.
//
func (v *VDFFile) ReadSource() (buf []byte, err error) {
	v.log(fmt.Sprintf("ReadSource() - %s", v.name))

	// Open file
	f, err := os.Open(v.name)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("ReadSource() - Can't open file %s - %v", v.name, err))
	}

	// Make a Reader
	// unicodeReader, v.encoding, err := UTFReader(f, "")
	unicodeReader, enc, err := UTFReader(f, "")
	if err != nil {
		return nil, errors.New(fmt.Sprintf("ReadSource() - %v", err))
	}
	v.encoding = enc

	// Read, decode (if needed) and store file content in a slice (utf8 no bom)
	buf, err = ioutil.ReadAll(unicodeReader)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("ReadSource() - Fail to read file %v", err))
	}

	f.Close()

	return buf, err
}

// SkipHeader() Skip vdf "header" by removing it from the buffer.
// Returns the same buffer but without header.
//
func (v *VDFFile) SkipHeader(buf []byte) (res []byte) {
	v.log("SkipHeader()")

	var getPattern = regexp.MustCompile(`(?mi)^\s*"[a-z]{1,15}"\s*\{`)

	res = buf
	for {
		if idx := getPattern.FindIndex(res); idx != nil {
			// Pattern found
			res = res[idx[1]+1:]
		} else {
			// Done
			break
		}
	}
	return res
}

func (v *VDFFile) Parse(buf []byte) (m_token map[string]string, err error) {
	v.log(fmt.Sprintf("Parse"))

	var tokenPattern = regexp.MustCompile(`(?mi)^\s*"[a-z]{1,15}"\s*\{`)


}

/*
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
*/

/*
// Parse the file into a map
//
//
func (v *VDFFile) ParseFile() (fileMapped map[string]string, err error) {
	v.log(fmt.Sprintf("ParseFile(%s)", v.fileName))

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
*/