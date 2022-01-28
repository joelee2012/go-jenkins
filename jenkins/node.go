package jenkins

import "strings"

type ComputerSetShortJson struct {
	Class     string               `json:"_class"`
	Computers []*ComputerShortJson `json:"computer"`
}

type ComputerShortJson struct {
	Class              string      `json:"_class"`
	Description        string      `json:"description"`
	DisplayName        string      `json:"displayName"`
	Executors          []*Executor `json:"executors"`
	Offline            bool        `json:"offline"`
	OfflineCause       interface{} `json:"offlineCause"`
	OfflineCauseReason string      `json:"offlineCauseReason"`
	OneOffExecutors    []*Executor `json:"oneOffExecutors"`
}

type NodeService struct {
	*Item
}

func NewNodeService(client *Client) *NodeService {
	return &NodeService{Item: NewItem(client.URL+"computer/", "Nodes", client)}
}

func (ns *NodeService) GetBuilds() ([]*BuildService, error) {
	var compSet ComputerSet
	var builds []*BuildService
	tree := "computer[executors[currentExecutable[url]],oneOffExecutors[currentExecutable[url]]]"
	if err := ns.BindAPIJson(ReqParams{"tree": tree, "depth": "2"}, &compSet); err != nil {
		return nil, err
	}
	buildURLs := map[string]string{}
	parseBuild := func(executors []*Executor) {
		for _, e := range executors {
			if e.CurrentExecutable == nil {
				continue
			}
			if strings.HasSuffix(e.CurrentExecutable.Class, "PlaceholderExecutable") {
				e.CurrentExecutable.Class = "org.jenkinsci.plugins.workflow.job.WorkflowRun"
			}
			buildURLs[e.CurrentExecutable.URL] = e.CurrentExecutable.Class
		}
	}
	for _, c := range compSet.Computers {
		parseBuild(c.Executors)
		parseBuild(c.OneOffExecutors)
	}
	for k, v := range buildURLs {
		builds = append(builds, NewBuild(k, v, ns.client))
	}
	return builds, nil
}

func (ns *NodeService) Get(name string) (*Computer, error) {
	compSet := &ComputerSet{}
	if err := ns.BindAPIJson(ReqParams{"tree": "computer[displayName]"}, compSet); err != nil {
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
	if err := ns.BindAPIJson(ReqParams{"tree": "computer[displayName]"}, compSet); err != nil {
		return nil, err
	}
	return compSet.Computers, nil
}

func (ns *NodeService) Enable(name string) error {
	_, err := ns.Request("POST", name+"/toggleOffline", ReqParams{"offlineMessage": ""})
	return err
}

func (ns *NodeService) Disable(name, msg string) error {
	_, err := ns.Request("POST", name+"/toggleOffline", ReqParams{"offlineMessage": msg})
	return err
}

func (ns *NodeService) Delete(name string) error {
	_, err := ns.Request("POST", name+"/doDelete")
	return err
}
