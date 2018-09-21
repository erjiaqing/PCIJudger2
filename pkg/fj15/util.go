package fj15

import (
	"fmt"
	"io/ioutil"
	"os"
)

func ReadFile(path string) (string, error) {
	data, err := ioutil.ReadFile(path)
	return string(data), err
}

func ReadFirstBytes(path string, maxSize int64) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	data := make([]byte, maxSize)
	count, err := file.Read(data)
	if err != nil {
		return "", err
	}
	fileStat, err := file.Stat()
	if err != nil {
		return "", err
	}
	fileSize := fmt.Sprintf("...(total %s bytes)", fileStat.Size())
	if fileStat.Size() > maxSize-int64(len(fileSize)) {
		return string(data[0:maxSize-int64(len(fileSize))]) + fileSize, nil
	} else {
		return string(data[0:count]), nil
	}
}
