package jenkins

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
)

type JenkinsOpts struct {
	URL      string
	User     string
	Password string
	http.Client
}

func (o *JenkinsOpts) DoRequest(url, method string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(o.User, o.Password)
	return o.Do(req)
}

type Jenkins struct {
	URL         string
	User        string
	Password    string
	client      *http.Client
	Header      http.Header
	Crumb       *Crumb
	ctx         *context.Context
	Credentials *CredentialService
	Nodes       *NodeService
	Queue       *QueueService
	Views       *ViewService
}

type Crumb struct {
	RequestFields string `json:"crumbRequestField"`
	Value         string `json:"crumb"`
}

// Init Jenkins client and create job to build
//
//	package main
//
//	import (
//		"log"
//		"time"
//
//		"github.com/joelee2012/go-jenkins/jenkins"
//	)
//
//	func main() {
//		client, err := jenkins.NewClient("http://localhost:8080/", "admin", "1234")
//		if err != nil {
//			log.Fatalln(err)
//		}
//		xml := `<?xml version='1.1' encoding='UTF-8'?>
//		<flow-definition plugin="workflow-job">
//		  <definition class="org.jenkinsci.plugins.workflow.cps.CpsFlowDefinition" plugin="workflow-cps">
//			<script>#!groovy
//				pipeline {
//				agent any
//				stages {
//					stage('build'){
//					steps{
//						sh 'echo $JENKINS_VERSION'
//					}
//					}
//				}
//				}</script>
//			<sandbox>true</sandbox>
//		  </definition>
//		  <disabled>false</disabled>
//		</flow-definition>`
//	  	// create jenkins job
//		if err := client.CreateJob("pipeline", xml); err != nil {
//			log.Fatalln(err)
//		}
//		qitem, err := client.BuildJob("pipeline", nil)
//		if err != nil {
//			log.Fatalln(err)
//		}
//		var build *Build
//		for {
//			time.Sleep(1 * time.Second)
//			build, err = qitem.GetBuild()
//			if err != nil {
//				log.Fatalln(err)
//			}
//			if build != nil {
//				break
//			}
//		}
//		// tail the build log to end
//		build.LoopProgressiveLog("text", func(line string) error {
//			log.Println(line)
//			time.Sleep(1 * time.Second)
//			return nil
//		})
//	}
func NewClient(url, user, password string) (*Jenkins, error) {
	url = appendSlash(url)
	c := &Jenkins{URL: url, Header: make(http.Header)}
	// disable redirect for Job.Rename() and Move()
	c.client = &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	auth := user + ":" + password
	c.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(auth)))
	c.Header.Set("Accept", "application/json")
	c.Credentials = NewCredentialService(c)
	c.Nodes = NewNodeService(c)
	c.Queue = NewQueueService(c)
	c.Views = NewViewService(c)
	return c, nil
}

// Set content type for request, default is 'application/json'
func (c *Jenkins) SetContentType(ctype string) {
	if ctype == "" {
		c.Header.Set("Accept", "application/json")
	} else {
		c.Header.Set("Accept", ctype)
	}
}

func (c *Jenkins) SetBasicAuth(username, password string) {
	auth := username + ":" + password
	c.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(auth)))
}

// func (j *Jenkins) DoRequest(url, method string, body io.Reader) (*http.Response, error) {
// 	req, err := http.NewRequest(method, url, body)
// 	if err != nil {
// 		return nil, err
// 	}
// 	req.SetBasicAuth(j.User, j.Password)
// 	return j.client.Do(req)
// }

func (c *Jenkins) GetCrumb() (*Crumb, error) {
	if c.Crumb != nil {
		return c.Crumb, nil
	}
	req, err := http.NewRequest("GET", c.URL+"crumbIssuer/api/json", nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(c.User, c.Password)
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s: %s", resp.Status, c.URL+"crumbIssuer/api/json")
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(body, &c.Crumb); err != nil {
		return nil, err
	}
	c.Header.Set(c.Crumb.RequestFields, c.Crumb.Value)
	return c.Crumb, nil
}

// Send request to jenkins,
//
//	// send request to get JSON data of jenkins
//	client.Request("GET", "api/json")
func (c *Jenkins) Request(method, entry string, body io.Reader) (*http.Response, error) {
	return c.doRequest(method, c.URL+entry, body)
}

func (c *Jenkins) doRequest(method, url string, body io.Reader) (*http.Response, error) {
	if _, err := c.GetCrumb(); err != nil {
		return nil, err
	}
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	req.Header = c.Header

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("%s: %s", resp.Status, url)
	}
	return resp, nil
}

// Get job with fullname:
//
//	job, err := client.GetJob("path/to/job")
//	if err != nil {
//		return err
//	}
//	fmt.Println(job)
func (c *Jenkins) GetJob(fullName string) (*JobItem, error) {
	folder, shortName := c.resolveJob(fullName)
	return folder.Get(shortName)
}

// Create job with given xml config:
//
//	xml := `<?xml version='1.1' encoding='UTF-8'?>
//	<flow-definition plugin="workflow-job">
//	  <definition class="org.jenkinsci.plugins.workflow.cps.CpsFlowDefinition" plugin="workflow-cps">
//		 <script>#!groovy
//			pipeline {
//			  agent any
//			  stages {
//			    stage('build'){
//				  steps{
//				    echo "test message"
//				  }
//			    }
//			  }
//			}
//	    </script>
//		<sandbox>true</sandbox>
//	  </definition>
//	  <disabled>false</disabled>
//	</flow-definition>`
//	// create jenkins job
//	if err := client.CreateJob("path/to/name", xml); err != nil {
//		log.Fatalln(err)
//	}
func (c *Jenkins) CreateJob(fullName string, xml io.Reader) (*http.Response, error) {
	folder, shortName := c.resolveJob(fullName)
	return folder.Create(shortName, xml)
}

