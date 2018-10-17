package fj15

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	shutil "github.com/termie/go-shutil"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
)

type BuildResult struct {
	Success bool    `json:"success"`
	Output  string  `json:"output"`
	Log     *PCILog `json:"log"`
}

func BuildProblem(problem, dest string, conf *Config) (*BuildResult, error) {
	logrus.Infof("Build problem %s -> %s", problem, dest)
	result := &BuildResult{
		Success: false,
		Output:  "",
		Log:     NewPCILog("problem-builder"),
	}
	if err := shutil.CopyTree(problem, dest, nil); err != nil {
		return nil, err
	}
	if currentDir, err := os.Getwd(); err != nil {
		result.Log.Append(fmt.Sprintf("Failed to get working directory %v", err))
		return result, err
	} else if err := os.Chdir(dest); err != nil {
		result.Log.Append(fmt.Sprintf("Failed to change working directory %v", err))
		return result, err
	} else {
		defer os.Chdir(currentDir)
	}
	currdir, _ := os.Getwd()
	logrus.Infof("Current directory: %s", currdir)

	problemMeta := &ProblemConfig{}
	if err := loadYAML("problem.yaml", problemMeta); err != nil {
		result.Log.Append(fmt.Sprintf("Failed to load problem.yaml: %v", err))
		return result, err
	}

	if problemMeta.Checker != nil {
		result.Log.Append(fmt.Sprintf("Compiling checker..."))
		logrus.Infof("Compiling checker...")
		compilerOutput, err := problemMeta.Checker.Compile(conf, dest)
		result.Log.Append("Compiler Output:")
		result.Log.Append(compilerOutput)
		logrus.Infof("Compiler Output:")
		fmt.Fprintf(os.Stderr, "%s", compilerOutput)
		if err != nil {
			result.Success = false
			return result, err
		}
	}

	if problemMeta.Interactor != nil {
		result.Log.Append(fmt.Sprintf("Compiling interactor"))
		compilerResult, err := problemMeta.Interactor.Compile(conf, dest)
		result.Log.Append("Compiler Output:")
		result.Log.Append(compilerResult)
		if err != nil {
			result.Success = false
			return result, err
		}
	}

	return result, nil
}

func GetProblem(conf *Config, problem, problemGit, problemVersion string) error {
	logrus.Info("Get problem: %s:%s", problemGit, problemVersion)

	problemDir := filepath.Join(conf.ProblemPath, problem)
	tmpDir := filepath.Join(conf.Tmp, GetRandomString())

	if currentDir, err := os.Getwd(); err != nil {
		return err
	} else if err := os.Chdir(conf.Tmp); err != nil {
		return err
	} else {
		defer os.Chdir(currentDir)
	}

	repo, err := git.PlainClone(tmpDir, false, &git.CloneOptions{
		URL: problemGit,
	})

	if err != nil {
		return err
	}

	defer os.RemoveAll(tmpDir)

	if currentDir, err := os.Getwd(); err != nil {
		return err
	} else if err := os.Chdir(tmpDir); err != nil {
		return err
	} else {
		defer os.Chdir(currentDir)
	}

	repoTree, err := repo.Worktree()
	if err != nil {
		return err
	}

	if err := repoTree.Checkout(&git.CheckoutOptions{
		Hash:   plumbing.NewHash(problemVersion),
		Create: false,
		Force:  false,
	}); err != nil {
		return err
	}

	_, err = BuildProblem(tmpDir, problemDir, conf)
	if err != nil {
		return err
	}
	return nil
}

func CheckProblem(conf *Config, problem, problemVersion string) bool {
	problemDir := filepath.Join(conf.ProblemPath, problem)
	repo, err := git.PlainOpen(problemDir)
	if err != nil {
		return false
	}
	head, err := repo.Head()
	if err != nil {
		return false
	}
	return head.Hash().String() == problemVersion
}
