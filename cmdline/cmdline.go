package cmdline

import (
	"bytes"
	"fmt"
	"strings"
)

type CMDLine struct {
	labels    map[string]*clStep
	firstStep *clStep
	step      *clStep
}

type clStep struct {
	typeTips   string
	retypeTips string
	verify     func(line string) interface{}
	next       *clStep
}

func (cl *CMDLine) AddStep(tips string, verify func(line string) interface{}) *CMDLine {
	var label string
	var typeTips string
	var retypeTips string

	sp := strings.SplitN(tips, "::", 2)
	if len(sp) == 2 {
		label = strings.TrimSpace(sp[0])
		tips = sp[1]
	}

	sp = strings.SplitN(tips, "||", 2)
	typeTips = strings.TrimSpace(sp[0])
	if len(sp) == 2 {
		retypeTips = strings.TrimSpace(sp[1])
	}

	if len(typeTips) == 0 || verify == nil {
		return cl
	}

	step := &clStep{typeTips: typeTips, retypeTips: retypeTips, verify: verify}
	if len(label) > 0 {
		if cl.labels == nil {
			cl.labels = map[string]*clStep{}
		}
		if _, ok := cl.labels[label]; !ok {
			cl.labels[label] = step
		}
	}

	if cl.firstStep == nil {
		cl.firstStep = step
	}
	if cl.step != nil {
		cl.step.next = step
	}
	cl.step = step

	return cl
}

func (cl *CMDLine) GotoStep(s int) bool {
	if s <= 0 || cl.firstStep == nil {
		return false
	}
	step := cl.firstStep
	for i := 0; i < s; i++ {
		if step == nil {
			return false
		}
		step = step.next
	}
	cl.step = step
	return true
}

func (cl *CMDLine) Scan(callback func()) {
	var c byte
	buf := bytes.NewBuffer(nil)

	if cl.firstStep == nil {
		return
	}

	cl.step = cl.firstStep
	fmt.Print(cl.step.typeTips, " ")

SCAN:
	for {
		if _, err := fmt.Scanf("%c", &c); err != nil {
			break
		}
		if c == '\n' {
			vr := cl.step.verify(buf.String())
			switch r := vr.(type) {
			case bool:
				if r {
					if cl.step.next == nil {
						callback()
						break SCAN
					}
					cl.step = cl.step.next
					fmt.Print(cl.step.typeTips, " ")
				} else if len(cl.step.retypeTips) > 0 {
					fmt.Print(cl.step.retypeTips, " ")
				} else {
					fmt.Print(cl.step.typeTips, " ")
				}

			case int:
				if !cl.GotoStep(r) {
					break SCAN
				}
				fmt.Print(cl.step.typeTips, " ")

			case string:
				if len(cl.labels) > 0 {
					if step, ok := cl.labels[r]; ok {
						cl.step = step
						fmt.Print(cl.step.typeTips, " ")
					}
				}

			default:
				fmt.Print(cl.step.retypeTips, " ")
			}
			buf.Reset()
			continue
		}
		buf.WriteByte(c)
	}
}
