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
	Conf *Config
	Root string
}

var Checker checker

func (c *checker) ProcessWork() error {
	if oldCwd, err := os.Getwd(); err != nil {
		return err
	} else {
		defer os.Chdir(oldCwd)
	}

	log := NewPCILog("checker")

	tmpDirName := GetRandomString()
	BaseDir := filepath.Join(c.Conf.Tmp, tmpDirName)
	ProblemSourceDir := c.Conf.Problem
	if err := shutil.CopyTree(ProblemSourceDir, BaseDir, nil); err != nil {
		return err
	}
	if err := os.Chdir(BaseDir); err != nil {
		return err
	}

	compileProblemLogger := NewPCILog("problem-compiler")
	compileProblemLogger.Append("Building problem")
	problemCompileDir := filepath.Join(c.Conf.Tmp, GetRandomString())

	return nil
}
