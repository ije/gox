package term

import (
	"bytes"
	"errors"
	"io"
	"os/exec"
)

type Command struct {
	started    bool
	stoped     bool
	stdinPipe  io.WriteCloser
	stdoutPipe io.ReadCloser
	stderrPipe io.ReadCloser
	stdError   error
	inputError error
	*exec.Cmd
}

func CMD(name string, args ...string) (cmd *Command) {
	cmd = &Command{Cmd: exec.Command(name, args...)}
	cmd.stdinPipe, cmd.stdError = cmd.Cmd.StdinPipe()
	if cmd.stdError != nil {
		return
	}
	cmd.stdoutPipe, cmd.stdError = cmd.Cmd.StdoutPipe()
	if cmd.stdError != nil {
		return
	}
	cmd.stderrPipe, cmd.stdError = cmd.Cmd.StderrPipe()
	return
}

func (cmd *Command) Input(v interface{}) *Command {
	if cmd.stdError != nil {
		return cmd
	}

	if !cmd.started {
		cmd.inputError = cmd.Cmd.Start()
		if cmd.inputError != nil {
			return cmd
		}
		cmd.started = true
	}

	switch d := v.(type) {
	case string:
		_, cmd.inputError = cmd.stdinPipe.Write([]byte(d))
	case []byte:
		_, cmd.inputError = cmd.stdinPipe.Write(d)
	case io.Reader:
		_, cmd.inputError = io.Copy(cmd.stdinPipe, d)
	}

	return cmd
}

func (cmd *Command) Output(wr io.Writer) (err error) {
	if cmd.stdError != nil {
		err = cmd.stdError
		return
	}

	if cmd.inputError != nil {
		err = cmd.inputError
		return
	}

	if !cmd.started {
		err = cmd.Cmd.Start()
		if err != nil {
			return
		}
		cmd.started = true
	}

	if cmd.stoped {
		err = errors.New("stoped")
		return
	}

	err = cmd.stdinPipe.Close()
	if err != nil {
		return
	}

	buf := bytes.NewBuffer(nil)
	io.Copy(buf, cmd.stderrPipe)
	err = cmd.stderrPipe.Close()
	if err != nil {
		return
	} else if buf.Len() > 0 {
		err = errors.New(buf.String())
		return
	}

	io.Copy(wr, cmd.stdoutPipe)
	err = cmd.stdoutPipe.Close()
	if err != nil {
		return
	}

	err = cmd.Wait()
	if err != nil {
		return
	}

	cmd.stoped = true
	return
}
