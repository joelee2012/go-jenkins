package jenkins

import (
	"log"
	"os"
	"testing"
)

var jenk *Jenkins
var PipelineConfig string


func SetUp() error {
	var err error
	jenk, err = NewJenkins(os.Getenv("JENKINS_URL"), os.Getenv("JENKINS_USER"), os.Getenv("JENKINS_PSW"))
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
	return nil
}


func TestNewJenkins(t *testing.T) {
	expect := "Jenkins-Crumb"
	if jenk.RequestFields != expect {
		t.Errorf("expect j.RequestFields has value %q, but got %q", expect, jenk.RequestFields)
	}
}

func TestGetVersion(t *testing.T) {
	version, err := jenk.GetVersion()
	if err != nil {
		t.Errorf("NewJenkins failed with error: %v", err)
	}
	if version != os.Getenv("JENKINS_VERSION") {
		t.Errorf("expect version %s, but got %s", os.Getenv("JENKINS_VERSION"), version)
	}
}

func TestCreateJob(t *testing.T) {
	if err := jenk.CreateJob("go-test1", PipelineConfig); err != nil {
		t.Errorf("expect create job successful, but got error:\n %v", err)
	}
	defer jenk.DeleteJob("go-test1")
	job, _ := jenk.GetJob("go-test1")
	if job == nil {
		t.Error("expect get job, but got nil")
	}
}

func TestGetJob(t *testing.T) {
	job, _ := jenk.GetJob("notexist")
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
		if url := jenk.NameToUrl(test.given); url != jenk.Url+test.expect {
			t.Errorf("expect NameToUrl(%q) return %q, but got %q", test.given, jenk.Url+test.expect, url)
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
		if name, _ := jenk.UrlToName(jenk.Url + test.given); name != test.expect {
			t.Errorf("expect UrlToName(%q) return %q, but got %q", jenk.Url+test.given, test.expect, name)
		}
	}
	given := "http://0.0.0.1/job/folder1/"
	if name, err := jenk.UrlToName("http://0.0.0.1/job/folder1/"); err == nil {
		t.Errorf("expect UrlToName(%q) return error, but got %q", given, name)
	}
}

func TestMain(m *testing.M) {
	if err := SetUp(); err !=nil {
		log.Fatalln(err)
	}
	// call flag.Parse() here if TestMain uses flags
	os.Exit(m.Run())
}