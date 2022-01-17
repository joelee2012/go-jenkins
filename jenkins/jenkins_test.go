package jenkins

import (
	"fmt"
	"log"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	client   *Client
	folder   *Job
	pipeline *Job
	jobConf  = `<?xml version='1.1' encoding='UTF-8'?>
<flow-definition plugin="workflow-job">
  <definition class="org.jenkinsci.plugins.workflow.cps.CpsFlowDefinition" plugin="workflow-cps">
    <script>echo  &quot;JENKINS_VERSION&quot;</script>
    <sandbox>true</sandbox>
  </definition>
  <disabled>false</disabled>
</flow-definition>`
	folderConf = `<?xml version='1.0' encoding='UTF-8'?>
<com.cloudbees.hudson.plugins.folder.Folder>
  <actions/>
  <description></description>
  <properties/>
  <folderViews/>
  <healthMetrics/>
</com.cloudbees.hudson.plugins.folder.Folder>`
	credConf = `<com.cloudbees.plugins.credentials.impl.UsernamePasswordCredentialsImpl>
<scope>GLOBAL</scope>
<id>user-id</id>
<username>user-name</username>
<password>user-password</password>
<description>user id for testing</description>
</com.cloudbees.plugins.credentials.impl.UsernamePasswordCredentialsImpl>`
)

func setup() error {
	log.Println("execute setup function")
	var err error
	client, err = NewClient(os.Getenv("JENKINS_URL"), os.Getenv("JENKINS_USER"), os.Getenv("JENKINS_PASSWORD"))
	if err != nil {
		return err
	}

	jobConf = strings.ReplaceAll(jobConf, "JENKINS_VERSION", os.Getenv("JENKINS_VERSION"))
	confs := []string{folderConf, folderConf, jobConf, jobConf}
	names := []string{"folder", "folder/folder1", "folder/pipeline", "folder/pipeline2"}

	for index, name := range names {
		log.Printf("create %s", name)
		if err = client.CreateJob(name, confs[index]); err != nil {
			return err
		}
	}

	folder, err = client.GetJob("folder")
	if err != nil {
		return err
	}

	pipeline, err = client.GetJob("folder/pipeline")
	if err != nil {
		return err
	}
	return nil
}

func tearsdown() {
	client.DeleteJob("folder")
}

func TestNewJenkins(t *testing.T) {
	assert.Equal(t, fmt.Sprint(client), fmt.Sprintf("<Jenkins: %s>", client.URL))
	expect := "Jenkins-Crumb"
	crumb, err := client.GetCrumb()
	assert.Nil(t, err)
	assert.Equal(t, crumb.RequestFields, expect)
	crumb1, err := client.GetCrumb()
	assert.Nil(t, err)
	assert.Equal(t, crumb, crumb1)
}

func TestGetVersion(t *testing.T) {
	version, err := client.GetVersion()
	assert.Nil(t, err)
	assert.Equal(t, os.Getenv("JENKINS_VERSION"), version)
}

func TestNameToUrl(t *testing.T) {
	var tests = []struct {
		given, expect string
	}{
		{"", ""},
		{"/job/", "job/job/"},
		{"job/", "job/job/"},
		{"/job", "job/job/"},
		{"job", "job/job/"},
		{"/job/job/", "job/job/job/job/"},
		{"job/job/", "job/job/job/job/"},
		{"/job/job", "job/job/job/job/"},
		{"job/job", "job/job/job/job/"},
	}
	for _, test := range tests {
		assert.Equal(t, client.URL+test.expect, client.NameToURL(test.given))
	}
}

func TestUrlToName(t *testing.T) {
	var tests = []struct {
		expect, given string
	}{
		{"job", "job/job/"},
		{"job/job", "job/job/job/job/"},
		{"job/job", "job/job/job/job"},
	}
	for _, test := range tests {
		name, _ := client.URLToName(client.URL + test.given)
		assert.Equal(t, test.expect, name)
	}
	_, err := client.URLToName("http://0.0.0.1/job/folder1/")
	assert.NotNil(t, err)
}

