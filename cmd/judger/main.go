package main

import (
	"flag"
	"os"

	"github.com/erjiaqing/PCIJudger2/pkg/executor"
	"github.com/sirupsen/logrus"
)

var (
	problem  string
	source   string
	language string
)

func init() {
	flag.StringVar(&problem, "problem", "/problem", "Specific the path of problem.")
	flag.StringVar(&source, "source", "/code", "Specific the path of source code.")
	flag.StringVar(&language, "language", "", "Specific the language of source code.")
}

func main() {
	logrus.Infof("[Final Judger 2]")
	uid := os.Getuid()
	if uid == -1 {
		logrus.Fatal("Final Judger cannot run on Windows, please use the container version or an vm")
	} else if uid != 0 {
		logrus.Fatalf("Final Judger should run in root, however, your uid is %d, 0 required", uid)
	}
	cgroup, err := executor.NewCGroup()
	if err != nil {
		logrus.Fatalf("Cannot create cgroup: %v", err)
	}
	logrus.Infof("Created cgroup: %s", cgroup.Name)
	defer func() {
		logrus.Infof("Cleanning up cgroup: %v", cgroup.Name)
		err := cgroup.CleanUp()
		if err != nil {
			logrus.Errorf("Failed to clean up: %v", err)
		}
	}()
}
