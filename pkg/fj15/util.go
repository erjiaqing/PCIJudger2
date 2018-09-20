package fj15

import "io/ioutil"

func ReadFile(path string) (string, error) {
	data, err := ioutil.ReadFile(path)
	return string(data), err
}
