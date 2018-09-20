package fj15

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
)

type ExecuteResult struct {
	RealTime   float32 `json:"realtime"`
	CPUTime    float32 `json:"cputime"`
	ExeMemory  uint64  `json:"memory"`
	ExitCode   int32   `json:"exitcode,omitempty"`
	ExitSignal int32   `json:"exitsig,omitempty"`
	TermSignal int32   `json:"termsig,omitempty"`
	ExitReason string  `json:"exceeded"`
}

func Execute(cmd []string, timeLimit float32, memoryLimit uint64, timeRatio float32, chroot string, limitSyscall bool, stdin, stdout, stderr string) (*ExecuteResult, error) {
	cpuTimelimit := timeLimit * timeRatio
	realTimelimit := cpuTimelimit * 1.5
	runCommand := []string{
		"/usr/local/bin/lrun",
		"--max-real-time",
		fmt.Sprintf("%.3f", realTimelimit),
		"--max-cpu-time",
		fmt.Sprintf("%.3f", cpuTimelimit),
		"--max-stack",
		"1073741824",
		"--max-memory",
		strconv.FormatUint(memoryLimit, 10),
		"--network",
		"false",
		"--result-fd",
		"3",
	}
	if limitSyscall {
		runCommand = append(runCommand, []string{
			"--chroot",
			chroot,
			"--remount-dev",
			"true",
			"--chdir",
			"/fj_tmp",
			"--syscalls",
			"!execve,flock,ptrace,sync,fdatasync,fsync,msync,sync_file_range,syncfs,unshare,setns,clone[a&268435456==268435456],query_module,sysinfo,syslog,sysfs",
		}...)
	}
	runCommand = append(runCommand, "--")
	runCommand = append(runCommand, cmd...)
	exe := exec.Command(runCommand[0], runCommand[1:]...)
	if stdin != "-" {
		fp, err := os.Open(stdin)
		if err != nil {
			return nil, err
		}
		exe.Stdin = fp
	}
	if stdout != "-" {
		fp, err := os.OpenFile(stdout, os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			return nil, err
		}
		exe.Stdout = fp
		if stderr == stdout {
			exe.Stderr = exe.Stdout
		}
	}
	if stderr != "-" && stderr != stdout {
		fp, err := os.OpenFile(stderr, os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			return nil, err
		}
		exe.Stderr = fp
	}
	resultYaml, err := ioutil.TempFile("", "runres")
	if err != nil {
		return nil, err
	}
	resultYamlName := resultYaml.Name()
	exe.ExtraFiles = []*os.File{resultYaml}
	err = exe.Run()
	resultYaml.Close()
	if err != nil {
		return nil, err
	}
	executorOutput := &ExecuteResult{}
	err = loadYAML(resultYamlName, executorOutput)
	if err != nil {
		return nil, err
	}
	return executorOutput, nil
}

func ExecuteInteractor(cmd, interactor []string, timeLimit float32, memoryLimit uint64, timeRatio float32, chroot string, limitSyscall bool) (*ExecuteResult, *ExecuteResult, error) {
	cpuTimelimit := timeLimit * timeRatio
	realTimelimit := cpuTimelimit * 1.5
	runCommand := []string{
		"/usr/local/bin/lrun",
		"--max-real-time",
		fmt.Sprintf("%.3f", realTimelimit),
		"--max-cpu-time",
		fmt.Sprintf("%.3f", cpuTimelimit),
		"--max-stack",
		"1073741824",
		"--max-memory",
		strconv.FormatUint(memoryLimit, 10),
		"--network",
		"false",
		"--result-fd",
		"3",
	}
	if limitSyscall {
		runCommand = append(runCommand, []string{
			"--chroot",
			chroot,
			"--remount-dev",
			"true",
			"--chdir",
			"/fj_tmp",
			"--syscalls",
			"!execve,flock,ptrace,sync,fdatasync,fsync,msync,sync_file_range,syncfs,unshare,setns,clone[a&268435456==268435456],query_module,sysinfo,syslog,sysfs",
		}...)
	}
	runCommand = append(runCommand, "--")
	runCommand = append(runCommand, cmd...)
	//------
	interactorCommand := []string{
		"/usr/local/bin/lrun",
		"--max-real-time",
		fmt.Sprintf("%.3f", realTimelimit),
		"--max-cpu-time",
		fmt.Sprintf("%.3f", cpuTimelimit),
		"--max-stack",
		"1073741824",
		"--max-memory",
		strconv.FormatUint(memoryLimit, 10),
		"--network",
		"false",
		"--result-fd",
		"3",
	}
	interactorCommand = append(interactorCommand, "--")
	interactorCommand = append(interactorCommand, cmd...)
	//------
	exeProgram := exec.Command(runCommand[0], runCommand[1:]...)
	exeInteractor := exec.Command(interactorCommand[0], interactorCommand[1:]...)
	//------
	if pr, iw, err := os.Pipe(); err != nil {
		return nil, nil, err
	} else {
		exeProgram.Stdin = pr
		exeInteractor.Stdout = iw
	}

	if pw, ir, err := os.Pipe(); err != nil {
		return nil, nil, err
	} else {
		exeProgram.Stdout = pw
		exeInteractor.Stdin = ir
	}
	///
	programResultYaml, err := ioutil.TempFile("", "runres0")
	if err != nil {
		return nil, nil, err
	}
	programResultYamlName := programResultYaml.Name()
	exeProgram.ExtraFiles = []*os.File{programResultYaml}
	///
	interactorResultYaml, err := ioutil.TempFile("", "runres1")
	if err != nil {
		return nil, nil, err
	}
	interactorResultYamlName := interactorResultYaml.Name()
	exeInteractor.ExtraFiles = []*os.File{interactorResultYaml}
	///
	exeInteractor.Start()
	exeProgram.Start()
	///
	exeInteractor.Wait()
	exeProgram.Wait()
	///
	interactorResultYaml.Close()
	programResultYaml.Close()
	if err != nil {
		return nil, nil, err
	}
	executorOutput := &ExecuteResult{}
	if err := loadYAML(programResultYamlName, executorOutput); err != nil {
		return nil, nil, err
	}
	interactorOutput := &ExecuteResult{}
	if err := loadYAML(interactorResultYamlName, interactorOutput); err != nil {
		return nil, nil, err
	}
	return executorOutput, interactorOutput, nil
}
