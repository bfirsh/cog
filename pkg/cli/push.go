package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/replicate/cog/pkg/config"
	"github.com/replicate/cog/pkg/docker"
)

func newPushCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use: "push IMAGE[:TAG]",

		Short:   "Push model in current directory to a Docker registry",
		Example: `cog push registry.hooli.corp/hotdog-detector`,
		RunE:    push,
		Args:    cobra.ExactArgs(1),
	}
	addProjectDirFlag(cmd)

	return cmd
}

func push(cmd *cobra.Command, args []string) error {
	image := args[0]

	cfg, projectDir, err := config.GetConfig(projectDirFlag)
	if err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "Building Docker image from environment in cog.yaml as %s...\n\n", buildTag)

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

	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "Pushing image '%s'...\n", image)

	return docker.Push(image)
}
