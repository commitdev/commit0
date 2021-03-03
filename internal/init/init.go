package init

import (
	"fmt"
	"os"
	"path"
	"sync"

	"github.com/commitdev/zero/internal/config/moduleconfig"
	"github.com/commitdev/zero/internal/config/projectconfig"
	"github.com/commitdev/zero/internal/module"
	"github.com/commitdev/zero/internal/registry"
	"github.com/commitdev/zero/pkg/util/exit"
	"github.com/commitdev/zero/pkg/util/flog"
	"github.com/manifoldco/promptui"
)

// Create cloud provider context
func Init(outDir string, localModulePath string) *projectconfig.ZeroProjectConfig {
	projectConfig := defaultProjConfig()

	projectRootParams := map[string]string{}
	promptName := getProjectNamePrompt()
	promptName.RunPrompt(projectRootParams)
	projectConfig.Name = projectRootParams[promptName.Field]

	rootDir := path.Join(outDir, projectConfig.Name)
	flog.Infof(":tada: Initializing project")

	err := os.MkdirAll(rootDir, os.ModePerm)
	if os.IsExist(err) {
		exit.Fatal("Directory %v already exists! Error: %v", projectConfig.Name, err)
	} else if err != nil {
		exit.Fatal("Error creating root: %v ", err)
	}

	moduleSources := chooseStack(registry.GetRegistry(localModulePath))
	moduleConfigs, mappedSources := loadAllModules(moduleSources)

	prompts := getProjectPrompts(projectConfig.Name, moduleConfigs)

	initParams := make(map[string]string)
	projectConfig.ShouldPushRepositories = true
	prompts["ShouldPushRepositories"].RunPrompt(initParams)
	if initParams["ShouldPushRepositories"] == "n" {
		projectConfig.ShouldPushRepositories = false
	}

	// Prompting for push-up stream, then conditionally prompting for github
	prompts["GithubRootOrg"].RunPrompt(initParams)

	projectData := promptAllModules(moduleConfigs)

	// Map parameter values back to specific modules
	for moduleName, module := range moduleConfigs {
		prompts[moduleName].RunPrompt(initParams)
		repoName := initParams[prompts[moduleName].Field]
		repoURL := fmt.Sprintf("%s/%s", initParams["GithubRootOrg"], repoName)
		projectModuleParams := moduleconfig.SummarizeParameters(module, projectData)
		projectModuleConditions := moduleconfig.SummarizeConditions(module)

		projectConfig.Modules[moduleName] = projectconfig.NewModule(
			projectModuleParams,
			repoName,
			repoURL,
			mappedSources[moduleName],
			module.DependsOn,
			projectModuleConditions,
		)
	}

	return &projectConfig
}

// loadAllModules takes a list of module sources, downloads those modules, and parses their config
func loadAllModules(moduleSources []string) (map[string]moduleconfig.ModuleConfig, map[string]string) {
	modules := make(map[string]moduleconfig.ModuleConfig)
	mappedSources := make(map[string]string)

	wg := sync.WaitGroup{}
	wg.Add(len(moduleSources))
	for _, moduleSource := range moduleSources {
		go module.FetchModule(moduleSource, &wg)
	}
	wg.Wait()

	for _, moduleSource := range moduleSources {
		mod, err := module.ParseModuleConfig(moduleSource)
		if err != nil {
			exit.Fatal("Unable to load module:  %v\n", err)
		}
		modules[mod.Name] = mod
		mappedSources[mod.Name] = moduleSource
	}
	return modules, mappedSources
}

// Project name is prompt individually because the rest of the prompts
// requires the projectName to populate defaults
func getProjectNamePrompt() PromptHandler {
	return PromptHandler{
		Parameter: moduleconfig.Parameter{
			Field:   "projectName",
			Label:   "Project Name",
			Default: "",
		},
		Condition: NoCondition,
		Validate:  ValidateProjectName,
	}
}

func getProjectPrompts(projectName string, modules map[string]moduleconfig.ModuleConfig) map[string]PromptHandler {
	handlers := map[string]PromptHandler{
		"ShouldPushRepositories": {
			Parameter: moduleconfig.Parameter{
				Field:   "ShouldPushRepositories",
				Label:   "Should the created projects be checked into github automatically? (y/n)",
				Default: "y",
			},
			Condition: NoCondition,
			Validate:  SpecificValueValidation("y", "n"),
		},
		"GithubRootOrg": {
			Parameter: moduleconfig.Parameter{
				Field:   "GithubRootOrg",
				Label:   "What's the root of the github org to create repositories in?",
				Default: "github.com/",
			},
			Condition: KeyMatchCondition("ShouldPushRepositories", "y"),
			Validate:  NoValidation,
		},
	}

	for moduleName, module := range modules {
		label := fmt.Sprintf("What do you want to call the %s project?", moduleName)

		handlers[moduleName] = PromptHandler{
			Parameter: moduleconfig.Parameter{
				Field:   moduleName,
				Label:   label,
				Default: module.OutputDir,
			},
			Condition: NoCondition,
			Validate:  NoValidation,
		}
	}

	return handlers
}

func chooseCloudProvider(projectConfig *projectconfig.ZeroProjectConfig) {
	// @TODO move options into configs
	providerPrompt := promptui.Select{
		Label: "Select Cloud Provider",
		Items: []string{"Amazon AWS", "Google GCP", "Microsoft Azure"},
	}

	_, providerResult, err := providerPrompt.Run()
	if err != nil {
		exit.Fatal("Prompt failed %v\n", err)
	}

	if providerResult != "Amazon AWS" {
		exit.Fatal("Only the AWS provider is available at this time")
	}
}

func chooseStack(reg registry.Registry) []string {
	providerPrompt := promptui.Select{
		Label: "Pick a stack you'd like to use",
		Items: registry.AvailableLabels(reg),
	}
	_, providerResult, err := providerPrompt.Run()
	if err != nil {
		exit.Fatal("Prompt failed %v\n", err)
	}

	return registry.GetModulesByName(reg, providerResult)
}

func defaultProjConfig() projectconfig.ZeroProjectConfig {
	return projectconfig.ZeroProjectConfig{
		Name:       "",
		Parameters: map[string]string{},
		Modules:    projectconfig.Modules{},
	}
}
