package wzputils

import "fmt"
import "github.com/gen2brain/beeep"

var soundAlert bool
var notifAlert bool

func GetSoundState() bool {
	return soundAlert
}

func GetNotifyState() bool {
	return notifAlert
}

func SoundStartTone() {
	err := beeep.Beep(700, 600)
	if err != nil {
		panic(err)
	}

}

func SoundMsgTone() {
	if soundAlert {
		err := beeep.Beep(beeep.DefaultFreq, beeep.DefaultDuration)
		if err != nil {
			panic(err)
		}
	}
}

// Toogle function
func ToggleAlert(t int) {
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
