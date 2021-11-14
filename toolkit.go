package vdfloc
// Publicly available high level functions

import (
	"fmt"
	"strings"
	"path/filepath"
)

// GetTokenNames()
//
// Return a slice with all the token names.
// Excludes the ones prefixed with [english]. They appear
// in some of the loc files holding the english source but are of no use.
//
func (v *VDFFile) GetTokenNames() (s []string, err error) {

	buf, err := v.ReadSource()
	if err != nil {
		return s, err
	}

	res, err := v.SkipHeader(buf)
	if err != nil {
		return s, err
	}

	tokens, err := v.ParseInSlice(res)

	for _,tkn := range tokens {
		// Skip token names begining with [english].
		if !strings.HasPrefix(tkn[0], "[english]") {	s = append(s, tkn[0]) }
	}

	return s, err
}


// GetStringsWithConditionalStatement()
//
// Return a slice with the details of all strings with conditional statements (e.g.[$WIN32]).
// Excludes the ones prefixed with [english].
// Returns a slice with key, value, conditional statement, comment.
//
func (v *VDFFile) GetStringsWithConditionalStatement() (s [][]string, err error) {

	buf, err := v.ReadSource()
	if err != nil {
		return s, err
	}

	res, err := v.SkipHeader(buf)
	if err != nil {
		return s, err
	}

	tokens, err := v.ParseInSlice(res)

	for _,tkn := range tokens {
		// Skip token names begining with [english] and the ones with no cond statements.
		if !strings.HasPrefix(tkn[0], "[english]") && len(tkn[2]) > 0 {
				s = append(s, []string{tkn[0], tkn[1], tkn[2], strings.TrimRight(tkn[3], "\r\n")})
		}
	}

	return s, err
}


// GetTokenInMap()
//
// Return a map of all token/content.
//
func (v *VDFFile) GetTokenInMap() (s map[string]string, err error) {

	buf, err := v.ReadSource()
	if err != nil {
		return s, err
	}

	res, err := v.SkipHeader(buf)
	if err != nil {
		return s, err
	}

	tokens, err := v.ParseInMap(res)

	return tokens, err
}


// GetEnFileName()
//
// Returns the name of the english file (source) corresponding to the current loc file name.
//  err != nil if loc file name is empty
//  A loc file name is formed like this xxxx_<language>.yyy or <language>.yyy
//
func (v *VDFFile)GetEnFileName() (enFileName string, err error) {
	v.log(fmt.Sprintf("GetEnFileName(%s)", v.fileName))
	enFileName, err = GetEnFileName(v.fileName)
	return enFileName, err
}


// GetEnFileName()
//
// Returns the name of the english file (source) corresponding to the loc file name passed as a parameter.
//  err != nil if loc file name is empty
//  A loc file name is formed like this xxxx_<language>.yyy or <language>.yyy
//
func GetEnFileName(locFileName string) (enFileName string, err error) {

	if len(locFileName) == 0 { return "", fmt.Errorf("Paramer shoudn't be empty.") }

	extension := filepath.Ext(locFileName)
	base := strings.TrimRight(filepath.Base(locFileName),extension)

	if lastUnderscore := strings.LastIndex(base, "_"); lastUnderscore == -1 {
		return "english" + extension, nil
	} else
	{
		return base[0:lastUnderscore] + "_english" + extension, nil
	}
}
