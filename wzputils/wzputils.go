package wzputils

import (
	"fmt"

	"github.com/spf13/viper"
)

const FILE_NAME = "config"

func InitConfig() {
	// viper.SetConfigName("config.yaml") // name of config file (without extension)
	// viper.SetConfigType("yaml")       // REQUIRED if the config file does not have the extension in the name
	// viper.AddConfigPath("/etc/appname/")  // path to look for the config file in
	// viper.AddConfigPath("$HOME/.appname") // call multiple times to add many search paths
	viper.AddConfigPath(".")    // optionally look for config in the working directory
	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {             // Handle errors reading the config file
		panic(fmt.Errorf("Fatal error config file: %w \n", err))
	}
}

func GetNumberFromFile() string {
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error if desired
			fmt.Println("Config file not found")
		} else {
			// Config file was found but another error was produced
			panic(fmt.Errorf("Fatal error config file: %w \n", err))
		}
	}

	phone := viper.Get("Phone") // this would be "steve"
	fmt.Printf("[!] %v\n", phone)
	if phone.(string) == "" {
		panic("Phone not found")
	}

	return phone.(string)

	// b, err := ioutil.ReadFile(FILE_NUMBER) // just pass the file name
	// if err != nil {
	// 	log.Panicf("Error: %v\n[!] Create a file called contact.numb with your favorite number. e.g: 56999050091\n", err)
	// }

	// number := strings.TrimSpace(string(b))
	// return number
}
