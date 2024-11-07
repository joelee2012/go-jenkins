package jenkins

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNodeGet(t *testing.T) {
	node, err := jenkins.Nodes().Get("Built-In Node")
	assert.Nil(t, err)
	assert.NotNil(t, node)
}

func TestNodeList(t *testing.T) {
	nodes, err := jenkins.Nodes().List()
	assert.Nil(t, err)
	assert.Len(t, nodes, 1)
}

func TestDisableNode(t *testing.T) {
	// check node status
	node, err := jenkins.Nodes().Get("Built-In Node")
	assert.Nil(t, err)
	assert.NotNil(t, node)
	assert.False(t, node.Offline)

	// disable and then check
	_, err = jenkins.Nodes().Disable("Built-In Node", "test")
	assert.Nil(t, err)
	node, err = jenkins.Nodes().Get("Built-In Node")
	assert.Nil(t, err)
	assert.True(t, node.Offline)

	// enable again
	_, err = jenkins.Nodes().Enable("Built-In Node")
	assert.Nil(t, err)
}
