package cli

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/replicate/cog/pkg/config"
	"github.com/replicate/cog/pkg/docker"
	"github.com/spf13/cobra"
)

func newRunCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run <command> [arg...]",
		Short: "Run a command inside a Docker environment",
		RunE:  run,
		Args:  cobra.MinimumNArgs(1),
	}

	flags := cmd.Flags()
	// Flags after first argment are considered args and passed to command
	flags.SetInterspersed(false)

	return cmd
}

func run(cmd *cobra.Command, args []string) error {
	// TODO: support multiple run architectures, or automatically select arch based on host
	arch := "cpu"

	cfg, projectDir, err := config.GetConfig(projectDirFlag)
	if err != nil {
		return err
	}

	// FIXME: refactor to share with predict

	// TODO: ditch tag for run so that prune works?
	image := "cog-" + path.Base(projectDir) + "-base:latest"

	fmt.Fprintf(os.Stderr, "Building Docker image from environment in cog.yaml...\n\n")

	generator := docker.NewDockerfileGenerator(cfg, arch, projectDir)
	defer generator.Cleanup()

	dockerfileContents, err := generator.GenerateBase()
	if err != nil {
		return fmt.Errorf("Failed to generate Dockerfile for %s: %w", arch, err)
	}

	if err := docker.Build(projectDir, dockerfileContents, image); err != nil {
		return fmt.Errorf("Failed to build Docker image: %w", err)
	}

	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "Running '%s' in Docker with the current directory mounted as a volume...\n", strings.Join(args, " "))
	return docker.Run(projectDir, image, args)

	// TODO: delete image
}
