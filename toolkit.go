package vdfloc
// Publicly available high level functions

import (
	"strings"
)

// GetTokenNames()
//
// Return a slice with all the token names.
// Eexcludes the ones prefixed with [english]. They appear 
// in some of the loc files holding the english source but are of no use.
// 
func (v *VDFFile) GetTokenNames() (s []string, err error) {

	buf, err := v.ReadSource()
	if err != nil {
		return s, err
	}
	
	res := v.SkipHeader(buf)
	
	tokens, err := v.ParseInSlice(res)

	for _,tkn := range tokens {
		// Skip token names begining with [english].
		if strings.HasPrefix(tkn[0], "[english]") {	s = append(s, tkn[0]) }
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
	
	res := v.SkipHeader(buf)
	
	tokens, err := v.ParseInMap(res)
		
	return tokens, err	
}

