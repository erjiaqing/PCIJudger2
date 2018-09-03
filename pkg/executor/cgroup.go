package executor

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"syscall"

	"github.com/erjiaqing/problem-ci-judger-2/pkg/util"
	"github.com/sirupsen/logrus"
)

func init() {
	if _, err := os.Stat(path.Join("/sys", "fs", "cgroup", "cpu,cpuacct", "FinalJudger2")); os.IsNotExist(err) {
		os.Mkdir(path.Join("/sys", "fs", "cgroup", "cpu,cpuacct", "FinalJudger2"), 0755)
	}
	if _, err := os.Stat(path.Join("/sys", "fs", "cgroup", "memory", "FinalJudger2")); os.IsNotExist(err) {
		os.Mkdir(path.Join("/sys", "fs", "cgroup", "memory", "FinalJudger2"), 0755)
	}
}

// NewCGroup Create a new cgroup, return with the CGroup Object
// WARNING: remember to call defer CGroup.CleanUp
func NewCGroup() (*CGroup, error) {
	ret := &CGroup{
		Name: util.RandSeq(16),
	}
	ret.setState(0)
	// create new CGroup
	if _, err := os.Stat(path.Join("/sys", "fs", "cgroup", "cpu,cpuacct", "FinalJudger2", ret.Name)); os.IsNotExist(err) {
		os.Mkdir(path.Join("/sys", "fs", "cgroup", "cpu,cpuacct", "FinalJudger2", ret.Name), 0755)
	}
	if _, err := os.Stat(path.Join("/sys", "fs", "cgroup", "memory", "FinalJudger2", ret.Name)); os.IsNotExist(err) {
		os.Mkdir(path.Join("/sys", "fs", "cgroup", "memory", "FinalJudger2", ret.Name), 0755)
	}
	return ret, nil
}

// UpdateMemoryLimit set/overwrites the memory limit of a cgroup
func (c *CGroup) UpdateMemoryLimit(memMiB int64) error {
	// write memory.limit_in_bytes
	c.Mutex.Lock()
	defer c.Mutex.Unlock()
	err := ioutil.WriteFile(
		path.Join("/sys", "fs", "cgroup", "memory", "FinalJudger2", c.Name, "memory.limit_in_bytes"),
		[]byte(fmt.Sprintf("%d", memMiB*1024*1024)),
		0644,
	)
	if err != nil {
		return fmt.Errorf("Failed to set memory limit: %v", err)
	}
	return nil
}

// UpdateCPULimit set/overwrites the cpu core limit of a cgroup
func (c *CGroup) UpdateCPULimit(cpuCore int) error {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()
	err := ioutil.WriteFile(
		path.Join("/sys", "fs", "cgroup", "cpu,cpuacct", "FinalJudger2", c.Name, "cpu.cfs_period_us"),
		[]byte(fmt.Sprintf("%d", 100000)),
		0644,
	)
	if err != nil {
		return fmt.Errorf("Failed to set cpu core limit: %v", err)
	}
	err = ioutil.WriteFile(
		path.Join("/sys", "fs", "cgroup", "cpu,cpuacct", "FinalJudger2", c.Name, "cpu.cfs_period_us"),
		[]byte(fmt.Sprintf("%d", cpuCore*100000)),
		0644,
	)
	if err != nil {
		return fmt.Errorf("Failed to set cpu core limit: %v", err)
	}
	return nil
}

func (c *CGroup) AddProcess(pid int) error {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()
	fcpu, err := os.OpenFile(path.Join("/sys", "fs", "cgroup", "cpu,cpuacct", "FinalJudger2", c.Name, "cgroup.procs"), os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("Failed to add process in CPU limit cgroup: %v", err)
	}
	defer fcpu.Close()
	if _, err = fcpu.WriteString(fmt.Sprintln(pid)); err != nil {
		return fmt.Errorf("Failed to add process in CPU limit cgroup: %v", err)
	}
	//--
	fmem, err := os.OpenFile(path.Join("/sys", "fs", "cgroup", "memory", "FinalJudger2", c.Name, "cgroup.procs"), os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("Failed to add process in memory limit cgroup: %v", err)
	}
	defer fmem.Close()
	if _, err = fmem.WriteString(fmt.Sprintln(pid)); err != nil {
		return fmt.Errorf("Failed to add process in memory limit cgroup: %v", err)
	}
	return nil
}

// CleanUp cleans up the cgroup created
func (c *CGroup) CleanUp() error {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()
	// kill all processes
	content, err := ioutil.ReadFile(path.Join("/sys", "fs", "cgroup", "cpu,cpuacct", "FinalJudger2", c.Name, "cgroup.procs"))
	if err != nil {
		return fmt.Errorf("Failed to read processes in cgroup: %v", err)
	}
	stringContent := strings.Split(string(content), "\n")
	for _, proc := range stringContent {
		pid, err := strconv.Atoi(proc)
		if err != nil {
			continue
		}
		err = syscall.Kill(-pid, syscall.SIGKILL)
		if err != nil {
			logrus.Warningf("Failed to kill process: %v", err)
		}
	}
	// kill by nemory
	content, err = ioutil.ReadFile(path.Join("/sys", "fs", "cgroup", "memory", "FinalJudger2", c.Name, "cgroup.procs"))
	if err != nil {
		return fmt.Errorf("Failed to read processes in cgroup: %v", err)
	}
	stringContent = strings.Split(string(content), "\n")
	for _, proc := range stringContent {
		pid, err := strconv.Atoi(proc)
		if err != nil {
			continue
		}
		err = syscall.Kill(-pid, syscall.SIGKILL)
		if err != nil {
			logrus.Warningf("Failed to kill process: %v", err)
		}
	}
	// remove directories
	return nil
}

// Exec returns a Cmd object for Run
func (c *CGroup) Exec(name string, command ...string) *exec.Cmd {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()
	cmd := exec.Command(name, command...)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	return cmd
}

// Run starts a command and add it to process list
func (c *CGroup) Run(cmd *exec.Cmd) error {
	err := cmd.Start()
	if err != nil {
		return err
	}
	return c.AddProcess(cmd.Process.Pid)
}
