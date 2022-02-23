![Tests](https://github.com/joelee2012/go-jenkins/workflows/Tests/badge.svg?branch=main)
![CodeQL](https://github.com/joelee2012/go-jenkins/workflows/CodeQL/badge.svg?branch=main)
[![codecov](https://codecov.io/gh/joelee2012/go-jenkins/branch/main/graph/badge.svg?token=SEWFVZ7UT3)](https://codecov.io/gh/joelee2012/go-jenkins)
![GitHub](https://img.shields.io/github/license/joelee2012/go-jenkins)

# go-jenkins
[Golang](https://golang.org/) client library for accessing [Jenkins API](https://www.jenkins.io/doc/book/using/remote-access-api/).
> Ported from [api4jenkins](https://github.com/joelee2012/api4jenkins>), a [Python3](https://www.python.org/) client library for [Jenkins API](https://www.jenkins.io/doc/book/using/remote-access-api/).



# Usage

```go
import "github.com/joelee2012/go-jenkins/jenkins"
```

## Client
Construct new client
```go
client, err := jenkins.NewClient("http://localhost:8080/", "admin", "1234")
if err != nil {
	log.Fatalln(err)
}
```

Create Job with given xml configuration
```go
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
```

Build Job and get BuildItem
```go
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

```

Tail the build log to end
```go
build.LoopProgressiveLog("text", func(line string) error {
	log.Println(line)
	time.Sleep(1 * time.Second)
	return nil
})
```

Get Job with full name
```go
job, err := client.GetJob("path/to/name")
if err != nil {
	log.Fatalln(err)
}
```

List job with depth
```go
jobs, err := client.ListJobs(1)
if err != nil {
	log.Fatalln(err)
}
```

## JobItem
Rename job

```go
if err := job.Rename("new name"); err != nil {
	log.Fatalln(err)
}
```

