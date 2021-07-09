package init

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"

	tm "github.com/buger/goterm"
	"github.com/commitdev/zero/internal/config/moduleconfig"
	"github.com/commitdev/zero/internal/constants"
	"github.com/commitdev/zero/internal/util"
	"github.com/commitdev/zero/pkg/util/exit"
	"github.com/commitdev/zero/pkg/util/flog"
	"github.com/manifoldco/promptui"
	"gopkg.in/yaml.v2"
)

const cyanArrow = "\033[36m\U000025B6\033[0m"
const greenCheckMark = "\033[32m\U00002714\033[0m"

const awsPickProfile = "Existing AWS Profiles"
const awsManualInputCredentials = "Enter my own AWS credentials"

type PromptHandler struct {
	moduleconfig.Parameter
	Condition CustomConditionSignature
	Validate  func(string) error
}

type CredentialPrompts struct {
	Vendor  string
	Prompts []PromptHandler
}

type CustomConditionSignature func(map[string]string) bool

func NoCondition(map[string]string) bool {
	return true
}

func KeyMatchCondition(key string, value string) CustomConditionSignature {
	return func(param map[string]string) bool {
		return param[key] == value
	}
}

func CustomCondition(fn CustomConditionSignature) CustomConditionSignature {
	return func(param map[string]string) bool {
		return fn(param)
	}
}

func NoValidation(string) error {
	return nil
}

func SpecificValueValidation(values ...string) func(string) error {
	return func(checkValue string) error {
		for _, allowedValue := range values {
			if checkValue == allowedValue {
				return nil
			}
		}
		return fmt.Errorf("Please choose one of %s", strings.Join(values, "/"))
	}
}

func ValidateAKID(input string) error {
	// 20 uppercase alphanumeric characters
	var awsAccessKeyIDPat = regexp.MustCompile(`^[A-Z0-9]{20}$`)
	if !awsAccessKeyIDPat.MatchString(input) {
		return errors.New("Invalid aws_access_key_id")
	}
	return nil
}

func ValidateSAK(input string) error {
	// 40 base64 characters
	var awsSecretAccessKeyPat = regexp.MustCompile(`^[A-Za-z0-9/+=]{40}$`)
	if !awsSecretAccessKeyPat.MatchString(input) {
		return errors.New("Invalid aws_secret_access_key")
	}
	return nil
}

// ValidateProjectName validates Project Name field user input.
func ValidateProjectName(input string) error {
	// the first 62 char out of base64 and -
	var pName = regexp.MustCompile(`^[A-Za-z0-9-]{1,16}$`)
	if !pName.MatchString(input) {
		// error if char len is greater than 16
		if len(input) > constants.MaxPnameLength {
			return errors.New("Invalid, Project Name: (cannot exceed a max length of 16)")
		}
		return errors.New("Invalid, Project Name: (can only contain alphanumeric chars & '-')")
	}
	return nil
}

const infoBoxHeight = 4

var currentLine int = infoBoxHeight

// showInfoBox prints a box with some text in it, and the title "Info"
func showInfoBox(infoText string) {
	box := tm.NewBox(100|tm.PCT, 4, 0)
	fmt.Fprint(box, infoText)
	tm.Print(tm.MoveTo(box.String(), 1, 1))
	tm.MoveCursor(4, 1)
	tm.Printf("Info")
}

// RunPrompt obtains the value of PromptHandler depending on the parameter's definition
// for the project config,  there are multiple ways of obtaining the value
// values go into params depending on `Condition` as the highest precedence (Whether it gets this value)
// then follows this order to determine HOW it obtains that value
// 1. Execute (this could potentially be refactored into type + data)
// 2. type: specific ways of obtaining values (in AWS credential case it will set 2 values to the map)
// 3. value: directly assigns a value to a parameter
// 4. prompt: requires users to select an option OR input a string
func (p PromptHandler) RunPrompt(projectParams map[string]string, envVarTranslationMap map[string]string) error {
	var err error
	var result string

	if p.Condition(projectParams) {

		// If we start printing below the bottom of the terminal screen, go back to the top
		if currentLine+infoBoxHeight+1 > tm.Height() {
			tm.Clear()
			currentLine = infoBoxHeight
		}
		showInfoBox(p.Parameter.Info)

		// TODO: figure out scope of projectParams per project
		// potentially dangerous to have cross module env leaking
		// so if community module has an `execute: twitter tweet $ENV`
		// it wouldnt leak things the module shouldnt have access to
		if p.Parameter.Execute != "" {
			result = executeCmd(p.Parameter.Execute, projectParams, envVarTranslationMap)
		} else if p.Parameter.Type != "" {
			err = CustomPromptHandler(p.Parameter.Type, projectParams)
		} else if p.Parameter.Value != "" {
			result = p.Parameter.Value
		} else {
			// Move down to the next line to show the prompt
			currentLine++
			tm.MoveCursor(1, currentLine)
			tm.Flush() // Call it every time at the end of rendering

			err, result = promptParameter(p)
		}
		if err != nil {
			return err
		}

		// Append the result to parameter map
		projectParams[p.Field] = sanitizeParameterValue(result)
	} else {
		flog.Debugf("Skipping prompt(%s) due to condition failed", p.Field)
	}
	return nil
}

