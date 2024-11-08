package jenkins

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestName(t *testing.T) {
	assert.Equal(t, "folder", folder.Name)
	assert.Equal(t, "folder", folder.FullName)
	assert.Equal(t, "folder", folder.FullDisplayName)

	assert.Equal(t, "pipeline", pipeline.Name)
	assert.Equal(t, "folder/pipeline", pipeline.FullName)
	assert.Equal(t, "folder » pipeline", pipeline.FullDisplayName)
}

func TestRename(t *testing.T) {
	_, err := pipeline.Rename("pipeline1")
	assert.Nil(t, err)
	newPipeline, err := folder.Get("pipeline1")
	assert.Nil(t, err)
	assert.Equal(t, pipeline.URL, newPipeline.URL)
	assert.Equal(t, pipeline.Name, newPipeline.Name)

	// old job 'pipeline' should not exist
	old, err := folder.Get("pipeline")
	assert.NotNil(t, err)
	assert.Nil(t, old)

	// revert
	_, err = pipeline.Rename("pipeline")
	assert.Nil(t, err)
}

func TestIsBuildable(t *testing.T) {
	buildable, err := pipeline.IsBuildable()
	assert.Nil(t, err)
	assert.True(t, buildable)

	// disable and check
	_, err = pipeline.Disable()
	assert.Nil(t, err)
	buildable, err = pipeline.IsBuildable()
	assert.Nil(t, err)
	assert.False(t, buildable)

	// enable and check
	_, err = pipeline.Enable()
	assert.Nil(t, err)
	buildable, err = pipeline.IsBuildable()
	assert.Nil(t, err)
	assert.True(t, buildable)

	// test foler
	buildable, err = folder.IsBuildable()
	assert.Nil(t, err)
	assert.False(t, buildable)
}

func TestList(t *testing.T) {
	// test job.List for folder
	jobs, err := folder.List(0)
	assert.Nil(t, err)
	assert.Len(t, jobs, 3)

	// test job.List for job
	jobs, err = pipeline.List(0)
	assert.NotNil(t, err)
	assert.Nil(t, jobs)
}

func TestGetParent(t *testing.T) {
	fParent, err := folder.GetParent()
	assert.NotNil(t, err)
	assert.Nil(t, fParent)

	pParent, err := pipeline.GetParent()
	assert.Nil(t, err)
	assert.Equal(t, folder.URL, pParent.URL)
}

func TestGetConfig(t *testing.T) {
	conf, err := pipeline.GetConfigure()
	assert.Nil(t, err)
	assert.Contains(t, conf, os.Getenv("JENKINS_VERSION"))
}

func TestListBuilds(t *testing.T) {
	builds, err := pipeline.ListBuilds()
	assert.Nil(t, err)
	assert.Len(t, builds, 1)
	builds, err = folder.ListBuilds()
	assert.NotNil(t, err)
	assert.Nil(t, builds)
}

func TestFolderCredentials(t *testing.T) {
	cm := folder.Credentials()
	creds, err := cm.List()
	assert.Nil(t, err)
	assert.Len(t, creds, 0)
	_, err = cm.Create(strings.NewReader(credConf))
	assert.Nil(t, err)
	creds, err = cm.List()
	assert.Nil(t, err)
	assert.Len(t, creds, 1)
	_, err = cm.Delete("user-id")
	assert.Nil(t, err)
	creds, err = cm.List()
	assert.Nil(t, err)
	assert.Len(t, creds, 0)
}

func TestSetDescription(t *testing.T) {
	description, err := pipeline.GetDescription()
	assert.Nil(t, err)
	assert.Empty(t, description)
	msg := "testing job for go jenkins"
	_, err = pipeline.SetDescription(msg)
	assert.Nil(t, err)
	description, err = pipeline.GetDescription()
	assert.Nil(t, err)
	assert.Equal(t, msg, description)
}

func TestGetBuildFunctions(t *testing.T) {
	expect_build := setupBuild(t)
	// test job.GetBuild
	build, err := pipeline.GetBuild(expect_build.Number)
	assert.Nil(t, err)
	assert.Equal(t, expect_build.Number, build.Number)

	// test job.GetLastBuild
	build, err = pipeline.GetLastBuild()
	assert.Nil(t, err)
	assert.Equal(t, expect_build.Number, build.Number)

	// test job.GetLastBuild
	build, err = pipeline.GetFirstBuild()
	assert.Nil(t, err)
	assert.Equal(t, expect_build.Number, build.Number)

	// test for folder
	build, err = folder.GetFirstBuild()
	assert.NotNil(t, err)
	assert.Nil(t, build)
}

func TestMove(t *testing.T) {
	_, err := pipeline.Move("/folder/folder1")
	assert.Nil(t, err)
	job, err := jenkins.GetJob("folder/pipeline")
	assert.NotNil(t, err)
	assert.Nil(t, job)
	job, err = jenkins.GetJob("folder/folder1/pipeline")
	assert.Nil(t, err)
	assert.Contains(t, job.URL, "folder1/job/pipeline")

	//revert change
	_, err = pipeline.Move("folder")
	assert.Nil(t, err)
}

func TestCopy(t *testing.T) {
	_, err := folder.Copy("pipeline", "new_pipeline")
	assert.Nil(t, err)
	job, err := jenkins.GetJob("folder/new_pipeline")
	assert.Nil(t, err)
	assert.Equal(t, job.Class, pipeline.Class)
	assert.Contains(t, job.URL, "new_pipeline")

	// clean
	jenkins.DeleteJob("folder/new_pipeline")
}
