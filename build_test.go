package jenkins

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func setupBuild(t *testing.T) *Build {
	var build *Build
	build, err := pipeline.GetLastCompleteBuild()
	assert.Nil(t, err)
	if build != nil {
		return build
	}
	qitem, err := pipeline.Build(nil)
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
	err = build.LoopProgressiveLog("text", func(line []byte) error {
		output = append(output, string(line))
		time.Sleep(1 * time.Second)
		return nil
	})
	assert.Nil(t, err)
	assert.Contains(t, strings.Join(output, ""), os.Getenv("JENKINS_VERSION"))
	return build
}

func TestBuildItemIsBuilding(t *testing.T) {
	build := setupBuild(t)
	building, err := build.IsBuilding()
	assert.Nil(t, err)
	assert.False(t, building)
}

func TestBuildItemGetJob(t *testing.T) {
	build := setupBuild(t)
	job, err := build.GetJob()
	assert.Nil(t, err)
	assert.Equal(t, job.URL, pipeline.URL)
}

func TestBuildItemGetDescription(t *testing.T) {
	build := setupBuild(t)
	discription, err := build.GetDescription()
	assert.Nil(t, err)
	assert.Empty(t, discription)
	_, err = build.SetDescription("test")
	assert.Nil(t, err)
	discription, err = build.GetDescription()
	assert.Nil(t, err)
	assert.Equal(t, discription, "test")
}

func TestBuildItemDelete(t *testing.T) {
	build := setupBuild(t)
	assert.NotNil(t, build)
	_, err := build.Delete()
	assert.Nil(t, err)
	build, err = pipeline.GetBuild(build.Number)
	assert.Nil(t, err)
	assert.Nil(t, build)
}

func TestStopBuildItem(t *testing.T) {
	// change config
	conf := `<?xml version='1.1' encoding='UTF-8'?>
<flow-definition plugin="workflow-job">
  <definition class="org.jenkinsci.plugins.workflow.cps.CpsFlowDefinition" plugin="workflow-cps">
    <script>sleep(20)</script>
    <sandbox>true</sandbox>
  </definition>
  <disabled>false</disabled>
</flow-definition>`
	_, err := pipeline.SetConfigure(strings.NewReader(conf))
	assert.Nil(t, err)

	// start build to sleep 20s
	qitem, err := pipeline.Build(nil)
	assert.Nil(t, err)
	job, err := qitem.GetJob()
	assert.Nil(t, err)
	assert.Equal(t, job.FullName, pipeline.FullName)
	var build *Build
	for {
		time.Sleep(1 * time.Second)
		build, err = qitem.GetBuild()
		assert.Nil(t, err)
		if build != nil {
			break
		}
	}
	building, err := build.IsBuilding()
	assert.Nil(t, err)
	assert.True(t, building)
	_, err = build.Stop()
	assert.Nil(t, err)
	building, err = build.IsBuilding()
	assert.Nil(t, err)
	assert.False(t, building)
	result, err := build.GetResult()
	assert.Nil(t, err)
	assert.Equal(t, result, "ABORTED")
	// delete build and revert configure
	_, err = build.Delete()
	assert.Nil(t, err)
	_, err = pipeline.SetConfigure(strings.NewReader(jobConf))
	assert.Nil(t, err)
}
