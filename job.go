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

type Job struct {
	*Item
	credentials     *Credentials
	views           *Views
	Name            string
	FullName        string
	FullDisplayName string
}

func NewJob(url, class string, jenkins *Jenkins) *Job {
	j := &Job{Item: NewItem(url, class, jenkins)}
	j.setName()
	return j
}

func (j *Job) Views() *Views {
	if j.views == nil {
		j.views = &Views{Item: NewItem(j.URL, "Views", j.jenkins)}
	}
	return j.views
}

func (j *Job) Credentials() *Credentials {
	if j.credentials == nil {
		j.credentials = &Credentials{Item: NewItem(j.URL+"credentials/store/folder/domain/_/", "Credentials", j.jenkins)}
	}
	return j.credentials
}

func (j *Job) Rename(name string) (newUrl *url.URL, err error) {
	v := url.Values{}
	v.Add("newName", name)
	resp, err := j.Request("POST", "confirmRename?"+v.Encode(), nil)
	if err != nil {
		return nil, err
	}

	defer func() {
		j.URL = appendSlash(newUrl.String())
		j.setName()
	}()

	return resp.Location()
}

func (j *Job) Move(path string) (newUrl *url.URL, err error) {
	v := url.Values{}
	v.Add("destination", "/"+strings.Trim(path, "/"))
	resp, err := j.Request("POST", "move/move?"+v.Encode(), nil)
	if err != nil {
		return
	}

	defer func() {
		j.URL = appendSlash(newUrl.String())
		j.setName()
	}()

	return resp.Location()
}

func (j *Job) Copy(src, dest string) (*http.Response, error) {
	v := url.Values{}
	v.Add("name", dest)
	v.Add("mode", "copy")
	v.Add("from", src)
	return j.Request("POST", "createItem?"+v.Encode(), nil)
}

func (j *Job) GetParent() (*Job, error) {
	fullName, _ := j.jenkins.URL2Name(j.URL)
	dir, _ := path.Split(strings.Trim(fullName, "/"))
	if dir == "" {
		return nil, fmt.Errorf("%s have no parent", j)
	}
	return j.jenkins.GetJob(dir)
}

func (j *Job) GetConfigure() (string, error) {
	return readResponseToString(j, "GET", "config.xml", nil)
}

func (j *Job) SetConfigure(xml io.Reader) (*http.Response, error) {
	return j.Request("POST", "config.xml", xml)
}

func (j *Job) Disable() (*http.Response, error) {
	return j.Request("POST", "disable", nil)
}

func (j *Job) Enable() (*http.Response, error) {
	return j.Request("POST", "enable", nil)
}

func (j *Job) IsBuildable() (bool, error) {
	var job struct {
		Class     string `json:"_class"`
		Buildable bool   `json:"buildable"`
	}
	err := j.ApiJson(&job, &ApiJsonOpts{Tree: "buildable"})
	return job.Buildable, err
}

func (j *Job) setName() {
	urlPath, _ := j.jenkins.URL2Name(j.URL)
	j.FullName, _ = url.PathUnescape(urlPath)
	_, j.Name = path.Split(j.FullName)
	j.FullDisplayName, _ = url.PathUnescape(strings.ReplaceAll(j.FullName, "/", " » "))
}

func (j *Job) GetDescription() (string, error) {
	data := make(map[string]string)
	if err := j.ApiJson(&data, &ApiJsonOpts{Tree: "description"}); err != nil {
		return "", err
	}
	return data["description"], nil
}

func (j *Job) SetDescription(description string) (*http.Response, error) {
	v := url.Values{}
	v.Add("description", description)
	return j.Request("POST", "submitDescription?"+v.Encode(), nil)
}

