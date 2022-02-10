package jenkins

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestViewServiceGet(t *testing.T) {
	v, err := client.Views.Get("all")
	assert.Nil(t, err)
	assert.Equal(t, v.Name, "all")
}
