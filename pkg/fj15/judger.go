package fj15

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/sirupsen/logrus"
	shutil "github.com/termie/go-shutil"
)

type JudgeResult struct {
	Success   bool           `json:"success"`
	Verdict   string         `json:"verdict"`
	ExeTime   uint64         `json:"exe_time"`
	ExeMemory uint64         `json:"exe_memory"`
	ExitCode  int32          `json:"exit_code"`
	UsedTime  uint64         `json:"used_time"`
	Detail    []*JudgeDetail `json:"detail"`
}

type JudgeDetail struct {
	Name       string `json:"name"`
	Input      string `json:"input"`
	Output     string `json:"output"`
	Answer     string `json:"answer"`
	Verdict    string `json:"verdict"`
	ExeTime    uint64 `json:"exe_time"`
	ExeMemory  uint64 `json:"exe_memory"`
	ExitCode   int32  `json:"exit_code"`
	ExitSignal int32  `json:"exit_signal"`
}

func Judge(conf *Config, code *SourceCode, problem string) (*JudgeResult, error) {
	judgeResult := &JudgeResult{
		Success: true,
	}

	problemConf := &ProblemConfig{}
	if err := loadYAML(filepath.Join(problem, "problem.yaml"), problemConf); err != nil {
		return nil, err
	}

	workDir := filepath.Join(conf.Tmp, GetRandomString())
	if err := os.MkdirAll(workDir, 0777); err != nil {
		return nil, err
	}
	defer os.RemoveAll(workDir)

	var sourceDir string

	if currentDir, err := os.Getwd(); err != nil {
		return nil, err
	} else if err := os.Chdir(workDir); err != nil {
		return nil, err
	} else {
		sourceDir = currentDir
		defer os.Chdir(currentDir)
	}

	execCommand, codeLanguage, err := GetExecuteCommand(code, conf)
	if err != nil {
		return nil, err
	}
	if _, err := shutil.Copy(code.Source, filepath.Join(workDir, execCommand.Source), false); err != nil {
		return nil, err
	}

	newCode := &SourceCode{
		Source:   execCommand.Source,
		Language: code.Language,
	}

	for _, extraFile := range problemConf.ExtraFile {
		if _, err := shutil.Copy(filepath.Join(problem, extraFile), filepath.Join(workDir, extraFile), false); err != nil {
			return nil, err
		}
	}

	compilerOutput, err := newCode.Compile(conf, workDir)
	if err != nil && newCode.CompileResult != nil {
		judgeResult.Verdict = "CE"
		judgeResult.Detail = append(judgeResult.Detail, &JudgeDetail{
			Name:       "compile",
			Verdict:    "CE",
			Output:     compilerOutput,
			ExeTime:    uint64(newCode.CompileResult.RealTime * 1000),
			ExeMemory:  newCode.CompileResult.ExeMemory,
			ExitCode:   newCode.CompileResult.ExitCode,
			ExitSignal: newCode.CompileResult.ExitSignal,
		})
		return judgeResult, nil
	} else if err != nil {
		return nil, errors.New("Syetem Error")
	}

	timeLimit := float32(problemConf.TimeLimit) / 1000.
	if timeLimit > 120 {
		timeLimit = 120.
	}
	judgeResult.Verdict = "AC"
	chrootName := GetRandomString()
	if err := func() error {
		chrootCmd := exec.Command("/usr/local/bin/lrun-mirrorfs", "--name", chrootName, "--setup", conf.MirrorFSConfig)
		err := chrootCmd.Run()
		return err
	}(); err != nil {
		return nil, err
	}

	for testId, testInfo := range problemConf.Case {
		logrus.Info("Judge test %d", testId+1)
		resDetail := &JudgeDetail{
			Name:    fmt.Sprintf("Test #%d", testId+1),
			Verdict: "AC",
		}
		if problemConf.Interactor == nil {
			execResult, err := Execute(execCommand.Execute, timeLimit, problemConf.MemoryLimit, codeLanguage.Execute.TimeRatio, filepath.Join(workDir, chrootName), true, filepath.Join(problem, testInfo.Input), "_stdout", "-")
		}
	}
	return judgeResult, nil
}
