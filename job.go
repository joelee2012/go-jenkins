package jenkins

import (
	"encoding/json"
	"fmt"
	"net/url"
	"path"
	"strings"
)

type JobItem struct {
	*Client
	BaseURL         string
	Class           string
	Credentials     *CredentialService
	Views           *ViewService
	Name            string
	FullName        string
	FullDisplayName string
}

func NewJobItem(url, class string, client *Client) *JobItem {
	j := &JobItem{BaseURL: url, Client: client, Class: class}
	j.Credentials = NewCredentialService(j)
	j.Views = NewViewService(j)
	j.setName()
	return j
}

func (job *JobItem) Rename(name string) error {
	resp, err := job.R().SetQueryParam("newName", name).Post("confirmRename")
	if err != nil {
		return err
	}
	url, _ := resp.Location()
	job.BaseURL = appendSlash(url.String())
	job.setName()
	return nil
}

func (j *JobItem) Move(path string) error {
	path = strings.Trim(path, "/")
	resp, err := j.R().SetQueryParam("destination", "/"+path).Post("move/move")
	if err != nil {
		return err
	}
	url, _ := resp.Location()
	j.BaseURL = appendSlash(url.String())
	j.setName()
	return nil
}

func (j *JobItem) Copy(src, dest string) error {
	_, err := j.R().SetQueryParams(map[string]string{"name": dest, "mode": "copy", "from": src}).Post("createItem")
	return err
}

func (j *JobItem) GetParent() (*JobItem, error) {
	fullName, _ := j.URL2Name(j.BaseURL)
	dir, _ := path.Split(strings.Trim(fullName, "/"))
	if dir == "" {
		return nil, nil
	}
	return j.GetJob(dir)
}

func (j *JobItem) GetConfigure() (string, error) {
	resp, err := j.R().Get("config.xml")
	return resp.String(), err
}

func (j *JobItem) SetConfigure(xml string) error {
	_, err := j.R().SetBody(xml).Post("config.xml")
	return err
}

func (j *JobItem) Disable() error {
	_, err := j.R().Post("disable")
	return err
}

func (j *JobItem) Enable() error {
	_, err := j.R().Post("enable")
	return err
}

func (j *JobItem) BindAPIJson(params map[string]string, v interface{}) error {
	_, err := j.R().SetQueryParams(params).SetResult(v).Get("api/json")
	return err
}

func (j *JobItem) IsBuildable() (bool, error) {
	var job struct {
		Class     string `json:"_class"`
		Buildable bool   `json:"buildable"`
	}
	err := j.BindAPIJson(map[string]string{"tree": "buildable"}, &job)
	return job.Buildable, err
}

func (j *JobItem) setName() {
	urlPath, _ := j.URL2Name(j.BaseURL)
	j.FullName, _ = url.PathUnescape(urlPath)
	_, j.Name = path.Split(j.FullName)
	j.FullDisplayName, _ = url.PathUnescape(strings.ReplaceAll(j.FullName, "/", " » "))
}

func (j *JobItem) GetDescription() (string, error) {
	data := make(map[string]string)
	if err := j.BindAPIJson(map[string]string{"tree": "description"}, &data); err != nil {
		return "", err
	}
	return data["description"], nil
}

func (j *JobItem) SetDescription(description string) error {
	_, err := j.R().SetQueryParam("description", description).Post("submitDescription")
	return err
}

func (j *JobItem) Build(params map[string]string) (*OneQueueItem, error) {
	entry := func() string {
		reserved := []string{"token", "delay"}
		for k := range params {
			for _, e := range reserved {
				if k != e {
					return "buildWithParameters"
				}
			}
		}
		return "build"
	}()

	resp, err := j.R().SetQueryParams(params).Post(entry)
	if err != nil {
		return nil, err
	}
	url, err := resp.Location()
	if err != nil {
		return nil, err
	}
	return NewQueueItem(url.String(), j.Client), nil
}

func (j *JobItem) GetBuild(number int) (*BuildItem, error) {
	if j.Class == "Folder" || j.Class == "WorkflowMultiBranchProject" {
		return nil, fmt.Errorf("%s have no builds", j)
	}
	jobJson := &Job{}
	if err := j.BindAPIJson(map[string]string{"tree": "builds[number,url]"}, &jobJson); err != nil {
		return nil, err
	}

	for _, build := range jobJson.Builds {
		if number == build.Number {
			return NewBuildItem(build.URL, build.Class, j.Client), nil
		}
	}
	return nil, nil
}

