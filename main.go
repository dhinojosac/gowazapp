// Package main provides various examples of Fyne API capabilities
package main

import (
	"bufio"
	"fmt"
	"os"
	"time"

	"fyne.io/fyne/theme"

	"fyne.io/fyne"
	"fyne.io/fyne/app"
	"fyne.io/fyne/driver/desktop"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/widget"
)

var index = 0
var groupScroller *widget.Group
var scrollChat *widget.ScrollContainer
var hidden = 0

func addTextToChat() {
	fmt.Printf("Enter pressed!\n")
	v := widget.NewLabel(mEntry.Text)
	groupScroller.Append(v)
	mEntry.SetText("")
	index += 1
	scrollChat.ScrollToEnd()
}

type ChatEntry struct {
	widget.Entry
}

func (e *ChatEntry) TypedKey(key *fyne.KeyEvent) {
	// Call the function as defined in widget.Entry
	e.Entry.TypedKey(key)

	// Do something else on hitting the enter key.
	if key.Name == fyne.KeyReturn || key.Name == fyne.KeyEnter {
		fmt.Printf("Chant Entry enter pressed!\n")
		addTextToChat()
	}
}

var mEntry *ChatEntry

func main() {

	//os.Setenv("FYNE_SCALE", "0.9")

	//gui
	a := app.New()
	wzpTheme := theme.WzpTheme()
	a.Settings().SetTheme(wzpTheme)
	w := a.NewWindow("GoWAZAPP")
	fs := false

	go func() {
		fmt.Printf("Listening...\n")
		reader := bufio.NewReader(os.Stdin)
		for {
			text, _ := reader.ReadString('\n')
			time.Sleep(1 * time.Second)
			fmt.Printf(text)
			if text == "show\n" {
				w.Show()
				hidden = 0

			}
		}
	}()

	w.SetMainMenu(fyne.NewMainMenu(
		fyne.NewMenu("File",
			fyne.NewMenuItem("New", func() { fmt.Println("Menu New") })), // a quit item will be appended to our first menu
		fyne.NewMenu("Edit",
			fyne.NewMenuItem("Cut", func() { fmt.Println("Menu Cut") }),
			fyne.NewMenuItem("Copy", func() { fmt.Println("Menu Copy") }),
			fyne.NewMenuItem("Paste", func() { fmt.Println("Menu Paste") }),
		),
		fyne.NewMenu("Window",
			fyne.NewMenuItem("Hide", func() {
				w.Hide()
				hidden = 1
			}),
			fyne.NewMenuItem("FullScreen", func() {
				if fs == false {
					w.SetFullScreen(true)
					fs = true
				} else {
					w.SetFullScreen(false)
					w.CenterOnScreen()
					fs = false
				}

			}),
		)))

	groupScroller, scrollChat = widget.NewGroupWithScroller("WZP Console")

	mEntry = &ChatEntry{}
	mEntry.ExtendBaseWidget(mEntry)

	w.Canvas().(desktop.Canvas).SetOnKeyDown(func(ev *fyne.KeyEvent) {

		fmt.Printf("Key pressed: %v\n", ev.Name)

		if hidden == 0 {
			if ev.Name == "LeftControl" {
				fmt.Printf("Press Space to hide!\n")
				hidden = 2
			}
		}

		if hidden == 2 {
			if ev.Name == "Space" {
				w.Hide()
				hidden = 1
			}
		}

	})

	button := widget.NewButton("SEND", func() {
		fmt.Printf("Button pressed!\n")
		addTextToChat()
	})

	c1 := fyne.NewContainerWithLayout(layout.NewBorderLayout(nil, nil, nil, button), mEntry, button)
	content := fyne.NewContainerWithLayout(layout.NewBorderLayout(nil, c1, nil, nil), groupScroller, c1)

	w.SetContent(content)

	w.Resize(fyne.NewSize(520, 300))
	w.SetFixedSize(true)

	w.ShowAndRun()
}
