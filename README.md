![Tests](https://github.com/joelee2012/go-jenkins/workflows/Tests/badge.svg?branch=main)
![CodeQL](https://github.com/joelee2012/go-jenkins/workflows/CodeQL/badge.svg?branch=main)
[![codecov](https://codecov.io/gh/joelee2012/go-jenkins/branch/main/graph/badge.svg?token=SEWFVZ7UT3)](https://codecov.io/gh/joelee2012/go-jenkins)
![GitHub](https://img.shields.io/github/license/joelee2012/go-jenkins)

# go-jenkins
[Jenkins](https://www.jenkins.io/) REST API client for [Golang](https://golang.org/), ported from [api4jenkins](https://github.com/joelee2012/api4jenkins>)


# Usage

```go
package main

import (
	"log"
	"time"

	"github.com/imroc/req"
	"github.com/joelee2012/go-jenkins/jenkins"
)

func main() {
	j, err := jenkins.NewJenkins("http://localhost:8080/", "admin", "1234")
	log.Println(j)
	if err != nil {
		log.Fatalln(err)
	}
	xml := `<?xml version='1.1' encoding='UTF-8'?>
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
    // create jenkins job
    if err := j.CreateJob("workflowjob1", xml); err != nil {
		log.Fatalln(err)
	}
    // get job
    job, err := j.GetJob("workflowjob1")
	if err != nil {
		log.Fatalln(err)
	}
    workflow := job.(*jenkins.WorkflowJob)
    // build job
    queue, err := workflow.Build(req.Param{})
    if err != nil {
        log.Fatalln(err)
    }

    // get build object from queue item
    var build interface{}
    for {
        time.Sleep(1 * time.Second)
        build, err = queue.GetBuild()
        if err != nil {
            log.Fatalln(err)
        }
        if build != nil {
            break
        }
    }
    // waiting build to finish
    run := build.(*jenkins.WorkflowRun)
    for run.IsBuilding() {
        time.Sleep(1 * time.Second)
    }
    // get console output
    text, err := run.GetConsoleText()
    if err != nil {
        log.Fatalln(err)
    }
    log.Println(string(text))
}
```