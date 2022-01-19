package jenkins

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"path"
	"strings"

	"github.com/imroc/req"
)

type Client struct {
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

// Create new Client
// client, err = NewClient(os.Getenv("JENKINS_URL"), os.Getenv("JENKINS_USER"), os.Getenv("JENKINS_PASSWORD"))
// if err != nil {
// 	return err
// }
// fmt.Println(client)
func NewClient(url, user, password string) (*Client, error) {
	url = appendSlash(url)
	header := make(http.Header)
	header.Set("Accept", "application/json")
	header.Set("Authorization", "Basic "+basicAuth(user, password))

	c := Client{URL: url, Crumb: nil, Header: header, Req: req.New()}
	// disable redirect for Job.Rename() and Move()
	c.Req.Client().CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
	return &c, nil
}

func (c *Client) GetCrumb() (*Crumb, error) {
	if c.Crumb != nil {
		return c.Crumb, nil
	}
	resp, err := c.Req.Get(c.URL+"crumbIssuer/api/json", c.Header)
	if err != nil {
		return nil, err
	}
	if resp.Response().StatusCode != http.StatusOK {
		return nil, fmt.Errorf(resp.String())
	}
	if err := resp.ToJSON(&c.Crumb); err != nil {
		return nil, err
	}
	c.Header.Set(c.Crumb.RequestFields, c.Crumb.Value)
	return c.Crumb, nil
}

func (c *Client) GetJob(fullName string) (*Job, error) {
	folder, shortName := c.resolveJob(fullName)
	return folder.Get(shortName)
}

func (c *Client) CreateJob(fullName, xml string) error {
	folder, shortName := c.resolveJob(fullName)
	return folder.Create(shortName, xml)
}

func (c *Client) DeleteJob(fullName string) error {
	folder, shortName := c.resolveJob(fullName)
	job, err := folder.Get(shortName)
	if err != nil {
		return err
	}
	return job.Delete()
}

func (c *Client) String() string {
	return fmt.Sprintf("<Jenkins: %s>", c.URL)
}

func (c *Client) resolveJob(fullName string) (*Job, string) {
	dir, base := path.Split(strings.Trim(fullName, "/"))
	url := c.NameToURL(dir)
	return NewJob(url, "Folder", c), base
}

func (c *Client) NameToURL(fullName string) string {
	if fullName == "" {
		return c.URL
	}
	path := strings.ReplaceAll(strings.Trim(fullName, "/"), "/", "/job/")
	return appendSlash(c.URL + "job/" + path)
}

func (c *Client) URLToName(url string) (string, error) {
	if !strings.HasPrefix(url, c.URL) {
		return "", fmt.Errorf("%s is not in %s", url, c.URL)
	}
	path := strings.ReplaceAll(url, c.URL, "/")
	return strings.Trim(strings.ReplaceAll(path, "/job/", "/"), "/"), nil
}

func (c *Client) ComputerSet() *ComputerSet {
	return NewComputerSet(c.URL+"computer/", c)
}

func (c *Client) GetVersion() (string, error) {
	resp, err := c.Req.Get(c.URL)
	if err != nil {
		return "", err
	}
	return resp.Response().Header.Get("X-Jenkins"), nil
}

func (c *Client) BuildJob(fullName string, params ReqParams) (*QueueItem, error) {
	job, err := c.GetJob(fullName)
	if err != nil {
		return nil, err
	}
	return job.Build(params)
}

func (c *Client) ListJobs(depth int) ([]*Job, error) {
	job := NewJob(c.URL, "Folder", c)
	return job.List(depth)
}

func (c *Client) Credentials() *Credentials {
	return &Credentials{Item: NewItem(c.URL+"credentials/store/system/domain/_/", "Credentials", c)}
}

func (c *Client) Request(method, entry string, v ...interface{}) (*req.Resp, error) {
	return doRequest(c, method, c.URL+entry, v...)
}

func (c *Client) Restart() error {
	return doRequestAndDropResp(c, "POST", "restart")
}

func (c *Client) SafeRestart() error {
	return doRequestAndDropResp(c, "POST", "safeRestart")
}

func (c *Client) Exit() error {
	return doRequestAndDropResp(c, "POST", "exit")
}

func (c *Client) SafeExit() error {
	return doRequestAndDropResp(c, "POST", "safeExit")
}

func (c *Client) QuiteDown() error {
	return doRequestAndDropResp(c, "POST", "quietDown")
}

func (c *Client) CancelQuiteDown() error {
	return doRequestAndDropResp(c, "POST", "cancelQuietDown")
}

func (c *Client) ReloadJCasC() error {
	return doRequestAndDropResp(c, "POST", "configuration-as-code/reload")
}

func (c *Client) ExportJCasC(name string) error {
	resp, err := c.Request("POST", "configuration-as-code/export")
	if err != nil {
		return err
	}
	return resp.ToFile(name)
}

func (c *Client) BindAPIJson(params ReqParams, v interface{}) error {
	return doBindAPIJson(c, params, v)
}

func (c *Client) ValidateJenkinsfile(content string) (string, error) {
	resp, err := c.Request("POST", "pipeline-model-converter/validate", ReqParams{"jenkinsfile": content})
	if err != nil {
		return "", err
	}
	return resp.String(), nil
}

func (c *Client) RunScript(script string) (string, error) {
	return doRunScript(c, script)
}
