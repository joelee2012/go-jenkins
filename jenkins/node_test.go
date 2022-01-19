package jenkins

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNodeGet(t *testing.T) {
	node, err := client.ComputerSet().Get("Built-In Node")
	assert.Nil(t, err)
	assert.NotNil(t, node)
}

func TestNodeList(t *testing.T) {
	nodes, err := client.ComputerSet().List()
	assert.Nil(t, err)
	assert.Len(t, nodes, 1)
}
