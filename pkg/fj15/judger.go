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
	Comment    string `json:"comment"`
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

	if source, err := shutil.Copy(code.Source, filepath.Join(workDir, filepath.Base(code.Source)), false); err != nil {
		return nil, fmt.Errorf("failed to copy source: %v", err)
	} else {
		code.Source = source
	}

	if currentDir, err := os.Getwd(); err != nil {
		return nil, err
	} else if err := os.Chdir(workDir); err != nil {
		return nil, err
	} else {
		defer os.Chdir(currentDir)
	}

	execCommand, codeLanguage, err := GetExecuteCommand(code, conf)
	if err != nil {
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

	checkerCmd := []string{filepath.Join(problem, problemConf.Checker.Executable)}
	if problemConf.Checker.Executable == "" {
		checkerCmd = []string{filepath.Join(problem, problemConf.Checker.Source+".exe")}
	}

	interCmd := []string{}
	if problemConf.Interactor != nil {
		interCmd = []string{filepath.Join(problem, problemConf.Interactor.Executable)}
		if problemConf.Interactor.Executable == "" {
			interCmd = []string{filepath.Join(problem, problemConf.Interactor.Source+".exe")}
		}
	}

	for testId, testInfo := range problemConf.Case {
		logrus.Infof("Judging test %d", testId+1)
		resDetail := &JudgeDetail{
			Name:    fmt.Sprintf("Test #%d", testId+1),
			Verdict: "AC",
		}
		judgeResult.Detail = append(judgeResult.Detail, resDetail)

		var execResult, interactorResult *ExecuteResult
		if problemConf.Interactor == nil {
			execResult, err = Execute(execCommand.Execute, timeLimit, problemConf.MemoryLimit*1024*1024, codeLanguage.Execute.TimeRatio, filepath.Join(workDir, chrootName), false, filepath.Join(problem, testInfo.Input), "_stdout", "-")
			if err != nil {
				resDetail.Verdict = "SE"
				resDetail.Comment = fmt.Sprintf("Failed to execute code: %v", err)
				break
			}
		} else {
			execResult, interactorResult, err = ExecuteInteractor(execCommand.Execute, append(interCmd, filepath.Join(problem, testInfo.Input), "_stdout", filepath.Join(problem, testInfo.Output)), timeLimit, problemConf.MemoryLimit*1024*1024, codeLanguage.Execute.TimeRatio, filepath.Join(workDir, chrootName), false)
			if err != nil {
				resDetail.Verdict = "SE"
				resDetail.Comment = fmt.Sprintf("Failed to execute code: %v", err)
				break
			}
			if interactorResult.ExitReason != "none" {
				execResult.ExitReason = "WA"
			} else if interactorResult.ExitCode != 0 {
				execResult.ExitReason = "WA"
			}
		}
		resDetail.ExeTime = uint64(1000 * execResult.CPUTime)
		resDetail.ExeMemory = execResult.ExeMemory
		if resDetail.ExeTime > judgeResult.ExeTime {
			judgeResult.ExeTime = resDetail.ExeTime
		}
		if execResult.ExeMemory > judgeResult.ExeMemory {
			judgeResult.ExeMemory = execResult.ExeMemory
		}

		resDetail.Input, _ = ReadFirstBytes(filepath.Join(problem, testInfo.Input), 128)
		resDetail.Output, _ = ReadFirstBytes("_stdout", 128)
		resDetail.Answer, _ = ReadFirstBytes(filepath.Join(problem, testInfo.Output), 128)

		if execResult.ExitReason != "none" {
			resDetail.Verdict = execResult.ExitReason
			judgeResult.Verdict = execResult.ExitReason
			break
		} else if execResult.ExitCode != 0 || execResult.ExitSignal != 0 || execResult.TermSignal != 0 {
			resDetail.Verdict = "RE"
			judgeResult.Verdict = "RE"
			break
		}

		tcheckerCmd := append(checkerCmd, filepath.Join(problem, testInfo.Input), "_stdout", filepath.Join(problem, testInfo.Output))

		checkerResult, err := Execute(tcheckerCmd, 10., problemConf.MemoryLimit*1024*1024, 1., "", false, "-", "checker.stderr", "checker.stderr")
		resDetail.Comment, _ = ReadFirstBytes("checker.stderr", 128)

		if err != nil {
			resDetail.Verdict = "SE"
			judgeResult.Verdict = "SE"
			resDetail.Comment = fmt.Sprintf("Failed to run checker: %v", err)
			break
		} else if checkerResult.ExitCode != 0 {
			resDetail.Verdict = "WA"
			judgeResult.Verdict = "WA"
			break
		}
	}
	return judgeResult, nil
}
