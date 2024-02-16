// spin.go // Implements the spin function which fetch_binary uses to report status.

package main

import (
	"fmt"
	"strings"
	"time"
)

var stopSpinner chan struct{}
var SpinCompleteFlag bool

// Spin starts the spinner
func Spin(text string) {
	stopSpinner = make(chan struct{})
	SpinCompleteFlag = false
	go func() {
		spinChars := `-/|\\`
		spinIndex := 0
		for {
			select {
			case <-stopSpinner:
				// When the spinner is stopped, print spaces to clean up the output
				fmt.Printf("\r%s\r", strings.Repeat(" ", len(text)+5+len(spinChars))) //   5 is the length of "Working...", len(spinChars) is the length of the spinner characters
				SpinCompleteFlag = true
				// Reinitialize the stopSpinner channel to nil
				stopSpinner = nil
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
	// Check if the channel is already closed
	if stopSpinner != nil {
		close(stopSpinner)
		SpinCompleteFlag = true
		// Wait for the spinner to finish clearing the line
		time.Sleep(100 * time.Millisecond)
	}
}
