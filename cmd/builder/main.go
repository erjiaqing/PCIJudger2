package main

import (
	"flag"

	"github.com/sirupsen/logrus"
)

var (
	problem     string
	problemDest string
)

func init() {
	flag.StringVar(&problem, "problem", "/problem", "Specific the path of problem.")
	flag.StringVar(&problemDest, "dest", "/dest", "Specific the path of built problem.")
}

func main() {
	logrus.Infof("[Final Judger 2]")
}
