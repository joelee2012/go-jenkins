package jenkins

import (
	"github.com/imroc/req"
)

type ComputerSet struct {
	Item
}

func (cs *ComputerSet) GetBuilds() ([]Build, error) {
	result, err := cs.APIJson(req.Param{"tree": "computer[oneOffExecutors[currentExecutable[url]]]"})
	var builds []Build
	if err != nil {
		return builds, err
	}
	executors := result.Get("computer.#.oneOffExecutors.#.currentExecutable")
	for _, executor := range executors.Array() {
		for _, build := range executor.Array() {
			url := build.Get("url").String()
			class := GetClassName(build.Get("_class").String())
			builds = append(builds, NewBuild(url, class, cs.jenkins))
		}
	}
	return builds, nil
}
