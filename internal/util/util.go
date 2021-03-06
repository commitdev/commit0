package util

// @TODO split up and move into /pkg directory

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"text/template"

	"github.com/google/uuid"
)

func CreateDirIfDoesNotExist(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err = os.MkdirAll(path, os.ModePerm)
		return err
	}
	return nil
}

func CleanGoIdentifier(identifier string) string {
	return strings.ReplaceAll(identifier, "-", "")
}

// @TODO how can we make these type of helpers extensible?
var FuncMap = template.FuncMap{
	"Title":             strings.Title,
	"ToLower":           strings.ToLower,
	"CleanGoIdentifier": CleanGoIdentifier,
	"GenerateUUID":      uuid.New,
}

func GetCwd() string {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Getting working directory failed: %v\n", err)
		panic(err)
	}

	return dir
}

func ExecuteCommand(cmd *exec.Cmd, pathPrefix string, envars []string, shouldPipeStdErr bool) error {

	cmd.Dir = pathPrefix
	if !filepath.IsAbs(pathPrefix) {
		dir := GetCwd()
		cmd.Dir = path.Join(dir, pathPrefix)
	}

	stdoutPipe, _ := cmd.StdoutPipe()
	stderrPipe, _ := cmd.StderrPipe()

	var errStdout, errStderr error
	errContent := new(bytes.Buffer)

	cmd.Env = os.Environ()
	if envars != nil {
		cmd.Env = append(os.Environ(), envars...)
	}

	err := cmd.Start()
	if err != nil {
		return err
	}

	go func() {
		_, errStdout = io.Copy(os.Stdout, stdoutPipe)
	}()
	go func() {
		stderrStreams := []io.Writer{errContent}
		if shouldPipeStdErr {
			stderrStreams = append(stderrStreams, os.Stderr)
		}
		stdErr := io.MultiWriter(stderrStreams...)
		_, errStderr = io.Copy(stdErr, stderrPipe)
	}()

	err = cmd.Wait()
	if err != nil {
		// Detecting and returning the makefile error to cmd
		// Passing alone makefile stderr as error message, otherwise it just says "exit status 2"
		if exitError, ok := err.(*exec.ExitError); ok {
			ws := exitError.Sys().(syscall.WaitStatus)
			exitCode := ws.ExitStatus()
			if exitCode == 2 {
				stderrOut := errContent.String()
				isMissingTarget, _ := regexp.MatchString("No rule to make target", stderrOut)
				if isMissingTarget {
					return errors.New("Module missing mandatory targets, this is likely an issue with the module itself.")
				}
				return errors.New(stderrOut)
			}
		}

		return errors.New(errContent.String())
	}

	if errStdout != nil {
		log.Printf("Failed to capture stdout: %v\n", errStdout)
	}

	if errStderr != nil {
		log.Printf("Failed to capture stderr: %v\n", errStderr)
	}
	return nil
}

// ExecuteCommandOutput runs the command and returns its
// combined standard output and standard error.
func ExecuteCommandOutput(cmd *exec.Cmd, pathPrefix string, envars []string) string {

	cmd.Dir = pathPrefix
	if !filepath.IsAbs(pathPrefix) {
		dir := GetCwd()
		cmd.Dir = path.Join(dir, pathPrefix)
	}

	cmd.Env = os.Environ()
	if envars != nil {
		cmd.Env = append(os.Environ(), envars...)
	}

	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Executing command with output failed: (%v) %s\n", err, out)
	}
	return string(out)
}

// AppendProjectEnvToCmdEnv converts a key-value pair map into a slice of `key=value`s
// allow module definition to use an alternative env-var-name than field while apply
func AppendProjectEnvToCmdEnv(envMap map[string]string, envList []string, translationMap map[string]string) []string {

	for key, val := range envMap {
		if val != "" {
			// overwrite key if exist in translation map
			if val, ok := translationMap[key]; ok {
				key = val
			}
			envList = append(envList, fmt.Sprintf("%s=%s", key, val))
		}
	}
	return envList
}

// IndentString will Add x space char padding at the beginging of each line.
func IndentString(content string, spaces int) string {
	var result string
	subStr := strings.Split(content, "\n")
	for _, s := range subStr {
		result += fmt.Sprintf("%"+strconv.Itoa(spaces)+"s%s\n", "", s)
	}
	return result
}

func ItemInSlice(slice []string, target string) bool {
	for _, item := range slice {
		if item == target {
			return true
		}
	}
	return false
}

// ReflectStructValueIntoMap receives a resource of struct type as
// type AWSCreds struct{
// 	AccessKeyID  string `yaml:"accessKeyId,omitempty"`
// 	SecretAccessKey  string `yaml:"secretAccessKey,omitempty"`
// }{
// 	AccessKeyID: "FOO",
// 	SecretAccessKey: "BAR",
// }
// It will base on the tag, fill in the value to supplied map[string]string
func ReflectStructValueIntoMap(resource interface{}, tagName string, paramsToFill map[string]string) {
	t := reflect.ValueOf(resource)

	for i := 0; i < t.NumField(); i++ {

		childStruct := t.Type().Field(i)
		childValue := t.Field(i)
		if childValue.Kind().String() != "string" {
			continue
		}
		tag, _ := parseTag(childStruct.Tag.Get(tagName))
		paramsToFill[tag] = childValue.String()
	}
}

func parseTag(tag string) (string, string) {
	if idx := strings.Index(tag, ","); idx != -1 {
		return tag[:idx], tag[idx+1:]
	}
	return tag, ""
}
