package jenkins

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"path"
	"strings"

	"github.com/imroc/req"
)

type Jenkins struct {
	URL    string
	Crumb  *Crumb
	Header http.Header
	Req    *req.Req
}

type Crumb struct {
	RequestFields string `json:"crumbRequestField"`
	Value         string `json:"crumb"`
}

func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

// Create new Jenkins instance
// J, err = NewJenkins(os.Getenv("JENKINS_URL"), os.Getenv("JENKINS_USER"), os.Getenv("JENKINS_PASSWORD"))
// if err != nil {
// 	return err
// }
// fmt.Println(J)
func NewJenkins(url, user, password string) (*Jenkins, error) {
	url = appendSlash(url)
	header := make(http.Header)
	header.Set("Accept", "application/json")
	header.Set("Authorization", "Basic "+basicAuth(user, password))

	j := Jenkins{URL: url, Crumb: nil, Header: header, Req: req.New()}
	// disable redirect for Job.Rename() and Move()
	j.Req.Client().CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
	return &j, nil
}

func (j *Jenkins) GetCrumb() (*Crumb, error) {
	if j.Crumb != nil {
		return j.Crumb, nil
	}
	resp, err := j.Req.Get(j.URL+"crumbIssuer/api/json", j.Header)
	if err != nil {
		return nil, err
	}
	if resp.Response().StatusCode != http.StatusOK {
		return nil, fmt.Errorf(resp.String())
	}
	if err := resp.ToJSON(&j.Crumb); err != nil {
		return nil, err
	}
	j.Header.Set(j.Crumb.RequestFields, j.Crumb.Value)
	return j.Crumb, nil
}

func (j *Jenkins) GetJob(fullName string) (*Job, error) {
	folder, shortName := j.resolveJob(fullName)
	return folder.Get(shortName)
}

func (j *Jenkins) CreateJob(fullName, xml string) error {
	folder, shortName := j.resolveJob(fullName)
	return folder.Create(shortName, xml)
}

func (j *Jenkins) DeleteJob(fullName string) error {
	folder, shortName := j.resolveJob(fullName)
	job, err := folder.Get(shortName)
	if err != nil {
		return err
	}
	return job.Delete()
}

func (j *Jenkins) String() string {
	return fmt.Sprintf("<Jenkins: %s>", j.URL)
}

func (j *Jenkins) resolveJob(fullName string) (*Job, string) {
	dir, base := path.Split(strings.Trim(fullName, "/"))
	url := j.NameToURL(dir)
	return NewJob(url, "Folder", j), base
}

func (j *Jenkins) NameToURL(fullName string) string {
	if fullName == "" {
		return j.URL
	}
	path := strings.ReplaceAll(strings.Trim(fullName, "/"), "/", "/job/")
	return appendSlash(j.URL + "job/" + path)
}

func (j *Jenkins) URLToName(url string) (string, error) {
	if !strings.HasPrefix(url, j.URL) {
		return "", fmt.Errorf("%s is not in %s", url, j.URL)
	}
	path := strings.ReplaceAll(url, j.URL, "/")
	return strings.Trim(strings.ReplaceAll(path, "/job/", "/"), "/"), nil
}

func (j *Jenkins) ComputerSet() *ComputerSet {
	return NewComputerSet(j.URL+"computer/", j)
}

func (j *Jenkins) GetVersion() (string, error) {
	resp, err := j.Req.Get(j.URL)
	if err != nil {
		return "", err
	}
	return resp.Response().Header.Get("X-Jenkins"), nil
}

func (j *Jenkins) BuildJob(fullName string, params ReqParams) (*QueueItem, error) {
	job, err := j.GetJob(fullName)
	if err != nil {
		return nil, err
	}
	return job.Build(params)
}

func (j *Jenkins) ListJobs(depth int) ([]*Job, error) {
	job := NewJob(j.URL, "Folder", j)
	return job.List(depth)
}

func (j *Jenkins) Credentials() *Credentials {
	return &Credentials{Item: NewItem(j.URL+"credentials/store/system/domain/_/", "Credentials", j)}
}

func (j *Jenkins) Request(method, entry string, vs ...interface{}) (*req.Resp, error) {
	return doRequest(j, method, j.URL+entry, vs...)
}

func (j *Jenkins) Restart() error {
	return doRequestAndDropResp(j, "POST", j.URL+"restart")
}

func (j *Jenkins) SafeRestart() error {
	return doRequestAndDropResp(j, "POST", j.URL+"safeRestart")
}

func (j *Jenkins) Exit() error {
	return doRequestAndDropResp(j, "POST", j.URL+"exit")
}

func (j *Jenkins) SafeExit() error {
	return doRequestAndDropResp(j, "POST", j.URL+"safeExit")
}

func (j *Jenkins) QuiteDown() error {
	return doRequestAndDropResp(j, "POST", j.URL+"quietDown")
}

func (j *Jenkins) CancelQuiteDown() error {
	return doRequestAndDropResp(j, "POST", j.URL+"cancelQuietDown")
}

func (j *Jenkins) ReloadJCasC() error {
	return doRequestAndDropResp(j, "POST", j.URL+"configuration-as-code/reload")
}

func (j *Jenkins) ExportJCasC(name string) error {
	resp, err := j.Request("POST", j.URL+"configuration-as-code/export")
	if err != nil {
		return err
	}
	return resp.ToFile(name)
}

func (j *Jenkins) ValidateJenkinsfile(content string) (string, error) {
	data := map[string]string{"jenkinsfile": content}
	resp, err := j.Request("POST", j.URL+"pipeline-model-converter/validate", data)
	if err != nil {
		return "", err
	}
	return resp.String(), nil
}

func (j *Jenkins) RunScript(script string) (string, error) {
	return doRunScript(j, script)
}
