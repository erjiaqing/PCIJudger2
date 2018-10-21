package builtin_cmp

import (
	"errors"
	"os/exec"
)

func init() {
	Diff["!diff"] = diff
}

func diff(outputFile, ansFile string) (bool, error) {
	path, err := exec.LookPath("diff")
	if err != nil {
		return false, errors.New("diff not found")
	}
	diffCmd := []string{
		"-bsqZB",
		ansFile,
		outputFile,
	}
	exe := exec.Command(path, diffCmd...)
	err = exe.Start()
	if err != nil {
		return false, err
	}
	exe.Wait()
	res := exe.ProcessState.Success()
	return res, nil
}
