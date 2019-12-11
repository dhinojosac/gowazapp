package wzpback

import (
	"encoding/gob"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Rhymen/go-whatsapp"
	"github.com/Rhymen/go-whatsapp/binary/proto"
	"github.com/gen2brain/beeep"

	qrcodeTerminal "github.com/Baozisoftware/qrcode-terminal-go"
)

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
