package jenkins

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"slices"
	"strings"
)

type JobItem struct {
	*Item
	credentials     *CredentialService
	views           *ViewService
	Name            string
	FullName        string
	FullDisplayName string
}

func NewJobItem(url, class string, client *Jenkins) *JobItem {
	j := &JobItem{Item: NewItem(url, class, client)}
	j.setName()
	return j
}

func (j *JobItem) Views() *ViewService {
	if j.views == nil {
		j.views = &ViewService{Item: NewItem(j.URL, "Views", j.jenkins)}
	}
	return j.views
}

func (j *JobItem) Credentials() *CredentialService {
	if j.credentials == nil {
		j.credentials = &CredentialService{Item: NewItem(j.URL+"credentials/store/folder/domain/_/", "Credentials", j.jenkins)}
	}
	return j.credentials
}

func (j *JobItem) Rename(name string) error {
	v := url.Values{}
	v.Add("newName", name)
	resp, err := j.Request("POST", "confirmRename?"+v.Encode(), nil)
	if err != nil {
		return err
	}
	url, _ := resp.Location()
	j.URL = appendSlash(url.String())
	j.setName()
	return nil
}

func (j *JobItem) Move(path string) error {
	v := url.Values{}
	v.Add("destination", "/"+strings.Trim(path, "/"))
	resp, err := j.Request("POST", "move/move?"+v.Encode(), nil)
	if err != nil {
		return err
	}
	url, _ := resp.Location()
	j.URL = appendSlash(url.String())
	j.setName()
	return nil
}

func (j *JobItem) Copy(src, dest string) error {
	v := url.Values{}
	v.Add("name", dest)
	v.Add("mode", "copy")
	v.Add("from", src)
	_, err := j.Request("POST", "createItem?"+v.Encode(), nil)
	return err
}

func (j *JobItem) GetParent() (*JobItem, error) {
	fullName, _ := j.jenkins.URL2Name(j.URL)
	dir, _ := path.Split(strings.Trim(fullName, "/"))
	if dir == "" {
		return nil, nil
	}
	return j.jenkins.GetJob(dir)
}

func (j *JobItem) GetConfigure() (string, error) {
	return readResponseToString(j, "GET", "config.xml", nil)
}

func (j *JobItem) SetConfigure(xml io.Reader) (*http.Response, error) {
	return j.Request("POST", "config.xml", xml)
}

func (j *JobItem) Disable() (*http.Response, error) {
	return j.Request("POST", "disable", nil)
}

func (j *JobItem) Enable() (*http.Response, error) {
	return j.Request("POST", "enable", nil)
}

func (j *JobItem) IsBuildable() (bool, error) {
	var job struct {
		Class     string `json:"_class"`
		Buildable bool   `json:"buildable"`
	}
	err := j.ApiJson(&job, &ApiJsonOpts{Tree: "buildable"})
	return job.Buildable, err
}

func (j *JobItem) setName() {
	urlPath, _ := j.jenkins.URL2Name(j.URL)
	j.FullName, _ = url.PathUnescape(urlPath)
	_, j.Name = path.Split(j.FullName)
	j.FullDisplayName, _ = url.PathUnescape(strings.ReplaceAll(j.FullName, "/", " » "))
}

func (j *JobItem) GetDescription() (string, error) {
	data := make(map[string]string)
	if err := j.ApiJson(&data, &ApiJsonOpts{Tree: "description"}); err != nil {
		return "", err
	}
	return data["description"], nil
}

func (j *JobItem) SetDescription(description string) error {
	v := url.Values{}
	v.Add("description", description)
	_, err := j.Request("POST", "submitDescription?"+v.Encode(), nil)
	return err
}

func (j *JobItem) Build(param url.Values) (*OneQueueItem, error) {
	entry := func() string {
		reserved := []string{"token", "delay"}
		for k := range param {
			if !slices.Contains(reserved, k) {
				return "buildWithParameters"
			}
		}
		return "build"
	}()

	resp, err := j.Request("POST", entry+"?"+param.Encode(), nil)
	if err != nil {
		return nil, err
	}
	url, err := resp.Location()
	if err != nil {
		return nil, err
	}
	return NewQueueItem(url.String(), j.jenkins), nil
}

