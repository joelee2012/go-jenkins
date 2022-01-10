package jenkins

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetName(t *testing.T) {
	assert.Equal(t, "folder", folder.GetName())
	assert.Equal(t, "pipeline", pipeline.GetName())
}

func TestGetFullDisplayName(t *testing.T) {
	assert.Equal(t, "folder", folder.GetFullDisplayName())
	assert.Equal(t, "folder » pipeline", pipeline.GetFullDisplayName())
}

func TestGetFullName(t *testing.T) {
	assert.Equal(t, "folder", folder.GetFullName())
	assert.Equal(t, "folder/pipeline", pipeline.GetFullName())
}

func TestRename(t *testing.T) {
	err := pipeline.Rename("pipeline1")
	assert.Nil(t, err)
	newPipeline, err := folder.Get("pipeline1")
	assert.Nil(t, err)
	assert.Equal(t, pipeline.URL, newPipeline.URL)

	// old job 'pipeline' should not exist
	old, err := folder.Get("pipeline")
	assert.NotNil(t, err)
	assert.Nil(t, old)

	// revert
	err = pipeline.Rename("pipeline")
	assert.Nil(t, err)
}

func TestIsBuildable(t *testing.T) {
	buildable, err := pipeline.IsBuildable()
	assert.Nil(t, err)
	assert.True(t, buildable)

	// disable and check
	assert.Nil(t, pipeline.Disable())
	buildable, err = pipeline.IsBuildable()
	assert.Nil(t, err)
	assert.False(t, buildable)

	// enable and check
	assert.Nil(t, pipeline.Enable())
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
	assert.Nil(t, err)
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
	credsManager := folder.Credentials()
	creds, err := credsManager.List()
	assert.Nil(t, err)
	assert.Len(t, creds, 0)
	assert.Nil(t, credsManager.Create(credConf))
	creds, err = credsManager.List()
	assert.Nil(t, err)
	assert.Len(t, creds, 1)
	assert.Nil(t, credsManager.Delete("user-id"))
	creds, err = credsManager.List()
	assert.Nil(t, err)
	assert.Len(t, creds, 0)
}

func TestSetDescription(t *testing.T) {
	description, err := pipeline.GetDescription()
	assert.Nil(t, err)
	assert.Empty(t, description)
	msg := "testing job for go jenkins"
	assert.Nil(t, pipeline.SetDescription(msg))
	description, err = pipeline.GetDescription()
	assert.Nil(t, err)
	assert.Equal(t, msg, description)
}

func TestGetBuildFunctions(t *testing.T) {
	// test job.GetBuild
	build, err := pipeline.GetBuild(1)
	assert.Nil(t, err)
	assert.Equal(t, 1, build.ID)

	// test job.GetLastBuild
	build, err = pipeline.GetLastBuild()
	assert.Nil(t, err)
	assert.Equal(t, 1, build.ID)

	// test job.GetLastBuild
	build, err = pipeline.GetFirstBuild()
	assert.Nil(t, err)
	assert.Equal(t, 1, build.ID)

	// test for folder
	build, err = folder.GetFirstBuild()
	assert.NotNil(t, err)
	assert.Nil(t, build)
}

func TestMove(t *testing.T) {
	assert.Nil(t, pipeline.Move("/folder/folder1"))
	job, err := client.GetJob("folder/pipeline")
	assert.Contains(t, err.Error(), "not contain job")
	assert.Nil(t, job)
	job, err = client.GetJob("folder/folder1/pipeline")
	assert.Nil(t, err)
	assert.Contains(t, job.URL, "folder1/job/pipeline")

	//revert change
	assert.Nil(t, pipeline.Move("folder"))
}

func TestCopy(t *testing.T) {
	assert.Nil(t, folder.Copy("pipeline", "new_pipeline"))
	job, err := client.GetJob("folder/new_pipeline")
	assert.Nil(t, err)
	assert.Equal(t, job.Class, pipeline.Class)
	assert.Contains(t, job.URL, "new_pipeline")

	// clean
	client.DeleteJob("folder/new_pipeline")
}
