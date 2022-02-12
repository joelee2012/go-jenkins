package jenkins

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"path"
	"strings"

	"github.com/imroc/req"
)

type Client struct {
	URL         string
	Header      http.Header
	Crumb       *Crumb
	Req         *req.Req
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
// 		package main
//
// 		import (
// 			"log"
// 			"time"
//
// 			"github.com/joelee2012/go-jenkins/jenkins"
// 		)
//
// 		func main() {
// 			client, err := jenkins.NewClient("http://localhost:8080/", "admin", "1234")
// 			if err != nil {
// 				log.Fatalln(err)
// 			}
// 			xml := `<?xml version='1.1' encoding='UTF-8'?>
// 			<flow-definition plugin="workflow-job">
// 			  <definition class="org.jenkinsci.plugins.workflow.cps.CpsFlowDefinition" plugin="workflow-cps">
// 				<script>#!groovy
// 					pipeline {
// 					agent any
// 					stages {
// 						stage('build'){
// 						steps{
// 							sh 'echo $JENKINS_VERSION'
// 						}
// 						}
// 					}
// 					}</script>
// 				<sandbox>true</sandbox>
// 			  </definition>
// 			  <disabled>false</disabled>
// 			</flow-definition>`
// 		  	// create jenkins job
// 			if err := client.CreateJob("pipeline", xml); err != nil {
// 				log.Fatalln(err)
// 			}
// 			qitem, err := client.BuildJob("pipeline", nil)
// 			if err != nil {
// 				log.Fatalln(err)
// 			}
// 			var build *Build
// 			for {
// 				time.Sleep(1 * time.Second)
// 				build, err = qitem.GetBuild()
// 				if err != nil {
// 					log.Fatalln(err)
// 				}
// 				if build != nil {
// 					break
// 				}
// 			}
// 			// tail the build log to end
// 			build.LoopProgressiveLog("text", func(line string) error {
// 				log.Println(line)
// 				time.Sleep(1 * time.Second)
// 				return nil
// 			})
// 		}
func NewClient(url, user, password string) (*Client, error) {
	url = appendSlash(url)
	c := &Client{URL: url, Header: make(http.Header), Req: req.New()}
	// disable redirect for Job.Rename() and Move()
	c.Req.Client().CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
	c.SetBasicAuth(user, password)
	c.SetContentType("")
	c.Credentials = NewCredentialService(c)
	c.Nodes = NewNodeService(c)
	c.Queue = NewQueueService(c)
	c.Views = NewViewService(c)
	return c, nil
}

// Set content type for request, default is 'application/json'
func (c *Client) SetContentType(ctype string) {
	if ctype == "" {
		c.Header.Set("Accept", "application/json")
	} else {
		c.Header.Set("Accept", ctype)
	}
}

func (c *Client) SetBasicAuth(username, password string) {
	auth := username + ":" + password
	c.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(auth)))
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
		return nil, fmt.Errorf("%s: %s", resp.Response().Status, c.URL+"crumbIssuer/api/json")
	}
	if err := resp.ToJSON(&c.Crumb); err != nil {
		return nil, err
	}
	c.Header.Set(c.Crumb.RequestFields, c.Crumb.Value)
	return c.Crumb, nil
}

// Get job with fullname:
// 		job, err := client.GetJob("path/to/job")
//		if err != nil {
// 			return err
// 		}
// 		fmt.Println(job)
func (c *Client) GetJob(fullName string) (*JobItem, error) {
	folder, shortName := c.resolveJob(fullName)
	return folder.Get(shortName)
}

// Create job with given xml config:
// 		xml := `<?xml version='1.1' encoding='UTF-8'?>
// 		<flow-definition plugin="workflow-job">
// 		  <definition class="org.jenkinsci.plugins.workflow.cps.CpsFlowDefinition" plugin="workflow-cps">
// 			 <script>#!groovy
// 				pipeline {
// 				  agent any
// 				  stages {
// 				    stage('build'){
// 					  steps{
// 					    echo "test message"
// 					  }
// 				    }
// 				  }
// 				}
// 		    </script>
// 			<sandbox>true</sandbox>
// 		  </definition>
// 		  <disabled>false</disabled>
// 		</flow-definition>`
// 		// create jenkins job
// 		if err := client.CreateJob("path/to/name", xml); err != nil {
// 			log.Fatalln(err)
// 		}
func (c *Client) CreateJob(fullName, xml string) error {
	folder, shortName := c.resolveJob(fullName)
	return folder.Create(shortName, xml)
}

func (c *Client) DeleteJob(fullName string) error {
	return NewJobItem(c.Name2URL(fullName), "Job", c).Delete()
}

func (c *Client) String() string {
	return fmt.Sprintf("<Jenkins: %s>", c.URL)
}

