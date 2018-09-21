package fj15

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
	if code == nil {
		return nil, nil, errors.New("code is nil")
	}
	Variables := make(map[string]string)
	SourceCode, err := ReadFile(code.Source)
	if err != nil {
		return nil, nil, err
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
				if variable.MatchTo < len(matchRes) && variable.MatchTo >= 0 {
					Variables[name] = matchRes[variable.MatchTo]
				}
			case "string":
				Variables[name] = variable.Value
			}
		}
	}
	ret := &ExecuteCommand{}
	for _, str := range language.Compile.Cmd {
		for k, v := range Variables {
			str = strings.Replace(str, "{"+k+"}", v, -1)
		}
		ret.Compile = append(ret.Compile, str)
	}
	for _, str := range language.Execute.Cmd {
		for k, v := range Variables {
			str = strings.Replace(str, "{"+k+"}", v, -1)
		}
		ret.Execute = append(ret.Execute, str)
	}
	ret.Source = language.Source
	for k, v := range Variables {
		ret.Source = strings.Replace(ret.Source, "{"+k+"}", v, -1)
	}
	ret.Executable = language.Executable
	for k, v := range Variables {
		ret.Executable = strings.Replace(ret.Executable, "{"+k+"}", v, -1)
	}
	return ret, language, nil
}

func (code *SourceCode) Compile(conf *Config, workdir string) (string, error) {
	logrus.Infof("Language: %s", code.Language)

	if currentDir, err := os.Getwd(); err != nil {
		return "", err
	} else if err := os.Chdir(workdir); err != nil {
		return "", err
	} else {
		defer os.Chdir(currentDir)
	}

	compileCfg, lang, err := GetExecuteCommand(code, conf)
	if err != nil {
		return "", err
	}

	compileRes, err := Execute(compileCfg.Compile, lang.Compile.TimeLimit, 1024*1024*1024, 1.0, "", false, "-", "-", "compile_error")
	if err != nil {
		return "", err
	}
	code.CompileResult = compileRes
	if code.CompileResult.ExitReason != "NONE" {
		return fmt.Sprintf("Compiler exited with %s", code.CompileResult.ExitReason), errors.New("CE")
	}
	executableInfo, err := os.Stat(compileCfg.Executable)
	if code.CompileResult.ExitCode != 0 || code.CompileResult.ExitSignal != 0 || code.CompileResult.TermSignal != 0 || err != nil {
		compilerStderr, _ := ioutil.ReadFile("compile_error")
		return string(compilerStderr), errors.New("CE")
	}
	if code.CompileResult.ExitReason == "NONE" {
		compilerStderr, _ := ioutil.ReadFile("compile_error")
		return string(compilerStderr), nil
	}
	return "", nil
}
