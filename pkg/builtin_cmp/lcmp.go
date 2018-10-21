package builtin_cmp

func init() {
	Diff["!lcmp"] = lcmp
}

func lcmp(outputFile, ansFile string) (bool, error) {
	return true, nil
}
