package jenkins

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNodeGet(t *testing.T) {
	node, err := client.Nodes.Get("Built-In Node")
	assert.Nil(t, err)
	assert.NotNil(t, node)
}

func TestNodeList(t *testing.T) {
	nodes, err := client.Nodes.List()
	assert.Nil(t, err)
	assert.Len(t, nodes, 1)
}

func TestDisableNode(t *testing.T) {
	// check node status
	node, err := client.Nodes.Get("Built-In Node")
	assert.Nil(t, err)
	assert.NotNil(t, node)
	assert.False(t, node.Offline)

	// disable and then check
	_, err = client.Nodes.Disable("Built-In Node", "test")
	assert.Nil(t, err)
	node, err = client.Nodes.Get("Built-In Node")
	assert.Nil(t, err)
	assert.True(t, node.Offline)

	// enable again
	_, err = client.Nodes.Enable("Built-In Node")
	assert.Nil(t, err)
}
