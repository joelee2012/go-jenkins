package jenkins

import (
	"context"
	"fmt"
	"path"
	"strings"

	"github.com/imroc/req/v3"
)

type APIError struct {
	Message          string `json:"message"`
	DocumentationUrl string `json:"documentation_url"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API Error: %s (refer to %s)", e.Message, e.DocumentationUrl)
}

type Client struct {
	Crumb *Crumb
	*req.Client
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
func NewClient(url, user, password string) (*Client, error) {
	reqClient := req.NewClient()
	// disable redirect for Job.Rename() and Move()
	reqClient.SetBaseURL(url).
		SetRedirectPolicy(req.NoRedirectPolicy()).
		SetCommonBasicAuth(user, password).
		SetCommonHeader("Accept", "application/json").
		SetCommonError(&APIError{}).
		OnAfterResponse(func(client *req.Client, resp *req.Response) error {
			if err, ok := resp.Error().(*APIError); ok {
				resp.Err = err
				return err
			}

			if !resp.IsSuccess() {
				resp.Err = fmt.Errorf("bad response, raw content:\n%s", resp.Dump())
				return nil
			}
			return nil
		})
	c := &Client{
		Crumb:  &Crumb{},
		Client: reqClient,
	}
	c.Credentials = NewCredentialService(c)
	c.Nodes = NewNodeService(c)
	c.Queue = NewQueueService(c)
	c.Views = NewViewService(c)
	return c, nil
}

func (c *Client) GetCrumb() (*Crumb, error) {
	if c.Crumb != nil {
		return c.Crumb, nil
	}
	_, err := c.R().SetResult(&c.Crumb).Get("crumbIssuer/api/json")
	c.SetCommonHeader(c.Crumb.RequestFields, c.Crumb.Value)
	return c.Crumb, err
}

// Get job with fullname:
//
//	job, err := client.GetJob("path/to/job")
//	if err != nil {
//		return err
//	}
//	fmt.Println(job)
func (c *Client) GetJob(fullName string) (*JobItem, error) {
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
func (c *Client) CreateJob(fullName, xml string) error {
	folder, shortName := c.resolveJob(fullName)
	return folder.Create(shortName, xml)
}

func (c *Client) DeleteJob(fullName string) error {
	return NewJobItem(c.Name2URL(fullName), "Job", c).Delete()
}

func (c *Client) String() string {
	return fmt.Sprintf("<Jenkins: %s>", c.BaseURL)
}

func (c *Client) resolveJob(fullName string) (*JobItem, string) {
	dir, name := path.Split(strings.Trim(fullName, "/"))
	url := c.Name2URL(dir)
	return NewJobItem(url, "Folder", c), name
}

// Covert fullname to url, eg:
//
//	path/to/name -> http://jenkins/job/path/job/to/job/name
func (c *Client) Name2URL(fullName string) string {
	if fullName == "" {
		return c.BaseURL
	}
	path := strings.ReplaceAll(strings.Trim(fullName, "/"), "/", "/job/")
	return appendSlash(c.BaseURL + "job/" + path)
}

// Covert url to full name, eg:
//
//	http://jenkins/job/path/job/to/job/name -> path/to/name
func (c *Client) URL2Name(url string) (string, error) {
	if !strings.HasPrefix(url, c.BaseURL) {
		return "", fmt.Errorf("%s is not in %s", url, c.BaseURL)
	}
	path := strings.ReplaceAll(url, c.BaseURL, "/")
	return strings.Trim(strings.ReplaceAll(path, "/job/", "/"), "/"), nil
}

// Get jenkins version number
func (c *Client) GetVersion() (string, error) {
	resp, err := c.R().Get("")
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
func (c *Client) BuildJob(fullName string, params map[string]string) (*OneQueueItem, error) {
	return NewJobItem(c.Name2URL(fullName), "Job", c).Build(params)
}

// List job with depth
func (c *Client) ListJobs(depth int) ([]*JobItem, error) {
	job := NewJobItem(c.BaseURL, "Folder", c)
	return job.List(depth)
}

func (c *Client) Restart() error {
	_, err := c.R().Post("restart")
	return err
}

func (c *Client) SafeRestart() error {
	_, err := c.R().Post("safeRestart")
	return err
}

func (c *Client) Exit() error {
	_, err := c.R().Post("exit")
	return err
}

func (c *Client) SafeExit() error {
	_, err := c.R().Post("safeExit")
	return err
}

func (c *Client) QuiteDown() error {
	_, err := c.R().Post("quietDown")
	return err
}

func (c *Client) CancelQuiteDown() error {
	_, err := c.R().Post("cancelQuietDown")
	return err
}

func (c *Client) ReloadJCasC() error {
	_, err := c.R().Post("configuration-as-code/reload")
	return err
}

func (c *Client) ExportJCasC(name string) error {
	_, err := c.R().SetOutputFile(name).Get("configuration-as-code/export")
	return err
}

// Bind jenkins JSON data to interface,
//
//	// bind json data to map
//	data := make(map[string]string)
//	client.BindAPIJson(jenkins.ReqParams{"tree":"description"}, &data)
//	fmt.Println(data["description"])
func (c *Client) BindAPIJson(params map[string]string, v interface{}) error {
	_, err := c.R().SetQueryParams(params).SetResult(v).Get("api/json")
	return err
}

func (c *Client) ValidateJenkinsfile(content string) (string, error) {
	resp, err := c.R().SetQueryParam("jenkinsfile", content).Post("pipeline-model-converter/validate")
	if err != nil {
		return "", err
	}
	return resp.String(), nil
}

func (c *Client) RunScript(script string) (string, error) {
	resp, err := c.R().SetQueryParam("script", script).Post("scriptText")
	if err != nil {
		return "", err
	}
	return resp.String(), nil
}
