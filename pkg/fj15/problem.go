package fj15

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	shutil "github.com/termie/go-shutil"
)

type BuildResult struct {
	Success bool    `json:"success"`
	Output  string  `json:"output"`
	Log     *PCILog `json:"log"`
}

func BuildProblem(problem, dest string) (*BuildResult, error) {
	logrus.Infof("Build problem %s -> %s", problem, dest)
	result := &BuildResult{
		Success: false,
		Output:  "",
		Log:     NewPCILog("problem-builder"),
	}
	shutil.CopyTree(problem, dest, nil)
	if currentDir, err := os.Getwd(); err != nil {
		result.Log.Append(fmt.Sprintf("Failed to get working directory %v", err))
		return result, err
	} else if err := os.Chdir(dest); err != nil {
		result.Log.Append(fmt.Sprintf("Failed to change working directory %v", err))
		return result, err
	} else {
		defer os.Chdir(currentDir)
	}

	problemMeta := &ProblemConfig{}
	if err := loadYAML("problem.yaml", problemMeta); err != nil {
		result.Log.Append(fmt.Sprintf("Failed to load problem.yaml: %v", err))
		return result, err
	}

	result.Log.Append(fmt.Sprintf("Compiling checker..."))

	return result, nil
}
