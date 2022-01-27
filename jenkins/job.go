package jenkins

import (
	"encoding/json"
	"fmt"
	"path"
	"strings"

	"github.com/imroc/req"
)

type JobService struct {
	*Item
	Credentials *CredentialService
}

type JobShortJson struct {
	Class  string           `json:"_class"`
	Builds []BuildShortJson `json:"builds"`
	Name   string           `json:"name"`
	URL    string           `json:"url"`
	Jobs   []JobShortJson   `json:"jobs"`
}

func NewJobService(url, class string, client *Client) *JobService {
	j := &JobService{Item: NewItem(url, class, client)}
	j.Credentials = NewCredentialService(j)
	return j
}

func (j *JobService) Rename(name string) error {
	resp, err := j.Request("POST", "confirmRename", ReqParams{"newName": name})
	if err != nil {
		return err
	}
	url, _ := resp.Response().Location()
	j.URL = appendSlash(url.String())
	return nil
}

func (j *JobService) Move(path string) error {
	path = strings.Trim(path, "/")
	resp, err := j.Request("POST", "move/move", ReqParams{"destination": "/" + path})
	if err != nil {
		return err
	}
	url, _ := resp.Response().Location()
	j.URL = appendSlash(url.String())
	return nil
}

func (j *JobService) Copy(src, dest string) error {
	_, err := j.Request("POST", "createItem", ReqParams{"name": dest, "mode": "copy", "from": src})
	return err
}

func (j *JobService) GetParent() (*JobService, error) {
	fullName, _ := j.client.URLToName(j.URL)
	dir, _ := path.Split(strings.Trim(fullName, "/"))
	if dir == "" {
		return nil, nil
	}
	return j.client.GetJob(dir)
}

func (j *JobService) GetConfigure() (string, error) {
	resp, err := j.Request("GET", "/config.xml")
	return resp.String(), err
}

func (j *JobService) SetConfigure(xml string) error {
	_, err := j.Request("POST", "/config.xml", req.BodyXML(xml))
	return err
}

func (j *JobService) Disable() error {
	_, err := j.Request("POST", "disable")
	return err
}

func (j *JobService) Enable() error {
	_, err := j.Request("POST", "enable")
	return err
}

func (j *JobService) IsBuildable() (bool, error) {
	var apiJson struct {
		Class     string `json:"_class"`
		Buildable bool   `json:"buildable"`
	}
	err := j.BindAPIJson(ReqParams{"tree": "buildable"}, &apiJson)
	return apiJson.Buildable, err
}

func (j *JobService) GetName() string {
	_, name := path.Split(strings.Trim(j.URL, "/"))
	return name
}

func (j *JobService) GetFullName() string {
	fullname, _ := j.client.URLToName(j.URL)
	return fullname
}

func (j *JobService) GetFullDisplayName() string {
	fullname, _ := j.client.URLToName(j.URL)
	return strings.ReplaceAll(fullname, "/", " » ")
}

func (j *JobService) GetDescription() (string, error) {
	data := make(map[string]string)
	if err := j.BindAPIJson(ReqParams{"tree": "description"}, &data); err != nil {
		return "", err
	}
	return data["description"], nil
}

func (j *JobService) SetDescription(description string) error {
	_, err := j.Request("POST", "submitDescription", ReqParams{"description": description})
	return err
}

func (j *JobService) Build(param ReqParams) (*QueueItem, error) {
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

func (j *JobService) GetBuild(number int) (*BuildService, error) {
	if j.Class == "Folder" || j.Class == "WorkflowMultiBranchProject" {
		return nil, fmt.Errorf("%s is a folder", j)
	}
	var jobJson JobShortJson
	if err := j.BindAPIJson(ReqParams{"tree": "builds[number,url]"}, &jobJson); err != nil {
		return nil, err
	}

	for _, build := range jobJson.Builds {
		if number == build.Number {
			return NewBuild(build.URL, parseClass(build.Class), j.client), nil
		}
	}
	return nil, nil
}

func (j *JobService) Get(name string) (*JobService, error) {
	var folderJson JobShortJson
	if err := j.BindAPIJson(ReqParams{"tree": "jobs[url,name]"}, &folderJson); err != nil {
		return nil, err
	}
	for _, job := range folderJson.Jobs {
		if job.Name == name {
			return NewJobService(job.URL, job.Class, j.client), nil
		}
	}
	return nil, fmt.Errorf("%s does not contain job: %s", j, name)
}

func (j *JobService) Create(name, xml string) error {
	_, err := j.Request("POST", "createItem", ReqParams{"name": name}, req.BodyXML(xml))
	return err
}

func (j *JobService) List(depth int) ([]*JobService, error) {
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
	var jobs []*JobService
	var _resolve func(item *JobShortJson)
	_resolve = func(item *JobShortJson) {
		for _, job := range item.Jobs {
			if len(job.Jobs) > 0 {
				_resolve(&job)
			}
			jobs = append(jobs, NewJobService(job.URL, job.Class, j.client))
		}
	}
	_resolve(&folderJson)
	return jobs, nil
}

func (j *JobService) GetFirstBuild() (*BuildService, error) {
	return j.getBuildByName("firstBuild")
}
func (j *JobService) GetLastBuild() (*BuildService, error) {
	return j.getBuildByName("lastBuild")
}
func (j *JobService) GetLastCompleteBuild() (*BuildService, error) {
	return j.getBuildByName("lastCompletedBuild")
}
func (j *JobService) GetLastFailedBuild() (*BuildService, error) {
	return j.getBuildByName("lastFailedBuild")
}
func (j *JobService) GetLastStableBuild() (*BuildService, error) {
	return j.getBuildByName("lastStableBuild")
}
func (j *JobService) GetLastUnstableBuild() (*BuildService, error) {
	return j.getBuildByName("lastUnstableBuild")
}
func (j *JobService) GetLastSuccessfulBuild() (*BuildService, error) {
	return j.getBuildByName("lastSuccessfulBuild")
}
func (j *JobService) GetLastUnsucessfulBuild() (*BuildService, error) {
	return j.getBuildByName("lastUnsuccessfulBuild")
}

func (j *JobService) getBuildByName(name string) (*BuildService, error) {
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
	return NewBuild(build.URL, build.Class, j.client), nil
}

func (j *JobService) Delete() error {
	_, err := j.Request("POST", "doDelete")
	return err
}

func (j *JobService) ListBuilds() ([]*BuildService, error) {
	if j.Class == "Folder" || j.Class == "WorkflowMultiBranchProject" {
		return nil, fmt.Errorf("%s is a folder", j)
	}
	var jobJson JobShortJson
	var builds []*BuildService
	if err := j.BindAPIJson(ReqParams{"tree": "builds[url]"}, &jobJson); err != nil {
		return nil, err
	}

	for _, build := range jobJson.Builds {
		builds = append(builds, NewBuild(build.URL, parseClass(build.Class), j.client))
	}
	return builds, nil
}
