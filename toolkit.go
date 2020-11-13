package vdfloc

import (
	// "bufio"
	// "bytes"
	// "flag"
	// "errors"
	// "fmt"
	// "io"
	// "io/ioutil"
	// "log"
	// "os"
	"strings"
)

// Publicly available high level functions



// Return a slice with all the token names.
// Token names ending with:
func (v *VDFFile) GetTokenNames() (s []string, err error) {

	buf, err := v.ReadSource()
	if err != nil {
		return s, err
	}
	
	res := v.SkipHeader(buf)
	
	tokens, err = v.ParseInSlice(res)

	for _,tkn := range tokens {
		s = append(s, tkn[0])
	}
		
	return s, nil	
}
