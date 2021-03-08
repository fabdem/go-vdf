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

// var suffixesPluralGender []string
var pluralTag 	string
var genderTags 	[]string
var json *config.Config

const defaultJson = "pluralgender.json"  // located along with the exe or bin


func init() {

	// suffixesPluralGender = []string {
	// 									":p",  // plural
	// 									":n",  // gender sender
	// 									":g",  // gender receiver
	// 									":np", // gender sender plural
	// 									":gp", // gender receiver plural
	// 								}

	// Defines each token suffixe and its associated check function
	m_pluralGender = map[string]interface{} {
			":p":  checkPlural,								// plural
			":n":  checkGenderSender,					// gender sender
			":g":  checkGenderReceiver,				// gender receiver
			":np": checkGenderSenderPlural,		// gender sender with plural
			":gp": checkGenderReceiverPlural,	// gender receiver with plural
		}

	genderTags = []string {
							"#|f|#",
							"#|n|#",
							"#|c|#",
							"#|m|#",
							"#|ma|#",
							"#|mi|#",
							"#|mp|#",
						}

	pluralTag = "#|#"

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
		res = fmt.Sprintf("Expected number of plural forms: %d - found: %d", n + 1, ct + 1)
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

	var total int

	for _, gender := range genderTags {

		ct := strings.Count(v, gender)

		if ok := strings.Contains(list,gender);(ct > 1) || (ct == 1 && !ok) { // bad syntax cases
			if len(list) > 0 {
				res = fmt.Sprintf("Error with gender form: %s - expected only one of: %s", gender, list)
			} else {
				res = fmt.Sprintf("Error with gender form: %s - no gender expected", gender)
			}
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

	var total int

	for _, gender := range genderTags {

		ct := strings.Count(v, gender)

		if ok := strings.Contains(list,gender);(ct != 1 || !ok)&&(ct != 0 || ok) { // bad syntax cases
			if len(list) > 0 {
				res = fmt.Sprintf("Error with gender form: %s - expected one of each: %s", gender, list)
			} else {
				res = fmt.Sprintf("Error with gender form: %s - no gender expected", gender)
			}
			break
		} else {
			if ok && ct == 1 {
				total++
			}
		}
	}

	if total != len(l) {  // If we don't have one of each -> syntax problem
		res = fmt.Sprintf("Error with gender form - expected %s", list)
	}

	return res, err
}

// checkGenderSenderPlural()
//
// Check gender syntax in a sender token value with plural. Needs as many gender
// tags valid for the language as they are plurals.
// 	Input:
//		- token name
//		- token value
//		- Language name
// 	Output:
//		- issue == nil if no syntax issue
//		- err
//
//	E.g. "Valve_TestPluralGenders_Noun1:np"    "#|m|#Trésor#|m|#Trésors"
//
func checkGenderSenderPlural(k string, v string, lang string) (res string, err error) {
	l, err := json.GetGenders(lang) // Get the list of gender tags
	if err != nil {
		return res, err
	}

	nbPluralExpected, err := json.GetPlural(lang) // Get the number of plurals
	if err != nil  {
		return res, err
	}

	var list string  // Convert slice into a single string
	for _, val := range l {
		list += val + ","
	}

	pluralCount := 0

	for _, gender := range genderTags {
		if ct := strings.Count(v, gender); ct > 0 && !strings.Contains(list,gender) {
			res = fmt.Sprintf("Error with gender/plural form: this tag was unexpected %s", gender)
			break
		} else {
			pluralCount += ct
		}
	}

	if pluralCount != nbPluralExpected  {  // If incorrect number of plural forms ->  error
		res = fmt.Sprintf("Error with gender/plural forms - expected %d, counted %d", nbPluralExpected, pluralCount)
	}

	return res, err
}


// checkGenderReceiverplural()
//
// Check gender syntax in a receiver token value with plural.
// Each gender list must be repeated as many time as there are plurals for the language.
// 	Input:
//		- token name
//		- token value
//		- Language name
// 	Output:
//		- issue == nil if no syntax issue
//		- err
//
// E.g. "Valve_TestPluralGenders_Adjective1:gp" "#|m|#peu Commun#|f|#peu Commune#|m|#peu Communs#|f|#peu Communes"
//
func checkGenderReceiverPlural(k string, v string, lang string) (res string, err error) {
	// WIP
	return res, err
}


// FilterPlrGdr()
//
// Keeps only plural and gender tokens.
// 	Input:
//		- Slice of token names
// 	Output:
//		- Slice of plural and gender token names
//
func (v *VDFFile) FilterPlrGdr(in []string) (out []string) {
	v.log(fmt.Sprintf("FilterPlrGdr()"))
	for _,tkn := range in {
		for sufx,_ := range m_pluralGender {
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
