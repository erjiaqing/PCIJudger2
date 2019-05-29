package pci15

import (
	"os"
)

type CheckResult struct {
	Success      bool         `json:"success"`
	Build        *BuildResult `json:"build"`
	RunSolutions []*CheckRun  `json:"sols"`
}

type CheckRun struct {
	*JudgeResult
	Expected []string `json:"expected"`
}

func CheckProblemRepo(conf *Config, source string) (res *CheckResult, err error) {
	res = &CheckResult{}
	res.Success = true

	var oldCwd string
	if oldCwd, err = os.Getwd(); err != nil {
		return res, err
	}

	defer os.Chdir(oldCwd)

	if res.Build, err = BuildProblem(source, "", conf); err != nil {
		return res, err
	}

	if err = os.Chdir(source); err != nil {
		return nil, err
	}

	problemMeta := &ProblemConfig{}
	if err := loadYAML("problem.yaml", problemMeta); err != nil {
		// problem.yaml should be valid and well-formatted yaml file
		return nil, err
	}

	for _, r := range problemMeta.TestSolutions {
		judgerConf := &Config{
			Tmp:             conf.Tmp,
			IsDocker:        conf.IsDocker,
			Problem:         source,
			LanguageStorage: conf.LanguageStorage,
			ProblemPath:     source,
			SupportFiles:    conf.SupportFiles,
			MirrorFSConfig:  conf.MirrorFSConfig,
			MaxJudgeThread:  conf.MaxJudgeThread,
			RunAll:          true,
		}
		runRes, err := Judge(judgerConf, &r.SourceCode, judgerConf.Problem)
		if err != nil {
			return nil, err
		}

		for _, r2 := range runRes.Detail {
			findValid := false
			for _, v := range r.ExpectedVerdict {
				// No IG should found
				if v == r2.Verdict {
					findValid = true
					break
				}
			}
			if !findValid {
				res.Success = false
				runRes.Success = false
			}
		}
		res.RunSolutions = append(res.RunSolutions, &CheckRun{runRes, append([]string{}, r.ExpectedVerdict...)})
	}

	return res, nil
}
