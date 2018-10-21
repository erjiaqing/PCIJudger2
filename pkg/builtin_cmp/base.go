package builtin_cmp

type CompFunc = func(outputFile, ansFile string) (bool, error)

var Diff = make(map[string]CompFunc)
