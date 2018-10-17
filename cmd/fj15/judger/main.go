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
	MirrorFSConfig:  "/.mirrorfs.conf",
}
var code = &fj15.SourceCode{
	Language: "",
	Source:   "/code",
}

func init() {
	flag.StringVar(&conf.Tmp, "tempdir", conf.Tmp, "tempory directory")
	flag.StringVar(&conf.Problem, "problem", conf.Problem, "problem path")
	flag.StringVar(&conf.LanguageStorage, "langconf", conf.LanguageStorage, "path to store languages")
	flag.StringVar(&conf.ProblemPath, "output", conf.ProblemPath, "path to output problem")
	flag.StringVar(&conf.MirrorFSConfig, "mirrorfsconf", conf.MirrorFSConfig, "path to mirrorfs config")
	flag.StringVar(&code.Source, "source", code.Source, "source code")
	flag.StringVar(&code.Language, "language", code.Language, "code language")
}

func main() {
	flag.Parse()
	res, err := fj15.Judge(conf, code, conf.Problem)
	if err != nil {
		logrus.Fatalf("Failed to judge code: %v", err)
	}
	resjson, err := json.MarshalIndent(res, "  ", "  ")
	if err != nil {
		logrus.Fatalf("Failed to generate output: %v", err)
	}
	fmt.Printf(string(resjson))
}
