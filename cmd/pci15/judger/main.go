package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/erjiaqing/PCIJudger2/pkg/hostconn"

	"github.com/erjiaqing/PCIJudger2/pkg/pci15"
	"github.com/sirupsen/logrus"
)

var conf = &pci15.Config{
	Tmp:             os.TempDir(),
	IsDocker:        false,
	Problem:         "/input",
	LanguageStorage: "/language",
	ProblemPath:     "/output",
	SupportFiles:    "/assets",
	MirrorFSConfig:  "/.mirrorfs.conf",
	MaxJudgeThread:  1,
}
var code = &pci15.SourceCode{
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
	flag.BoolVar(&conf.IsDocker, "docker", conf.IsDocker, "is running in docker?")
	flag.StringVar(&conf.Problem, "problem", conf.Problem, "problem path")
	flag.StringVar(&conf.LanguageStorage, "langconf", conf.LanguageStorage, "path to store languages")
	flag.StringVar(&conf.ProblemPath, "output", conf.ProblemPath, "path to output problem")
	flag.StringVar(&conf.MirrorFSConfig, "mirrorfsconf", conf.MirrorFSConfig, "path to mirrorfs config")
	flag.StringVar(&code.Source, "source", code.Source, "source code")
	flag.StringVar(&code.Language, "language", code.Language, "code language")
	flag.StringVar(&hostUDPConnIP, "udp.ip", "", "host ip")
	flag.StringVar(&judgeUid, "udp.uid", "", "judge id")
	flag.StringVar(&conf.SupportFiles, "assets", conf.SupportFiles, "path to place supporting files")
	flag.IntVar(&hostUDPConnPort, "udp.port", 0, "host port")
	flag.IntVar(&conf.MaxJudgeThread, "thread", conf.MaxJudgeThread, "code language")
}

func main() {
	flag.Parse()
	conf.HostSocket = hostconn.NewUDP(hostUDPConnIP, hostUDPConnPort, judgeUid)
	if conf.MaxJudgeThread <= 0 {
		conf.MaxJudgeThread = 1
	}
	res, err := pci15.Judge(conf, code, conf.Problem)
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
