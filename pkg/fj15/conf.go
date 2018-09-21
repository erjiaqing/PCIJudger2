package fj15

type Config struct {
	Tmp             string `json:"tmp"`
	Problem         string `json:"problem"`
	LanguageStorage string `json:"lang"`
	ProblemPath     string `json:"datapath"`
	MirrorFSConfig  string `json:"mirrorfs"`
}
