package jenkins

import "strings"

type ComputerSet struct {
	*Item
}

func NewComputerSet(url string, jenkins *Jenkins) *ComputerSet {
	return &ComputerSet{Item: NewItem(url, "ComputerSet", jenkins)}
}

func (cs *ComputerSet) GetBuilds() ([]*Build, error) {
	var csJson ComputerSetJson
	var builds []*Build
	if err := cs.BindAPIJson(ReqParams{"tree": "computer[executors[currentExecutable[url]],oneOffExecutors[currentExecutable[url]]]", "depth": "2"}, &csJson); err != nil {
		return nil, err
	}
	var buildConf map[string]string
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
	for _, c := range csJson.Computers {
		parseBuild(c.Executors)
		parseBuild(c.OneOffExecutors)
	}
	for k, v := range buildConf {
		builds = append(builds, NewBuild(k, v, cs.jenkins))
	}
	return builds, nil
}

func (cs *ComputerSet) Get(name string) (*Computer, error) {
	var csJson ComputerSetJson
	if err := cs.BindAPIJson(ReqParams{"tree": "computer[displayName]"}, &csJson); err != nil {
		return nil, err
	}
	nodeName := map[string]string{"master": "(master)", "Built-In Node": "(built-in)"}
	for _, c := range csJson.Computers {
		if name == c.DisplayName {
			if name, ok := nodeName[c.DisplayName]; ok {
				return &Computer{Item: NewItem(cs.jenkins.URL+name, c.Class, cs.jenkins)}, nil
			} else {
				return &Computer{Item: NewItem(cs.jenkins.URL+c.DisplayName, c.Class, cs.jenkins)}, nil
			}
		}
	}
	return nil, nil
}

func (cs *ComputerSet) List() ([]*Computer, error) {
	var csJson ComputerSetJson
	var computers []*Computer
	if err := cs.BindAPIJson(ReqParams{"tree": "computer[displayName]"}, &csJson); err != nil {
		return nil, err
	}
	nodeName := map[string]string{"master": "(master)", "Built-In Node": "(built-in)"}
	for _, c := range csJson.Computers {
		if name, ok := nodeName[c.DisplayName]; ok {
			computers = append(computers, &Computer{Item: NewItem(cs.jenkins.URL+name, c.Class, cs.jenkins)})
		} else {
			computers = append(computers, &Computer{Item: NewItem(cs.jenkins.URL+c.DisplayName, c.Class, cs.jenkins)})
		}
	}
	return computers, nil
}

type Computer struct {
	*Item
}

func (c *Computer) Enable() error {
	return doRequestAndDropResp(c, "POST", "toggleOffline", ReqParams{"offlineMessage": ""})
}

func (c *Computer) Disable(msg string) error {
	return doRequestAndDropResp(c, "POST", "toggleOffline", ReqParams{"offlineMessage": msg})
}
