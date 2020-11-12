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


// Parse all key/values in a map
func (v *VDFFile) ParseInMap(buf []byte) (m_token map[string]string, err error) {
	v.log(fmt.Sprintf("ParseInMap()"))

	// var pairPattern = regexp.MustCompile(`(?mi)(?:^\s*")([a-z\d_:#\$]{1,})(?:"\s*")([^"\\]*(?:\\.[^"\\]*)*)"`)
	var pairPattern = regexp.MustCompile(`(?mi)^\s*"([a-z\d_:#\$]{1,})"\s*"([^"\\]*(?:\\.[^"\\]*)*)"`)
	
	if pairPattern == nil {
		return m_token, errors.New(fmt.Sprintf("ParseInMap() - no match"))
	}

	kvPairs := pairPattern.FindAllSubmatch(buf, -1)

	m_token = make(map[string]string)
	
	for _, kv := range kvPairs {
		key := string(kv[1])    // kv[0] is the full match
		value := string(kv[2])
		m_token[key] = value
	}
	return m_token, nil
}

// Parse all key/values in a slice
func (v *VDFFile) ParseInSlice(buf []byte) (s_token [][]string, err error) {
	v.log(fmt.Sprintf("ParseInSlice()"))

	// var pairPattern = regexp.MustCompile(`(?mi)(?:^\s*")([a-z\d_:#\$]{1,})(?:"\s*")([^"\\]*(?:\\.[^"\\]*)*)"`)
	var pairPattern = regexp.MustCompile(`(?mi)^\s*"([a-z\d_:#\$]{1,})"\s*"([^"\\]*(?:\\.[^"\\]*)*)"`)
	
	if pairPattern == nil {
		return s_token, errors.New(fmt.Sprintf("ParseinSlice() - no match"))
	}

	kvPairs := pairPattern.FindAllSubmatch(buf, -1)
	
	for _, kv := range kvPairs {
		s_token = append(s_token, []string{string(kv[1]), string(kv[2])}) // kv[0] is the full match
	}
	return s_token, nil
}


// Returns encoding of current file
func (v *VDFFile) GetEncoding() (string) {
	v.log("GetEncoding()")
	return v.encoding
}	
	