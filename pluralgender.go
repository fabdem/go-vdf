package vdfloc
// Publicly available high level functions

import (
	// "bufio"
	// "bytes"
	// "flag"
	// "errors"
	"fmt"
	// "io"
	// "io/ioutil"
	// "log"
	// "os"
	"strings"
	"go-vdfloc/config"

)

// type t_PluralGender struct {
// 	suffix	string
// 	check	interface{}
// 	}

var m_pluralGender map[string]interface{}

var suffixesPluralGender []string
var pluralTag 	string
var genderTags 	[]string
var json *config.Config

const defaultJson = "pluralgender.json"  // located along with the exe or bin


func init() {

	suffixesPluralGender = []string {	":p", // plural
										":n", // gender sender
										":g", // gender receiver
									}  

	genderTags = []string {"#|f|#",
							"#|n|#",
							"#|c|#",
							"#|m|#",
							"#|ma|#",
							"#|mi|#",
							"#|mp|#",
						}
						

	pluralTag = "#|#"
	
	m_pluralGender = map[string]interface{} {
			":p": checkPlural,
			":n": checkGenderSender,
			":g": checkGenderReceiver,
		}

	// Try to load the default config file
	json, _ = config.New(defaultJson)

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

	json, err = config.New(f)
	
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
//		- issue == nil if no syntax issue
//		- err
//
func checkPlural(k string, v string, lang string) (res string, err error) {
	n, err := json.GetPlural(lang)
	if err != nil  {
		return res, err
	}
	
	if n > 0 { n-- }  // e.g. 2 form plural -> 1 separator
	
	if ct := strings.Count(v, pluralTag); ct != n {
		res = fmt.Sprintf("Expected number of plural forms: %d - found: %d", n+1, ct)
	}
	return res, err
}

// checkGenderSender()
//
// Check gender syntax in a sender token value. Needs either 1 of tag list for that language.
// 	Input:
//		- token name
//		- token value
//		- Language name
// 	Output:
//		- issue == nil if no syntax issue
//		- err
//
func checkGenderSender(k string, v string, lang string) (res string, err error) {
	l, err := json.GetGenders(lang)
	if err != nil {
		return res, err
	}

	var list string  // Convert slice to a single string
	for _, val := range l { 
		list += val + ","
	}

	fmt.Printf("checkGenderSender lang: %s tkn: %s val: %s list:%s len=%d\n",lang, k, v, list, len(l))
	
	var total int
	
	for _, gender := range genderTags {
		
		ct := strings.Count(v, gender)
		fmt.Printf("	checkGenderSender gender:%s ct: %d\n",gender, ct)
		
		if ok := strings.Contains(list,gender);(ct > 1) || (ct == 1 && !ok) { // bad syntax cases
			res = fmt.Sprintf("Error with gender form: %s - expected only one of: %s", gender, list)
			break
		} else {
			if ct >0 { // found one good match
				total++
			}
		}
	}

	if len(l) > 0 && total != 1 {  // If we have not found exactly 1 match when there are genders
		res = fmt.Sprintf("Error with gender form - expected %s", list)
	}
	
	return res, err
}


// checkGenderReceiver()
//
// Check gender syntax in a receiver token value. Needs 1 of each tag for that language.
// 	Input:
//		- token name
//		- token value
//		- Language name
// 	Output:
//		- issue == nil if no syntax issue
//		- err
//
func checkGenderReceiver(k string, v string, lang string) (res string, err error) {
	l, err := json.GetGenders(lang)
	if err != nil {
		return res, err
	}

	var list string  // Convert slice to a single string
	for _, val := range l { 
		list += (val + ",")
	}

	fmt.Printf("checkGenderReceiver list:%s len=%d\n",list, len(l))
	fmt.Printf("checkGenderReceiver lang:%s\n",lang)
	
	var total int
	
	for _, gender := range genderTags {
		
		ct := strings.Count(v, gender)
		
		if ok := strings.Contains(list,gender);(ct != 1 || !ok)&&(ct != 0 || ok) { // bad syntax cases
			res = fmt.Sprintf("Error with gender form: %s - expected one of each: %s", gender, list)
			break
		} else {
			if ok && ct == 1 {
				total++
			}
		}
	}

	fmt.Printf("checkGenderReceiver total:%d\n",total)
	
	if total != len(l) {  // If we don't have one of each -> syntax problem
		res = fmt.Sprintf("Error with gender form - expected %s", list)
	}
	
	return res, err


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
				out = append(out,tkn)
			}
		}
	}
	return out
}


// CheckPlrlGendrTokenVal()
//
// Check plural and gender syntax of a token value.
// If it's not a plural or gender token just ignore (return empty string).
// 	Input:
//		- token name
//		- token value
//		- Language name
// 	Output:
//		- issue == nil if no syntax issue
//		- err
//
func (v *VDFFile) CheckPlrlGendrTokenVal(tkn string, val string, language string) (issue string, err error) {
	v.log(fmt.Sprintf("CheckPlrlGendrTokenVal(%s, %s, %s)", tkn, val, language))
	
	if idx := strings.LastIndex(tkn,":"); idx > 0 {
		if f,ok := m_pluralGender[tkn[idx:]]; ok {
			issue, err = f.(func(string, string, string)(string, error))(tkn, val, language)  // Check syntax
			// bOK,bArrayRes := record.fctOpen.(func (string) (bool,[]byte))(openingTag)    
		}
	}	
	return issue, err
}


