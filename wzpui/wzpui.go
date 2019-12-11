package wzpui

import (
	"fmt"

	"fyne.io/fyne"
	"fyne.io/fyne/widget"
)

type ChatEntry struct {
	widget.Entry
}

func (e *ChatEntry) TypedKey(key *fyne.KeyEvent, f func()) {
	// Call the function as defined in widget.Entry
	e.Entry.TypedKey(key)

	// Do something else on hitting the enter key.
	if key.Name == fyne.KeyReturn || key.Name == fyne.KeyEnter {
		fmt.Printf("Chat Entry enter pressed!\n")
		f() //execute funcion callback
	}
}
