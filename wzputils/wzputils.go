package wzputils

import (
	"fmt"
	"io/ioutil"
	"strings"
)

const FILE_NUMBER = "contact.numb"

func GetNumberFromFile() string {
	b, err := ioutil.ReadFile(FILE_NUMBER) // just pass the file name
	if err != nil {
		fmt.Print(err)
	}

	number := strings.TrimSpace(string(b))
	return number
}
