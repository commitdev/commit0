package init

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/commitdev/zero/internal/config/globalconfig"
	"github.com/commitdev/zero/internal/config/moduleconfig"
	"github.com/commitdev/zero/internal/util"
	"github.com/commitdev/zero/pkg/credentials"
	"github.com/commitdev/zero/pkg/util/exit"
	"github.com/commitdev/zero/pkg/util/flog"
	"github.com/manifoldco/promptui"
	"gopkg.in/yaml.v2"
)

// Constant to maintain prompt orders so users can have the same flow,
// modules get downloaded asynchronously therefore its easier to just hardcode an order
var AvailableVendorOrders = []string{"aws", "github", "circleci"}

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
	var pName = regexp.MustCompile(`^[A-Za-z0-9-]{1,15}$`)
	if !pName.MatchString(input) {
		return errors.New("Invalid project-name (cannot contain special chars except '-' & max len of 15)")
	}
	return nil
}

func validateRootDomain(input string) error {
	// FIXME: this regex would still accept subdomains a.co.uk a permanent solution matching all TLDs
	// is not pretty https://publicsuffix.org/list/public_suffix_list.dat another solution would be shortlisting
	// TLDs to the most commonly used.
	var rootDomain = regexp.MustCompile(`^([a-z0-9]+(-[a-z0-9]+)*\.{1})+[a-z]{2,}$`)
	if !rootDomain.MatchString(input) {
		return errors.New("Invalid root domain name")
	}
	return nil
}

func validateSubDomain(input string) error {
	// match all char a-z and 0-9 can contain a - must end with a .
	var subDomainName = regexp.MustCompile(`^([a-z0-9]+(-[a-z0-9]+)*\.)$`)
	if !subDomainName.MatchString(input) {
		return errors.New("Invalid subdomain (cannot contain special chars & must end with a '.')")
	}
	return nil
}

// TODO: validation / allow prompt retry ...etc
func (p PromptHandler) GetParam(projectParams map[string]string) string {
	var err error
	var result string

	if p.Condition(projectParams) {
		if p.Parameter.Info != "" {
			flog.Guidef(p.Parameter.Info)
		}
		// TODO: figure out scope of projectParams per project
		// potentially dangerous to have cross module env leaking
		// so if community module has an `execute: twitter tweet $ENV`
		// it wouldnt leak things the module shouldnt have access to
		if p.Parameter.Execute != "" {
			result = executeCmd(p.Parameter.Execute, projectParams)
		} else if p.Parameter.Value != "" {
			result = p.Parameter.Value
		} else {
			err, result = promptParameter(p)
		}
		if err != nil {
			exit.Fatal("Exiting prompt:  %v\n", err)
		}

		return sanitizeParameterValue(result)
	}
	return ""
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
		prompt := promptui.Select{
			Label: label,
			Items: param.Options,
		}
		_, result, err = prompt.Run()

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
		exit.Fatal("Exiting prompt:  %v\n", err)
	}

	return nil, result
}

func executeCmd(command string, envVars map[string]string) string {
	cmd := exec.Command("bash", "-c", command)
	cmd.Env = util.AppendProjectEnvToCmdEnv(envVars, os.Environ())
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

// PromptParams renders series of prompt UI based on the config
func PromptModuleParams(moduleConfig moduleconfig.ModuleConfig, parameters map[string]string, projectCredentials globalconfig.ProjectCredential) (map[string]string, error) {
	// map module field names to corresponding validate functions.
	// not optimal solution as changing a field name in the EKS stack would render the map invalid.
	m := map[string]func(string) error{
		"productionHostRoot": validateRootDomain,
		"stagingHostRoot":    validateRootDomain,

		"productionFrontendSubdomain": validateSubDomain,
		"productionBackendSubdomain":  validateSubDomain,
		"stagingFrontendSubdomain":    validateSubDomain,
		"stagingBackendSubdomain":     validateSubDomain,
	}

	credentialEnvs := projectCredentials.SelectedVendorsCredentialsAsEnv(moduleConfig.RequiredCredentials)
	for _, promptConfig := range moduleConfig.Parameters {
		// deduplicate fields already prompted and received
		if _, isAlreadySet := parameters[promptConfig.Field]; isAlreadySet {
			continue
		}

		// evaluate which validation method to use
		evalFunc := m[promptConfig.Field]

		promptHandler := PromptHandler{
			Parameter: promptConfig,
			Condition: NoCondition,
			Validate:  evalFunc,
		}
		// merging the context of param and credentals
		// this treats credentialEnvs as throwaway, parameters is shared between modules
		// so credentials should not be in parameters as it gets returned to parent
		for k, v := range parameters {
			credentialEnvs[k] = v
		}
		result := promptHandler.GetParam(credentialEnvs)

		parameters[promptConfig.Field] = result
	}
	return parameters, nil
}

func promptCredentialsAndFillProjectCreds(credentialPrompts []CredentialPrompts, creds globalconfig.ProjectCredential) globalconfig.ProjectCredential {
	promptsValues := map[string]map[string]string{}

	for _, prompts := range credentialPrompts {
		vendor := prompts.Vendor
		vendorPromptValues := map[string]string{}

		// vendors like AWS have multiple prompts (accessKeyId and secretAccessKey)
		for _, prompt := range prompts.Prompts {
			vendorPromptValues[prompt.Field] = prompt.GetParam(vendorPromptValues)
		}
		promptsValues[vendor] = vendorPromptValues
	}

	// FIXME: what is a good way to dynamically modify partial data of a struct
	// current just marashing to yaml, then unmarshaling into the base struct
	yamlContent, _ := yaml.Marshal(promptsValues)
	yaml.Unmarshal(yamlContent, &creds)

	// Fill AWS credentials based on profile from ~/.aws/credentials
	if val, ok := promptsValues["aws"]; ok {
		if val["use_aws_profile"] == awsPickProfile {
			creds = credentials.GetAWSProfileProjectCredentials(val["aws_profile"], creds)
		}
	}
	return creds
}

func appendToSet(set []string, toAppend []string) []string {
	for _, appendee := range toAppend {
		if !util.ItemInSlice(set, appendee) {
			set = append(set, appendee)
		}
	}
	return set
}
