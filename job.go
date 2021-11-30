package jenkins

import (
	"fmt"
	"path"
	"strings"

	"github.com/imroc/req"
)

type Job struct {
	Item
}

func NewJob(url, class string, jenkins *Jenkins) *Job {
	return &Job{Item: *NewItem(url, class, jenkins)}
}

func (j *Job) Rename(name string) error {
	resp, err := j.Request("POST", "confirmRename", ReqParams{"newName": name})
	if err != nil {
		return err
	}
	url, _ := resp.Response().Location()
	j.URL = url.String()
	return nil
}

func (j *Job) Move(path string) error {
	parms := fmt.Sprintf(`{"destination": "/%s", "json": {"destination": "/%s"}}`, path, path)
	resp, err := j.Request("POST", "move/move", req.BodyJSON(parms))
	if err != nil {
		return err
	}
	url, _ := resp.Response().Location()
	j.URL = url.String()
	return nil
}

func (j *Job) Delete() error {
	return doDelete(j)
}

func (j *Job) GetParent() (*Job, error) {
	fullName, _ := j.jenkins.URLToName(j.URL)
	dir, _ := path.Split(strings.Trim(fullName, "/"))
	if dir == "" {
		return nil, nil
	}
	return j.jenkins.GetJob(dir)
}

func (j *Job) GetConfigure() (string, error) {
	return doGetConfigure(j)
}

func (j *Job) SetConfigure(xml string) error {
	return doSetConfigure(j, xml)
}

func (j *Job) IsBuildable() (bool, error) {
	var apiJson struct {
		Class     string `json:"_class"`
		Buildable bool   `json:"buildable"`
	}
	err := j.BindAPIJson(ReqParams{"tree": "buildable"}, &apiJson)
	return apiJson.Buildable, err
}

func (j *Job) GetName() string {
	_, name := path.Split(strings.Trim(j.URL, "/"))
	return name
}

func (j *Job) GetFullName() string {
	fullname, _ := j.jenkins.URLToName(j.URL)
	return fullname
}

func (j *Job) GetFullDisplayName() string {
	fullname, _ := j.jenkins.URLToName(j.URL)
	return strings.ReplaceAll(fullname, "/", " » ")
}

func (j *Job) GetDescription() (string, error) {
	return doGetDescription(j)
}

func (j *Job) SetDescription(description string) error {
	return doSetDescription(j, description)
}

func (j *Job) Build(param ReqParams) (*QueueItem, error) {
	buildable, err := j.IsBuildable()
	if err != nil {
		return nil, err
	}
	if !buildable {
		return nil, fmt.Errorf("%v is not buildable", j)
	}
	entry := func() string {
		reserved := []string{"token", "delay"}
		for k := range param {
			for _, e := range reserved {
				if k != e {
					return "buildWithParameters"
				}
			}
		}
		return "build"
	}()

	resp, err := j.Request("POST", entry, param)
	if err != nil {
		return nil, err
	}
	url, err := resp.Response().Location()
	if err != nil {
		return nil, err
	}
	return NewQueueItem(url.String(), j.jenkins), nil
}

type JobShortJson struct {
	Class  string           `json:"_class"`
	Builds []BuildShortJson `json:"builds"`
	Name   string           `json:"name"`
	URL    string           `json:"url"`
	Jobs   []JobShortJson   `json:"jobs"`
}

func (j *Job) GetBuild(number int) (*Build, error) {
	var jobJson JobShortJson
	err := j.BindAPIJson(ReqParams{"tree": "builds[number,url]"}, &jobJson)
	if err != nil {
		return nil, err
	}

	for _, build := range jobJson.Builds {
		if number == build.Number {
			return NewBuild(build.URL, parseClass(build.Class), j.jenkins), nil
		}
	}
	return nil, nil
}

func (j *Job) Get(name string) (*Job, error) {
	var folderJson JobShortJson
	if err := j.BindAPIJson(ReqParams{"tree": "jobs[url,name]"}, &folderJson); err != nil {
		return nil, err
	}
	for _, job := range folderJson.Jobs {
		if job.Name == name {
			return NewJob(job.URL, job.Class, j.jenkins), nil
		}
	}
	return nil, fmt.Errorf("no such job %s", name)
}

func (j *Job) Create(name, xml string) error {
	_, err := j.Request("POST", "createItem", ReqParams{"name": name}, req.BodyXML(xml))
	return err
}

func (j *Job) List(depth int) ([]*Job, error) {
	if j.Class != "Folder" && j.Class != "WorkflowMultiBranchProject" {
		return nil, fmt.Errorf("only Folder or WorkflowMultiBranchProject can be listed")
	}
	query := "jobs[url]"
	qf := "jobs[url,%s]"
	for i := 0; i < depth; i++ {
		query = fmt.Sprintf(qf, query)
	}
	var folderJson JobShortJson
	if err := j.BindAPIJson(ReqParams{"tree": query}, &folderJson); err != nil {
		return nil, err
	}
	var jobs []*Job
	var _resolve func(item *JobShortJson)
	_resolve = func(item *JobShortJson) {
		for _, job := range item.Jobs {
			if len(job.Jobs) > 0 {
				_resolve(&job)
			}
			jobs = append(jobs, NewJob(job.URL, job.Class, j.jenkins))
		}
	}
	_resolve(&folderJson)
	return jobs, nil
}
