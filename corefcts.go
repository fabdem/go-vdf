package vdfloc

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
)

// ReadSource() Read entire source in a buffer.
//
// Process utf8 with or without bom and utf16 be/le
// Determine the encoding.
// Store the file in a slice for further procesing.
//
func (v *VDFFile) ReadSource() (buf []byte, err error) {
	v.log(fmt.Sprintf("ReadSource() - %s", v.pathAndName))

	// Open file
	f, err := os.Open(v.pathAndName)
	if err != nil {
		return nil, fmt.Errorf("ReadSource() - Can't open file %s - %v", v.pathAndName, err)
	}

	// Make a Reader
	// unicodeReader, v.encoding, err := UTFReader(f, "")
	unicodeReader, enc, err := UTFReader(f, "")
	if err != nil {
		return nil, fmt.Errorf("ReadSource() - %v", err)
	}
	v.encoding = enc

	// Read, decode (if needed) and store file content in a slice (utf8 no bom)
	buf, err = ioutil.ReadAll(unicodeReader)
	if err != nil {
		return nil, fmt.Errorf("ReadSource() - Fail to read file %v", err)
	}

	f.Close()

	return buf, err
}

// SkipHeader() Skip vdf "header" by removing it from the buffer.
// Returns the same buffer but without header.
// And a very unlikely Error
//
func (v *VDFFile) SkipHeader(buf []byte) (res []byte, err error) {
	v.log("SkipHeader()")

	getPattern, err := regexp.Compile(`(?mi)^\s*"[a-z]{1,15}"\s*\{`)
	if err != nil {
		return res, fmt.Errorf("Err in regEx: %v", err)
	}

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
	return res, nil
}

// GetHeader() Returns the "vdf header".
// And a very unlikely Error
//
func (v *VDFFile) GetHeader(buf []byte) (res []byte, err error) {
	v.log("GetHeader()")

	getPattern, err := regexp.Compile(`(?mi)^\s*"[a-z]{1,15}"\s*\{`)
	if err != nil {
		return res, fmt.Errorf("Err in regEx: %v", err)
	}

	idxes := getPattern.FindAllIndex(buf, -1)

	if idxes == nil {
		return nil, nil
	}
	lastIdx := idxes[len(idxes)-1][1]
	return buf[0:lastIdx], nil
}

// ParseInMap()
//
// Parse all key/values in a map
func (v *VDFFile) ParseInMap(buf []byte) (m_token map[string]string, err error) {
	v.log(fmt.Sprintf("ParseInMap()"))

	regex := `(?mi)^\s*"([a-z\d_:#\$\[\]!&\|]{1,})"\s*"([^"\\]*(?:\\.[^"\\]*)*)"`

	// var pairPattern = regexp.MustCompile(`(?mi)(?:^\s*")([a-z\d_:#\$]{1,})(?:"\s*")([^"\\]*(?:\\.[^"\\]*)*)"`)
	pairPattern, err := regexp.Compile(regex)

	if err != nil {
		return m_token, fmt.Errorf("Err in regEx: %v", err)
	}

	kvPairs := pairPattern.FindAllSubmatch(buf, -1)

	m_token = make(map[string]string)

	for _, kv := range kvPairs {
		key := string(kv[1]) // kv[0] is the full match
		value := string(kv[2])
		if !strings.HasPrefix(key, "[english]") || v.sourceTkn { // Add token if key doesn't start with [english] or we want to capture everything
			m_token[key] = value
		}
	}
	return m_token, nil
}

