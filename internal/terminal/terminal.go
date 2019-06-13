package terminal

import (
	"fmt"
	"golang.org/x/crypto/ssh/terminal"
	"io"
	"os"
)

// TermInfo tells you info about a terminal.
type TermInfo struct {
	IsTerminal  bool
	HeightLines int
	WidthCols   int
}

// GetTermInfo returns info about the current file descriptor; is it a terminal,
// and if so, what's it's height and width?
func GetTermInfo(w io.Writer) (TermInfo, error) {
	switch v := w.(type) {
	case *os.File:
		if terminal.IsTerminal(int(v.Fd())) {
			width, height, err := terminal.GetSize(int(v.Fd()))
			if err != nil {
				return TermInfo{}, err
			}

			return TermInfo{
				IsTerminal:  true,
				HeightLines: height,
				WidthCols:   width,
			}, nil
		}
		return TermInfo{IsTerminal: false}, nil

	default:
		fmt.Println("No terminal attached")
		return TermInfo{IsTerminal: false}, nil
	}
}
