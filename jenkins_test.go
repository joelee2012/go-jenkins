package jenkins

import (
	"log"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var J *Jenkins
var PipelineConfig string
var FolderConfig string
var CredentialConfig string
var folder, pipeline *Job

func ReadFile(name string) string {
	data, err := os.ReadFile(name)
	if err != nil {
		log.Fatal(err)
	}
	return string(data)
}

func SetUp() error {
	log.Println("execute setup function")
	var err error
	J, err = NewJenkins(os.Getenv("JENKINS_URL"), os.Getenv("JENKINS_USER"), os.Getenv("JENKINS_PASSWORD"))
	if err != nil {
		return err
	}

	PipelineConfig = strings.ReplaceAll(ReadFile("tests_data/pipeline.xml"), "JENKINS_VERSION", os.Getenv("JENKINS_VERSION"))

	FolderConfig = ReadFile("tests_data/folder.xml")

	CredentialConfig = ReadFile("tests_data/credential.xml")
	confs := []string{FolderConfig, FolderConfig, PipelineConfig, PipelineConfig}
	names := []string{"folder", "folder/folder1", "folder/pipeline", "folder/pipeline2"}

	for index, name := range names {
		log.Printf("create %s", name)
		if err = J.CreateJob(name, confs[index]); err != nil {
			return err
		}
	}

	folder, err = J.GetJob("folder")
	if err != nil {
		return err
	}

	pipeline, err = J.GetJob("folder/pipeline")
	if err != nil {
		return err
	}
	return nil
}

func TearsDown() {
	J.DeleteJob("folder")
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
	qitem, err := J.BuildJob("folder/pipeline", ReqParams{})
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
	job, err := J.GetJob("folder/pipeline2")
	assert.Nil(t, err)
	assert.Equal(t, job.Class, "WorkflowJob")

	job, err = J.GetJob("folder/notexist")
	assert.Nil(t, job)
	assert.Contains(t, err.Error(), "not contain job")
	// wrong path
	job, err = J.GetJob("folder/pipeline2/notexist")
	assert.Contains(t, err.Error(), "not contain job")
	assert.Nil(t, job)
}

func TestListJobs(t *testing.T) {
	jobs, err := J.ListJobs(0)
	assert.Nil(t, err)
	assert.Len(t, jobs, 1)
	jobs, err = J.ListJobs(1)
	assert.Nil(t, err)
	assert.Equal(t, jobs, 4)
}

func TestSystemCredentials(t *testing.T) {
	credsManager := J.Credentials()
	creds, err := credsManager.List()
	assert.Nil(t, err)
	assert.Len(t, creds, 0)
	assert.Nil(t, credsManager.Create(CredentialConfig))
	creds, err = credsManager.List()
	assert.Nil(t, err)
	assert.Len(t, creds, 1)
	assert.Nil(t, credsManager.Delete("user-id"))
	creds, err = credsManager.List()
	assert.Nil(t, err)
	assert.Len(t, creds, 0)
}

func TestMain(m *testing.M) {
	if err := SetUp(); err != nil {
		TearsDown()
		log.Fatal(err)
	}
	exitCode := m.Run()
	TearsDown()
	os.Exit(exitCode)
}