func (j *JobItem) Get(name string) (*JobItem, error) {
	if j.Class != "Folder" && j.Class != "WorkflowMultiBranchProject" {
		return nil, fmt.Errorf("%s have no jobs", j)
	}
	var folderJson Job
	if err := j.BindAPIJson(map[string]string{"tree": "jobs[url,name]"}, &folderJson); err != nil {
		return nil, err
	}
	for _, job := range folderJson.Jobs {
		if job.Name == name {
			return NewJobItem(job.URL, job.Class, j.Client), nil
		}
	}
	return nil, nil
}

func (j *JobItem) Create(name, xml string) error {
	_, err := j.R().SetQueryParams(map[string]string{"name": name}).SetBody(xml).Post("createItem")
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
	if err := j.BindAPIJson(map[string]string{"tree": query}, &folderJson); err != nil {
		return nil, err
	}
	var jobs []*JobItem
	var _resolve func(item *Job)
	_resolve = func(item *Job) {
		for _, job := range item.Jobs {
			if len(job.Jobs) > 0 {
				_resolve(job)
			}
			jobs = append(jobs, NewJobItem(job.URL, job.Class, j.Client))
		}
	}
	_resolve(&folderJson)
	return jobs, nil
}

func (j *JobItem) GetFirstBuild() (*BuildItem, error) {
	return j.GetBuildByName("firstBuild")
}
func (j *JobItem) GetLastBuild() (*BuildItem, error) {
	return j.GetBuildByName("lastBuild")
}
func (j *JobItem) GetLastCompleteBuild() (*BuildItem, error) {
	return j.GetBuildByName("lastCompletedBuild")
}
func (j *JobItem) GetLastFailedBuild() (*BuildItem, error) {
	return j.GetBuildByName("lastFailedBuild")
}
func (j *JobItem) GetLastStableBuild() (*BuildItem, error) {
	return j.GetBuildByName("lastStableBuild")
}
func (j *JobItem) GetLastUnstableBuild() (*BuildItem, error) {
	return j.GetBuildByName("lastUnstableBuild")
}
func (j *JobItem) GetLastSuccessfulBuild() (*BuildItem, error) {
	return j.GetBuildByName("lastSuccessfulBuild")
}
func (j *JobItem) GetLastUnsucessfulBuild() (*BuildItem, error) {
	return j.GetBuildByName("lastUnsuccessfulBuild")
}

func (j *JobItem) GetBuildByName(name string) (*BuildItem, error) {
	if j.Class == "Folder" || j.Class == "WorkflowMultiBranchProject" {
		return nil, fmt.Errorf("%s have no builds", j)
	}
	var jobJson map[string]json.RawMessage
	if err := j.BindAPIJson(map[string]string{"tree": name + "[url]"}, &jobJson); err != nil {
		return nil, err
	}
	if string(jobJson[name]) == "null" {
		return nil, nil
	}
	build := &Build{}
	if err := json.Unmarshal(jobJson[name], build); err != nil {
		return nil, err
	}
	return NewBuildItem(build.URL, build.Class, j.Client), nil
}

func (j *JobItem) Delete() error {
	_, err := j.R().Post("doDelete")
	return err
}

func (j *JobItem) ListBuilds() ([]*BuildItem, error) {
	if j.Class == "Folder" || j.Class == "WorkflowMultiBranchProject" {
		return nil, fmt.Errorf("%s have no builds", j)
	}
	var jobJson Job
	var builds []*BuildItem
	if err := j.BindAPIJson(map[string]string{"tree": "builds[url]"}, &jobJson); err != nil {
		return nil, err
	}

	for _, build := range jobJson.Builds {
		builds = append(builds, NewBuildItem(build.URL, build.Class, j.Client))
	}
	return builds, nil
}

func (j *JobItem) SetNextBuildNumber(number string) error {
	_, err := j.R().SetPathParam("nextBuildNumber", number).Post("nextbuildnumber/submit")
	return err
}

func (j *JobItem) GetParameters() ([]*ParameterDefinition, error) {
	jobJson := &Job{}
	if err := j.BindAPIJson(nil, jobJson); err != nil {
		return nil, err
	}
	for _, p := range jobJson.Property {
		if p.Class == "hudson.model.ParametersDefinitionProperty" {
			return p.ParameterDefinitions, nil
		}
	}
	return nil, nil
}

func (j *JobItem) SCMPolling() error {
	_, err := j.R().Post("polling")
	return err
}

func (j *JobItem) GetMultibranchPipelineScanLog() (string, error) {
	if j.Class != "WorkflowMultiBranchProject" {
		return "", fmt.Errorf("%s is not a WorkflowMultiBranchProject", j)
	}
	resp, err := j.R().Post("indexing/consoleText")
	if err != nil {
		return "", err
	}
	return resp.String(), nil
}