// ParseInSlice()
//
// Parse all keys/values/cond statements/comments in a slice
// 		E.g. "a_key"	"a value" [$WIN32]	// A comment
//		slice[0]: the entire line: "a_key"	"a value" [$WIN32]	// A comment
//		slice[1]: a_key
//		slice[2]: a value
//		slice[3]: [$WIN32]
//		slice[4]: // A comment
//
func (v *VDFFile) ParseInSlice(buf []byte) (s_token [][]string, err error) {
	v.log(fmt.Sprintf("ParseInSlice()"))

	// regex = `(?mi)^\s*"([a-z\d_:#\$\[\]]{1,})"\s*"([^"\\]*(?:\\.[^"\\]*)*)"`
	// regex = `(?mi)^\s*"([a-z\d_:#\$\[\]]{1,})"\s*"([^"\\]*(?:\\.[^"\\]*)*)"(?:(?: |\t)*)(\[[^\]]*\])?(?:(?: |\t)*)(//.*)?`
	regex := `(?mi)^\s*"([a-z\d_:#\$\[\]!&\|.\-\+/ \^']{1,})"\s*"([^"\\]*(?:\\.[^"\\]*)*)"(?:(?: |\t)*)(\[[^\]]*\])?(?:(?: |\t)*)(//.*)?`

	pairPattern, err := regexp.Compile(regex)

	if err != nil {
		return s_token, fmt.Errorf("Err in regEx: %v", err)
	}

	kvPairs := pairPattern.FindAllSubmatch(buf, -1)

	for _, kv := range kvPairs {
		// fmt.Printf("key=%s\nvalue=%s\nstatement=%s\ncomment=%s\n\n",string(kv[1]), string(kv[2]), strings.TrimRight(string(kv[3]), "\r\n"), strings.TrimRight(string(kv[4]), "\r\n"))
		if !strings.HasPrefix(string(kv[1]), "[english]") || v.sourceTkn { // Add token if key doesn't start with [english] or we want to capture everything
			s_token = append(s_token, []string{string(kv[0]), string(kv[1]), string(kv[2]), strings.TrimRight(string(kv[3]), "\r\n"), strings.TrimRight(string(kv[4]), "\r\n")})
		}
	}

	return s_token, nil
}


// FuzzyParseInSlice()
//
// Parse all keys/values/cond statements/comments in a slice
// Use a less restrictive regex than ParseInSlice() in order to identify VDF structure breaks.
// To be used to primarily check the validity of a VDF file.  
// 		E.g. "a_key"	"a value" [$WIN32]	// A comment
//		slice[0]: the entire line: "a_key"	"a value" [$WIN32]	// A comment
//		slice[1]: a_key
//		slice[2]: a value
//		slice[3]: [$WIN32]
//		slice[4]: // A comment
//
func (v *VDFFile) FuzzyParseInSlice(buf []byte) (s_token [][]string, err error) {
	v.log(fmt.Sprintf("FuzzyParseInSlice()"))

	regex := `(?mi)(?:[/]{2,}.*)|(?:\s)*"([^"]{1,})"\s*"([^"\\]*(?:\\.[^"\\]*)*)"(?:(?: |\t)*)(\[[^\]]*\])?(?:(?: |\t)*)(//.*)?`

	pairPattern, err := regexp.Compile(regex)

	if err != nil {
		return s_token, fmt.Errorf("Err in regEx: %v", err)
	}

	kvPairs := pairPattern.FindAllSubmatch(buf, -1)

	for _, kv := range kvPairs {
		// fmt.Printf("key=%s\nvalue=%s\nstatement=%s\ncomment=%s\n\n",string(kv[1]), string(kv[2]), strings.TrimRight(string(kv[3]), "\r\n"), strings.TrimRight(string(kv[4]), "\r\n"))
		// Comment lines are captured as matches so need to filter them:
		// add token if key is not null and (doesn't start with [english] or we want to capture everything)
		if len(kv[1]) > 0 && (!strings.HasPrefix(string(kv[1]), "[english]") || v.sourceTkn) { 
			s_token = append(s_token, []string{string(kv[0]), string(kv[1]), string(kv[2]), strings.TrimRight(string(kv[3]), "\r\n"), strings.TrimRight(string(kv[4]), "\r\n")})
		}
	}

	return s_token, nil
}


// GetEncoding()
//
// Returns encoding of current file
//
func (v *VDFFile) GetEncoding() string {
	v.log("GetEncoding()")
	return v.encoding
}
