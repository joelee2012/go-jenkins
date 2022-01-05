package jenkins

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetJobOfBuild(t *testing.T) {
	t.Skip()
	build, err := pipeline.GetBuild(1)
	assert.Nil(t, err)
	job, err := build.GetJob()
	assert.Nil(t, err)
	assert.Equal(t, job, pipeline)
}
