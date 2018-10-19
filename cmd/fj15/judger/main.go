package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/erjiaqing/problem-ci-judger-2/pkg/hostconn"

	"github.com/erjiaqing/problem-ci-judger-2/pkg/fj15"
	"github.com/sirupsen/logrus"
)

var conf = &fj15.Config{
	Tmp:             os.TempDir(),
	Problem:         "/input",
	LanguageStorage: "/language",
	ProblemPath:     "/output",
	MirrorFSConfig:  "/.mirrorfs.conf",
	MaxJudgeThread:  1,
}
var code = &fj15.SourceCode{
	Language: "",
	Source:   "/code",
}

var (
	hostUDPConnIP   string
	hostUDPConnPort int
	judgeUid        string
)

func init() {
	flag.StringVar(&conf.Tmp, "tempdir", conf.Tmp, "tempory directory")
	flag.StringVar(&conf.Problem, "problem", conf.Problem, "problem path")
	flag.StringVar(&conf.LanguageStorage, "langconf", conf.LanguageStorage, "path to store languages")
	flag.StringVar(&conf.ProblemPath, "output", conf.ProblemPath, "path to output problem")
	flag.StringVar(&conf.MirrorFSConfig, "mirrorfsconf", conf.MirrorFSConfig, "path to mirrorfs config")
	flag.StringVar(&code.Source, "source", code.Source, "source code")
	flag.StringVar(&code.Language, "language", code.Language, "code language")
	flag.StringVar(&hostUDPConnIP, "udp.ip", "", "host ip")
	flag.StringVar(&judgeUid, "udp.uid", "", "judge id")
	flag.IntVar(&hostUDPConnPort, "udp.port", 0, "host port")
	flag.IntVar(&conf.MaxJudgeThread, "thread", conf.MaxJudgeThread, "code language")
}

func main() {
	flag.Parse()
	conf.HostSocket = hostconn.NewUDP(hostUDPConnIP, hostUDPConnPort, judgeUid)
	if conf.MaxJudgeThread <= 0 {
		conf.MaxJudgeThread = 1
	}
	res, err := fj15.Judge(conf, code, conf.Problem)
	if err != nil {
		logrus.Fatalf("Failed to judge code: %v", err)
	}
	resjson, err := json.MarshalIndent(res, "  ", "  ")
	if err != nil {
		logrus.Fatalf("Failed to generate output: %v", err)
	}
	conf.HostSocket.SendStatus("FF", 100)
	fmt.Printf(string(resjson))
}
