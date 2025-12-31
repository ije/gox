package term

import (
	"fmt"
	"strings"
	"time"
)

// Spinner is a spinner for terminal apps.
type Spinner struct {
	chars  []string
	index  int
	ticker *time.Ticker
	fps    int
}

// NewSpinner creates a new spinner.
func NewSpinner(chars string, fps int) *Spinner {
	return &Spinner{
		chars: strings.Split(chars, ""),
		fps:   fps,
	}
}

// Start starts the spinner.
func (t *Spinner) Start() {
	SaveCursor()
	t.ticker = time.NewTicker(time.Second / time.Duration(t.fps))
	go func() {
		t.print()
		for {
			<-t.ticker.C
			t.print()
		}
	}()
}

// Stop stops the spinner.
func (t *Spinner) Stop() {
	RestoreCursor()
	if t.ticker != nil {
		t.ticker.Stop()
		t.ticker = nil
	}
}

func (t *Spinner) print() {
	RestoreCursor()
	fmt.Print(Dim(t.chars[t.index]))
	t.index = (t.index + 1) % len(t.chars)
}
