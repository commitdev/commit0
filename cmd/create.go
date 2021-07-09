package cmd

import (
	"fmt"
	"path"
	"strings"

	"github.com/commitdev/zero/internal/config/projectconfig"
	"github.com/commitdev/zero/internal/constants"
	"github.com/commitdev/zero/internal/generate"
	"github.com/commitdev/zero/internal/vcs"
	"github.com/commitdev/zero/pkg/util/exit"
	"github.com/commitdev/zero/pkg/util/flog"
	"github.com/spf13/cobra"
)

var (
	createConfigPath string
	overwriteFiles   bool
)

func init() {
	createCmd.PersistentFlags().StringVarP(&createConfigPath, "config", "c", constants.ZeroProjectYml, "The project.yml file to load. The default is the one in the current directory.")
	createCmd.PersistentFlags().BoolVarP(&overwriteFiles, "overwrite", "o", false, "overwrite pre-existing files")

	rootCmd.AddCommand(createCmd)
}

var createCmd = &cobra.Command{
	Use:   "create",
	Short: fmt.Sprintf("Create projects for modules and configuration specified in %s", constants.ZeroProjectYml),
	Run: func(cmd *cobra.Command, args []string) {
		Create(projectconfig.RootDir, createConfigPath)
	},
}

func Create(dir string, createConfigPath string) {
	if strings.Trim(createConfigPath, " ") == "" {
		exit.Fatal("config path cannot be empty!")
	}
	configFilePath := path.Join(dir, createConfigPath)
	projectConfig := projectconfig.LoadConfig(configFilePath)

	generate.Generate(*projectConfig, overwriteFiles)

	if projectConfig.ShouldPushRepositories {
		flog.Infof(":up_arrow: Done Rendering - committing repositories to version control.")

		for _, module := range projectConfig.Modules {
			err, githubApiKey := projectconfig.ReadVendorCredentialsFromModule(module, "github")
			if err != nil {
				flog.Errorf(err.Error())
			}
			vcs.InitializeRepository(module.Files.Repository, githubApiKey)
		}
	} else {
		flog.Infof(":up_arrow: Done Rendering - you will need to commit the created projects to version control.")
	}

	flog.Infof(":check_mark_button: Done - run zero apply to create any required infrastructure or execute any other remote commands to prepare your environments.")
}
