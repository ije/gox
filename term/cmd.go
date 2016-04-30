package term

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
)

type Command struct {
	started    bool
	stoped     bool
	stdinPipe  io.WriteCloser
	stdoutPipe io.ReadCloser
	*exec.Cmd
}

func NewCMD(name string, args ...string) *Command {
	c := exec.Command(name, args...)
	stdinPipe, _ := c.StdinPipe()
	stdoutPipe, _ := c.StdoutPipe()

	return &Command{
		stdinPipe:  stdinPipe,
		stdoutPipe: stdoutPipe,
		Cmd:        c,
	}
}

func (c *Command) Input(v interface{}) (cmd *Command, err error) {
	if !c.started {
		err = c.Start()
		if err != nil {
			return
		}
		c.started = true
	}

	switch d := v.(type) {
	case string:
		_, err = c.stdinPipe.Write([]byte(d))
	case []byte:
		_, err = c.stdinPipe.Write(d)
	case io.Reader:
		_, err = io.Copy(c.stdinPipe, d)
	default:
		_, err = fmt.Fprintf(c.stdinPipe, "%v", d)
	}

	cmd = c
	return
}

func (c *Command) Output(prefix []byte) (ret *bytes.Buffer, err error) {
	if !c.started {
		err = fmt.Errorf("waiting input")
		return
	}

	if c.stoped {
		err = fmt.Errorf("stoped")
		return
	}

	c.stdinPipe.Close()

	buffer := bytes.NewBuffer(prefix)
	io.Copy(buffer, c.stdoutPipe)
	c.stdoutPipe.Close()

	err = c.Wait()
	if err != nil {
		return
	}

	ret = buffer
	c.stoped = true
	return
}
