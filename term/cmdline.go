package term

import (
	"bytes"
	"fmt"
	"strings"
)

type CMDLine struct {
	labelSetps map[string]*clStep
	firstStep  *clStep
	step       *clStep
	callback   func()
}

func NewCMDLine(callback func()) *CMDLine {
	return &CMDLine{
		labelSetps: map[string]*clStep{},
		callback:   callback,
	}
}

type clStep struct {
	tips   string
	verify func(input string) interface{}
	next   *clStep
}

func (cl *CMDLine) AddStepWithLabel(label, tips string, verify func(input string) interface{}) *CMDLine {
	if len(tips) == 0 || verify == nil {
		return cl
	}

	step := &clStep{tips: tips, verify: verify}
	if cl.firstStep == nil {
		cl.firstStep = step
	}
	if cl.step != nil {
		cl.step.next = step
	}
	cl.step = step

	if len(label) > 0 {
		if cl.labelSetps == nil {
			cl.labelSetps = map[string]*clStep{}
		}
		if _, ok := cl.labelSetps[label]; !ok {
			cl.labelSetps[label] = step
		}
	}

	return cl
}

func (cl *CMDLine) AddStep(tips string, verify func(input string) interface{}) *CMDLine {
	return cl.AddStepWithLabel("", tips, verify)
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

func (cl *CMDLine) Scan() {
	if cl.firstStep == nil {
		return
	}

	cl.step = cl.firstStep
	fmt.Print(cl.step.tips, " ")

	var c byte
	buf := bytes.NewBuffer(nil)
SCAN:
	for {
		if _, err := fmt.Scanf("%c", &c); err != nil {
			if err.Error() == "unexpected newline" {
				err = nil
				c = '\n'
			} else {
				break
			}
		}

		if c == '\n' {
			vr := cl.step.verify(strings.TrimSpace(buf.String()))
			switch r := vr.(type) {
			case bool:
				if r {
					if cl.step.next == nil {
						if cl.callback != nil {
							cl.callback()
						}
						break SCAN
					}
					cl.step = cl.step.next
					fmt.Print(cl.step.tips, " ")
				} else {
					fmt.Print(cl.step.tips, " ")
				}

			case int:
				if !cl.GotoStep(r) {
					break SCAN
				}

				fmt.Print(cl.step.tips, " ")

			case string:
				if len(r) == 0 {
					fmt.Print(cl.step.tips, " ")
					break
				}

				if len(cl.labelSetps) > 0 {
					if step, ok := cl.labelSetps[r]; ok {
						cl.step = step
						fmt.Print(cl.step.tips, " ")
						break
					}
				}

				fmt.Print(r, " ")

			default:
				fmt.Print(cl.step.tips, " ")
			}

			buf.Reset()
			continue
		}
		buf.WriteByte(c)
	}
}
