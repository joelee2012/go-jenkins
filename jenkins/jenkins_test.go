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
	client    *Client
	folder    *JobItem
	pipeline  *JobItem
	pipeline2 *JobItem
	jobConf   = `<?xml version='1.1' encoding='UTF-8'?>
<flow-definition plugin="workflow-job">
  <definition class="org.jenkinsci.plugins.workflow.cps.CpsFlowDefinition" plugin="workflow-cps">
    <script>echo  &quot;JENKINS_VERSION&quot;</script>
    <sandbox>true</sandbox>
  </definition>
  <disabled>false</disabled>
</flow-definition>`
	paramsJobConf = `<?xml version='1.1' encoding='UTF-8'?>
	<flow-definition plugin="workflow-job">
	  <description></description>
	  <keepDependencies>false</keepDependencies>
	  <properties>
		<hudson.model.ParametersDefinitionProperty>
		  <parameterDefinitions>
			<hudson.model.StringParameterDefinition>
			  <name>ARG1</name>
			  <trim>false</trim>
			</hudson.model.StringParameterDefinition>
		  </parameterDefinitions>
		</hudson.model.ParametersDefinitionProperty>
	  </properties>
	  <definition class="org.jenkinsci.plugins.workflow.cps.CpsFlowDefinition" plugin="workflow-cps">
		<script>echo params.ARG1</script>
		<sandbox>true</sandbox>
	  </definition>
	  <triggers/>
	  <disabled>false</disabled>
	</flow-definition>`
	folderConf = `<?xml version='1.0' encoding='UTF-8'?>
<com.cloudbees.hudson.plugins.folder.Folder>
  <actions/>
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

	viewConf = `<?xml version="1.1" encoding="UTF-8"?>
<hudson.model.ListView>
    <description>test</description>
    <filterExecutors>false</filterExecutors>
    <filterQueue>false</filterQueue>
    <properties class="hudson.model.View$PropertyList"/>
    <jobNames>
        <comparator class="hudson.util.CaseInsensitiveComparator"/>
    </jobNames>
    <jobFilters/>
    <columns>
        <hudson.views.StatusColumn/>
        <hudson.views.WeatherColumn/>
        <hudson.views.JobColumn/>
        <hudson.views.LastSuccessColumn/>
        <hudson.views.LastFailureColumn/>
        <hudson.views.LastDurationColumn/>
        <hudson.views.BuildButtonColumn/>
        <hudson.plugins.favorite.column.FavoriteColumn plugin="favorite@2.3.2"/>
    </columns>
    <recurse>false</recurse>
</hudson.model.ListView>`
)

func setup() error {
	log.Println("execute setup function")
	var err error
	client, err = NewClient(os.Getenv("JENKINS_URL"), os.Getenv("JENKINS_USER"), os.Getenv("JENKINS_PASSWORD"))
	if err != nil {
		return err
	}

	jobConf = strings.ReplaceAll(jobConf, "JENKINS_VERSION", os.Getenv("JENKINS_VERSION"))
	confs := []string{folderConf, folderConf, jobConf, paramsJobConf}
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

	pipeline2, err = client.GetJob("folder/pipeline2")
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

func TestName2Url(t *testing.T) {
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
		assert.Equal(t, client.URL+test.expect, client.Name2URL(test.given))
	}
}

func TestUrl2Name(t *testing.T) {
	var tests = []struct {
		expect, given string
	}{
		{"job", "job/job/"},
		{"job/job", "job/job/job/job/"},
		{"job/job", "job/job/job/job"},
	}
	for _, test := range tests {
		name, _ := client.URL2Name(client.URL + test.given)
		assert.Equal(t, test.expect, name)
	}
	_, err := client.URL2Name("http://0.0.0.1/job/folder1/")
	assert.NotNil(t, err)
}

func TestGetJob(t *testing.T) {
	// check job exist
	job, err := client.GetJob(pipeline.FullName)
	assert.Nil(t, err)
	assert.Equal(t, job.Class, "WorkflowJob")

	// check job does not exist
	job, err = client.GetJob("folder/notexist")
	assert.Nil(t, err)
	assert.Nil(t, job)
	// wrong path
	job, err = client.GetJob(pipeline.FullName + "/notexist")
	assert.Nil(t, err)
	assert.Nil(t, job)
}

func TestDeleteJob(t *testing.T) {
	assert.NotNil(t, client.DeleteJob(""))
	assert.Nil(t, client.CreateJob("folder/pipeline3", jobConf))
	assert.Nil(t, client.DeleteJob("folder/pipeline3"))
	assert.NotNil(t, client.DeleteJob("folder/pipeline3"))
}

func TestListJobs(t *testing.T) {
	jobs, err := client.ListJobs(0)
	assert.Nil(t, err)
	assert.Len(t, jobs, 1)
	jobs, err = client.ListJobs(1)
	assert.Nil(t, err)
	assert.Len(t, jobs, 4)
}
func TestBuildJob(t *testing.T) {
	build := setupBuild(t)

	// test build.IsBuilding
	building, err := build.IsBuilding()
	assert.Nil(t, err)
	assert.False(t, building)

	// test build.GetResult
	result, err := build.GetResult()
	assert.Nil(t, err)
	assert.Equal(t, result, "SUCCESS")

	// test build.GetConsoleText
	output := []string{}
	err = build.LoopLog(func(line string) error {
		output = append(output, line)
		return nil
	})
	assert.Nil(t, err)
	assert.Contains(t, output, os.Getenv("JENKINS_VERSION"))

	// test job.GetBuild
	build1, err := pipeline.GetBuild(build.ID)
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

func TestBuildJobWithParameters(t *testing.T) {
	qitem, err := client.BuildJob(pipeline2.FullName, ReqParams{"ARG1": "ARG1_VALUE"})
	var build *BuildItem
	assert.Nil(t, err)
	for {
		time.Sleep(1 * time.Second)
		build, err = qitem.GetBuild()
		assert.Nil(t, err)
		if build != nil {
			break
		}
	}
	var output []string
	err = build.LoopProgressiveLog("text", func(line string) error {
		output = append(output, line)
		time.Sleep(1 * time.Second)
		return nil
	})
	assert.Nil(t, err)
	assert.Contains(t, strings.Join(output, ""), "ARG1_VALUE")
}

func TestSystemCredentials(t *testing.T) {
	cm := client.Credentials
	creds, err := cm.List()
	assert.Nil(t, err)
	assert.Len(t, creds, 0)
	assert.Nil(t, cm.Create(credConf))
	cred, err := cm.Get("user-id")
	assert.NotNil(t, cred)
	assert.Nil(t, err)
	assert.Equal(t, cred.ID, "user-id")
	conf, err := cm.GetConfigure("user-id")
	assert.Nil(t, err)
	assert.NotEmpty(t, conf)
	creds, err = cm.List()
	assert.Nil(t, err)
	assert.Len(t, creds, 1)
	assert.Nil(t, cm.Delete("user-id"))
	creds, err = cm.List()
	assert.Nil(t, err)
	assert.Len(t, creds, 0)
}

func TestRunScript(t *testing.T) {
	output, err := client.RunScript(`println("hi, go-jenkins")`)
	assert.Nil(t, err)
	assert.Equal(t, "hi, go-jenkins\n", output)
}

func TestValidateJenkinsfile(t *testing.T) {
	output, err := client.ValidateJenkinsfile("")
	assert.Nil(t, err)
	assert.Contains(t, output, "did not contain the 'pipeline' step")

	output, err = client.ValidateJenkinsfile("pipeline { }")
	assert.Nil(t, err)
	assert.Contains(t, output, "Missing required section")
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