func (j *Job) Build(param url.Values) (*OneQueueItem, error) {
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

func (j *Job) GetBuild(number int) (*Build, error) {
	if j.Class == "Folder" || j.Class == "WorkflowMultiBranchProject" {
		return nil, fmt.Errorf("%s have no builds", j)
	}
	jobJson := &JobJson{}
	if err := j.ApiJson(&jobJson, &ApiJsonOpts{Tree: "builds[number,url]"}); err != nil {
		return nil, err
	}

	for _, build := range jobJson.Builds {
		if number == build.Number {
			return NewBuild(build.URL, build.Class, j.jenkins), nil
		}
	}
	return nil, fmt.Errorf("%s have no builds #%d", j, number)
}

func (j *Job) Get(name string) (*Job, error) {
	if j.Class != "Folder" && j.Class != "WorkflowMultiBranchProject" {
		return nil, fmt.Errorf("%s have no jobs", j)
	}
	var folderJson JobJson
	if err := j.ApiJson(&folderJson, &ApiJsonOpts{Tree: "jobs[url,name]"}); err != nil {
		return nil, err
	}
	for _, job := range folderJson.Jobs {
		if job.Name == name {
			return NewJob(job.URL, job.Class, j.jenkins), nil
		}
	}
	return nil, fmt.Errorf("no such job [%s%s]", j.URL, name)
}

func (j *Job) Create(name string, xml io.Reader) (*http.Response, error) {
	v := url.Values{}
	v.Add("name", name)
	return j.Request("POST", "createItem?"+v.Encode(), xml)
}

func (j *Job) List(depth int) ([]*Job, error) {
	if j.Class != "Folder" && j.Class != "WorkflowMultiBranchProject" {
		return nil, fmt.Errorf("%s have no jobs", j)
	}
	query := "jobs[url]"
	qf := "jobs[url,%s]"
	for i := 0; i < depth; i++ {
		query = fmt.Sprintf(qf, query)
	}
	var folderJson JobJson

	if err := j.ApiJson(&folderJson, &ApiJsonOpts{Tree: query}); err != nil {
		return nil, err
	}
	var jobs []*Job
	var _resolve func(item *JobJson)
	_resolve = func(item *JobJson) {
		for _, job := range item.Jobs {
			if len(job.Jobs) > 0 {
				_resolve(job)
			}
			jobs = append(jobs, NewJob(job.URL, job.Class, j.jenkins))
		}
	}
	_resolve(&folderJson)
	return jobs, nil
}

func (j *Job) GetFirstBuild() (*Build, error) {
	return j.GetBuildByName("firstBuild")
}
func (j *Job) GetLastBuild() (*Build, error) {
	return j.GetBuildByName("lastBuild")
}
func (j *Job) GetLastCompleteBuild() (*Build, error) {
	return j.GetBuildByName("lastCompletedBuild")
}
func (j *Job) GetLastFailedBuild() (*Build, error) {
	return j.GetBuildByName("lastFailedBuild")
}
func (j *Job) GetLastStableBuild() (*Build, error) {
	return j.GetBuildByName("lastStableBuild")
}
func (j *Job) GetLastUnstableBuild() (*Build, error) {
	return j.GetBuildByName("lastUnstableBuild")
}
func (j *Job) GetLastSuccessfulBuild() (*Build, error) {
	return j.GetBuildByName("lastSuccessfulBuild")
}
func (j *Job) GetLastUnsucessfulBuild() (*Build, error) {
	return j.GetBuildByName("lastUnsuccessfulBuild")
}

func (j *Job) GetBuildByName(name string) (*Build, error) {
	if j.Class == "Folder" || j.Class == "WorkflowMultiBranchProject" {
		return nil, fmt.Errorf("%s have no builds", j)
	}
	var jobJson map[string]json.RawMessage
	if err := j.ApiJson(&jobJson, &ApiJsonOpts{Tree: name + "[url]"}); err != nil {
		return nil, err
	}
	if string(jobJson[name]) == "null" {
		// build is null but no http error
		return nil, nil
	}
	build := &BuildJson{}
	if err := json.Unmarshal(jobJson[name], build); err != nil {
		return nil, err
	}
	return NewBuild(build.URL, build.Class, j.jenkins), nil
}

func (j *Job) Delete() (*http.Response, error) {
	return j.Request("POST", "doDelete", nil)
}

func (j *Job) ListBuilds() ([]*Build, error) {
	if j.Class == "Folder" || j.Class == "WorkflowMultiBranchProject" {
		return nil, fmt.Errorf("%s have no builds", j)
	}
	var jobJson JobJson
	var builds []*Build
	if err := j.ApiJson(&jobJson, &ApiJsonOpts{Tree: "builds[url]"}); err != nil {
		return nil, err
	}

	for _, build := range jobJson.Builds {
		builds = append(builds, NewBuild(build.URL, build.Class, j.jenkins))
	}
	return builds, nil
}

func (j *Job) SetNextBuildNumber(number int) (*http.Response, error) {
	return j.Request("POST", fmt.Sprintf("nextbuildnumber/submit?nextBuildNumber=%d", number), nil)
}

func (j *Job) GetParameters() ([]*ParameterDefinition, error) {
	jobJson := &JobJson{}
	if err := j.ApiJson(jobJson, nil); err != nil {
		return nil, err
	}
	for _, p := range jobJson.Property {
		if p.Class == "hudson.model.ParametersDefinitionProperty" {
			return p.ParameterDefinitions, nil
		}
	}
	return nil, fmt.Errorf("%s has no parameters", j)
}

func (j *Job) SCMPolling() (*http.Response, error) {
	return j.Request("POST", "polling", nil)
}

func (j *Job) GetMultibranchPipelineScanLog() (string, error) {
	if j.Class != "WorkflowMultiBranchProject" {
		return "", fmt.Errorf("%s is not a WorkflowMultiBranchProject", j)
	}
	return readResponseToString(j, "POST", "indexing/consoleText", nil)
}
