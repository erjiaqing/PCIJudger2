package pci15

import (
	"io/ioutil"

	"github.com/ghodss/yaml"
)

type ProblemConfig struct {
	Version     int         `json:"version"`
	TimeLimit   uint64      `json:"timelimit"`
	TimeLimitBK uint64      `json:"time"`
	MemoryLimit uint64      `json:"memorylimit"`
	Name        string      `json:"name,omitempty"`
	Template    string      `json:"template"`
	Checker     *SourceCode `json:"checker"`
	Interactor  *SourceCode `json:"interactor,omitempty"`
	ExtraFile   []string    `json:"additionalLibrary,omitempty"`
	Case        []TestCase  `json:"case"`
}

type SourceCode struct {
	Source        string         `json:"source"`
	Language      string         `json:"lang"`
	Executable    string         `json:"-"`
	CompileResult *ExecuteResult `json:"-"`
}

type TestCase struct {
	Dependencies []string `json:"dep"`
	Score        int      `json:"score"`
	Input        string   `json:"input"`
	Output       string   `json:"output"`
	Example      bool     `json:"example"`
	TimeLimit    uint64   `json:"time,omitempty"`
	MemoryLimit  uint64   `json:"memoryLimit,omitempty"`
}

type Language struct {
	Meta struct {
		Name string `json:"name"`
	} `json:"meta"`
	Variable []*struct {
		Name    string `json:"name"`
		Match   int    `json:"match"`
		Type    string `json:"type"`
		Value   string `json:"value"`
		Default string `json:"default"`
	} `json:"variable"`
	Source     string `json:"source"`
	Executable string `json:"executable"`
	Compile    *struct {
		Cmd       []string `json:"args"`
		TimeLimit float32  `json:"timelimit"`
	} `json:"compile"`
	Execute *struct {
		Cmd       []string `json:"cmd"`
		TimeRatio float32  `json:"timeratio"`
	}
}

type CompileResult struct {
	ExeTime        uint64   `json:"exe_time"`
	ExitCode       uint64   `json:"exit_code"`
	CompilerOutput string   `json:"compiler_output"`
	CompileResult  string   `json:"compile_result"`
	Log            *PCILog  `json:"log"`
	Success        bool     `json:"success"`
	Executable     string   `json:"executable"`
	ExecuteCommand []string `json:"execute_cmd"`
}

func loadYAML(path string, to interface{}) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(data, to)
	if err != nil {
		return err
	}
	return nil
}