func (c *Jenkins) DeleteJob(fullName string) error {
	return NewJobItem(c.Name2URL(fullName), "Job", c).Delete()
}

func (c *Jenkins) String() string {
	return fmt.Sprintf("<Jenkins: %s>", c.URL)
}

func (c *Jenkins) resolveJob(fullName string) (*JobItem, string) {
	dir, name := path.Split(strings.Trim(fullName, "/"))
	url := c.Name2URL(dir)
	return NewJobItem(url, "Folder", c), name
}

// Covert fullname to url, eg:
//
//	path/to/name -> http://jenkins/job/path/job/to/job/name
func (c *Jenkins) Name2URL(fullName string) string {
	if fullName == "" {
		return c.URL
	}
	path := strings.ReplaceAll(strings.Trim(fullName, "/"), "/", "/job/")
	return appendSlash(c.URL + "job/" + path)
}

// Covert url to full name, eg:
//
//	http://jenkins/job/path/job/to/job/name -> path/to/name
func (c *Jenkins) URL2Name(url string) (string, error) {
	if !strings.HasPrefix(url, c.URL) {
		return "", fmt.Errorf("%s is not in %s", url, c.URL)
	}
	path := strings.ReplaceAll(url, c.URL, "/")
	return strings.Trim(strings.ReplaceAll(path, "/job/", "/"), "/"), nil
}

// Get jenkins version number
func (c *Jenkins) GetVersion() (string, error) {
	resp, err := c.Request("HEAD", "", nil)
	if err != nil {
		return "", err
	}
	return resp.Header.Get("X-Jenkins"), nil
}

// Trigger job to build:
//
//	// without parameters
//	client.BuildJob("your job", nil)
//	client.BuildJob("your job", jenkins.ReqParams{})
//	// with parameters
//	client.BuildJob("your job", jenkins.ReqParams{"ARG1": "ARG1_VALUE"})
func (c *Jenkins) BuildJob(fullName string, params url.Values) (*OneQueueItem, error) {
	return NewJobItem(c.Name2URL(fullName), "Job", c).Build(params)
}

// List job with depth
func (c *Jenkins) ListJobs(depth int) ([]*JobItem, error) {
	job := NewJobItem(c.URL, "Folder", c)
	return job.List(depth)
}

func (c *Jenkins) Restart() (*http.Response, error) {
	return c.Request("POST", "restart", nil)
}

func (c *Jenkins) SafeRestart() (*http.Response, error) {
	return c.Request("POST", "safeRestart", nil)
}

func (c *Jenkins) Exit() (*http.Response, error) {
	return c.Request("POST", "exit", nil)
}

func (c *Jenkins) SafeExit() (*http.Response, error) {
	return c.Request("POST", "safeExit", nil)
}

func (c *Jenkins) QuiteDown() (*http.Response, error) {
	return c.Request("POST", "quietDown", nil)
}

func (c *Jenkins) CancelQuiteDown() (*http.Response, error) {
	return c.Request("POST", "cancelQuietDown", nil)
}

func (c *Jenkins) ReloadJCasC() (*http.Response, error) {
	return c.Request("POST", "configuration-as-code/reload", nil)
}

type ApiJsonOpts struct {
	Tree  string
	Depth int
}

func (o *ApiJsonOpts) Encode() string {
	v := url.Values{}
	v.Add("tree", o.Tree)
	v.Add("depth", strconv.Itoa(o.Depth))
	return v.Encode()
}

// func (c *Jenkins) ExportJCasC(name string) error {
// 	resp, err := c.Request("POST", "configuration-as-code/export")
// 	if err != nil {
// 		return err
// 	}
// 	return resp.ToFile(name)
// }

// Bind jenkins JSON data to interface,
//
//	// bind json data to map
//	data := make(map[string]string)
//	client.BindAPIJson(jenkins.ReqParams{"tree":"description"}, &data)
//	fmt.Println(data["description"])
func (c *Jenkins) BindAPIJson(v interface{}, opts *ApiJsonOpts) error {
	return unmarshalApiJson(c, v, opts)
}

func unmarshalApiJson(r Requester, v interface{}, opts *ApiJsonOpts) error {
	entry := "api/json"
	if opts != nil {
		entry = "api/json?" + opts.Encode()
	}
	resp, err := r.Request("GET", entry, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(data, v); err != nil {
		return err
	}
	return nil
}

func (c *Jenkins) ValidateJenkinsfile(content string) (string, error) {
	data, err := json.Marshal(map[string]string{"jenkinsfile": content})
	if err != nil {
		return "", err
	}
	return readResponseToString(c, "POST", "pipeline-model-converter/validate", bytes.NewBuffer(data))
}
func readResponseToString(r Requester, method, url string, body io.Reader) (string, error) {
	resp, err := r.Request(method, url, body)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (c *Jenkins) RunScript(content string) (string, error) {
	data, err := json.Marshal(map[string]string{"jenkinsfile": content})
	if err != nil {
		return "", err
	}
	return readResponseToString(c, "POST", "scriptText", bytes.NewBuffer(data))
}
