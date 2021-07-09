package generate

import (
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sync"
	"text/template"

	"github.com/commitdev/zero/internal/condition"
	"github.com/commitdev/zero/internal/config/projectconfig"
	"github.com/commitdev/zero/internal/constants"
	"github.com/commitdev/zero/internal/module"
	"github.com/commitdev/zero/internal/util"
	"github.com/commitdev/zero/pkg/util/flog"
	"github.com/commitdev/zero/pkg/util/fs"

	"github.com/gabriel-vasile/mimetype"
)

// Generate accepts a projectconfig struct and renders the templates for all referenced modules
func Generate(projectConfig projectconfig.ZeroProjectConfig, overwriteFiles bool) error {
	flog.Infof(":clock: Fetching Modules")

	// Make sure module sources are on disk
	wg := sync.WaitGroup{}
	wg.Add(len(projectConfig.Modules))
	for _, mod := range projectConfig.Modules {
		go module.FetchModule(mod.Files.Source, &wg)
	}
	wg.Wait()

	flog.Infof(":memo: Rendering Modules")
	for _, mod := range projectConfig.Modules {
		// Load module configuration
		moduleConfig, err := module.ParseModuleConfig(mod.Files.Source)
		if err != nil {
			return fmt.Errorf("unable to load module (%s):  %v", mod.Files.Source, err)
		}

		moduleDir := path.Join(module.GetSourceDir(mod.Files.Source), moduleConfig.InputDir)
		delimiters := moduleConfig.Delimiters
		outputDir := mod.Files.Directory

		// Data that will be passed in to each template
		templateData := struct {
			Name       string
			Params     projectconfig.Parameters
			Files      projectconfig.Files
			Conditions []projectconfig.Condition
		}{
			projectConfig.Name,
			mod.Parameters,
			mod.Files,
			mod.Conditions,
		}

		txtTypeFiles, binTypeFiles := sortFileType(moduleDir, outputDir, overwriteFiles)

		executeTemplates(txtTypeFiles, templateData, delimiters)
		copyBinFiles(binTypeFiles)

		for _, cond := range mod.Conditions {
			condition.Perform(cond, mod)
		}
	}
	return nil
}

type fileConfig struct {
	source      string
	destination string
	modeBits    os.FileMode
}

// sortFileType walks the module directory to find and classify all files into bin / text/plain (non-bin) types.
func sortFileType(moduleDir string, outputDir string, overwrite bool) ([]*fileConfig, []*fileConfig) {
	binTypeFiles := []*fileConfig{}
	txtTypeFiles := []*fileConfig{}

	paths, err := getAllFilePathsInDirectory(moduleDir)
	if err != nil {
		panic(err)
	}

	for _, path := range paths {
		ignoredPaths, _ := regexp.Compile(constants.IgnoredPaths)
		if ignoredPaths.MatchString(path) {
			continue
		}

		outputPath := fs.ReplacePath(path, moduleDir, outputDir)

		if !overwrite {
			if exists, _ := fs.FileExists(outputPath); exists {
				flog.Warnf("%v already exists. skipping.", outputPath)
				continue
			}
		}

		fileInfo, err := os.Stat(path)
		if err != nil {
			panic(err)
		}

		// detect the file type
		detectedMIME, err := mimetype.DetectFile(path)
		if err != nil {
			panic(err)
		}

		// detect root file type
		isBinary := true
		for mime := detectedMIME; mime != nil; mime = mime.Parent() {
			if mime.Is("text/plain") {
				isBinary = false
			}
		}

		if isBinary {
			binTypeFiles = append(binTypeFiles, &fileConfig{
				source:      path,
				destination: outputPath,
				modeBits:    fileInfo.Mode().Perm(),
			})
			continue
		}

		txtTypeFiles = append(txtTypeFiles, &fileConfig{
			source:      path,
			destination: outputPath,
			modeBits:    fileInfo.Mode().Perm(),
		})
	}
	return txtTypeFiles, binTypeFiles
}

// getAllFilePathsInDirectory Recursively get all file paths in directory, including sub-directories.
func getAllFilePathsInDirectory(moduleDir string) ([]string, error) {
	var paths []string
	err := filepath.Walk(moduleDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			paths = append(paths, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return paths, nil
}

func executeTemplates(templates []*fileConfig, data interface{}, delimiters []string) {
	var wg sync.WaitGroup
	leftDelim := delimiters[0]
	rightDelim := delimiters[1]
	if leftDelim == "" {
		leftDelim = "{{"
	}
	if rightDelim == "" {
		rightDelim = "}}"
	}
	// flog.Infof("Templating params:")
	// pp.Println(data)

	for _, tmpltConfig := range templates {
		source := tmpltConfig.source
		dest := tmpltConfig.destination

		outputDirPath, _ := path.Split(dest)
		err := fs.CreateDirs(outputDirPath)
		if err != nil {
			flog.Errorf("Error creating directory '%s': %v", source, err)
		}
		f, err := os.Create(dest)
		if err != nil {
			flog.Errorf("Error initializing file '%s'", err)
		}

		err = f.Chmod(tmpltConfig.modeBits)
		if err != nil {
			flog.Errorf("Error changing mode bits '%s'", err)
		}

		// @TODO if strict mode then only copy file
		name := path.Base(source)
		template, err := template.New(name).Delims(leftDelim, rightDelim).Funcs(util.FuncMap).ParseFiles(source)
		err = template.Execute(f, data)

		if err != nil {
			flog.Errorf("Error templating '%s': %v", source, err)
		} else {
			flog.Successf("Finished templating : %s", dest)
		}
	}

	wg.Wait()
}

func copyBinFiles(binTypeFiles []*fileConfig) {
	for _, binFile := range binTypeFiles {
		source := binFile.source
		dest := binFile.destination

		// create dir
		outputDirPath, _ := path.Split(dest)
		err := fs.CreateDirs(outputDirPath)
		if err != nil {
			flog.Errorf("Error creating directory '%s': %v", source, err)
		}

		// create refs to src and dest
		from, err := os.Open(source)
		if err != nil {
			flog.Errorf("Error opening file to read '%s' : %v", source, err)
		}
		defer from.Close()

		to, err := os.OpenFile(dest, os.O_RDWR|os.O_CREATE, binFile.modeBits)
		if err != nil {
			log.Fatal(err)
			flog.Errorf("Error creating file '%s': %v", dest, err)
		}
		defer to.Close()

		// copy file
		_, err = io.Copy(to, from)
		if err != nil {
			flog.Errorf("Error copying file '%s' : %v", source, err)
		} else {
			flog.Successf("Finished copying file : %s", dest)
		}
	}
}
