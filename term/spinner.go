package term

import (
	"os"
	"time"
)

type SpinnerConfig struct {
	Chars   []string
	Message string
	FPS     int
}

// Spinner is a spinner for terminal apps.
type Spinner struct {
	SpinnerConfig
	index int
	timer *time.Timer
}

// NewSpinner creates a new spinner.
func NewSpinner(config SpinnerConfig) *Spinner {
	if len(config.Chars) == 0 {
		config.Chars = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	}
	if config.FPS <= 0 {
		config.FPS = 5
	}
	return &Spinner{SpinnerConfig: config}
}

// Start starts the spinner.
func (t *Spinner) Start() {
	SaveCursor()
	t.print()
}

// Stop stops the spinner.
func (t *Spinner) Stop() {
	if t.timer != nil {
		t.timer.Stop()
		t.timer = nil
	}
	RestoreCursor()
	ClearLineRight()
}

func (t *Spinner) print() {
	RestoreCursor()
	os.Stdout.WriteString(Dim(t.Chars[t.index]))
	if len(t.Message) > 0 {
		os.Stdout.WriteString(" " + Dim(t.Message))
	}
	t.index = (t.index + 1) % len(t.Chars)
	t.timer = time.AfterFunc(time.Second/time.Duration(t.FPS), t.print)
}
