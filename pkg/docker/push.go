package docker

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/logrusorgru/aurora"
)

func Push(image string) error {
	cmd := exec.Command(
		"docker", "push", image)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Fprintln(os.Stderr, aurora.Faint("$ "+strings.Join(cmd.Args, " ")).String())
	return cmd.Run()
}
