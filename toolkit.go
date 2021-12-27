package vdfloc

// Publicly available high level functions

import (
	"errors"
	"fmt"
	"regexp"
	"path/filepath"
	"strings"
)

// GetTokenNames()
//
// Return a slice with all the token names.
// Excludes the ones prefixed with [english]. They appear
// in some of the loc files holding the english source but are of no use.
//
func (v *VDFFile) GetTokenNames() (s []string, err error) {
	v.log(fmt.Sprintf("GetTokenNames()"))

	buf, err := v.ReadSource()
	if err != nil {
		return s, err
	}

	res, err := v.SkipHeader(buf)
	if err != nil {
		return s, err
	}

	tokens, err := v.ParseInSlice(res)

	for _, tkn := range tokens {
		// Skip token names begining with [english].
		if !strings.HasPrefix(tkn[1], "[english]") {
			s = append(s, tkn[1])
		}
	}

	return s, err
}

// GetStringsWithConditionalStatement()
//
// Returns a slice with the details of all strings with conditional statements (e.g.[$WIN32]).
// Excludes the ones prefixed with [english].
// Returns a slice with full line minus leading lf/rc/spaces and tabs, key, value, conditional statement, comment.
//
func (v *VDFFile) GetStringsWithConditionalStatement() (s [][]string, err error) {
	v.log(fmt.Sprintf("GetStringsWithConditionalStatement()"))

	buf, err := v.ReadSource()
	if err != nil {
		return s, err
	}

	res, err := v.SkipHeader(buf)
	if err != nil {
		return s, err
	}

	tokens, err := v.ParseInSlice(res)
	if err != nil {
		return s, err
	}
	for _, tkn := range tokens {
		//fmt.Println("%s\n",strings.TrimLeft(tkn[0], "\t \r\n"))
		// Skip token names begining with [english] and the ones with no cond statements.
		if !strings.HasPrefix(tkn[1], "[english]") && len(tkn[3]) > 0 {
		
			s = append(s, []string{strings.TrimLeft(tkn[0], "\t \r\n"), tkn[1], tkn[2], tkn[3], tkn[4]})
		}
	}

	return s, err
}

// GetTokenInMap()
//
// Return a map of all token/content.
//
func (v *VDFFile) GetTokenInMap() (s map[string]string, err error) {
	v.log(fmt.Sprintf("GetTokenInMap()"))

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
func (v *VDFFile) GetEnFileName() (enFileName string, err error) {
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

	if len(locFileName) == 0 {
		return "", fmt.Errorf("Paramer shoudn't be empty.")
	}

	extension := filepath.Ext(locFileName)
	base := strings.TrimRight(filepath.Base(locFileName), extension)

	if lastUnderscore := strings.LastIndex(base, "_"); lastUnderscore == -1 {
		return "english" + extension, nil
	} else {
		return base[0:lastUnderscore] + "_english" + extension, nil
	}
}



// CheckKeyValidity()
//
// Tries to detect missing or wrongly escaped double quotes.
// Has better chances to work with non English files.
// Not bulletproof since key value pairs detection is based on valid characters. 
//  
// Parse all keys statements from a slice of tokens (use FuzzyParseInSlice())
// and returns an error if they are invalid (longer than autorized maxKeyLen or containing spaces/tabs or other non english characters)
// plus a list of the offending token keys if any.

func (v *VDFFile) CheckKeyValidity(tokens [][]string) (list []string, err error) {
	v.log(fmt.Sprintf("CheckKeyValidity()"))

	// Parse all keys
	err_flag := false

	var isKeyNameCharValid = regexp.MustCompile(`^[0-9a-zA-Z\[\]\$#_:&!\|.\-\+/ \^']+$`).MatchString 

	for _, tkn := range tokens {
		// fmt.Printf("|1>%s|2>%s|3>%s|4>%s\n",tkn[1],tkn[2],tkn[3],tkn[4] )
		if len(tkn[1]) > v.maxKeyLen || !isKeyNameCharValid(tkn[1]){
			list = append(list, tkn[1])
			err_flag = true
		}
		/* DEBUG if len(tkn[1]) > v.maxKeyLen {
			list = append(list, "1_" + fmt.Sprintf("%d",(i)) + " " + tkn[1])
			err_flag = true
		}
		if !isKeyNameCharValid(tkn[1]){
			list = append(list, "2_" + fmt.Sprintf("%d",(i)) + " " + tkn[1])
			err_flag = true
		} */
		
	}

	if err_flag {
		err = errors.New("Invalid key(s) found.")
	}
	return list, err
}


// CheckKeyUnicity()
//
// Parse all keys/conditional statements from a slice of tokens (use ParseInSlice() or FuzzyParseInSlice())
// and returns an error if they are non unique plus a list of non unique token keys if any.
// Would make sense to be ran after CheckKeyValidity()()
// If err is nil then all is good.
func (v *VDFFile) CheckKeyUnicity(tokens [][]string) (list []string, err error) {
	v.log(fmt.Sprintf("CheckKeyUnicity()"))

	// Move slice in a map and count occurrences
	// map key is string key + conditional statement
	s := make(map[string]int)

	for _, tkn := range tokens {
		s[tkn[1] + tkn[3]]++
	}

	// Now builds a list of keys for which unicity is broken if any
	err_flag := false
	for k, v := range s {
		if v > 1 {
			list = append(list, k)
			err_flag = true
		}
	}

	if err_flag {
		err = errors.New("Non unique key(s)")
	}
	return list, err
}



// CheckIsolatedConditionalStatements()
//
// Search in a byte buffer for isolated conditional statements which is an invalid VDF form.
// Would make sense to be ran after CheckKeyValidity()()
func (v *VDFFile) CheckIsolatedConditionalStatements(buf []byte) (list []string, err error) {
	v.log(fmt.Sprintf("CheckIsolatedConditionalStatements()"))

	// Look for occurrences of isolated conditional statements
	regex := `(?mi)^[ \t]*(\[[^\]]*\])`
	pattern, err := regexp.Compile(regex)
	vals := pattern.FindAllSubmatch(buf, -1)
	if err != nil {
		return list, fmt.Errorf("Err in regEx: %v", err)
	}
	err_flag := false
	for _, v := range vals {
		list = append(list, string(v[0]))
		err_flag = true
	}

	if err_flag {
		err = errors.New("Isolated conditional statement(s) found.")
	}

	return list, err
}


