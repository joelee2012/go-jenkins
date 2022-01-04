package jenkins

import (
	"encoding/json"
	"fmt"
	"path"
	"strings"

	"github.com/imroc/req"
)

type Job struct {
	*Item
}

func NewJob(url, class string, jenkins *Jenkins) *Job {
	return &Job{Item: NewItem(url, class, jenkins)}
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
func (j *Job) Copy(src, dest string) error {
	return doRequestAndDropResp(j, "POST", "createItem", ReqParams{"name": dest, "mode": "copy", "from": src})
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

func (j *Job) Disable() error {
	return doDisable(j)
}

func (j *Job) Enable() error {
	return doEnable(j)
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

func (j *Job) GetBuild(number int) (*Build, error) {
	if j.Class == "Folder" || j.Class == "WorkflowMultiBranchProject" {
		return nil, fmt.Errorf("%s is a folder", j)
	}
	var jobJson JobShortJson
	if err := j.BindAPIJson(ReqParams{"tree": "builds[number,url]"}, &jobJson); err != nil {
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
	if j.Class != "Folder" && j.Class != "WorkflowMultiBranchProject" {
		return nil, fmt.Errorf("%s is not a folder", j)
	}
	var folderJson JobShortJson
	if err := j.BindAPIJson(ReqParams{"tree": "jobs[url,name]"}, &folderJson); err != nil {
		return nil, err
	}
	for _, job := range folderJson.Jobs {
		if job.Name == name {
			return NewJob(job.URL, job.Class, j.jenkins), nil
		}
	}
	return nil, nil
}

func (j *Job) Create(name, xml string) error {
	return doRequestAndDropResp(j, "POST", "createItem", ReqParams{"name": name}, req.BodyXML(xml))
}

func (j *Job) IsFolder() error {
	if j.Class == "Folder" || j.Class == "WorkflowMultiBranchProject" {
		return nil
	}
	return fmt.Errorf("%s is not a folder", j)
}

func (j *Job) List(depth int) ([]*Job, error) {
	if j.Class != "Folder" && j.Class != "WorkflowMultiBranchProject" {
		return nil, fmt.Errorf("%s is not a folder", j)
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

func (j *Job) Credentials() *Credentials {
	return &Credentials{Item: NewItem(j.URL+"credentials/store/folder/domain/_/", "Credentials", j.jenkins)}
}

func (j *Job) GetFirstBuild() (*Build, error) {
	return j.getBuildByName("firstBuild")
}
func (j *Job) GetLastBuild() (*Build, error) {
	return j.getBuildByName("lastBuild")
}
func (j *Job) GetLastCompleteBuild() (*Build, error) {
	return j.getBuildByName("lastCompletedBuild")
}
func (j *Job) GetLastFailedBuild() (*Build, error) {
	return j.getBuildByName("lastFailedBuild")
}
func (j *Job) GetLastStableBuild() (*Build, error) {
	return j.getBuildByName("lastStableBuild")
}
func (j *Job) GetLastUnstableBuild() (*Build, error) {
	return j.getBuildByName("lastUnstableBuild")
}
func (j *Job) GetLastSuccessfulBuild() (*Build, error) {
	return j.getBuildByName("lastSuccessfulBuild")
}
func (j *Job) GetLastUnsucessfulBuild() (*Build, error) {
	return j.getBuildByName("lastUnsuccessfulBuild")
}

func (j *Job) getBuildByName(name string) (*Build, error) {
	if j.Class == "Folder" || j.Class == "WorkflowMultiBranchProject" {
		return nil, fmt.Errorf("%s is a folder", j)
	}
	var jobJson map[string]json.RawMessage
	if err := j.BindAPIJson(ReqParams{"tree": name + "[url]"}, &jobJson); err != nil {
		return nil, err
	}
	var build BuildShortJson
	if err := json.Unmarshal(jobJson[name], &build); err != nil {
		return nil, err
	}
	return NewBuild(build.URL, build.Class, j.jenkins), nil
}

func (j *Job) Delete() error {
	return doDelete(j)
}

func (j *Job) ListBuilds() ([]*Build, error) {
	if j.Class == "Folder" || j.Class == "WorkflowMultiBranchProject" {
		return nil, fmt.Errorf("%s is a folder", j)
	}
	var jobJson JobShortJson
	var builds []*Build
	if err := j.BindAPIJson(ReqParams{"tree": "builds[number,url]"}, &jobJson); err != nil {
		return nil, err
	}

	for _, build := range jobJson.Builds {
		builds = append(builds, NewBuild(build.URL, parseClass(build.Class), j.jenkins))
	}
	return builds, nil
}

type JobShortJson struct {
	Class  string           `json:"_class"`
	Builds []BuildShortJson `json:"builds"`
	Name   string           `json:"name"`
	URL    string           `json:"url"`
	Jobs   []JobShortJson   `json:"jobs"`
}