func (j *JobItem) GetBuild(number int) (*BuildItem, error) {
	if j.Class == "Folder" || j.Class == "WorkflowMultiBranchProject" {
		return nil, fmt.Errorf("%s have no builds", j)
	}
	jobJson := &Job{}
	if err := j.ApiJson(&jobJson, &ApiJsonOpts{Tree: "builds[number,url]"}); err != nil {
		return nil, err
	}

	for _, build := range jobJson.Builds {
		if number == build.Number {
			return NewBuildItem(build.URL, build.Class, j.jenkins), nil
		}
	}
	return nil, nil
}

func (j *JobItem) Get(name string) (*JobItem, error) {
	if j.Class != "Folder" && j.Class != "WorkflowMultiBranchProject" {
		return nil, fmt.Errorf("%s have no jobs", j)
	}
	var folderJson Job
	if err := j.ApiJson(&folderJson, &ApiJsonOpts{Tree: "jobs[url,name]"}); err != nil {
		return nil, err
	}
	for _, job := range folderJson.Jobs {
		if job.Name == name {
			return NewJobItem(job.URL, job.Class, j.jenkins), nil
		}
	}
	return nil, nil
}

func (j *JobItem) Create(name string, xml io.Reader) (*http.Response, error) {
	v := url.Values{}
	v.Add("name", name)
	return j.Request("POST", "createItem?"+v.Encode(), xml)
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

	if err := j.ApiJson(&folderJson, &ApiJsonOpts{Tree: query}); err != nil {
		return nil, err
	}
	var jobs []*JobItem
	var _resolve func(item *Job)
	_resolve = func(item *Job) {
		for _, job := range item.Jobs {
			if len(job.Jobs) > 0 {
				_resolve(job)
			}
			jobs = append(jobs, NewJobItem(job.URL, job.Class, j.jenkins))
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
	if err := j.ApiJson(&jobJson, &ApiJsonOpts{Tree: name + "[url]"}); err != nil {
		return nil, err
	}
	if string(jobJson[name]) == "null" {
		return nil, nil
	}
	build := &Build{}
	if err := json.Unmarshal(jobJson[name], build); err != nil {
		return nil, err
	}
	return NewBuildItem(build.URL, build.Class, j.jenkins), nil
}

func (j *JobItem) Delete() error {
	_, err := j.Request("POST", "doDelete", nil)
	return err
}

func (j *JobItem) ListBuilds() ([]*BuildItem, error) {
	if j.Class == "Folder" || j.Class == "WorkflowMultiBranchProject" {
		return nil, fmt.Errorf("%s have no builds", j)
	}
	var jobJson Job
	var builds []*BuildItem
	if err := j.ApiJson(&jobJson, &ApiJsonOpts{Tree: "builds[url]"}); err != nil {
		return nil, err
	}

	for _, build := range jobJson.Builds {
		builds = append(builds, NewBuildItem(build.URL, build.Class, j.jenkins))
	}
	return builds, nil
}

func (j *JobItem) SetNextBuildNumber(number int) (*http.Response, error) {
	return j.Request("POST", fmt.Sprintf("nextbuildnumber/submit?nextBuildNumber=%d", number), nil)
}

func (j *JobItem) GetParameters() ([]*ParameterDefinition, error) {
	jobJson := &Job{}
	if err := j.ApiJson(jobJson, nil); err != nil {
		return nil, err
	}
	for _, p := range jobJson.Property {
		if p.Class == "hudson.model.ParametersDefinitionProperty" {
			return p.ParameterDefinitions, nil
		}
	}
	return nil, nil
}

func (j *JobItem) SCMPolling() (*http.Response, error) {
	return j.Request("POST", "polling", nil)
}

func (j *JobItem) GetMultibranchPipelineScanLog() (string, error) {
	if j.Class != "WorkflowMultiBranchProject" {
		return "", fmt.Errorf("%s is not a WorkflowMultiBranchProject", j)
	}
	return readResponseToString(j, "POST", "indexing/consoleText", nil)
}