func (c *Client) resolveJob(fullName string) (*JobItem, string) {
	dir, name := path.Split(strings.Trim(fullName, "/"))
	url := c.Name2URL(dir)
	return NewJobItem(url, "Folder", c), name
}

// Covert fullname to url, eg:
//		path/to/name -> http://jenkins/job/path/job/to/job/name
func (c *Client) Name2URL(fullName string) string {
	if fullName == "" {
		return c.URL
	}
	path := strings.ReplaceAll(strings.Trim(fullName, "/"), "/", "/job/")
	return appendSlash(c.URL + "job/" + path)
}

// Covert url to full name, eg:
// 		http://jenkins/job/path/job/to/job/name -> path/to/name
func (c *Client) URL2Name(url string) (string, error) {
	if !strings.HasPrefix(url, c.URL) {
		return "", fmt.Errorf("%s is not in %s", url, c.URL)
	}
	path := strings.ReplaceAll(url, c.URL, "/")
	return strings.Trim(strings.ReplaceAll(path, "/job/", "/"), "/"), nil
}

// Get jenkins version number
func (c *Client) GetVersion() (string, error) {
	resp, err := c.Req.Get(c.URL)
	if err != nil {
		return "", err
	}
	return resp.Response().Header.Get("X-Jenkins"), nil
}

// Trigger job to build:
//		// without parameters
//		client.BuildJob("your job", nil)
//		client.BuildJob("your job", jenkins.ReqParams{})
//		// with parameters
//		client.BuildJob("your job", jenkins.ReqParams{"ARG1": "ARG1_VALUE"})
func (c *Client) BuildJob(fullName string, params ReqParams) (*OneQueueItem, error) {
	return NewJobItem(c.Name2URL(fullName), "Job", c).Build(params)
}

// List job with depth
func (c *Client) ListJobs(depth int) ([]*JobItem, error) {
	job := NewJobItem(c.URL, "Folder", c)
	return job.List(depth)
}

// Send request to jenkins,
//		// send request to get JSON data of jenkins
//		client.Request("GET", "api/json")
func (c *Client) Request(method, entry string, v ...interface{}) (*req.Resp, error) {
	return c.Do(method, c.URL+entry, v...)
}

func (c *Client) Do(method, url string, v ...interface{}) (*req.Resp, error) {
	if _, err := c.GetCrumb(); err != nil {
		return nil, err
	}
	v = append(v, c.Header)
	resp, err := c.Req.Do(method, url, v...)
	if err != nil {
		return nil, err
	}
	if resp.Response().StatusCode >= 400 {
		return nil, fmt.Errorf("%s: %s", resp.Response().Status, url)
	}
	return resp, nil
}

func (c *Client) Restart() error {
	_, err := c.Request("POST", "restart")
	return err
}

func (c *Client) SafeRestart() error {
	_, err := c.Request("POST", "safeRestart")
	return err
}

func (c *Client) Exit() error {
	_, err := c.Request("POST", "exit")
	return err
}

func (c *Client) SafeExit() error {
	_, err := c.Request("POST", "safeExit")
	return err
}

func (c *Client) QuiteDown() error {
	_, err := c.Request("POST", "quietDown")
	return err
}

func (c *Client) CancelQuiteDown() error {
	_, err := c.Request("POST", "cancelQuietDown")
	return err
}

func (c *Client) ReloadJCasC() error {
	_, err := c.Request("POST", "configuration-as-code/reload")
	return err
}

func (c *Client) ExportJCasC(name string) error {
	resp, err := c.Request("POST", "configuration-as-code/export")
	if err != nil {
		return err
	}
	return resp.ToFile(name)
}

// Bind jenkins JSON data to interface,
//		// bind json data to map
//		data := make(map[string]string)
//		client.BindAPIJson(jenkins.ReqParams{"tree":"description"}, &data)
//		fmt.Println(data["description"])
func (c *Client) BindAPIJson(params ReqParams, v interface{}) error {
	resp, err := c.Request("GET", "api/json", params)
	if err != nil {
		return err
	}
	return resp.ToJSON(v)
}

func (c *Client) ValidateJenkinsfile(content string) (string, error) {
	resp, err := c.Request("POST", "pipeline-model-converter/validate", ReqParams{"jenkinsfile": content})
	if err != nil {
		return "", err
	}
	return resp.String(), nil
}

func (c *Client) RunScript(script string) (string, error) {
	resp, err := c.Request("POST", "scriptText", ReqParams{"script": script})
	if err != nil {
		return "", err
	}
	return resp.String(), nil
}

func (c *Client) WithContext(ctx context.Context) *Client {
	c.ctx = &ctx
	return c
}
