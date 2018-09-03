package executor

import "sync"

type ExecuteResult struct {
	ExeTime    int    `json:"exe_time"`
	ExeMemory  int64  `json:"exe_memory"`
	CPUTime    int    `json:"cpu_time"`
	RealTime   int    `json:"real_time"`
	ExitCode   int    `json:"exit_code"`
	ExitSignal int    `json:"exit_signal"`
	ExitReason string `json:"exit_reason"`
}

type BuildResult struct {
	BuildTime   int    `json:"build_time"`
	BuildMemory int64  `json:"build_memory"`
	BuildOutput string `json:"build_output"`
	Success     bool   `json:"success"`
}

type CGroupConfig struct {
	Memory *int64
	CPU    *int64
}

// CGroup defines a control group
// Name: the name of the cgroup
// Config currently exectime and execmemory
// Chroot change root, create overlayfs on path and then change root into it
// if Chroot = "/" or "", then will not change root
type CGroup struct {
	Name   string
	Config CGroupConfig
	Mutex  sync.Mutex
	Chroot string
	state  int
}

func (g *CGroup) setState(s int) {
	g.state = s
}

func (g *CGroup) getState() int {
	return g.state
}
