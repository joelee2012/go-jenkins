package jenkins

import (
	"fmt"
	"log"
	"path"
	"strings"

	"github.com/imroc/req/v3"
)

type APIError struct {
	Servlet string `json:"servlet"`
	Message string `json:"message"`
	URL     string `json:"url"`
	Status  string `json:"status"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API Error: %s , (refer to %s), Code: %s", e.Message, e.URL, e.Status)
}

var reqClient *req.Client

func Init() {
	reqClient = req.NewClient()
}

func R() *req.Request {
	return reqClient.R()
}

func doDelete(URL string) error {
	_, err := R().Post(URL + "doDelete")
	return err
}

type Jenkins struct {
	URL         string
	Credentials *CredentialService
	Crumb       *Crumb
	Nodes       *NodeService
	Queue       *QueueService
	// Views       *ViewService
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
//		client, err := jenkins.NewJenkins("http://localhost:8080/", "admin", "1234")
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
func NewJenkins(url, user, password string) (*Jenkins, error) {
	jenkins := &Jenkins{
		URL: appendSlash(url),
	}

	reqClient.SetUserAgent("Go-Jenkins").
		SetRedirectPolicy(req.NoRedirectPolicy()).
		SetCommonBasicAuth(user, password).
		SetCommonHeader("Accept", "application/json").
		SetCommonErrorResult(&APIError{}).
		OnBeforeRequest(func(client *req.Client, req *req.Request) error {
			if strings.HasSuffix(req.RawURL, "crumbIssuer/api/json") {
				return nil
			}
			if _, err := jenkins.GetCrumb(); err != nil {
				return err
			}
			return nil
		}).
		OnAfterResponse(func(client *req.Client, resp *req.Response) error {
			if err, ok := resp.Error().(*APIError); ok {
				log.Println(err)
				resp.Err = err
				return err
			}

			if !resp.IsSuccess() {
				resp.Err = fmt.Errorf("bad response, raw content:\n%s", resp.Dump())
				return nil
			}
			return nil
		})

	jenkins.Credentials = NewCredentialService(jenkins)
	jenkins.Nodes = NewNodeService(jenkins)
	jenkins.Queue = NewQueueService(jenkins)
	// c.Views = NewViewService(c)
	return jenkins, nil
}

func (j *Jenkins) GetCrumb() (*Crumb, error) {
	if j.Crumb != nil {
		return j.Crumb, nil
	}
	_, err := R().SetSuccessResult(&j.Crumb).Get(j.URL + "/crumbIssuer/api/json")
	reqClient.SetCommonHeader(j.Crumb.RequestFields, j.Crumb.Value)
	return j.Crumb, err
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
func (c *Jenkins) CreateJob(fullName, xml string) error {
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
	resp, err := R().Get(c.URL)
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
func (c *Jenkins) BuildJob(fullName string, params map[string]string) (*OneQueueItem, error) {
	return NewJobItem(c.Name2URL(fullName), "Job", c).Build(params)
}

// List job with depth
func (c *Jenkins) ListJobs(depth int) ([]*JobItem, error) {
	job := NewJobItem(c.URL, "Folder", c)
	return job.List(depth)
}

func (c *Jenkins) Restart() error {
	_, err := R().Post("/restart")
	return err
}

func (c *Jenkins) SafeRestart() error {
	_, err := R().Post("/safeRestart")
	return err
}

func (c *Jenkins) Exit() error {
	_, err := R().Post("/exit")
	return err
}

func (c *Jenkins) SafeExit() error {
	_, err := R().Post("/safeExit")
	return err
}

func (c *Jenkins) QuiteDown() error {
	_, err := R().Post("/quietDown")
	return err
}

func (c *Jenkins) CancelQuiteDown() error {
	_, err := R().Post("/cancelQuietDown")
	return err
}

func (c *Jenkins) ReloadJCasC() error {
	_, err := R().Post("/configuration-as-code/reload")
	return err
}

func (c *Jenkins) ExportJCasC(name string) error {
	_, err := R().SetOutputFile(name).Get(c.URL + "/configuration-as-code/export")
	return err
}

func (c *Jenkins) ValidateJenkinsfile(content string) (string, error) {
	resp, err := R().SetQueryParam("jenkinsfile", content).Post(c.URL + "/pipeline-model-converter/validate")
	if err != nil {
		return "", err
	}
	return resp.String(), nil
}

func (c *Jenkins) RunScript(script string) (string, error) {
	resp, err := R().SetQueryParam("script", script).Post(c.URL + "/scriptText")
	if err != nil {
		return "", err
	}
	return resp.String(), nil
}
