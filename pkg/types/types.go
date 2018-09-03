package types

import (
	"strings"

	"github.com/sirupsen/logrus"
)

type TestCase struct {
	Input       string `json:"input"`
	Output      string `json:"output"`
	TimeLimit   int    `json:"time"`
	MemoryLimit int64  `json:"memory"`
}

type LanguageConf struct {
	Name         string            `json:"name"`
	Build        []string          `json:"build"`
	Exec         []string          `json:"exec"`
	ConstRegexp  map[string]string `json:"const"`
	TimeRatio    float64           `json:"ratio"`
	Mounts       []string          `json:"mounts,omitempty"`
	ParsedMounts map[string]string `json:"-"`
}

func (lang *LanguageConf) parseMount() {
	for _, v := range lang.Mounts {
		splited := strings.Split(v, ":")
		if len(splited) != 2 {
			logrus.Warningf("Invalid mount point: ``%s'', ignore", v)
			continue
		}
		lang.ParsedMounts[splited[1]] = splited[0]
	}
}

type Subtask struct {
	Name      string   `json:"name,omitempty"`
	Score     int      `json:"score"`
	Case      []int    `json:"case,omitempty"`
	CaseInput []string `json:"caseInput,omitempty"`
}

type Problem struct {
	TimeLimit   int        `json:"time"`
	MemoryLimit int64      `json:"memory"`
	TestCase    []TestCase `json:"case"`
	Interector  Program    `json:"interactor,omitempty"`
	Checker     Program    `json:"checker,omitempty"`
	Subtasks    []Subtask  `json:"subtasks"`
}

type Program struct {
	Language   string `json:"lang"`
	Source     string `json:"src"`
	Binary     string `json:"binary,omitempty"`
	SourceHash string `json:"sourceHash,omitempty"`
}
