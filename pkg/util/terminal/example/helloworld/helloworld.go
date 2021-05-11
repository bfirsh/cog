package main

import (
	"context"
	"io"
	"time"

	"github.com/replicate/cog/pkg/util/terminal"
)

func main() {
	ui := terminal.ConsoleUI(context.Background())
	status := ui.Status()
	status.Update("reticulating splines...")
	time.Sleep(2 * time.Second)
	status.Step(terminal.StatusOK, "hello world")

	if closer, ok := ui.(io.Closer); ok && closer != nil {
		closer.Close()
	}

}
