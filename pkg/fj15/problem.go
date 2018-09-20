package fj15

type CompileResult struct {
	Success bool    `json:"success"`
	Output  string  `json:"output"`
	Log     *PCILog `json:"log"`
}

func CompileProblem() {

}
