package wzputils

import "fmt"

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
