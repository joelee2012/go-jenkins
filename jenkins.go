package jenkins

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
)

type JenkinsOpts struct {
	URL      string
	User     string
	Password string
	http.Client
}

type Jenkins struct {
	*Item
	client      *http.Client
	crumb       *Crumb
	credentials *Credentials
	nodes       *Nodes
	queue       *Queue
	views       *Views
	Header      http.Header
	Debug       bool
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
//		jenkins, err := jenkins.New("http://localhost:8080/", "admin", "1234")
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
//		if err := jenkins.CreateJob("pipeline", xml); err != nil {
//			log.Fatalln(err)
//		}
//		qitem, err := jenkins.BuildJob("pipeline", nil)
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
func New(url, user, password string) (*Jenkins, error) {
	url = appendSlash(url)
	c := &Jenkins{Header: make(http.Header)}
	c.Item = NewItem(url, "Jenkins", c)
	c.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", user, password))))
	c.Header.Set("Accept", "application/json")
	c.Header.Set("Content-Type", "application/xml; charset=UTF-8")
	return c, nil
}

func (j *Jenkins) SetClient(c *http.Client) {
	j.client = c
}

func (j *Jenkins) Client() *http.Client {
	if j.client == nil {
		// disable redirect for Job.Rename() and Move()
		j.client = &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}
	}
	return j.client
}

func (j *Jenkins) Nodes() *Nodes {
	if j.nodes == nil {
		j.nodes = &Nodes{Item: NewItem(j.URL+"computer/", "Nodes", j)}
	}
	return j.nodes
}

func (j *Jenkins) Views() *Views {
	if j.views == nil {
		j.views = &Views{Item: NewItem(j.URL, "Views", j)}
	}
	return j.views
}

func (j *Jenkins) Credentials() *Credentials {
	if j.credentials == nil {
		j.credentials = &Credentials{Item: NewItem(j.URL+"credentials/store/system/domain/_/", "Credentials", j)}
	}
	return j.credentials
}

func (j *Jenkins) Queue() *Queue {
	if j.queue == nil {
		j.queue = &Queue{Item: NewItem(j.URL+"queue/", "Queue", j)}
	}
	return j.queue
}

func (c *Jenkins) GetCrumb() (*Crumb, error) {
	if c.crumb != nil {
		return c.crumb, nil
	}
	req, err := http.NewRequest("GET", c.URL+"crumbIssuer/api/json", nil)
	if err != nil {
		return nil, err
	}
	req.Header = c.Header
	resp, err := c.Client().Do(req)
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
	if err := json.Unmarshal(body, &c.crumb); err != nil {
		return nil, err
	}
	c.Header.Set(c.crumb.RequestFields, c.crumb.Value)
	c.Header.Set("Cookie", resp.Header.Get("set-cookie"))
	return c.crumb, nil
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
	if c.Debug {
		printRequest(req)
	}
	resp, err := c.Client().Do(req)
	if c.Debug {
		defer printResponse(resp)
	}
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		data, err := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("%s: %s, %s, %s", resp.Status, url, data, err)
	}
	return resp, nil
}

func printRequest(req *http.Request) {
	fmt.Printf("> %s %s %s\n", req.Method, req.URL.RequestURI(), req.Proto)
	fmt.Printf("> Host: %s\n", req.Host)
	fmt.Printf("> Content-Length: %d\n", req.ContentLength)
	for k, v := range req.Header {
		fmt.Printf("> %s: %s\n", k, v[0])
	}
}

func printResponse(resp *http.Response) {
	fmt.Printf("< %s %s\n", resp.Proto, resp.Status)
	for k, v := range resp.Header {
		fmt.Printf("< %s: %s\n", k, v[0])
	}
}

// Get job with fullname:
//
//	job, err := jenkins.GetJob("path/to/job")
//	if err != nil {
//		return err
//	}
//	fmt.Println(job)
func (c *Jenkins) GetJob(fullName string) (*Job, error) {
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
//	if err := jenkins.CreateJob("path/to/name", xml); err != nil {
//		log.Fatalln(err)
//	}
func (c *Jenkins) CreateJob(fullName string, xml io.Reader) (*http.Response, error) {
	folder, shortName := c.resolveJob(fullName)
	return folder.Create(shortName, xml)
}

func (c *Jenkins) DeleteJob(fullName string) (*http.Response, error) {
	return NewJob(c.Name2URL(fullName), "Job", c).Delete()
}

func (c *Jenkins) resolveJob(fullName string) (*Job, string) {
	dir, name := path.Split(strings.Trim(fullName, "/"))
	url := c.Name2URL(dir)
	return NewJob(url, "Folder", c), name
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
//	jenkins.BuildJob("your job", nil)
//	jenkins.BuildJob("your job", jenkins.ReqParams{})
//	// with parameters
//	jenkins.BuildJob("your job", jenkins.ReqParams{"ARG1": "ARG1_VALUE"})
func (c *Jenkins) BuildJob(fullName string, params url.Values) (*OneQueueItem, error) {
	return NewJob(c.Name2URL(fullName), "Job", c).Build(params)
}

// List job with depth
func (c *Jenkins) ListJobs(depth int) ([]*Job, error) {
	job := NewJob(c.URL, "Folder", c)
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

// func (c *Jenkins) ExportJCasC(name string) error {
// 	resp, err := c.Request("POST", "configuration-as-code/export")
// 	if err != nil {
// 		return err
// 	}
// 	return resp.ToFile(name)
// }

func (c *Jenkins) ValidateJenkinsfile(content string) (string, error) {
	v := url.Values{}
	v.Add("jenkinsfile", content)
	return readResponseToString(c, "POST", "pipeline-model-converter/validate?"+v.Encode(), nil)
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

func (c *Jenkins) RunScript(script string) (string, error) {
	v := url.Values{}
	v.Add("script", script)
	return readResponseToString(c, "POST", "scriptText?"+v.Encode(), nil)
}
