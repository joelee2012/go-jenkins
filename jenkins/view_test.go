package jenkins

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestViewServiceGet(t *testing.T) {
	v, err := client.Views.Get("All")
	assert.Nil(t, err)
	assert.Equal(t, v.Name, "All")
}
