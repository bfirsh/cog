package terminal

import (
	"context"
	"os"

	"github.com/containerd/console"
	"github.com/mattn/go-isatty"
	sshterm "golang.org/x/crypto/ssh/terminal"
)

// Returns a UI which will write to the current processes
// stdout/stderr.
func ConsoleUI(ctx context.Context) UI {
	// We do both of these checks because some sneaky environments fool
	// one or the other and we really only want the glint-based UI in
	// truly interactive environments.
	glint := isatty.IsTerminal(os.Stdout.Fd()) && sshterm.IsTerminal(int(os.Stdout.Fd()))
	if glint {
		glint = false
		if c, err := console.ConsoleFromFile(os.Stdout); err == nil {
			if sz, err := c.Size(); err == nil {
				glint = sz.Height > 0 && sz.Width > 0
			}
		}
	}

	if glint {
		return GlintUI(ctx)
	} else {
		return NonInteractiveUI(ctx)
	}
}
