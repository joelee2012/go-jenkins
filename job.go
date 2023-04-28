package jenkins

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"path"
	"strings"
)

type JobItem struct {
	jenkins     *Jenkins
	URL         string
	Class       string
	Credentials *CredentialService
	// Views           *ViewService
	Name            string
	FullName        string
	FullDisplayName string
}

func NewJobItem(url, class string, jenkins *Jenkins) *JobItem {
	j := &JobItem{URL: url, jenkins: jenkins, Class: class}
	j.Credentials = NewCredentialService(j)
	// j.Views = NewViewService(j)
	j.setName()
	return j
}

func (j *JobItem) Rename(name string) error {
	resp, err := R().SetQueryParam("newName", name).Post(j.URL + "confirmRename")
	if err != nil {
		return err
	}
	url, _ := resp.Location()
	j.URL = appendSlash(url.String())
	j.setName()
	return nil
}

func (j *JobItem) Move(path string) error {
	path = strings.Trim(path, "/")
	resp, err := R().SetQueryParam("destination", "/"+path).Post(j.URL + "move/move")
	if err != nil {
		return err
	}
	url, _ := resp.Location()
	j.URL = appendSlash(url.String())
	j.setName()
	return nil
}

func (j *JobItem) Copy(src, dest string) error {
	_, err := R().SetQueryParams(map[string]string{"name": dest, "mode": "copy", "from": src}).Post(j.URL + "createItem")
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
	return getConfigure(j.URL)
}

func (j *JobItem) SetConfigure(xml string) error {
	return setConfigure(j.URL, xml)
}

func (j *JobItem) Disable() error {
	return disable(j.URL)
}

func (j *JobItem) Enable() error {
	return enable(j.URL)
}

func (j *JobItem) IsBuildable() (bool, error) {
	var job struct {
		Class     string `json:"_class"`
		Buildable bool   `json:"buildable"`
	}
	_, err := R().SetQueryParam("tree", "buildable").SetSuccessResult(job).Get(j.URL + "api/json")
	return job.Buildable, err
}

func (j *JobItem) setName() {
	urlPath, _ := j.jenkins.URL2Name(j.URL)
	j.FullName, _ = url.PathUnescape(urlPath)
	_, j.Name = path.Split(j.FullName)
	j.FullDisplayName, _ = url.PathUnescape(strings.ReplaceAll(j.FullName, "/", " » "))
}

func (j *JobItem) GetDescription() (string, error) {
	return getDescription(j.URL)
}

func (j *JobItem) SetDescription(description string) error {
	return setDescription(j.URL, description)
}

func (j *JobItem) Build(params map[string]string) (*OneQueueItem, error) {
	entry := func() string {
		reserved := []string{"token", "delay"}
		for k := range params {
			for _, e := range reserved {
				if k != e {
					return "/buildWithParameters"
				}
			}
		}
		return "/build"
	}()

	resp, err := R().SetQueryParams(params).Post(j.URL + entry)
	if err != nil {
		return nil, err
	}
	url, err := resp.Location()
	if err != nil {
		return nil, err
	}
	return NewQueueItem(url.String(), j.jenkins), nil
}

func (j *JobItem) String() string {
	return fmt.Sprintf("<%s: %s>", j.Class, j.URL)
}

func (j *JobItem) GetBuild(number int) (*BuildItem, error) {
	if j.Class == "Folder" || j.Class == "WorkflowMultiBranchProject" {
		return nil, fmt.Errorf("%s have no builds", j)
	}
	jobJson := &Job{}

	if _, err := R().SetQueryParam("tree", "builds[number,url]").SetSuccessResult(jobJson).Get(j.URL + "api/json"); err != nil {
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
	log.Println(j.URL)
	if _, err := R().SetQueryParam("tree", "jobs[url,name]").SetSuccessResult(folderJson).Get(j.URL + "api/json"); err != nil {
		return nil, err
	}
	log.Println(folderJson)
	for _, job := range folderJson.Jobs {
		if job.Name == name {
			return NewJobItem(job.URL, job.Class, j.jenkins), nil
		}
	}
	return nil, nil
}

func (j *JobItem) Create(name, xml string) error {
	_, err := R().SetQueryParam("name", name).SetBody(xml).Post(j.URL + "createItem")
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
	if _, err := R().SetQueryParam("tree", query).SetSuccessResult(folderJson).Get(j.URL + "api/json"); err != nil {
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
	log.Println(j)
	if j.Class == "Folder" || j.Class == "WorkflowMultiBranchProject" {
		return nil, fmt.Errorf("%s have no builds", j)
	}
	var jobJson map[string]json.RawMessage

	if _, err := R().SetQueryParam("tree", name+"[url]").SetSuccessResult(jobJson).Get(j.URL + "api/json"); err != nil {
		return nil, err
	}
	if string(jobJson[name]) == "null" {
		return nil, nil
	}
	log.Println(jobJson)
	build := &Build{}
	if err := json.Unmarshal(jobJson[name], build); err != nil {
		return nil, err
	}
	return NewBuildItem(build.URL, build.Class, j.jenkins), nil
}

func (j *JobItem) Delete() error {
	return doDelete(j.URL)
}

func (j *JobItem) ListBuilds() ([]*BuildItem, error) {
	if j.Class == "Folder" || j.Class == "WorkflowMultiBranchProject" {
		return nil, fmt.Errorf("%s have no builds", j)
	}
	var jobJson Job
	var builds []*BuildItem

	if _, err := R().SetQueryParam("tree", "builds[url]").SetSuccessResult(jobJson).Get(j.URL + "api/json"); err != nil {
		return nil, err
	}

	for _, build := range jobJson.Builds {
		builds = append(builds, NewBuildItem(build.URL, build.Class, j.jenkins))
	}
	return builds, nil
}

func (j *JobItem) SetNextBuildNumber(number string) error {
	_, err := R().SetPathParam("nextBuildNumber", number).Post("nextbuildnumber/submit")
	return err
}

func (j *JobItem) GetParameters() ([]*ParameterDefinition, error) {
	jobJson := &Job{}

	if _, err := R().SetSuccessResult(jobJson).Get(j.URL + "api/json"); err != nil {
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
	_, err := R().Post("polling")
	return err
}

func (j *JobItem) GetMultibranchPipelineScanLog() (string, error) {
	if j.Class != "WorkflowMultiBranchProject" {
		return "", fmt.Errorf("%s is not a WorkflowMultiBranchProject", j)
	}
	resp, err := R().Post("indexing/consoleText")
	if err != nil {
		return "", err
	}
	return resp.String(), nil
}
