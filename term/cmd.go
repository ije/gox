package term

import (
	"bytes"
	"errors"
	"io"
	"os/exec"
)

type Command struct {
	started    bool
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

func (cmd *Command) CD(dir string) *Command {
	cmd.Dir = dir
	return cmd
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

	switch data := v.(type) {
	case string:
		_, cmd.inputError = cmd.stdinPipe.Write([]byte(data))
	case []byte:
		_, cmd.inputError = cmd.stdinPipe.Write(data)
	case io.Reader:
		_, cmd.inputError = io.Copy(cmd.stdinPipe, data)
	}

	return cmd
}

func (cmd *Command) Output(wr io.Writer, ignoreStderr bool) (err error) {
	defer func() {
		if cmd.started {
			cmd.Wait()
		}
	}()

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

	cmd.stdinPipe.Close()

	if !ignoreStderr {
		buf := bytes.NewBuffer(nil)
		_, err = io.Copy(buf, cmd.stderrPipe)
		cmd.stderrPipe.Close()
		if err == nil && buf.Len() > 0 {
			err = errors.New(buf.String())
		}
		if err != nil {
			return
		}
	} else {
		cmd.stderrPipe.Close()
	}

	if wr != nil {
		_, err = io.Copy(wr, cmd.stdoutPipe)
	}
	cmd.stdoutPipe.Close()

	return
}
