package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/erjiaqing/problem-ci-judger-2/pkg/fj15"
	"github.com/sirupsen/logrus"
)

var conf = &fj15.Config{
	Tmp:             os.TempDir(),
	Problem:         "/input",
	LanguageStorage: "/language",
	ProblemPath:     "/output",
	MirrorFSConfig:  ".mirrorfs.conf",
}

func init() {
	flag.StringVar(&conf.Tmp, "tempdir", conf.Tmp, "tempory directory")
	flag.StringVar(&conf.Problem, "input", conf.Problem, "problem path")
	flag.StringVar(&conf.LanguageStorage, "langconf", conf.LanguageStorage, "path to store languages")
	flag.StringVar(&conf.ProblemPath, "output", conf.ProblemPath, "path to output problem")
}

func main() {
	flag.Parse()
	src := conf.Problem
	dst := conf.ProblemPath
	res, err := fj15.BuildProblem(src, dst, conf)
	if err != nil {
		logrus.Fatalf("Failed to build problem: %v", err)
	}
	resjson, err := json.MarshalIndent(res, "  ", "  ")
	if err != nil {
		logrus.Fatalf("Failed to generate output: %v", err)
	}
	fmt.Printf(string(resjson))
}
