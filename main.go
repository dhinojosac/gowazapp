// Package main provides various examples of Fyne API capabilities
package main

import (
	"bufio"
	"fmt"
	"os"
	"time"

	"github.com/dhinojosac/gowazapp/wzpback"
	"github.com/dhinojosac/gowazapp/wzpui"
	"github.com/dhinojosac/gowazapp/wzputils"
)

func main() {

	//read number from file
	wzputils.InitConfig()
	number := wzputils.GetNumberFromFile()
	wzpback.SetNumberWZP(number)
	fmt.Printf("[!]%v\n", wzpback.GetNumberWZP())
	time.Sleep(100 * time.Millisecond)
	wzputils.SoundStartTone()

	w := wzpui.CreateWindowApp()
	chatchan := make(chan string)
	wzpui.SetChatChan(chatchan)

	//Listen command line
	go func() {
		fmt.Printf("Listening...\n")
		reader := bufio.NewReader(os.Stdin)
		for {
			text, _ := reader.ReadString('\n')
			time.Sleep(1 * time.Second)
			fmt.Printf(text)
			if text == "show\n" {
				wzpui.ShowWindowApp(w)
			}
		}
	}()

	// wzpback.StartWZP()

	w.ShowAndRun()

}
