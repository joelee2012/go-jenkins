package jenkins

import (
	"encoding/json"
	"fmt"
	"path"
	"strings"

	"github.com/imroc/req"
)

type JobItem struct {
	*Item
	Credentials *CredentialService
}

func NewJobItem(url, class string, client *Client) *JobItem {
	j := &JobItem{Item: NewItem(url, class, client)}
	j.Credentials = NewCredentialService(j)
	return j
}

func (j *JobItem) Rename(name string) error {
	resp, err := j.Request("POST", "confirmRename", ReqParams{"newName": name})
	if err != nil {
		return err
	}
	url, _ := resp.Response().Location()
	j.URL = appendSlash(url.String())
	return nil
}

func (j *JobItem) Move(path string) error {
	path = strings.Trim(path, "/")
	resp, err := j.Request("POST", "move/move", ReqParams{"destination": "/" + path})
	if err != nil {
		return err
	}
	url, _ := resp.Response().Location()
	j.URL = appendSlash(url.String())
	return nil
}

func (j *JobItem) Copy(src, dest string) error {
	_, err := j.Request("POST", "createItem", ReqParams{"name": dest, "mode": "copy", "from": src})
	return err
}

func (j *JobItem) GetParent() (*JobItem, error) {
	fullName, _ := j.client.URL2Name(j.URL)
	dir, _ := path.Split(strings.Trim(fullName, "/"))
	if dir == "" {
		return nil, nil
	}
	return j.client.GetJob(dir)
}

func (j *JobItem) GetConfigure() (string, error) {
	resp, err := j.Request("GET", "/config.xml")
	return resp.String(), err
}

func (j *JobItem) SetConfigure(xml string) error {
	_, err := j.Request("POST", "/config.xml", req.BodyXML(xml))
	return err
}

func (j *JobItem) Disable() error {
	_, err := j.Request("POST", "disable")
	return err
}

func (j *JobItem) Enable() error {
	_, err := j.Request("POST", "enable")
	return err
}

func (j *JobItem) IsBuildable() (bool, error) {
	var job struct {
		Class     string `json:"_class"`
		Buildable bool   `json:"buildable"`
	}
	err := j.BindAPIJson(ReqParams{"tree": "buildable"}, &job)
	return job.Buildable, err
}

func (j *JobItem) GetName() string {
	_, name := path.Split(strings.Trim(j.URL, "/"))
	return name
}

func (j *JobItem) GetFullName() string {
	fullname, _ := j.client.URL2Name(j.URL)
	return fullname
}

func (j *JobItem) GetFullDisplayName() string {
	fullname, _ := j.client.URL2Name(j.URL)
	return strings.ReplaceAll(fullname, "/", " » ")
}

func (j *JobItem) GetDescription() (string, error) {
	data := make(map[string]string)
	if err := j.BindAPIJson(ReqParams{"tree": "description"}, &data); err != nil {
		return "", err
	}
	return data["description"], nil
}

func (j *JobItem) SetDescription(description string) error {
	_, err := j.Request("POST", "submitDescription", ReqParams{"description": description})
	return err
}

func (j *JobItem) Build(param ReqParams) (*OneQueueItem, error) {
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
	return NewQueueItem(url.String(), j.client), nil
}

func (j *JobItem) GetBuild(number int) (*BuildItem, error) {
	if j.Class == "Folder" || j.Class == "WorkflowMultiBranchProject" {
		return nil, fmt.Errorf("%s have no builds", j)
	}
	jobJson := &Job{}
	if err := j.BindAPIJson(ReqParams{"tree": "builds[number,url]"}, &jobJson); err != nil {
		return nil, err
	}

	for _, build := range jobJson.Builds {
		if number == build.Number {
			return NewBuildItem(build.URL, parseClass(build.Class), j.client), nil
		}
	}
	return nil, nil
}

