package wzpback

import (
	"encoding/gob"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sort"
	"syscall"
	"time"

	"github.com/Rhymen/go-whatsapp"
	"github.com/Rhymen/go-whatsapp/binary/proto"
	"github.com/dhinojosac/gowazapp/wzpui"
	"github.com/dhinojosac/gowazapp/wzputils"

	qrcodeTerminal "github.com/Baozisoftware/qrcode-terminal-go"
)

var number string
var historyMessages []whatsapp.TextMessage //TODO:implement and order history messages FIFO
var lastMsg string
var wac *whatsapp.Conn
var ctime time.Time //time of login

// ByTime implements sort.Interface based on the timestamp field.
type byTime []whatsapp.TextMessage

func (a byTime) Len() int           { return len(a) }
func (a byTime) Less(i, j int) bool { return a[i].Info.Timestamp < a[j].Info.Timestamp }
func (a byTime) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

func SetNumberWZP(s string) error {
	//TODO: check error number
	number = s
	return nil
}

func GetNumberWZP() string {
	return number
}

type waHandler struct {
	c *whatsapp.Conn
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
		//lastMsg = message.Text
		diff := ctime.Sub(t)
		if diff < 0 { //new messages
			fmt.Printf("[new]")
			//sound
			wzputils.SoundMsgTone()
			//notify
			if message.Info.FromMe == true {
				wzpui.AddWzpTextToChat(t.Format("01/02/2006 15:04:05") + ">> " + message.Text)
			} else {
				wzpui.AddWzpTextToChat(t.Format("01/02/2006 15:04:05") + "<< " + message.Text)
			}
		} else {
			// Append text to chat
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

func printHistory() {
	sort.Sort(byTime(historyMessages))
	fmt.Printf("History\n----------------------\n")
	for i := range historyMessages {
		message := historyMessages[i]
		t := time.Unix(int64(message.Info.Timestamp), 0)
		if message.Info.FromMe == true {
			wzpui.AddWzpTextToChat(t.Format("01/02/2006 15:04:05") + ">> " + message.Text)
		} else {
			wzpui.AddWzpTextToChat(t.Format("01/02/2006 15:04:05") + "<< " + message.Text)
		}
	}
	fmt.Printf("----------------------\n")
}

func StartWZP() {
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
			wzpui.ChangeState("Logged in")
		}

		//verifies phone connectivity
		pong, err := wac.AdminTest()

		if !pong || err != nil {
			log.Fatalf("error pinging in: %v\n", err)
		} else {
			fmt.Printf("[!] WZP pinged in!\n")
			wzpui.ChangeState("Connected")
		}

		// wait while chat jids are acquired through incoming initial messages
		fmt.Println("[!] Waiting for chats info...")
		wzpui.ChangeState("Loading history...")
		<-time.After(6 * time.Second)

		printHistory()
		wzpui.ChangeState("Ready")

		ch := wzpui.GetChatChan()
		go func() {
			fmt.Printf("Ready to listen chat\n")
			for msg := range ch {
				SendMessage(wac, msg)
			}
		}()

		wzpui.EnableEntryChat()

		//wait signal to shut down application
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		<-c

		//Disconnect safe
		fmt.Println("Shutting down now.")
		wzpui.ChangeState("Offline")
		session, err := wac.Disconnect()
		if err != nil {
			log.Fatalf("error disconnecting: %v\n", err)
		}
		if err := writeSession(session); err != nil {
			log.Fatalf("error saving session: %v", err)
		}

	}()
}
