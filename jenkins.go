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
	Url string
	Crumb
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

func NewJenkins(url, user, password string) (*Jenkins, error) {
	url = AppendSlash(url)
	header := make(http.Header)
	header.Set("Accept", "application/json")
	header.Set("Authorization", "Basic "+basicAuth(user, password))
	client := req.New()
	// req.Debug = true
	resp, err := client.Get(url+"crumbIssuer/api/json", header)
	if err != nil {
		return nil, err
	}
	if resp.Response().StatusCode != 200 {
		return nil, fmt.Errorf(resp.String())
	}
	var crumb Crumb
	if err := resp.ToJSON(&crumb); err != nil {
		return nil, err
	}
	header.Set(crumb.RequestFields, crumb.Value)
	return &Jenkins{Url: url, Crumb: crumb, Header: header, Req: client}, nil
}

func (j *Jenkins) GetJob(fullName string) (Job, error) {
	folder, shortName := j.resolveJob(fullName)
	return folder.GetJob(shortName)
}

func (j *Jenkins) CreateJob(fullName, xml string) error {
	folder, shortName := j.resolveJob(fullName)
	return folder.CreateJob(shortName, xml)
}

func (j *Jenkins) DeleteJob(fullName string) error {
	folder, shortName := j.resolveJob(fullName)
	job, err := folder.GetJob(shortName)
	if err != nil {
		return err
	}
	return job.Delete()
}

func (j *Jenkins) String() string {
	return fmt.Sprintf("<Jenkins: %s>", j.Url)
}

func (j *Jenkins) resolveJob(fullName string) (*Folder, string) {
	dir, base := path.Split(strings.Trim(fullName, "/"))
	url := j.NameToUrl(dir)
	job := NewJob(url, "Folder", j)
	return job.(*Folder), base
}

func (j *Jenkins) NameToUrl(fullName string) string {
	if fullName == "" {
		return j.Url
	}
	path := strings.ReplaceAll(strings.Trim(fullName, "/"), "/", "/job/")
	return AppendSlash(j.Url + "job/" + path)
}

func (j *Jenkins) UrlToName(url string) (string, error) {
	if !strings.HasPrefix(url, j.Url) {
		return "", fmt.Errorf("%s is not in %s", url, j.Url)
	}
	path := strings.ReplaceAll(url, j.Url, "/")
	return strings.Trim(strings.ReplaceAll(path, "/job/", "/"), "/"), nil
}

func (j *Jenkins) GetComputerSet() *ComputerSet {
	return &ComputerSet{
		Item: Item{
			Url:     j.Url + "computer/",
			Class:   "ComputerSet",
			jenkins: j,
		},
	}
}

func (j *Jenkins) GetVersion() (string, error) {
	resp, err := j.Req.Get(j.Url)
	if err != nil {
		return "", err
	}
	return resp.Response().Header.Get("X-Jenkins"), nil
}
