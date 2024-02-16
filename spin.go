// spin.go // Implement the spin function which fetch_binary uses to report status.

package main

import (
	"fmt"
	"strings"
	"time"
)

var stopSpinner chan struct{}
var SpinCompleteFlag bool

// errMsg is a custom error type for error messages
type errMsg error

// Spin starts the spinner
// Spin starts the spinner
func Spin(text string) {
	stopSpinner = make(chan struct{})
	SpinCompleteFlag = false
	go func() {
		spinChars := `-\|/`
		spinIndex := 0
		for {
			select {
			case <-stopSpinner:
				// When the spinner is stopped, print spaces to clean up the output
				fmt.Printf("\r%s\r", strings.Repeat(" ", len(text)+5)) //  5 is the length of "Working..."
				SpinCompleteFlag = true
				return
			default:
				fmt.Printf("\r%c Working... %s", spinChars[spinIndex], text)
				spinIndex = (spinIndex + 1) % len(spinChars)
				time.Sleep(100 * time.Millisecond)
			}
		}
	}()
}

// StopSpinner stops the spinner
func StopSpinner() {
	close(stopSpinner)
	SpinCompleteFlag = true
}
