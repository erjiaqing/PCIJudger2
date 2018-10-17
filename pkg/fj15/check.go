package fj15

import (
	"os"
	"path/filepath"

	shutil "github.com/termie/go-shutil"
)

type CheckerResult struct {
	Success bool `json:"success"`
}

type checker struct {
	Root string
}

var Checker checker

func (c *checker) ProcessWork(conf *Config) error {
	if oldCwd, err := os.Getwd(); err != nil {
		return err
	} else {
		defer os.Chdir(oldCwd)
	}

	//log := NewPCILog("checker")

	tmpDirName := GetRandomString()
	BaseDir := filepath.Join(conf.Tmp, tmpDirName)
	ProblemSourceDir := conf.Problem
	if err := shutil.CopyTree(ProblemSourceDir, BaseDir, nil); err != nil {
		return err
	}
	if err := os.Chdir(BaseDir); err != nil {
		return err
	}

	compileProblemLogger := NewPCILog("problem-compiler")
	compileProblemLogger.Append("Building problem")
	problemCompileDir := filepath.Join(conf.Tmp, GetRandomString())
	BuildProblem(BaseDir, problemCompileDir, conf)

	return nil
}
