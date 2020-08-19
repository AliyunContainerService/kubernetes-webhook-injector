package mock

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMockPlugin(t *testing.T) {
	mp := NewMockPlugin()
	assert.True(t, mp.Name() == MockPluginName, "name not match")
}
