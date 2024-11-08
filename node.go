package jenkins

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type Nodes struct {
	*Item
}

var nodeNameMap = map[string]string{"master": "(master)", "Built-In Node": "(built-in)"}

func (ns *Nodes) GetBuilds() ([]*Build, error) {
	compSet := &ComputerSet{}
	var builds []*Build
	tree := "computer[executors[currentExecutable[url]],oneOffExecutors[currentExecutable[url]]]"
	if err := ns.ApiJson(compSet, &ApiJsonOpts{Tree: tree, Depth: 2}); err != nil {
		return nil, err
	}
	buildConf := map[string]string{}
	parseBuild := func(executors []*Executor) {
		for _, e := range executors {
			if e.CurrentExecutable == nil {
				continue
			}
			if strings.HasSuffix(e.CurrentExecutable.Class, "PlaceholderExecutable") {
				e.CurrentExecutable.Class = "org.jenkinsci.plugins.workflow.job.WorkflowRun"
			}
			buildConf[e.CurrentExecutable.URL] = e.CurrentExecutable.Class
		}
	}
	for _, c := range compSet.Computers {
		parseBuild(c.Executors)
		parseBuild(c.OneOffExecutors)
	}
	for k, v := range buildConf {
		builds = append(builds, NewBuild(k, v, ns.jenkins))
	}
	return builds, nil
}

func (ns *Nodes) Get(name string) (*Computer, error) {
	compSet := &ComputerSet{}
	if err := ns.ApiJson(compSet, nil); err != nil {
		return nil, err
	}

	for _, c := range compSet.Computers {
		if name == c.DisplayName {
			return c, nil
		}
	}
	return nil, fmt.Errorf("no such node [%s]", name)
}

func (ns *Nodes) List() ([]*Computer, error) {
	compSet := &ComputerSet{}
	if err := ns.ApiJson(compSet, nil); err != nil {
		return nil, err
	}
	return compSet.Computers, nil
}

func (ns *Nodes) covertName(name string) string {
	if displayName, ok := nodeNameMap[name]; ok {
		return displayName
	}
	return name
}

func (ns *Nodes) Enable(name string) (*http.Response, error) {
	return ns.Request("POST", ns.covertName(name)+"/toggleOffline?offlineMessage=", nil)
}

func (ns *Nodes) Disable(name, msg string) (*http.Response, error) {
	v := url.Values{}
	v.Add("offlineMessage", msg)
	return ns.Request("POST", ns.covertName(name)+"/toggleOffline?"+v.Encode(), nil)
}

func (ns *Nodes) Delete(name string) (*http.Response, error) {
	return ns.Request("POST", ns.covertName(name)+"/doDelete", nil)
}
