package jenkins

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNodeGet(t *testing.T) {
	node, err := jenkins.ComputerSet().Get("Built-In Node")
	assert.Nil(t, err)
	assert.NotNil(t, node)
}

func TestNodeList(t *testing.T) {
	nodes, err := jenkins.ComputerSet().List()
	assert.Nil(t, err)
	assert.Len(t, nodes, 1)
}
