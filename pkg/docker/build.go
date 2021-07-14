package docker

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/logrusorgru/aurora"
)

func Build(dir, dockerfile, imageName string) error {
	cmd := exec.Command(
		"docker", "build", ".",
		"-f", "-",
		"--build-arg", "BUILDKIT_INLINE_CACHE=1",
		"-t", imageName,
	)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = strings.NewReader(dockerfile)
	cmd.Env = append(os.Environ(), "DOCKER_BUILDKIT=1")

	fmt.Fprintln(os.Stderr, aurora.Faint("$ "+strings.Join(cmd.Args, " ")).String())
	return cmd.Run()
}
