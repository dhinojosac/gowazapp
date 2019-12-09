// Package main provides various examples of Fyne API capabilities
package main

import (
	"bufio"
	"encoding/gob"
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
	"github.com/Rhymen/go-whatsapp/binary/proto"
	"github.com/gen2brain/beeep"

	qrcodeTerminal "github.com/Baozisoftware/qrcode-terminal-go"
)

var index = 0
var groupScroller *widget.Group
var scrollChat *widget.ScrollContainer
var hidden = 0
var ctime time.Time

var number string
var historyMessages []whatsapp.TextMessage //TODO:implement and order history messages
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

type ChatEntry struct {
	widget.Entry
}

func (e *ChatEntry) TypedKey(key *fyne.KeyEvent) {
	// Call the function as defined in widget.Entry
	e.Entry.TypedKey(key)

	// Do something else on hitting the enter key.
	if key.Name == fyne.KeyReturn || key.Name == fyne.KeyEnter {
		fmt.Printf("Chat Entry enter pressed!\n")
		addTextToChat()
	}
}

var mEntry *ChatEntry

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

func login(wac *whatsapp.Conn) error {
	//load saved session
	session, err := readSession()
	if err == nil {
		//restore session
		session, err = wac.RestoreWithSession(session)
		if err != nil {
			return fmt.Errorf("restoring failed: %v\n", err)
		}
	} else {
		//no saved session -> regular login
		qr := make(chan string)
		go func() {
			terminal := qrcodeTerminal.New()
			terminal.Get(<-qr).Print()
		}()
		session, err = wac.Login(qr)
		if err != nil {
			return fmt.Errorf("error during login: %v\n", err)
		}
	}

	//save session
	err = writeSession(session)
	if err != nil {
		return fmt.Errorf("error saving session: %v\n", err)
	}
	return nil
}

func readSession() (whatsapp.Session, error) {
	session := whatsapp.Session{}
	fmt.Printf("[!] Session saved in: %v as a %v\n", os.TempDir(), "whatsappSession.gob")
	file, err := os.Open(os.TempDir() + "/whatsappSession.gob")
	if err != nil {
		return session, err
	}
	defer file.Close()
	decoder := gob.NewDecoder(file)
	err = decoder.Decode(&session)
	if err != nil {
		return session, err
	}
	return session, nil
}

func writeSession(session whatsapp.Session) error {
	file, err := os.Create(os.TempDir() + "/whatsappSession.gob")
	if err != nil {
		return err
	}
	defer file.Close()
	encoder := gob.NewEncoder(file)
	err = encoder.Encode(session)
	if err != nil {
		return err
	}
	return nil
}

//HandleError needs to be implemented to be a valid WhatsApp handler
func (h *waHandler) HandleError(err error) {

	if e, ok := err.(*whatsapp.ErrConnectionFailed); ok {
		log.Printf("Connection failed, underlying error: %v", e.Err)
		log.Println("Waiting 30sec...")
		<-time.After(30 * time.Second)
		log.Println("Reconnecting...")
		err := h.c.Restore()
		if err != nil {
			log.Fatalf("Restore failed: %v", err)
		}
	} else {
		log.Printf("error occoured: %v\n", err)
	}
}

//Optional to be implemented. Implement HandleXXXMessage for the types you need.
func (*waHandler) HandleTextMessage(message whatsapp.TextMessage) {

	if message.Info.RemoteJid == number+"@s.whatsapp.net" {
		t := time.Unix(int64(message.Info.Timestamp), 0)
		lastMsg = message.Text
		diff := ctime.Sub(t)
		if diff < 0 { //new messages
			fmt.Printf("[new]")
			var err error
			if soundAlert {
				err = beeep.Beep(beeep.DefaultFreq, beeep.DefaultDuration)
				if err != nil {
					panic(err)
				}
			}
			if notifAlert {
				err = beeep.Alert("WZP", message.Text, "")
				if err != nil {
					panic(err)
				}
			}
			if message.Info.FromMe == true {
				addWzpTextToChat(">> " + message.Text)
			} else {
				addWzpTextToChat("<< " + message.Text)
			}
		} else {
			historyMessages = append(historyMessages, message) //Add to history message array
		}
	}
}

func SendMessage(w *whatsapp.Conn, m string) {

	previousMessage := "xD"
	quotedMessage := proto.Message{
		Conversation: &previousMessage,
	}

	ContextInfo := whatsapp.ContextInfo{
		QuotedMessage:   &quotedMessage,
		QuotedMessageID: "",
		Participant:     "", //Whot sent the original message
	}

	msg := whatsapp.TextMessage{
		Info: whatsapp.MessageInfo{
			RemoteJid: number + "@s.whatsapp.net",
		},
		ContextInfo: ContextInfo,
		Text:        m,
	}

	_, err := w.Send(msg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error sending message: %v", err)
		os.Exit(1)
	} else {

		fmt.Println(m)
	}
}
