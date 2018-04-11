package term

import (
	"bytes"
	"fmt"
	"strings"
)

type CMDLineApp struct {
	steps       map[string]*clStep
	firstStep   *clStep
	currentStep *clStep
	callback    func()
}

type clStep struct {
	tips   string
	verify func(input string) interface{}
	next   *clStep
}

func NewCMDLineApp(callback func()) *CMDLineApp {
	return &CMDLineApp{
		steps:    map[string]*clStep{},
		callback: callback,
	}
}

func (cl *CMDLineApp) AddStepWithLabel(label, tips string, verify func(input string) interface{}) *CMDLineApp {
	if len(tips) == 0 || verify == nil {
		return cl
	}

	step := &clStep{tips: tips, verify: verify}
	if cl.firstStep == nil {
		cl.firstStep = step
	}
	if cl.currentStep != nil {
		cl.currentStep.next = step
	}
	cl.currentStep = step

	if len(label) > 0 {
		if cl.steps == nil {
			cl.steps = map[string]*clStep{}
		}
		if _, ok := cl.steps[label]; !ok {
			cl.steps[label] = step
		}
	}

	return cl
}

func (cl *CMDLineApp) AddStep(tips string, verify func(input string) interface{}) *CMDLineApp {
	return cl.AddStepWithLabel("", tips, verify)
}

func (cl *CMDLineApp) GotoStep(n int) bool {
	if n <= 0 || cl.firstStep == nil {
		return false
	}

	step := cl.firstStep
	for i := 0; i < n; i++ {
		if step == nil {
			return false
		}
		step = step.next
	}
	cl.currentStep = step
	return true
}

func (cl *CMDLineApp) Run() {
	if cl.firstStep == nil {
		return
	}

	cl.currentStep = cl.firstStep
	fmt.Print(cl.currentStep.tips, " ")

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
			vr := cl.currentStep.verify(strings.TrimSpace(buf.String()))
			switch r := vr.(type) {
			case bool:
				if r {
					if cl.currentStep.next == nil {
						if cl.callback != nil {
							cl.callback()
						}
						break SCAN
					}
					cl.currentStep = cl.currentStep.next
					fmt.Print(cl.currentStep.tips, " ")
				} else {
					fmt.Print(cl.currentStep.tips, " ")
				}

			case int:
				if !cl.GotoStep(r) {
					break SCAN
				}

				fmt.Print(cl.currentStep.tips, " ")

			case string:
				if len(r) == 0 {
					fmt.Print(cl.currentStep.tips, " ")
					break
				}

				if len(cl.steps) > 0 {
					if step, ok := cl.steps[r]; ok {
						cl.currentStep = step
						fmt.Print(cl.currentStep.tips, " ")
						break
					}
				}

				fmt.Print(r, " ")

			default:
				fmt.Print(cl.currentStep.tips, " ")
			}

			buf.Reset()
			continue
		}
		buf.WriteByte(c)
	}
}
