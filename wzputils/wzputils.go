package wzputils

import (
	"io/ioutil"
	"log"
	"strings"
)

const FILE_NUMBER = "contact.numb"

func GetNumberFromFile() string {
	b, err := ioutil.ReadFile(FILE_NUMBER) // just pass the file name
	if err != nil {
		log.Panicf("Error: %v\n[!] Create a file called contact.numb with your favorite number. e.g: 56999050091\n", err)
	}

	number := strings.TrimSpace(string(b))
	return number
}
