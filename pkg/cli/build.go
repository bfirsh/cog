package cli

import (
	"fmt"
	"os"
	"path"

	"github.com/replicate/cog/pkg/config"
	"github.com/replicate/cog/pkg/docker"
	"github.com/spf13/cobra"
)

func newBuildCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "build",
		Short: "Build an image from cog.yaml",
		Args:  cobra.NoArgs,
		RunE:  buildCommand,
	}
	return cmd
}

func buildCommand(cmd *cobra.Command, args []string) error {

	cfg, projectDir, err := config.GetConfig(projectDirFlag)
	if err != nil {
		return err
	}

	image := cfg.Image
	if image == "" {
		image = "cog-" + path.Base(projectDir) + ":latest"
	}

	fmt.Fprintf(os.Stderr, "Building Docker image from environment in cog.yaml as %s...\n\n", image)

	arch := "cpu"
	generator := docker.NewDockerfileGenerator(cfg, arch, projectDir)
	defer generator.Cleanup()

	dockerfileContents, err := generator.Generate()
	if err != nil {
		return fmt.Errorf("Failed to generate Dockerfile for %s: %w", arch, err)
	}

	if err := docker.Build(projectDir, dockerfileContents, image); err != nil {
		return fmt.Errorf("Failed to build Docker image: %w", err)
	}

	fmt.Fprintf(os.Stderr, "\nImage built as %s\n", image)

	return nil
}
