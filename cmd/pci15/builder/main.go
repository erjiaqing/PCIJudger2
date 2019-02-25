package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/erjiaqing/PCIJudger2/pkg/pci15"
	"github.com/sirupsen/logrus"
)

var conf = &pci15.Config{
	Tmp:             os.TempDir(),
	Problem:         "/input",
	LanguageStorage: "/language",
	ProblemPath:     "/output",
	SupportFiles:    "/assets",
	MirrorFSConfig:  ".mirrorfs.conf",
}

func init() {
	flag.StringVar(&conf.Tmp, "tempdir", conf.Tmp, "tempory directory")
	flag.StringVar(&conf.Problem, "input", conf.Problem, "problem path")
	flag.StringVar(&conf.LanguageStorage, "langconf", conf.LanguageStorage, "path to store languages")
	flag.StringVar(&conf.SupportFiles, "assets", conf.SupportFiles, "path to place supporting files")
}

func main() {
	flag.Parse()
	src := conf.Problem
	res, err := pci15.BuildProblem(src, "", conf)
	if err != nil {
		logrus.Fatalf("Failed to build problem: %v", err)
	}
	resjson, err := json.MarshalIndent(res, "  ", "  ")
	if err != nil {
		logrus.Fatalf("Failed to generate output: %v", err)
	}
	fmt.Printf(string(resjson))
}
