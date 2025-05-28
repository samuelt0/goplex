package pane

import (
	"io"
	"os"
	"os/exec"

	"github.com/creack/pty"
)

type Pane struct {
	cmd  *exec.Cmd
	ptmx *os.File
	Out  chan []byte
}

func NewPane(shell string) (*Pane, error) {
	if shell == "" {
		shell = "bash"
	}
	cmd := exec.Command(shell)

	ptmx, tty, err := pty.Open()
	if err != nil {
		return nil, err
	}

	cmd.Stdin = tty
	cmd.Stdout = tty
	cmd.Stderr = tty

	if err := cmd.Start(); err != nil {
		_ = ptmx.Close()
		_ = tty.Close()
		return nil, err
	}
	_ = tty.Close()

	p := &Pane{
		cmd:  cmd,
		ptmx: ptmx,
		Out:  make(chan []byte, 128),
	}

	go p.readLoop()
	return p, nil
}

func (p *Pane) readLoop() {
	buf := make([]byte, 4096)
	for {
		n, err := p.ptmx.Read(buf)
		if n > 0 {
			chunk := make([]byte, n)
			copy(chunk, buf[:n])
			p.Out <- chunk
		}
		if err != nil {
			close(p.Out)
			return
		}
	}
}

func (p *Pane) WriteRune(r rune) error {
	_, err := io.WriteString(p.ptmx, string(r))
	return err
}

func (p *Pane) Close() error {
	_ = p.ptmx.Close()
	return p.cmd.Process.Kill()
}
