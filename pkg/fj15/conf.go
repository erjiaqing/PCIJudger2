package fj15

import "github.com/erjiaqing/PCIJudger2/pkg/hostconn"

type Config struct {
	Tmp             string        `json:"tmp"`
	Problem         string        `json:"problem"`
	LanguageStorage string        `json:"lang"`
	ProblemPath     string        `json:"datapath"`
	MirrorFSConfig  string        `json:"mirrorfs"`
	MaxJudgeThread  int           `json:"thread"`
	HostSocket      *hostconn.UDP `json:"-"`
}
