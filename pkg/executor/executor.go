package executor

import (
	"errors"

	"github.com/erjiaqing/problem-ci-judger-2/pkg/types"
)

func Execute(timeLimit int, memoryLimit int64, program types.Program, language types.LanguageConf) (*ExecuteResult, error) {
	return nil, errors.New("not implemented")
}

func Build(program *types.Program, language types.LanguageConf) (*BuildResult, error) {
	return nil, errors.New("not implemented")
}
