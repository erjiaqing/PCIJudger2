package pci15

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/erjiaqing/PCIJudger2/pkg/builtin_cmp"
	"github.com/sirupsen/logrus"
	shutil "github.com/termie/go-shutil"
)

type JudgeResult struct {
	Success     bool           `json:"success"`
	Verdict     string         `json:"verdict"`
	ExeTime     float32        `json:"exe_time"`
	ExeMemory   uint64         `json:"exe_memory"`
	ExitCode    int32          `json:"exit_code"`
	UsedTime    uint64         `json:"used_time"`
	Score       int            `json:"score"`
	FullScore   int            `json:"full_score"`
	Detail      []*JudgeDetail `json:"detail"`
	lastTest    int
	judgeResult map[int]*JudgeDetail
	judgeState  *sync.Map
}

type JudgeDetail struct {
	Name       string  `json:"name"`
	Input      string  `json:"input,omitempty"`
	Output     string  `json:"output,omitempty"`
	Answer     string  `json:"answer,omitempty"`
	Comment    string  `json:"comment,omitempty"`
	Score      int     `json:"score"`
	Verdict    string  `json:"verdict"`
	ExeTime    float32 `json:"exe_time"`
	ExeMemory  uint64  `json:"exe_memory"`
	ExitCode   int32   `json:"exit_code"`
	ExitSignal int32   `json:"exit_signal"`
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

func (j *JudgeResult) Collect(problemConf *ProblemConfig, testCase int) {
	for i := 0; i < testCase; i++ {
		if val, ok := j.judgeResult[i]; ok {
			shouldJudge := true
			for _, dep := range problemConf.Case[i].Dependencies {
				if !j.checkPass(dep) {
					shouldJudge = false
				}
			}

			if shouldJudge {
				j.Detail = append(j.Detail, val)
			} else {
				j.Detail = append(j.Detail, &JudgeDetail{
					Verdict: "IG",
					Name:    val.Name,
				})
			}
			j.judgeState.Store(problemConf.Case[i].Input, j.Detail[i].Verdict)

			if val.ExeTime > j.ExeTime {
				j.ExeTime = val.ExeTime
			}
			if val.ExeMemory > j.ExeMemory {
				j.ExeMemory = val.ExeMemory
			}

			j.Score += val.Score

			if j.Verdict == "AC" && val.Verdict != "AC" {
				j.Verdict = val.Verdict
			}
		}
	}
}

func (j *JudgeResult) checkPass(t string) bool {
	status, ok := j.judgeState.Load(t)
	if !ok {
		return false
	}
	statusStr := status.(string)
	return statusStr == "AC" || statusStr == "NJ"
}

func (j *JudgeResult) prepareProblemConf(problemConf *ProblemConfig) error {
	for i, _ := range problemConf.Case {
		if len(problemConf.Case[i].Dependencies) == 0 && i > 0 {
			problemConf.Case[i].Dependencies = append(problemConf.Case[i].Dependencies, problemConf.Case[i-1].Input)
		} else {
			for _, p := range problemConf.Case[i].Dependencies {
				if _, ok := j.judgeState.Load(p); !ok {
					return fmt.Errorf("Input ``%s'' appears not before test case ``%s'', cycle dependencies might happen, use checkpoint instead.", p, problemConf.Case[i].Input)
				}
			}
		}
		if _, ok := j.judgeState.Load(problemConf.Case[i].Input); ok {
			return fmt.Errorf("Input ``%s'' appears more than once, which is no sense.", problemConf.Case[i].Input)
		}
		j.judgeState.Store(problemConf.Case[i].Input, "NJ")
	}
	return nil
}

func fastMode(problemPath string, problemConf *ProblemConfig) error {
	outputs := make(map[string]string)
	inputList := make([]string, 0)
	err := filepath.Walk(problemPath, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		if path2, err := filepath.Rel(problemPath, path); err == nil {
			path = path2
		}
		if strings.HasSuffix(path, ".in") {
			baseName := path[0 : len(path)-3]
			inputList = append(inputList, baseName)
		} else if strings.HasSuffix(path, ".out") || strings.HasSuffix(path, ".ans") {
			baseName := path[0 : len(path)-4]
			outputs[baseName] = path
		}
		return nil
	})
	sort.Strings(inputList)
	if err != nil {
		return err
	}
	for _, input := range inputList {
		if output, ok := outputs[input]; ok {
			problemConf.Case = append(problemConf.Case, TestCase{
				Input:  input + ".in",
				Output: output,
				Score:  1,
			})
		}
	}
	return nil
}

