package cli

import (
	"fmt"
	"os"
	"path"

	"github.com/replicate/cog/pkg/config"
	"github.com/replicate/cog/pkg/docker"
	"github.com/spf13/cobra"
)

var buildTag string

func newBuildCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "build",
		Short: "Build an image from cog.yaml",
		Args:  cobra.NoArgs,
		RunE:  buildCommand,
	}
	cmd.Flags().StringVarP(&buildTag, "tag", "t", "", "A name for the built image in the form 'repository:tag'")
	return cmd
}

func buildCommand(cmd *cobra.Command, args []string) error {

	cfg, projectDir, err := config.GetConfig(projectDirFlag)
	if err != nil {
		return err
	}

	if buildTag == "" {
		buildTag = "cog-" + path.Base(projectDir) + ":latest"
	}

	fmt.Fprintf(os.Stderr, "Building Docker image from environment in cog.yaml as %s...\n\n", buildTag)

	arch := "cpu"
	generator := docker.NewDockerfileGenerator(cfg, arch, projectDir)
	defer generator.Cleanup()

	dockerfileContents, err := generator.Generate()
	if err != nil {
		return fmt.Errorf("Failed to generate Dockerfile for %s: %w", arch, err)
	}

	if err := docker.Build(projectDir, dockerfileContents, buildTag); err != nil {
		return fmt.Errorf("Failed to build Docker image: %w", err)
	}

	fmt.Fprintf(os.Stderr, "\nImage built as %s\n", buildTag)

	return nil
}
