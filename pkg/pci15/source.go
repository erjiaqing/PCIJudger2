package pci15

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/sirupsen/logrus"
)

type ExecuteCommand struct {
	Source     string
	Executable string
	Compile    []string
	Execute    []string
}

func GetExecuteCommand(code *SourceCode, conf *Config) (*ExecuteCommand, *Language, error) {
	return GetExecuteCommand2(code, conf, "", false)
}

func GetExecuteCommand2(code *SourceCode, conf *Config, path string, ignoreFileName bool) (*ExecuteCommand, *Language, error) {
	if code == nil {
		return nil, nil, errors.New("code is nil")
	}
	Variables := make(map[string]string)
	SourceCode, err := ReadFile(code.Source)
	if err != nil {
		return nil, nil, err
	}

	if path == "" {
		cwd, _ := os.Getwd()
		path = cwd
	}

	if !ignoreFileName && !filepath.IsAbs(code.Source) {
		if err == nil {
			code.Source = filepath.Join(path, code.Source)
		}
	}

	language := &Language{}
	if err := loadYAML(filepath.Join(conf.LanguageStorage, code.Language+".yaml"), language); err != nil {
		return nil, nil, err
	}
	if language.Variable != nil {
		for _, variable := range language.Variable {
			name := variable.Name
			switch variable.Type {
			case "regexp":
				res, err := regexp.Compile(variable.Value)
				if err != nil {
					logrus.Warningf("Unable to compile %s: %v", variable.Value, err)
					continue
				}
				matchRes := res.FindStringSubmatch(SourceCode)
				if variable.Match < len(matchRes) && variable.Match >= 0 {
					Variables[name] = matchRes[variable.Match]
					logrus.Infof("[%s] -> %s\n", name, matchRes[variable.Match])
				}
			case "string":
				Variables[name] = variable.Value
				logrus.Infof("[%s] -> %s\n", name, variable.Value)
			}
		}
	}
	ret := &ExecuteCommand{}

	if ignoreFileName {
		ret.Source = filepath.Join(path, language.Source)
		for k, v := range Variables {
			ret.Source = strings.Replace(ret.Source, "{"+k+"}", v, 1000000)
		}
	} else {
		ret.Source = code.Source
	}

	Variables["source"] = ret.Source
	sourceWOsuffix := strings.LastIndex(ret.Source, ".")
	if sourceWOsuffix == -1 {
		sourceWOsuffix = len(ret.Source)
	}
	Variables["source<"] = ret.Source[:sourceWOsuffix]

	ret.Executable = language.Executable
	for k, v := range Variables {
		ret.Executable = strings.Replace(ret.Executable, "{"+k+"}", v, 1000000)
	}
	// Variables["source"] = ret.Source
	Variables["executable"] = ret.Executable
	for _, str := range language.Compile.Cmd {
		for k, v := range Variables {
			str = strings.Replace(str, "{"+k+"}", v, 1000000)
		}
		ret.Compile = append(ret.Compile, str)
	}
	for _, str := range language.Execute.Cmd {
		for k, v := range Variables {
			str = strings.Replace(str, "{"+k+"}", v, 1000000)
		}
		ret.Execute = append(ret.Execute, str)
	}
	return ret, language, nil
}

func (code *SourceCode) Compile(conf *Config, workdir string) (string, error) {
	return code.Compile2(conf, workdir, false)
}

func (code *SourceCode) Compile2(conf *Config, workdir string, ignoreFileName bool) (string, error) {
	logrus.Infof("Language: %s", code.Language)

	if currentDir, err := os.Getwd(); err != nil {
		return "", err
	} else if err := os.Chdir(workdir); err != nil {
		return "", err
	} else {
		defer os.Chdir(currentDir)
	}

	compileCfg, lang, err := GetExecuteCommand2(code, conf, workdir, ignoreFileName)
	if err != nil {
		return "", err
	}

	compileRes, err := Execute(compileCfg.Compile, lang.Compile.TimeLimit, 1024*1024*1024, 1.0, "", "", false, "-", "-", "compile_error")
	if err != nil {
		return "", err
	}
	code.CompileResult = compileRes
	_, err = os.Stat(compileCfg.Executable)
	compilerStderr, _ := ioutil.ReadFile("compile_error")
	if code.CompileResult.ExitCode != 0 || code.CompileResult.ExitSignal != 0 || code.CompileResult.TermSignal != 0 || err != nil {
		return string(compilerStderr), errors.New("CE")
	}
	if code.CompileResult.ExitReason == "none" {
		return string(compilerStderr), nil
	} else if code.CompileResult.ExitReason != "none" {
		return fmt.Sprintf("Compiler exited with %s", code.CompileResult.ExitReason), errors.New("CE")
	}
	return "", nil
}
