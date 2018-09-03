package main

import (
	"flag"

	"github.com/sirupsen/logrus"
)

var (
	problem string
)

func init() {
	flag.StringVar(&problem, "problem", "/problem", "Specific the path of problem.")
}

func main() {
	logrus.Infof("[Final Judger 2]")
}
