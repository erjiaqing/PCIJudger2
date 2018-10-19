package pci15

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/sirupsen/logrus"
	shutil "github.com/termie/go-shutil"
)

type JudgeResult struct {
	Success     bool           `json:"success"`
	Verdict     string         `json:"verdict"`
	ExeTime     uint64         `json:"exe_time"`
	ExeMemory   uint64         `json:"exe_memory"`
	ExitCode    int32          `json:"exit_code"`
	UsedTime    uint64         `json:"used_time"`
	Detail      []*JudgeDetail `json:"detail"`
	lastTest    int
	judgeResult map[int]*JudgeDetail
}

type JudgeDetail struct {
	Name       string `json:"name"`
	Input      string `json:"input,omitempty"`
	Output     string `json:"output,omitempty"`
	Answer     string `json:"answer,omitempty"`
	Comment    string `json:"comment,omitempty"`
	Verdict    string `json:"verdict"`
	ExeTime    uint64 `json:"exe_time"`
	ExeMemory  uint64 `json:"exe_memory"`
	ExitCode   int32  `json:"exit_code"`
	ExitSignal int32  `json:"exit_signal"`
}

type JudgeRequest struct {
	Id   int
	Case TestCase
}

func (j *JudgeResult) Append(testCase int, detail *JudgeDetail) int {
	j.judgeResult[testCase] = detail
	if testCase > j.lastTest {
		for {
			if _, ok := j.judgeResult[j.lastTest+1]; ok {
				j.lastTest++
			} else {
				break
			}
		}
	}
	return j.lastTest
}

func (j *JudgeResult) Collect(testCase int) {
	for i := 0; i < testCase; i++ {
		if val, ok := j.judgeResult[i]; ok {
			j.Detail = append(j.Detail, val)
			if val.ExeTime > j.ExeTime {
				j.ExeTime = val.ExeTime
			}
			if val.ExeMemory > j.ExeMemory {
				j.ExeMemory = val.ExeMemory
			}
			if val.Verdict != "AC" {
				j.Verdict = val.Verdict
				break
			}
		}
	}
}

func doJudge(testId int, testInfo TestCase, problemConf *ProblemConfig, execCommand *ExecuteCommand, timeLimit float32, codeLanguage *Language, chrootName, problem string, checkerCmd, interCmd []string) (*JudgeDetail, bool) {
	logrus.Infof("Judging test %d", testId+1)

	judgeUid := GetRandomString()

	resDetail := &JudgeDetail{
		Name:    fmt.Sprintf("Test #%d", testId+1),
		Verdict: "AC",
	}

	var execResult, interactorResult *ExecuteResult
	var err error
	if problemConf.Interactor == nil {
		execResult, err = Execute(execCommand.Execute, timeLimit, problemConf.MemoryLimit*1024*1024, codeLanguage.Execute.TimeRatio, filepath.Join("/run", chrootName), true, filepath.Join(problem, testInfo.Input), judgeUid+".stdout", "-")
		if err != nil {
			resDetail.Verdict = "SE"
			resDetail.Comment = fmt.Sprintf("Failed to execute code: %v", err)
			return resDetail, false
		}
	} else {
		execResult, interactorResult, err = ExecuteInteractor(execCommand.Execute, append(interCmd, filepath.Join(problem, testInfo.Input), judgeUid+".stdout", filepath.Join(problem, testInfo.Output)), timeLimit, problemConf.MemoryLimit*1024*1024, codeLanguage.Execute.TimeRatio, filepath.Join("/run", chrootName), true)
		if err != nil {
			resDetail.Verdict = "SE"
			resDetail.Comment = fmt.Sprintf("Failed to execute code: %v", err)
			return resDetail, false
		}
		if execResult.ExitReason == "none" {
			if interactorResult.ExitReason != "none" {
				execResult.ExitReason = "WA"
			} else if interactorResult.ExitCode != 0 || interactorResult.ExitSignal != 0 || interactorResult.TermSignal != 0 {
				execResult.ExitReason = "WA"
			}
		}
	}

	resDetail.ExeTime = uint64(1000 * execResult.CPUTime)
	resDetail.ExeMemory = execResult.ExeMemory

	resDetail.Input, _ = ReadFirstBytes(filepath.Join(problem, testInfo.Input), 128)
	resDetail.Output, _ = ReadFirstBytes(judgeUid+".stdout", 128)
	resDetail.Answer, _ = ReadFirstBytes(filepath.Join(problem, testInfo.Output), 128)

	if execResult.ExitReason != "none" {
		resDetail.Verdict = execResult.ExitReason
		return resDetail, false
	} else if execResult.ExitCode != 0 || execResult.ExitSignal != 0 || execResult.TermSignal != 0 {
		resDetail.Verdict = "RE"
		return resDetail, false
	}

	tcheckerCmd := append(checkerCmd, filepath.Join(problem, testInfo.Input), judgeUid+".stdout", filepath.Join(problem, testInfo.Output))

	checkerResult, err := Execute(tcheckerCmd, 10., problemConf.MemoryLimit*1024*1024, 1., "", false, "-", judgeUid+".checker.stderr", judgeUid+".checker.stderr")
	resDetail.Comment, _ = ReadFirstBytes(judgeUid+".checker.stderr", 128)

	if err != nil {
		resDetail.Verdict = "SE"
		resDetail.Comment = fmt.Sprintf("Failed to run checker: %v", err)
		return resDetail, false
	} else if checkerResult.ExitCode != 0 {
		resDetail.Verdict = "WA"
		return resDetail, false
	}
	return resDetail, true
}

