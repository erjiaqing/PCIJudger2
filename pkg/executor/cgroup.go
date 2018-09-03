package executor

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/erjiaqing/problem-ci-judger-2/pkg/util"
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
		return fmt.Errorf("Failed to add process in CPU limit cgroup: %v", err)
	}
	defer fmem.Close()
	if _, err = fmem.WriteString(fmt.Sprintln(pid)); err != nil {
		return fmt.Errorf("Failed to add process in memory limit cgroup: %v", err)
	}
	return nil
}

// CleanUp cleans up the cgroup created
func (*CGroup) CleanUp() error {
	// kill all processes
	// remove directories
	return nil
}
