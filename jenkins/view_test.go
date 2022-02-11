package jenkins

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestViewServiceGet(t *testing.T) {
	v, err := client.Views.Get("all")
	assert.Nil(t, err)
	assert.NotNil(t, v)
	assert.Equal(t, v.Name, "all")
}

func TestViewServiceCreate(t *testing.T) {
	v, err := folder.Views.Get("testview")
	assert.Nil(t, err)
	assert.Empty(t, v)

	// create view
	assert.Nil(t, folder.Views.Create("testview", viewConf))
	v, err = folder.Views.Get("testview")
	assert.Nil(t, err)
	assert.NotNil(t, v)
	assert.Equal(t, v.Name, "testview")
	assert.Equal(t, v.Description, "test")

	// list views
	views, err := folder.Views.List()
	assert.Nil(t, err)
	assert.Len(t, views, 2)

	// get job from view
	job, err := folder.Views.GetJobFromView("testview", pipeline.Name)
	assert.Nil(t, err)
	assert.Nil(t, job)

	// add job to view
	assert.Nil(t, folder.Views.AddJobToView("testview", pipeline.Name))
	job, err = folder.Views.GetJobFromView("testview", pipeline.Name)
	assert.Nil(t, err)
	assert.Equal(t, job.FullName, pipeline.FullName)

	jobs, err := folder.Views.ListJobInView("testview")
	assert.Nil(t, err)
	assert.Len(t, jobs, 1)

	// remove job from view
	assert.Nil(t, folder.Views.RemoveJobFromView("testview", pipeline.Name))
	jobs, err = folder.Views.ListJobInView("testview")
	assert.Nil(t, err)
	assert.Len(t, jobs, 0)

	// set description
	assert.Nil(t, folder.Views.SetDescription("testview", "new description"))
	v, err = folder.Views.Get("testview")
	assert.Nil(t, err)
	assert.Equal(t, v.Description, "new description")

	// set/get configuration
	assert.Nil(t, folder.Views.SetConfigure("testview", strings.ReplaceAll(viewConf, "test", "newtest")))
	config, err := folder.Views.GetConfigure("testview")
	assert.Nil(t, err)
	assert.Contains(t, config, "newtest")

	// delete view
	assert.Nil(t, folder.Views.Delete("testview"))
	v, err = folder.Views.Get("testview")
	assert.Nil(t, err)
	assert.Nil(t, v)
}
