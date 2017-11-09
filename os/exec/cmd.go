package exec

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
	pipeError  error
	inputError error
	*exec.Cmd
}

func CMD(name string, args ...string) (cmd *Command) {
	cmd = &Command{Cmd: exec.Command(name, args...)}
	cmd.stdinPipe, cmd.pipeError = cmd.Cmd.StdinPipe()
	if cmd.pipeError != nil {
		return
	}
	cmd.stdoutPipe, cmd.pipeError = cmd.Cmd.StdoutPipe()
	if cmd.pipeError != nil {
		return
	}
	cmd.stderrPipe, cmd.pipeError = cmd.Cmd.StderrPipe()
	return
}

func (cmd *Command) CD(dir string) *Command {
	cmd.Dir = dir
	return cmd
}

func (cmd *Command) Input(v interface{}) *Command {
	if cmd.pipeError != nil {
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

func (cmd *Command) Output(wr io.Writer) (err error) {
	defer func() {
		if cmd.started {
			cmd.Wait()
		}
	}()

	if cmd.pipeError != nil {
		err = cmd.pipeError
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

	waitChan := make(chan struct{}, 1)

	errBuf := bytes.NewBuffer(nil)
	go func() {
		_, err = io.Copy(errBuf, cmd.stderrPipe)
		cmd.stderrPipe.Close()
		waitChan <- struct{}{}
	}()
	go func() {
		_, err = io.Copy(wr, cmd.stdoutPipe)
		cmd.stdoutPipe.Close()
		waitChan <- struct{}{}
	}()

	<-waitChan
	if err != nil {
		return
	}

	if errBuf.Len() > 0 {
		err = errors.New(errBuf.String())
	}
	return
}
