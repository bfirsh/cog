package cli

import (
	"fmt"
	"os"
	"regexp"

	"github.com/spf13/cobra"

	"github.com/replicate/cog/pkg/global"
	"github.com/replicate/cog/pkg/model"
	"github.com/replicate/cog/pkg/settings"
	"github.com/replicate/cog/pkg/util/console"
)

var modelFlag string
var projectDirFlag string

var modelRegex = regexp.MustCompile("^(?:(https?://[^/]*)/)?(?:([-_a-zA-Z0-9]+)/)([-_a-zA-Z0-9]+)$")

func NewRootCommand() (*cobra.Command, error) {
	rootCmd := cobra.Command{
		Use:   "cog",
		Short: "Cog: Containers for machine learning",
		Long: `Containers for machine learning.
		
To get started, you first need to create a 'cog.yaml' file. Check out the docs to learn how to do that:

...`,
		Example: `   To run a command inside a Docker environment defined with Cog:
      $ cog run echo hello world`,
		Version: fmt.Sprintf("%s (built %s)", global.Version, global.BuildTime),
		// This stops errors being printed because we print them in cmd/cog/cog.go
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if global.Verbose {
				console.SetLevel(console.DebugLevel)
			}
			cmd.SilenceUsage = true
		},
		SilenceErrors: true,
	}
	setPersistentFlags(&rootCmd)

	rootCmd.AddCommand(
		newBuildCommand(),
		newDebugCommand(),
		newPushCommand(),
		newPredictCommand(),
		newRunCommand(),
	)

	return &rootCmd, nil
}

func setPersistentFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().BoolVar(&global.Verbose, "debug", false, "Show debugging output")
	cmd.PersistentFlags().BoolVar(&global.ProfilingEnabled, "profile", false, "Enable profiling")
	cmd.PersistentFlags().Bool("version", false, "Show version of Cog")
	_ = cmd.PersistentFlags().MarkHidden("profile")
}

func addModelFlag(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&modelFlag, "model", "m", "", "Model URL, e.g. https://cog.hooli.corp/hotdog-detector/hotdog-detector")
}

func getModel() (*model.Model, error) {
	if modelFlag != "" {
		model, err := parseModel(modelFlag)
		if err != nil {
			return nil, err
		}
		return model, nil
	} else {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		projectSettings, err := settings.LoadProjectSettings(cwd)
		if err != nil {
			return nil, err
		}
		if projectSettings.Model != nil {
			return projectSettings.Model, nil
		}
		return nil, fmt.Errorf("No model set. You need to either run `cog model set <url>` to set a model for the current directory, or pass --model <url> to the command")
	}
}

func parseModel(modelString string) (*model.Model, error) {
	matches := modelRegex.FindStringSubmatch(modelString)
	if len(matches) == 0 {
		return nil, fmt.Errorf("Model '%s' doesn't match [http[s]://<host>/]<user>/<name>", modelString)
	}
	return &model.Model{
		Host: matches[1],
		User: matches[2],
		Name: matches[3],
	}, nil
}

func addProjectDirFlag(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&projectDirFlag, "project-dir", "D", "", "Project directory, defaults to current working directory")
}
