package jenkins

import "strings"

type NodeService struct {
	*Jenkins
	BaseURL string
}

var nodeNameMap = map[string]string{"master": "(master)", "Built-In Node": "(built-in)"}

func NewNodeService(client *Jenkins) *NodeService {
	return &NodeService{BaseURL: client.URL + "computer/", Jenkins: client}
}

func (ns *NodeService) GetBuilds() ([]*BuildItem, error) {
	compSet := &ComputerSet{}
	var builds []*BuildItem
	tree := "computer[executors[currentExecutable[url]],oneOffExecutors[currentExecutable[url]]]"
	if _, err := R().SetQueryParams(map[string]string{"tree": tree, "depth": "2"}).SetSuccessResult(compSet).Post("api/json"); err != nil {
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
		builds = append(builds, NewBuildItem(k, v, ns.Jenkins))
	}
	return builds, nil
}

func (ns *NodeService) Get(name string) (*Computer, error) {
	compSet := &ComputerSet{}
	if _, err := R().SetSuccessResult(compSet).Get("api/json"); err != nil {
		return nil, err
	}

	for _, c := range compSet.Computers {
		if name == c.DisplayName {
			return c, nil
		}
	}
	return nil, nil
}

func (ns *NodeService) List() ([]*Computer, error) {
	compSet := &ComputerSet{}
	if _, err := R().SetSuccessResult(compSet).Get("api/json"); err != nil {
		return nil, err
	}
	return compSet.Computers, nil
}

func (ns *NodeService) covertName(name string) string {
	if displayName, ok := nodeNameMap[name]; ok {
		return displayName
	}
	return name
}

func (ns *NodeService) Enable(name string) error {
	_, err := R().SetQueryParam("offlineMessage", "").Post(ns.covertName(name) + "/toggleOffline")
	return err
}

func (ns *NodeService) Disable(name, msg string) error {
	_, err := R().SetQueryParam("offlineMessage", msg).Post(ns.covertName(name) + "/toggleOffline")
	return err
}

func (ns *NodeService) Delete(name string) error {
	_, err := R().Post(ns.covertName(name) + "/doDelete")
	return err
}