func TestBuildJob(t *testing.T) {
	qitem, err := client.BuildJob("folder/pipeline", ReqParams{})
	assert.Nil(t, err)
	var build *Build
	for {
		time.Sleep(1 * time.Second)
		build, err = qitem.GetBuild()
		assert.Nil(t, err)
		if build != nil {
			break
		}
	}

	// test build.IterateProgressConsoleText
	var output []string
	err = build.LoopProgressiveLog("text", func(line string) error {
		output = append(output, line)
		time.Sleep(1 * time.Second)
		return nil
	})
	assert.Nil(t, err)
	assert.Nil(t, err)
	assert.Contains(t, strings.Join(output, ""), os.Getenv("JENKINS_VERSION"))

	// test build.IsBuilding
	building, err := build.IsBuilding()
	assert.Nil(t, err)
	assert.False(t, building)

	// test build.GetResult
	result, err := build.GetResult()
	assert.Nil(t, err)
	assert.Equal(t, result, "SUCCESS")

	// test build.GetConsoleText
	output = []string{}
	err = build.LoopLog(func(line string) error {
		output = append(output, line)
		return nil
	})
	assert.Nil(t, err)
	assert.Contains(t, output, os.Getenv("JENKINS_VERSION"))

	// test job.GetBuild
	build1, err := pipeline.GetBuild(1)
	assert.Nil(t, err)
	assert.Equal(t, build, build1)

	// test job.GetLastBuild
	build1, err = pipeline.GetLastBuild()
	assert.Nil(t, err)
	assert.Equal(t, build, build1)

	// test job.GetLastBuild
	build1, err = pipeline.GetFirstBuild()
	assert.Nil(t, err)
	assert.Equal(t, build, build1)
}

func TestGetJob(t *testing.T) {
	job, err := client.GetJob("folder/pipeline2")
	assert.Nil(t, err)
	assert.Equal(t, job.Class, "WorkflowJob")

	job, err = client.GetJob("folder/notexist")
	assert.Nil(t, job)
	assert.Contains(t, err.Error(), "not contain job")
	// wrong path
	job, err = client.GetJob("folder/pipeline2/notexist")
	assert.Contains(t, err.Error(), "not contain job")
	assert.Nil(t, job)
}

func TestListJobs(t *testing.T) {
	jobs, err := client.ListJobs(0)
	assert.Nil(t, err)
	assert.Len(t, jobs, 1)
	jobs, err = client.ListJobs(1)
	assert.Nil(t, err)
	assert.Len(t, jobs, 4)
}

func TestSystemCredentials(t *testing.T) {
	cm := client.Credentials()
	creds, err := cm.List()
	assert.Nil(t, err)
	assert.Len(t, creds, 0)
	assert.Nil(t, cm.Create(credConf))
	creds, err = cm.List()
	assert.Nil(t, err)
	assert.Len(t, creds, 1)
	assert.Nil(t, cm.Delete("user-id"))
	creds, err = cm.List()
	assert.Nil(t, err)
	assert.Len(t, creds, 0)
}

func TestRunScript(t *testing.T) {
	output, err := client.RunScript("println('hi, go-jenkins')")
	assert.Nil(t, err)
	assert.Equal(t, output, "hi, go-jenkins")
}

func TestValidateJenkinsfile(t *testing.T) {
	output, err := client.ValidateJenkinsfile("")
	assert.Nil(t, err)
	assert.Equal(t, output, "No Jenkinsfile specified")
}

func TestQuiteDown(t *testing.T) {
	var status struct {
		Class        string `json:"_class"`
		QuietingDown bool   `json:"quietingDown"`
	}
	assert.Nil(t, client.BindAPIJson(ReqParams{}, &status))
	assert.False(t, status.QuietingDown)
	// set quite down
	assert.Nil(t, client.QuiteDown())
	assert.Nil(t, client.BindAPIJson(ReqParams{}, &status))
	assert.True(t, status.QuietingDown)
	// cancel quite down
	assert.Nil(t, client.CancelQuiteDown())
	assert.Nil(t, client.BindAPIJson(ReqParams{}, &status))
	assert.False(t, status.QuietingDown)
}

func TestMain(m *testing.M) {
	if err := setup(); err != nil {
		tearsdown()
		log.Fatal(err)
	}
	exitCode := m.Run()
	tearsdown()
	os.Exit(exitCode)
}