func Judge(conf *Config, code *SourceCode, problem string) (*JudgeResult, error) {
	judgeResult := &JudgeResult{
		Success:     true,
		judgeResult: make(map[int]*JudgeDetail),
	}

	problemConf := &ProblemConfig{}
	if err := loadYAML(filepath.Join(problem, "problem.yaml"), problemConf); err != nil {
		return nil, err
	}

	workDir := filepath.Join(conf.Tmp, GetRandomString())
	if err := os.MkdirAll(workDir, 0777); err != nil {
		return nil, err
	}
	//defer os.RemoveAll(workDir)

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

	conf.HostSocket.SendStatus("00", 0)

	for _, extraFile := range problemConf.ExtraFile {
		if _, err := shutil.Copy(filepath.Join(problem, extraFile), filepath.Join(workDir, extraFile), false); err != nil {
			return nil, err
		}
	}

	conf.HostSocket.SendStatus("01", 0)

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

	conf.HostSocket.SendStatus("02", 0)

	timeLimit := float32(problemConf.TimeLimit) / 1000.
	if timeLimit > 120 {
		timeLimit = 120.
	}
	judgeResult.Verdict = "AC"
	chrootName := GetRandomString()
	if err := func() error {
		logrus.Infof("Setting up mirrorfs: /run/%s ...", chrootName)
		chrootCmd := exec.Command("/usr/local/bin/lrun-mirrorfs", "--name", chrootName, "--setup", conf.MirrorFSConfig)
		return chrootCmd.Run()
	}(); err != nil {
		return nil, err
	} else {
		defer func() error {
			logrus.Infof("Tearing down mirrorfs: /run/%s ...", chrootName)
			chrootCmd := exec.Command("/usr/local/bin/lrun-mirrorfs", "--name", chrootName, "--teardown", conf.MirrorFSConfig)
			return chrootCmd.Run()
		}()
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

	judgeChan := make(chan *JudgeRequest, len(problemConf.Case))
	countTestCase := len(problemConf.Case)

	for testId, testInfo := range problemConf.Case {
		judgeChan <- &JudgeRequest{
			Id:   testId,
			Case: testInfo,
		}
	}
	close(judgeChan)

	endOfJudge := false

	var wg sync.WaitGroup

	mut := &sync.Mutex{}

	for i := 0; i < conf.MaxJudgeThread; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				mut.Lock()
				val, ok := <-judgeChan
				eoj := endOfJudge
				mut.Unlock()
				if !ok {
					break
				}
				if eoj {
					continue
				}
				detail, cont := doJudge(val.Id, val.Case, problemConf, execCommand, timeLimit, codeLanguage, chrootName, problem, checkerCmd, interCmd)

				mut.Lock()
				judged := judgeResult.Append(val.Id, detail) + 1

				conf.HostSocket.SendStatus("10", 100*judged/countTestCase)
				logrus.Infof("%d / %d Test Judged", judged, countTestCase)
				if !cont {
					endOfJudge = true
				}
				mut.Unlock()
			}
		}()
	}
	wg.Wait()

	judgeResult.Collect(countTestCase)
	return judgeResult, nil
}