func (j *JobItem) Get(name string) (*JobItem, error) {
	if j.Class != "Folder" && j.Class != "WorkflowMultiBranchProject" {
		return nil, fmt.Errorf("%s have no jobs", j)
	}
	var folderJson Job
	if err := j.BindAPIJson(ReqParams{"tree": "jobs[url,name]"}, &folderJson); err != nil {
		return nil, err
	}
	for _, job := range folderJson.Jobs {
		if job.Name == name {
			return NewJobItem(job.URL, job.Class, j.client), nil
		}
	}
	return nil, nil
}

func (j *JobItem) Create(name, xml string) error {
	_, err := j.Request("POST", "createItem", ReqParams{"name": name}, req.BodyXML(xml))
	return err
}

func (j *JobItem) List(depth int) ([]*JobItem, error) {
	if j.Class != "Folder" && j.Class != "WorkflowMultiBranchProject" {
		return nil, fmt.Errorf("%s have no jobs", j)
	}
	query := "jobs[url]"
	qf := "jobs[url,%s]"
	for i := 0; i < depth; i++ {
		query = fmt.Sprintf(qf, query)
	}
	var folderJson Job
	if err := j.BindAPIJson(ReqParams{"tree": query}, &folderJson); err != nil {
		return nil, err
	}
	var jobs []*JobItem
	var _resolve func(item *Job)
	_resolve = func(item *Job) {
		for _, job := range item.Jobs {
			if len(job.Jobs) > 0 {
				_resolve(job)
			}
			jobs = append(jobs, NewJobItem(job.URL, job.Class, j.client))
		}
	}
	_resolve(&folderJson)
	return jobs, nil
}

func (j *JobItem) GetFirstBuild() (*BuildItem, error) {
	return j.getBuildByName("firstBuild")
}
func (j *JobItem) GetLastBuild() (*BuildItem, error) {
	return j.getBuildByName("lastBuild")
}
func (j *JobItem) GetLastCompleteBuild() (*BuildItem, error) {
	return j.getBuildByName("lastCompletedBuild")
}
func (j *JobItem) GetLastFailedBuild() (*BuildItem, error) {
	return j.getBuildByName("lastFailedBuild")
}
func (j *JobItem) GetLastStableBuild() (*BuildItem, error) {
	return j.getBuildByName("lastStableBuild")
}
func (j *JobItem) GetLastUnstableBuild() (*BuildItem, error) {
	return j.getBuildByName("lastUnstableBuild")
}
func (j *JobItem) GetLastSuccessfulBuild() (*BuildItem, error) {
	return j.getBuildByName("lastSuccessfulBuild")
}
func (j *JobItem) GetLastUnsucessfulBuild() (*BuildItem, error) {
	return j.getBuildByName("lastUnsuccessfulBuild")
}

func (j *JobItem) getBuildByName(name string) (*BuildItem, error) {
	if j.Class == "Folder" || j.Class == "WorkflowMultiBranchProject" {
		return nil, fmt.Errorf("%s have no builds", j)
	}
	var jobJson map[string]json.RawMessage
	if err := j.BindAPIJson(ReqParams{"tree": name + "[url]"}, &jobJson); err != nil {
		return nil, err
	}
	var build Build
	if err := json.Unmarshal(jobJson[name], &build); err != nil {
		return nil, err
	}
	return NewBuildItem(build.URL, build.Class, j.client), nil
}

func (j *JobItem) Delete() error {
	_, err := j.Request("POST", "doDelete")
	return err
}

func (j *JobItem) ListBuilds() ([]*BuildItem, error) {
	if j.Class == "Folder" || j.Class == "WorkflowMultiBranchProject" {
		return nil, fmt.Errorf("%s have no builds", j)
	}
	var jobJson Job
	var builds []*BuildItem
	if err := j.BindAPIJson(ReqParams{"tree": "builds[url]"}, &jobJson); err != nil {
		return nil, err
	}

	for _, build := range jobJson.Builds {
		builds = append(builds, NewBuildItem(build.URL, parseClass(build.Class), j.client))
	}
	return builds, nil
}
