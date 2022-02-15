package wzpui

import (
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"github.com/dhinojosac/gowazapp/wzputils"
)

var index = 0

var groupScroller *container.Scroll
var contentScroller *fyne.Container

// var scrollChat *fyne.Widget.Scroll
var soundMenu *fyne.MenuItem
var hidden = 0
var fullScreenState = false

//Status bar
var statusBar *fyne.Container
var statusLabel *widget.Label
var soundLabel *widget.Label
var notifyLabel *widget.Label

//Entry Chat
var mEntry *ChatEntry

var chatchan chan string

func SetChatChan(ch chan string) {
	chatchan = ch
}

func GetChatChan() chan string {
	return chatchan
}

type ChatEntry struct {
	widget.Entry
}

func DisableEntryChat() {
	mEntry.Disable()
}

func EnableEntryChat() {
	mEntry.Enable()
}

func AddTextToChat() {
	fmt.Printf("Enter pressed!\n")
	s := mEntry.Text
	t := time.Now()
	v := widget.NewLabel(t.Format("01/02/2006 15:04:05") + ">> " + mEntry.Text)
	//v.SetColor(color.RGBA{0x33, 0x99, 0xff, 0xff}) //deprecated

	mEntry.SetText("")
	contentScroller.Add(v)
	chatchan <- s
	index += 1
	groupScroller.ScrollToBottom()

}

func AddWzpTextToChat(s string, fromMe bool) {
	v := widget.NewLabel(s)
	if fromMe {
		//v.SetColor(color.RGBA{0x77, 0x99, 0x77, 0x80}) //deprecated
	} else {
		//v.SetColor(color.RGBA{0x77, 0x77, 0x99, 0x80}) //deprecated
	}
	// groupScroller.Append(v)
	contentScroller.Add(v)
	mEntry.SetText("")
	index += 1
	//scrollChat.ScrollToEnd()  //deprecated
}

func (e *ChatEntry) TypedKey(key *fyne.KeyEvent) {
	// Call the function as defined in widget.Entry
	e.Entry.TypedKey(key)

	// Do something else on hitting the enter key.
	if key.Name == fyne.KeyReturn || key.Name == fyne.KeyEnter {
		fmt.Printf("Chat Entry enter pressed!\n")
		AddTextToChat() //execute funcion callback
	}
}

func CreateWindowApp() fyne.Window {
	a := app.New()

	w := a.NewWindow("GoWAZAPP")
	// SetMenuBar(w) // todo: uncomment this line to set menu bar
	// groupScroller = widget.NewGroupWithScroller("WZP Console")
	contentScroller = container.NewVBox()

	groupScroller = container.NewVScroll(contentScroller)

	mEntry = &ChatEntry{}
	mEntry.ExtendBaseWidget(mEntry)
	DisableEntryChat()

	// w.Canvas().(desktop.Canvas).SetOnKeyDown(func(ev *fyne.KeyEvent) {

	// 	fmt.Printf("Key pressed: %v\n", ev.Name)

	// 	if hidden == 0 {
	// 		if ev.Name == "LeftControl" {
	// 			fmt.Printf("Press Space to hide!\n")
	// 			hidden = 2
	// 		}
	// 	}

	// 	if hidden == 2 {
	// 		if ev.Name == "Space" {
	// 			w.Hide()
	// 			hidden = 1
	// 		}
	// 	}

	// })

	button := widget.NewButton("SEND", func() {
		fmt.Printf("Button pressed!\n")
		//addTextToChat()
	})

	statusLabel = widget.NewLabel("Status: Connecting...      ")
	soundLabel = widget.NewLabel("Sound: OFF")
	notifyLabel = widget.NewLabel("Alert: OFF")

	alertLabels := container.NewHBox(soundLabel, notifyLabel)
	//statusBar = widget.NewHBox(statusLabel, alertLabels)
	statusBar = fyne.NewContainerWithLayout(layout.NewBorderLayout(nil, nil, nil, alertLabels), statusLabel, alertLabels)
	c1 := fyne.NewContainerWithLayout(layout.NewBorderLayout(nil, nil, nil, button), mEntry, button)
	content := fyne.NewContainerWithLayout(layout.NewBorderLayout(nil, c1, nil, nil), groupScroller, c1)
	content2 := fyne.NewContainerWithLayout(layout.NewBorderLayout(nil, statusBar, nil, nil), content, statusBar)

	w.SetContent(content2)

	w.Resize(fyne.NewSize(520, 300))
	w.SetFixedSize(true)

	return w
}

func HiddenWindowApp(w fyne.Window) {
	w.Hide()
	hidden = 1
}

func ShowWindowApp(w fyne.Window) {
	w.Show()
	hidden = 0
}

func SetMenuBar(w fyne.Window) {
	soundMenu = fyne.NewMenuItem("Sound Enable", func() {
		if !wzputils.GetSoundState() {
			soundLabel.SetText("Sound: ON")
			wzputils.SoundStartTone()
			soundMenu.Label = "Sound Disable"
		} else {
			soundLabel.SetText("Sound: OFF")
			soundMenu.Label = "Sound Enable"
		}
		wzputils.ToggleAlert(0)

	})
	w.SetMainMenu(fyne.NewMainMenu(
		fyne.NewMenu("File",
			fyne.NewMenuItem("New", func() { fmt.Println("Menu New") })), // a quit item will be appended to our first menu
		fyne.NewMenu("Alert", soundMenu),
		fyne.NewMenu("Window",
			fyne.NewMenuItem("Hide", func() {
				w.Hide()
				hidden = 1
			}),
			fyne.NewMenuItem("FullScreen", func() {
				if fullScreenState == false {
					w.SetFullScreen(true)
					fullScreenState = true
				} else {
					w.SetFullScreen(false)
					w.CenterOnScreen()
					fullScreenState = false
				}
			}),
		)))
}

func ChangeState(s string) {
	statusLabel.SetText("Status: " + s)
}

/**************************************/
