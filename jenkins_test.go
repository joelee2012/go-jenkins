package jenkins

import (
	"log"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var J *Jenkins
var PipelineConfig string
var FolderConfig string
var CredentialConfig string

func SetUp() error {
	var err error
	J, err = NewJenkins(os.Getenv("JENKINS_URL"), os.Getenv("JENKINS_USER"), os.Getenv("JENKINS_PASSWORD"))
	if err != nil {
		return err
	}
	PipelineConfig = `<?xml version='1.1' encoding='UTF-8'?>
	<flow-definition plugin="workflow-job">
	  <definition class="org.jenkinsci.plugins.workflow.cps.CpsFlowDefinition" plugin="workflow-cps">
		<script>#!groovy
	pipeline {
	  agent any
	  stages {
		stage('build'){
		  steps{
			sh 'echo $JENKINS_VERSION'
		  }
		}
	  }
	}</script>
		<sandbox>true</sandbox>
	  </definition>
	  <disabled>false</disabled>
	</flow-definition>`
	FolderConfig = `<?xml version='1.0' encoding='UTF-8'?>
	<com.cloudbees.hudson.plugins.folder.Folder>
	  <actions/>
	  <description></description>
	  <properties/>
	  <folderViews/>
	  <healthMetrics/>
	</com.cloudbees.hudson.plugins.folder.Folder>`

	CredentialConfig = `<com.cloudbees.plugins.credentials.impl.UsernamePasswordCredentialsImpl>
	<scope>GLOBAL</scope>
	<id>user-id</id>
	<username>user-name</username>
	<password>user-password</password>
	<description>user id for testing</description>
  </com.cloudbees.plugins.credentials.impl.UsernamePasswordCredentialsImpl>`
	return nil
}

func TestNewJenkins(t *testing.T) {
	expect := "Jenkins-Crumb"
	crumb, err := J.GetCrumb()
	assert.Nil(t, err)
	assert.Equal(t, crumb.RequestFields, expect)
}

func TestGetVersion(t *testing.T) {
	version, err := J.GetVersion()
	assert.Nil(t, err)
	assert.Equal(t, os.Getenv("JENKINS_VERSION"), version)
}

func TestCreateJob(t *testing.T) {
	assert.Nil(t, J.CreateJob("folder", FolderConfig))
	assert.Nil(t, J.CreateJob("folder/pipeline", PipelineConfig))
	defer J.DeleteJob("folder")
	folder, err := J.GetJob("folder")
	assert.Nil(t, err)
	assert.IsType(t, &Job{}, folder)
	folderParent, err := folder.GetParent()
	assert.Nil(t, err)
	assert.Nil(t, folderParent)
	xml, err := folder.GetConfigure()
	assert.Nil(t, err)
	assert.NotEmpty(t, xml)

	pipeline, _ := J.GetJob("folder/pipeline")
	jobParent, err := pipeline.GetParent()
	assert.Nil(t, err)
	assert.IsType(t, &Job{}, jobParent)

	assert.Equal(t, "pipeline", pipeline.GetName())
	assert.Equal(t, "folder » pipeline", pipeline.GetFullDisplayName())
	assert.Equal(t, "folder/pipeline", pipeline.GetFullName())
	noexist, err := pipeline.Get("abc")
	assert.Nil(t, err)
	assert.Equal(t, noexist.Class, "abc")
}

func TestListJob(t *testing.T) {
	t.Skip()
	assert.Nil(t, J.CreateJob("folder", FolderConfig))
	assert.Nil(t, J.CreateJob("folder/pipeline", PipelineConfig))
	defer J.DeleteJob("folder")
	jobs, err := J.ListJobs(0)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(jobs))
	jobs, _ = J.ListJobs(1)
	assert.Equal(t, 2, len(jobs))
	pipeline, _ := J.GetJob("folder/pipeline")
	_, err = pipeline.List(0)
	assert.NotNil(t, err)
}

func TestGetParent(t *testing.T) {
	J.CreateJob("folder", FolderConfig)
	J.CreateJob("folder/pipeline", PipelineConfig)
	defer J.DeleteJob("folder")
	folder, _ := J.GetJob("folder")
	folderParent, err := folder.GetParent()
	assert.Nil(t, err)
	assert.Nil(t, folderParent)

	pipeline, _ := J.GetJob("folder/pipeline")
	jobParent, err := pipeline.GetParent()
	assert.Nil(t, err)
	assert.IsType(t, &Job{}, jobParent)
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
		assert.Equal(t, J.URL+test.expect, J.NameToURL(test.given))
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
		name, _ := J.URLToName(J.URL + test.given)
		assert.Equal(t, test.expect, name)
	}
	_, err := J.URLToName("http://0.0.0.1/job/folder1/")
	assert.NotNil(t, err)
}

func TestBuildJob(t *testing.T) {
	assert.Nil(t, J.CreateJob("go-test1", PipelineConfig))
	defer J.DeleteJob("go-test1")
	qitem, err := J.BuildJob("go-test1", ReqParams{})
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
	// waiting build to finish
	for {
		building, err := build.IsBuilding()
		assert.Nil(t, err)
		if !building {
			break
		}
		time.Sleep(1 * time.Second)
	}
	// get console output
	text, err := build.GetConsoleText()
	assert.Nil(t, err)
	version, err := J.GetVersion()
	assert.Nil(t, err)
	assert.Contains(t, text, version)
}

func TestRename(t *testing.T) {
	assert.Nil(t, J.CreateJob("go-test1", PipelineConfig))
	defer J.DeleteJob("go-test2")
	job, err := J.GetJob("go-test1")
	assert.Nil(t, err)
	assert.Nil(t, job.Rename("go-test2"))
	assert.Equal(t, "go-test2", job.GetName())
}

func TestBuildable(t *testing.T) {
	assert.Nil(t, J.CreateJob("go-test2", PipelineConfig))
	defer J.DeleteJob("go-test2")
	job, err := J.GetJob("go-test2")
	assert.Nil(t, err)
	assert.IsType(t, &Job{}, job)
	buildable, err := job.IsBuildable()
	assert.Nil(t, err)
	assert.True(t, buildable)
	// disable job
	assert.Nil(t, job.Disable())
	buildable, err = job.IsBuildable()
	assert.Nil(t, err)
	assert.False(t, buildable)
	// enable job
	assert.Nil(t, job.Enable())
	buildable, err = job.IsBuildable()
	assert.Nil(t, err)
	assert.True(t, buildable)
}

func TestCredentials(t *testing.T) {
	t.Skip()
	J.CreateJob("folder", FolderConfig)
	defer J.DeleteJob("folder")
	folder, err := J.GetJob("folder")
	assert.Nil(t, err)
	credsManager := folder.Credentials()
	creds, err := credsManager.List()
	assert.Nil(t, err)
	assert.Equal(t, 0, len(creds))
	assert.Nil(t, credsManager.Create(CredentialConfig))
	creds, err = credsManager.List()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(creds))
	assert.Nil(t, credsManager.Delete("user-id"))
	creds, err = credsManager.List()
	assert.Nil(t, err)
	assert.Equal(t, 0, len(creds))
}

func TestMain(m *testing.M) {
	if err := SetUp(); err != nil {
		log.Fatalln(err)
	}
	// call flag.Parse() here if TestMain uses flags
	os.Exit(m.Run())
}