func promptParameter(prompt PromptHandler) (error, string) {
	param := prompt.Parameter
	label := param.Label
	if param.Label == "" {
		label = param.Field
	}
	defaultValue := param.Default

	var err error
	var result string
	if len(param.Options) > 0 {
		var selectedIndex int
		// Scope of selected does not have the label data, so we need a dynamic
		// template with string format to put in the label in `selected`
		optionTemplate := &promptui.SelectTemplates{
			Label:    `{{ . }}`,
			Active:   fmt.Sprintf("%s {{ .Value | cyan }}", cyanArrow),
			Inactive: "  {{ .Value }}",
			Selected: fmt.Sprintf("%s %s: {{ .Value }}", greenCheckMark, label),
		}

		prompt := promptui.Select{
			Label:     label,
			Items:     param.Options,
			Templates: optionTemplate,
		}

		selectedIndex, _, err = prompt.Run()
		result = param.Options[selectedIndex].Key.(string)
	} else {
		prompt := promptui.Prompt{
			Label:     label,
			Default:   defaultValue,
			AllowEdit: true,
			Validate:  prompt.Validate,
		}
		result, err = prompt.Run()
	}
	if err != nil {
		return err, ""
	}

	return nil, result
}

func executeCmd(command string, envVars map[string]string, envVarTranslationMap map[string]string) string {
	cmd := exec.Command("bash", "-c", command)
	// Might need to pass down module's translation map as well,
	// currently only works in `zero apply`
	cmd.Env = util.AppendProjectEnvToCmdEnv(envVars, os.Environ(), envVarTranslationMap)
	out, err := cmd.Output()
	flog.Debugf("Running command: %s", command)
	if err != nil {
		log.Fatalf("Failed to execute  %v\n", err)
	}
	flog.Debugf("Command result: %s", string(out))
	return string(out)
}

// aws cli prints output with linebreak in them
func sanitizeParameterValue(str string) string {
	re := regexp.MustCompile("\\n")
	return re.ReplaceAllString(str, "")
}

// PromptModuleParams renders series of prompt UI based on the config
func PromptModuleParams(moduleConfig moduleconfig.ModuleConfig, parameters map[string]string) (map[string]string, error) {
	envVarTranslationMap := moduleConfig.GetParamEnvVarTranslationMap()
	for _, parameter := range moduleConfig.Parameters {
		// deduplicate fields already prompted and received
		if _, isAlreadySet := parameters[parameter.Field]; isAlreadySet {
			continue
		}

		var validateFunc func(input string) error = nil

		// type:regex field validation for zero-module.yaml
		if parameter.FieldValidation.Type == constants.RegexValidation {
			validateFunc = func(input string) error {
				var regexRule = regexp.MustCompile(parameter.FieldValidation.Value)
				if !regexRule.MatchString(input) {
					return errors.New(parameter.FieldValidation.ErrorMessage)
				}
				return nil
			}
		}
		// TODO: type:fuction field validation for zero-module.yaml

		promptHandler := PromptHandler{
			Parameter: parameter,
			Condition: paramConditionsMapper(parameter.Conditions),
			Validate:  validateFunc,
		}
		// merging the context of param and credentals
		// this treats credentialEnvs as throwaway, parameters is shared between modules
		// so credentials should not be in parameters as it gets returned to parent
		// for k, v := range parameters {
		// 	credentialEnvs[k] = v
		// }
		err := promptHandler.RunPrompt(parameters, envVarTranslationMap)
		if err != nil {
			return parameters, err
		}
	}
	flog.Debugf("Module %s prompt: \n %#v", moduleConfig.Name, parameters)
	return parameters, nil
}

// promptAllModules takes a map of all the modules and prompts the user for values for all the parameters
// Important: This is done here because in this step we share the parameter across modules,
// meaning if module A and B both asks for region, it will reuse the response for both (and is deduped during runtime)
func promptAllModules(modules map[string]moduleconfig.ModuleConfig) map[string]string {
	parameterValues := map[string]string{}
	for _, config := range modules {
		var err error

		parameterValues, err = PromptModuleParams(config, parameterValues)
		if err != nil {
			exit.Fatal("Exiting prompt(%s):  %v\n", config.Name, err)
		}
	}
	return parameterValues
}

func paramConditionsMapper(conditions []moduleconfig.Condition) CustomConditionSignature {
	if len(conditions) == 0 {
		return NoCondition
	} else {
		return func(params map[string]string) bool {
			// Prompts must pass every condition to proceed
			for i := 0; i < len(conditions); i++ {
				cond := conditions[i]
				if !conditionHandler(cond)(params) {
					flog.Debugf("Did not meet condition %v, expected %v to be %v", cond.Action, cond.MatchField, cond.WhenValue)
					return false
				}
			}
			return true
		}
	}
}
func conditionHandler(cond moduleconfig.Condition) CustomConditionSignature {
	if cond.Action == "KeyMatchCondition" {
		return KeyMatchCondition(cond.MatchField, cond.WhenValue)
	} else {
		flog.Errorf("Unsupported condition")
		return nil
	}
}

func appendToSet(set []string, toAppend []string) []string {
	for _, appendee := range toAppend {
		if !util.ItemInSlice(set, appendee) {
			set = append(set, appendee)
		}
	}
	return set
}

func listToPromptOptions(list []string) yaml.MapSlice {
	mapSlice := make(yaml.MapSlice, len(list))
	for i := 0; i < len(list); i++ {
		mapSlice[i] = yaml.MapItem{
			Key:   list[i],
			Value: list[i],
		}
	}
	return mapSlice
}
