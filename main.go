// Package main provides various examples of Fyne API capabilities
package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"sort"
	"strings"
	"syscall"
	"time"

	"fyne.io/fyne/theme"

	"fyne.io/fyne"
	"fyne.io/fyne/app"
	"fyne.io/fyne/driver/desktop"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/widget"

	"github.com/Rhymen/go-whatsapp"

	"github.com/dhinojosac/gowazapp/wzpui"
)

var index = 0
var groupScroller *widget.Group
var scrollChat *widget.ScrollContainer
var hidden = 0
var ctime time.Time

var number string
var historyMessages []whatsapp.TextMessage //TODO:implement and order history messages FIFO

//Alerts
var soundAlert bool
var notifAlert bool
var lastMsg string
var wac *whatsapp.Conn

type waHandler struct {
	c *whatsapp.Conn
}

// ByTime implements sort.Interface based on the timestamp field.
type byTime []whatsapp.TextMessage

func (a byTime) Len() int           { return len(a) }
func (a byTime) Less(i, j int) bool { return a[i].Info.Timestamp < a[j].Info.Timestamp }
func (a byTime) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

func printHistory() {
	sort.Sort(byTime(historyMessages))
	fmt.Printf("History\n----------------------\n")
	for i := range historyMessages {
		message := historyMessages[i]
		t := time.Unix(int64(message.Info.Timestamp), 0)
		if message.Info.FromMe == true {
			addWzpTextToChat(t.Format("01/02/2006 15:04:05") + ">> " + message.Text)
		} else {
			addWzpTextToChat(t.Format("01/02/2006 15:04:05") + "<< " + message.Text)
		}
	}
	fmt.Printf("----------------------\n")
}

// Toogle function
func toggleAlert(t int) {
	if t == 0 {
		if soundAlert {
			soundAlert = false
			fmt.Printf("Sound Alert: Disabled\n")
		} else {
			soundAlert = true
			fmt.Printf("Sound Alert: Enabled\n")
		}
	} else if t == 1 {
		if notifAlert {
			notifAlert = false
			fmt.Printf("Nofification: Disabled\n")
		} else {
			notifAlert = true
			fmt.Printf("Nofification: Enabled\n")
		}
	}
}

func addTextToChat() {
	fmt.Printf("Enter pressed!\n")
	s := mEntry.Text
	v := widget.NewLabel(mEntry.Text)
	mEntry.SetText("")
	groupScroller.Append(v)
	go SendMessage(wac, s) //send message to wzp
	index += 1
	scrollChat.ScrollToEnd()

}

func addWzpTextToChat(s string) {
	v := widget.NewLabel(s)
	groupScroller.Append(v)
	mEntry.SetText("")
	index += 1
	scrollChat.ScrollToEnd()
}

var mEntry *wzpui.ChatEntry

func main() {

	//os.Setenv("FYNE_SCALE", "0.9")

	//read number from file

	b, err := ioutil.ReadFile("contact.numb") // just pass the file name
	if err != nil {
		fmt.Print(err)
	}

	number = strings.TrimSpace(string(b))
	fmt.Printf("number %v\n", number)
	time.Sleep(500 * time.Millisecond)

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
			if text == "*s" {
				toggleAlert(0)
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

	go func() {
		fmt.Printf("[!] Start WZP client\n")
		ctime = time.Now()

		var err error
		//create new WhatsApp connection
		wac, err = whatsapp.NewConn(10 * time.Second)
		if err != nil {
			log.Fatalf("error creating connection: %v\n", err)
		} else {
			fmt.Printf("[!] WZP Connected!\n")
		}

		//Add handler
		wac.AddHandler(&waHandler{wac})

		//SendMessage(wac, "Client connected!\n") //testing send message

		//login or restore
		if err := login(wac); err != nil {
			log.Fatalf("error logging in: %v\n", err)
		} else {
			fmt.Printf("[!] WZP logged in!\n")
		}

		//verifies phone connectivity
		pong, err := wac.AdminTest()

		if !pong || err != nil {
			log.Fatalf("error pinging in: %v\n", err)
		} else {
			fmt.Printf("[!] WZP pinged in!\n")
		}

		// wait while chat jids are acquired through incoming initial messages
		fmt.Println("[!] Waiting for chats info...")
		<-time.After(6 * time.Second)
		printHistory()

		//wait signal to shut down application
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		<-c

		//Disconnect safe
		fmt.Println("Shutting down now.")
		session, err := wac.Disconnect()
		if err != nil {
			log.Fatalf("error disconnecting: %v\n", err)
		}
		if err := writeSession(session); err != nil {
			log.Fatalf("error saving session: %v", err)
		}

	}()

	w.ShowAndRun()

}
