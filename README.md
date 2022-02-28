![Tests](https://github.com/joelee2012/go-jenkins/workflows/Tests/badge.svg?branch=main)
![CodeQL](https://github.com/joelee2012/go-jenkins/workflows/CodeQL/badge.svg?branch=main)
[![codecov](https://codecov.io/gh/joelee2012/go-jenkins/branch/main/graph/badge.svg?token=SEWFVZ7UT3)](https://codecov.io/gh/joelee2012/go-jenkins)
![GitHub](https://img.shields.io/github/license/joelee2012/go-jenkins)

# go-jenkins
[Golang](https://golang.org/) client library for accessing [Jenkins API](https://www.jenkins.io/doc/book/using/remote-access-api/).
> Ported from [api4jenkins](https://github.com/joelee2012/api4jenkins>), a [Python3](https://www.python.org/) client library for [Jenkins API](https://www.jenkins.io/doc/book/using/remote-access-api/).

# Features
This API client package covers most of the existing Jenkins API calls and is updated regularly to add new and/or missing endpoints.

Currently, the following are supported:

- Job
- Build
- Credential
- View
- Queue
- Node

# Usage

```go
import "github.com/joelee2012/go-jenkins"
```

## Example
```go
package main

import (
	"log"
	"time"

	"github.com/joelee2012/go-jenkins"
)

func main() {
	// Construct new client
	client, err := jenkins.NewClient("http://localhost:8080/", "admin", "1234")
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
	if err := client.CreateJob("pipeline", xml); err != nil {
		log.Fatalln(err)
	}
	// Build Job and get BuildItem
	qitem, err := client.BuildJob("pipeline", nil)
	if err != nil {
		log.Fatalln(err)
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
	// tail the build log to end
	build.LoopProgressiveLog("text", func(line string) error {
		log.Println(line)
		time.Sleep(1 * time.Second)
		return nil
	})
}
```

# Documentation
See https://pkg.go.dev/github.com/joelee2012/go-jenkins