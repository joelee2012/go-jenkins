package jenkins

import (
	"fmt"
	"log"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/imroc/req"
)

var J *Jenkins
var PipelineConfig string
var FolderConfig string

func SetUp() error {
	var err error
	J, err = NewJenkins(os.Getenv("JENKINS_URL"), os.Getenv("JENKINS_USER"), os.Getenv("JENKINS_PSW"))
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
	return nil
}

func TestNewJenkins(t *testing.T) {
	expect := "Jenkins-Crumb"
	if J.Crumb.RequestFields != expect {
		t.Errorf("expect j.RequestFields has value %q, but got %q", expect, J.Crumb.RequestFields)
	}
}

func TestGetVersion(t *testing.T) {
	version, err := J.GetVersion()
	if err != nil {
		t.Errorf("NewJenkins failed with error: %v", err)
	}
	if version != os.Getenv("JENKINS_VERSION") {
		t.Errorf("expect version %s, but got %s", os.Getenv("JENKINS_VERSION"), version)
	}
}

func TestCreateJob(t *testing.T) {
	if err := J.CreateJob("go-test1", PipelineConfig); err != nil {
		t.Errorf("expect create job successful, but got error:\n %v", err)
	}
	defer J.DeleteJob("go-test1")
	job, err := J.GetJob("go-test1")
	if err != nil {
		t.Error(err)
	}
	if job.GetName() != "go-test1" {
		t.Error("expect get job, but got nil")
	}
}

func TestGetParent(t *testing.T) {
	J.CreateJob("folder", FolderConfig)
	J.CreateJob("folder/pipeline", PipelineConfig)
	defer J.DeleteJob("folder")
	folder, _ := J.GetJob("folder")
	folderParent, err := folder.GetParent()
	if err != nil {
		t.Errorf("expect get parent of folder, but got %v", err)
	}
	if folderParent != nil {
		t.Errorf("expect parent of folder is nil, but got %v", folderParent)
	}
	pipeline, _ := J.GetJob("folder/pipeline")
	jobParent, err := pipeline.GetParent()
	if err != nil {
		t.Errorf("expect get parent of pipeline, but got %v", err)
	}
	if jobParent == nil {
		t.Error("expect parent of pipeline, but got nil")
	}
}

func TestGetJob(t *testing.T) {
	job, _ := J.GetJob("notexist")
	if job != nil {
		t.Errorf("expect no such job, but got %v", job)
	}
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
		if url := J.NameToURL(test.given); url != J.URL+test.expect {
			t.Errorf("expect NameToUrl(%q) return %q, but got %q", test.given, J.URL+test.expect, url)
		}
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
		if name, _ := J.URLToName(J.URL + test.given); name != test.expect {
			t.Errorf("expect UrlToName(%q) return %q, but got %q", J.URL+test.given, test.expect, name)
		}
	}
	given := "http://0.0.0.1/job/folder1/"
	if name, err := J.URLToName("http://0.0.0.1/job/folder1/"); err == nil {
		t.Errorf("expect UrlToName(%q) return error, but got %q", given, name)
	}
}

func TestGetConfig(t *testing.T) {
	if err := J.CreateJob("go-test1", PipelineConfig); err != nil {
		t.Errorf("expect create job successful, but got error:\n %v", err)
	}
	defer J.DeleteJob("go-test1")
	job, _ := J.GetJob("go-test1")
	if job == nil {
		t.Errorf("expect no such job, but got %v", job)
	}
	xml, _ := job.GetConfigure()
	if xml == "" {
		t.Error("expect get job config, but got empty")
	}
}
func TestBuildJob(t *testing.T) {
	if err := J.CreateJob("go-test1", PipelineConfig); err != nil {
		t.Errorf("expect create job successful, but got error:\n %v", err)
	}
	defer J.DeleteJob("go-test1")
	qitem, err := J.BuildJob("go-test1", req.Param{})
	if err != nil {
		t.Errorf("expect build job successful, but got error:\n %v", err)
	}
	var build *Build
	for {
		time.Sleep(1 * time.Second)
		build, err = qitem.GetBuild()
		if err != nil {
			log.Fatalln(err)
		}
		if build != nil {
			break
		}
	}
	// waiting build to finish
	fmt.Println(build)
	for {
		building, err := build.IsBuilding()
		if err != nil {
			t.Error(err)
		}
		if !building {
			break
		}
		time.Sleep(1 * time.Second)
	}
	// get console output
	text, err := build.GetConsoleText()
	if err != nil {
		log.Fatalln(err)
	}
	version, _ := J.GetVersion()
	if !strings.Contains(text, version) {
		t.Errorf("expect console text contain %s, but got: %s", version, text)
	}

}

func TestRename(t *testing.T) {
	if err := J.CreateJob("go-test1", PipelineConfig); err != nil {
		t.Errorf("expect create job successful, but got error:\n %v", err)
	}
	defer J.DeleteJob("go-test2")
	job, err := J.GetJob("go-test1")
	if err != nil {
		t.Error(err)
	}

	if err := job.Rename("go-test2"); err != nil {
		t.Errorf("rename got error: %v", err)
	}
	if job.GetName() != "go-test2" {
		t.Errorf("expect job.GetName() == 'go-test2', but got %s", job.GetName())
	}
}

func TestMain(m *testing.M) {
	if err := SetUp(); err != nil {
		log.Fatalln(err)
	}
	// call flag.Parse() here if TestMain uses flags
	os.Exit(m.Run())
}
