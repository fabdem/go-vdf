package vdfloc
// Publicly available high level functions

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
	"vdfloc/config"

)

typedef t_PluralGender struct {
	suffix: string
	check: interface{}
	}

var m_pluralGender []t_PluralGender

var suffixesPluralGender []string
var pluralTag 	string
var genderTags 	[]string
var json config

const defaultJson = "pluralgender.json"


func init() {

	suffixesPluralGender := []string {	":p", // plural
										":n", // gender sender
										":g", // gender receiver
									}  

	genderTags := []string {"#|f|#",
							"#|n|#",
							"#|c|#",
							"#|m|#",
							"#|ma|#",
							"#|mi|#",
							"#|mp|#",
						}
						

	pluralTag := "#|#"
	
	m_pluralGender = map[string]interface{} {
			":p": checkPlural,
			":n": checkGenderSender,
			":g": checkGenderReceiver,
		}

	// Try to load the default config file
	json, err := config.New(defaultJson)

}


// LoadJsonConf()
//
// Load a json config file.
// 	Input:
//		- path and name or nil if default
// 	Output:
//		- err != nil if error
//		- update global var json
//
func LoadJsonConf(f string) (err error) {

	json, err := config.New(f)
	
	return err
}	


// checkPlural()
//
// Check plural syntax in a token value.
// 	Input:
//		- token name
//		- token value
//		- Language name
// 	Output:
//		- issue == nil if no error
//
func checkPlural(k string, v string, lang string) (res string, err error) {
	n, err := json.GetPlural(lang)
	if err != nil  {
		return res, err
	}
	if ct := strings.Count(v, suffixesPluralGender[0]); ct != n {
		res = fmt.Sprintf("Expected number of plural forms: %d - found: %d - token: %s - value: %s", n, ct, k, v)
	}
	return res, err
}

// checkGenderSender()
//
// Check gender syntax in a sender token value.
// 	Input:
//		- token name
//		- token value
//		- Language name
// 	Output:
//		- issue == nil if no error
//
func checkGenderSender(k string, v string, lang string) (res string, err error) {
	list, err := json.GetGender(lang)
	if err != nil {
		return res, err
	}

	for gender := range genderTags {
		
		ct := strings.Count(v, gender)
		
		if ok:= strings.Contains(list,gender);(ct != 1 || !ok)&&(ct != 0 || ok) { // bad syntax cases
			res = fmt.Sprintf("Error with gender form %s - expected %s - token: %s - value: %s", gender, list, k, v)
			break
		}
	}	
	return res, err
}


// checkGenderReceiver()
//
// Check gender syntax in a receiver token value.
// 	Input:
//		- token name
//		- token value
//		- Language name
// 	Output:
//		- issue == nil if no error
//
func checkGenderReceiver(k string, v string, lang string) (res string) {


}



// FilterPlrGdr()
//
// Filter out only plural and gender tokens.
// 	Input:
//		- Slice of token names
// 	Output:
//		- Slice of plural and gender token names
//
func (v *VDFFile) FilterPlrGdr(in []string) (out []string) {
	v.log(fmt.Sprintf("FilterPlrGdr()"))
	for _,tkn := range in {
		for _,sufx := range suffixesPluralGender {
			if strings.HasSuffix(tkn, sufx) {
				out := append(out,tkn)
			}
		}
	}
	return out
)


// CheckPlrlGendrTokenVal()
//
// Check plural and gender syntax of a token value.
// If it's not a plural or gender token just ignore (return empty string).
// 	Input:
//		- token name
//		- token value
//		- Language name
// 	Output:
//		- issue == nil if no error
//
func (v *VDFFile) CheckPlrlGendrTokenVal(tkn string, val string, language string) (issue string) {
	v.log(fmt.Sprintf("CheckPlrlGendrTokenVal(%s, %s, %s)", tkn, val, language))
	
	if idx := strings.LastIndex(tkn,":"); idx > 0 {
		if f,ok := m_pluralGender[tkn[idx:]]; ok {
			issue = f(tkn, val, language)  // Check syntax
		}
	}
	
	return issue
}


