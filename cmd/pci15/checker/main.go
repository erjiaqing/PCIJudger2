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
	SupportFiles:    "/assets",
	MirrorFSConfig:  ".mirrorfs.conf",
	MaxJudgeThread:  1,
}

func init() {
	flag.StringVar(&conf.Tmp, "tempdir", conf.Tmp, "tempory directory")
	flag.StringVar(&conf.Problem, "input", conf.Problem, "problem path")
	flag.StringVar(&conf.LanguageStorage, "langconf", conf.LanguageStorage, "path to store languages")
	flag.StringVar(&conf.SupportFiles, "assets", conf.SupportFiles, "path to place supporting files")
	flag.IntVar(&conf.MaxJudgeThread, "thread", conf.MaxJudgeThread, "code language")
}

func main() {
	flag.Parse()
	src := conf.Problem
	res, err := pci15.CheckProblemRepo(conf, src)
	if err != nil {
		logrus.Fatalf("Failed to build problem: %v", err)
	}
	resjson, err := json.MarshalIndent(res, "  ", "  ")
	if err != nil {
		logrus.Fatalf("Failed to generate output: %v", err)
	}
	fmt.Printf(string(resjson))
}
