package config
// Manage a json file defining source 2 plurals and genders by language

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

type Config struct {
	filename string
	attribs  langAttributes
}

	
type language struct {
		Name 	string `json:"name"`
		Plural  int    `json:"plural"`
		Genders []struct {
			Gender string `json:"gender"`
		} `json:"genders"`
	}

type langAttributes struct {
	Languages []language  `json:"languages"`
	}


var c 	Config					// config stored in memeory
var attribs map[string]language	// map to simplify access to language attributes


// New()
// Create a new instance.
// Open the json file and load its content in memory.
// 	Parameter:
//		- path and name of the file
//	Returns:
//		- err != null in case of error
//		- pointer to instance
func New(jsonfilename string) (*Config, error) {

	if !fileExists(jsonfilename) {
		return nil, errors.New(fmt.Sprintf("package config - Can't find file %s", jsonfilename))
	}

	c := &Config{}
	c.filename = jsonfilename

	// Try to load a json file
	jsonFile, err := os.Open(jsonfilename)
	if err != nil {
		fmt.Printf("Error - problem opening %s\n%s\n", jsonfilename, err)
		return nil, errors.New(fmt.Sprintf("package config - Can't open file %s", jsonfilename))
	}
	// defer the closing
	defer jsonFile.Close()

	// read the file in a byte slice.
	buffer, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("package config - Issue reading %s %v", jsonfilename, err))
	}

	// fmt.Printf("file read: %s", buffer)
	// fmt.Printf("struct: %v", c.attrib)

	// Unmarshal json buffer in struct
	err = json.Unmarshal(buffer, &c.attribs)
	// fmt.Printf("struct: %v", c.attrib)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("package config - Issue unmarshalling json %v", err))
	}

	// Move language attributes into a map to simplify access to data
	attribs = make(map[string]language)
	for _, v := range c.attribs.Languages {
		attribs[v.Name] = v
	}
	// fmt.Printf("c.attribs=%v\n", c.attribs)
	// fmt.Printf("attribs=%v\n", attribs)
	
	if len(attribs) <= 0 {
		return nil, errors.New(fmt.Sprintf("package config - at least one language needs to be defined"))
	}

	return c, nil
}

// Release instance
// Close the file and release the structure.
func Close(c *Config) (err error) {
	c = nil
	return nil
}


// GetPlural()
//
//	Get the language plural details
// 	Parameter:
//		- language 
//	Returns:
//		- err != null if fails to find language
//   	- number of plurals expected
//
func (c *Config) GetPlural(lang string) (plurals int, err error) {

	if _, ok := attribs[lang]; ok {
		// fmt.Printf("plural= %d   \n",attribs[lang].Plural  )
		return attribs[lang].Plural, nil
	} else { 
		return plurals, errors.New(fmt.Sprintf("package config - Can't find language for %s", lang))
	}
}


// GetGenders()
//
//	Get the language gender details
// 	Parameter:
//		- language 
//	Returns:
//		- err != null if fails to find language
//		- list of genders expected
//
func (c *Config) GetGenders(lang string) (genders []string, err error) {

	if _, ok := attribs[lang]; ok {
		for _, gender := range attribs[lang].Genders {
			// fmt.Printf("gender= %v ---- ",gender.Gender  )
			genders = append(genders, gender.Gender)
		}
		// fmt.Printf("plural= %d   \n",attribs[lang].Plural  )
		return genders, nil
	} else { 
		return genders, errors.New(fmt.Sprintf("package config - Can't find language for %s", lang))
	}
}



// fileExists checks if a file exists and is not a directory before we
// try using it to prevent further errors.
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