func (j *JudgeResult) doJudge(testId int, testInfo TestCase, problemConf *ProblemConfig, execCommand *ExecuteCommand, timeLimit float32, codeLanguage *Language, chrootName, workdir, problem string, checkerCmd, interCmd []string) (*JudgeDetail, bool) {
	logrus.Infof("Judging test %d", testId+1)

	judgeUid := GetRandomString()
	checkPoint := false

	input := testInfo.Input

	resDetail := &JudgeDetail{
		Name:    fmt.Sprintf("#%d (Test)", testId+1),
		Verdict: "AC",
	}

	if input[0] == '*' {
		checkPoint = true
		resDetail.Name = fmt.Sprintf("#%d (Checkpoint)", testId+1)
		resDetail.Verdict = "AC"
		resDetail.Score = testInfo.Score
	}

	for _, dep := range testInfo.Dependencies {
		if !j.checkPass(dep) {
			resDetail.Verdict = "IG"
			resDetail.Score = 0
			return resDetail, false
		}
	}

	if checkPoint {
		return resDetail, true
	}

	var execResult, interactorResult *ExecuteResult
	var err error
	if problemConf.Interactor == nil {
		execResult, err = Execute(execCommand.Execute, timeLimit, problemConf.MemoryLimit*1024*1024, codeLanguage.Execute.TimeRatio, filepath.Join("/fj_tmp/mirrorfs", chrootName), workdir, true, filepath.Join(problem, testInfo.Input), judgeUid+".stdout", judgeUid+".stderr")
		if err != nil {
			resDetail.Verdict = "SE"
			resDetail.Comment = fmt.Sprintf("Failed to execute code: %v", err)
			stderr, _ := ReadFirstBytes(judgeUid+".stderr", 1024)
			logrus.Errorf("Datail: %s", stderr)
			return resDetail, false
		}
	} else {
		execResult, interactorResult, err = ExecuteInteractor(execCommand.Execute, append(interCmd, filepath.Join(problem, testInfo.Input), judgeUid+".stdout", filepath.Join(problem, testInfo.Output)), timeLimit, problemConf.MemoryLimit*1024*1024, codeLanguage.Execute.TimeRatio, filepath.Join("/fj_tmp/mirrorfs", chrootName), workdir, true)
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

	resDetail.ExeTime = execResult.CPUTime
	resDetail.ExeMemory = execResult.ExeMemory / 1024

	resDetail.Input, _ = ReadFirstBytes(filepath.Join(problem, testInfo.Input), 128)
	resDetail.Output, _ = ReadFirstBytes(judgeUid+".stdout", 128)
	resDetail.Answer, _ = ReadFirstBytes(filepath.Join(problem, testInfo.Output), 128)

	if execResult.ExitReason != "none" {
		resDetail.Verdict = execResult.ExitReason
		return resDetail, false
	} else if execResult.ExitCode != 0 || execResult.ExitSignal != 0 || execResult.TermSignal != 0 {
		resDetail.ExitCode = execResult.ExitCode
		resDetail.ExitSignal = execResult.ExitSignal
		if execResult.ExitSignal == 0 {
			resDetail.ExitSignal = -execResult.TermSignal
		}
		resDetail.Verdict = "RE"
		return resDetail, false
	}

	if problemConf.Checker.Source[0] == '!' {
		checkerRes, err := builtin_cmp.Diff[problemConf.Checker.Source](judgeUid+".stdout", filepath.Join(problem, testInfo.Output))
		if err != nil {
			resDetail.Verdict = "SE"
			resDetail.Comment = err.Error()
		} else if !checkerRes {
			resDetail.Verdict = "WA"
		} else {
			resDetail.Score = testInfo.Score
		}
		return resDetail, checkerRes
	}

	tcheckerCmd := append(checkerCmd, filepath.Join(problem, testInfo.Input), judgeUid+".stdout", filepath.Join(problem, testInfo.Output))

	checkerResult, err := Execute(tcheckerCmd, 10., problemConf.MemoryLimit*1024*1024, 1., "", "-", false, "-", judgeUid+".checker.stderr", judgeUid+".checker.stderr")
	resDetail.Comment, _ = ReadFirstBytes(judgeUid+".checker.stderr", 128)

	if err != nil {
		resDetail.Verdict = "SE"
		resDetail.Comment = fmt.Sprintf("Failed to run checker: %v", err)
		return resDetail, false
	} else if checkerResult.ExitCode != 0 {
		resDetail.Verdict = "WA"
		return resDetail, false
	}

	resDetail.Score = testInfo.Score
	return resDetail, true
}

func Judge(conf *Config, code *SourceCode, problem string) (*JudgeResult, error) {
	judgeResult := &JudgeResult{
		Success:     true,
		judgeResult: make(map[int]*JudgeDetail),
		judgeState:  &sync.Map{},
	}

	problemConf := &ProblemConfig{}

	if err := loadYAML(filepath.Join(problem, "problem.yaml"), problemConf); err != nil {
		logrus.Warningf("Failed to find problem.yaml, enter fast mode...")
		problemConf.TimeLimit = 1000
		problemConf.MemoryLimit = 512
		if err := fastMode(problem, problemConf); err != nil {
			logrus.Errorf("Failed to generate fast mode test cases: %v", err)
			return nil, err
		}
	}

	judgeResult.prepareProblemConf(problemConf)

	if problemConf.TimeLimit == 0 {
		problemConf.TimeLimit = problemConf.TimeLimitBK
	}

	if problemConf.TimeLimit == 0 {
		problemConf.TimeLimit = 1000
	}

	if problemConf.Checker == nil {
		problemConf.Checker = &SourceCode{}
	}

	if problemConf.Checker.Source == "" {
		problemConf.Checker.Source = "!diff"
	}

	workDir := filepath.Join(conf.Tmp, GetRandomString())

	if !conf.IsDocker {
		if err := os.MkdirAll(workDir, 0777); err != nil {
			return nil, err
		}
	} else {
		workDir = conf.Tmp
	}
	//defer os.RemoveAll(workDir)

	header := make([]byte, 0)
	footer := make([]byte, 0)
	codeBin := make([]byte, 0)

	if problemConf.Template != "" {
		header, _ = ioutil.ReadFile(filepath.Join(problem, problemConf.Template+".header."+code.Language))
		footer, _ = ioutil.ReadFile(filepath.Join(problem, problemConf.Template+".footer."+code.Language))
	}

	codeBin, _ = ioutil.ReadFile(code.Source)

	execCommand, codeLanguage, err := GetExecuteCommand2(code, conf, workDir, true)
	if err != nil {
		return nil, err
	}

	if err := func() error {
		fp, err := os.OpenFile(execCommand.Source, os.O_CREATE|os.O_TRUNC|os.O_RDWR, os.ModePerm)
		if err != nil {
			return err
		}
		defer fp.Close()
		if _, err := fp.Write(header); err != nil {
			return err
		}
		if _, err := fp.Write(codeBin); err != nil {
			return err
		}
		if _, err := fp.Write(footer); err != nil {
			return err
		}
		return nil
	}(); err != nil {
		return nil, fmt.Errorf("failed to copy source: %v", err)
	} else {
		code.Source = execCommand.Source
	}

	if err := func() error {
		fp, err := os.OpenFile(filepath.Join(workDir, "mirrorfs.conf"), os.O_CREATE|os.O_TRUNC|os.O_RDWR, os.ModePerm)
		if err != nil {
			return err
		}
		defer fp.Close()
		mirrorfs := make([]byte, 0)
		mirrorfs, _ = ioutil.ReadFile(conf.MirrorFSConfig)
		fp.Write(mirrorfs)
		//fp.Write([]byte(fmt.Sprintf("mirror %s", workDir)))
		return nil
	}(); err != nil {
		return nil, fmt.Errorf("failed to copy source: %v", err)
	}

	logrus.Infof("Code source name: %s", code.Source)

	if currentDir, err := os.Getwd(); err != nil {
		return nil, err
	} else if err := os.Chdir(workDir); err != nil {
		return nil, err
	} else {
		defer os.Chdir(currentDir)
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

	compilerOutput, err := newCode.Compile2(conf, workDir, true)
	if err != nil && newCode.CompileResult != nil {
		judgeResult.Verdict = "CE"
		judgeResult.Detail = append(judgeResult.Detail, &JudgeDetail{
			Name:       "compile",
			Verdict:    "CE",
			Output:     compilerOutput,
			ExeTime:    newCode.CompileResult.RealTime,
			ExeMemory:  newCode.CompileResult.ExeMemory,
			ExitCode:   newCode.CompileResult.ExitCode,
			ExitSignal: newCode.CompileResult.ExitSignal,
		})
		return judgeResult, nil
	} else if err != nil {
		judgeResult.Verdict = "CE"
		judgeResult.Detail = append(judgeResult.Detail, &JudgeDetail{
			Name:       "compile",
			Verdict:    "CE",
			Output:     err.Error(),
			ExeTime:    0,
			ExeMemory:  0,
			ExitCode:   0,
			ExitSignal: 0,
		})
		return judgeResult, nil
	}

	conf.HostSocket.SendStatus("02", 0)

	timeLimit := float32(problemConf.TimeLimit) / 1000.
	if timeLimit > 120 {
		timeLimit = 120.
	}
	judgeResult.Verdict = "AC"
	chrootName := GetRandomString()
	if err := func() error {
		logrus.Infof("Setting up mirrorfs: /fj_tmp/mirrorfs/%s ...", chrootName)
		chrootCmd := exec.Command("/usr/local/bin/lrun-mirrorfs", "--name", chrootName, "--setup", conf.MirrorFSConfig)
		return chrootCmd.Run()
	}(); err != nil {
		return nil, err
	} else {
		defer func() error {
			logrus.Infof("Tearing down mirrorfs: /fj_tmp/mirrorfs/%s ...", chrootName)
			chrootCmd := exec.Command("/usr/local/bin/lrun-mirrorfs", "--name", chrootName, "--teardown", conf.MirrorFSConfig)
			return chrootCmd.Run()
		}()
	}

	checkerCmd := []string{filepath.Join(problem, problemConf.Checker.Executable)}
	if problemConf.Checker.Executable == "" {
		problemConf.Checker.Source = filepath.Join(problem, problemConf.Checker.Source)
		checkerExec, _, err := GetExecuteCommand(problemConf.Checker, conf)
		if err != nil {
			logrus.Warnf("Failed to get checker Exec: %v", err)
			checkerCmd = []string{filepath.Join(problem, problemConf.Checker.Source+".exe")}
		} else {
			checkerCmd = checkerExec.Execute
		}
	}

	interCmd := []string{}
	if problemConf.Interactor != nil {
		interCmd = []string{filepath.Join(problem, problemConf.Interactor.Executable)}
		if problemConf.Interactor.Executable == "" {
			problemConf.Interactor.Source = filepath.Join(problem, problemConf.Interactor.Source)
			interExec, _, err := GetExecuteCommand(problemConf.Interactor, conf)
			if err != nil {
				logrus.Warnf("Failed to get interactor Exec: %v", err)
				interCmd = []string{filepath.Join(problem, problemConf.Interactor.Source+".exe")}
			} else {
				interCmd = interExec.Execute
			}
		}
	}

	judgeChan := make(chan *JudgeRequest, len(problemConf.Case))
	countTestCase := len(problemConf.Case)

	for testId, testInfo := range problemConf.Case {
		judgeResult.FullScore += testInfo.Score
		judgeChan <- &JudgeRequest{
			Id:   testId,
			Case: testInfo,
		}
	}
	close(judgeChan)

	var wg sync.WaitGroup

	mut := &sync.Mutex{}

	for i := 0; i < conf.MaxJudgeThread; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				mut.Lock()
				val, ok := <-judgeChan
				mut.Unlock()
				if !ok {
					break
				}

				detail, _ := judgeResult.doJudge(val.Id, val.Case, problemConf, execCommand, timeLimit, codeLanguage, chrootName, workDir, problem, checkerCmd, interCmd)

				mut.Lock()

				judgeResult.judgeState.Store(val.Case.Input, detail.Verdict)
				judged := judgeResult.Append(val.Id, detail) + 1

				conf.HostSocket.SendStatus("10", 100*judged/countTestCase)
				logrus.Infof("%d / %d Test Judged", judged, countTestCase)
				mut.Unlock()
			}
		}()
	}
	wg.Wait()

	judgeResult.Collect(problemConf, countTestCase)

	if newCode.CompileResult != nil {
		judgeResult.Detail = append(judgeResult.Detail, &JudgeDetail{
			Name:       "compile",
			Verdict:    "AC",
			Output:     compilerOutput,
			ExeTime:    newCode.CompileResult.RealTime,
			ExeMemory:  newCode.CompileResult.ExeMemory,
			ExitCode:   newCode.CompileResult.ExitCode,
			ExitSignal: newCode.CompileResult.ExitSignal,
		})
	}
	return judgeResult, nil
}
