package pkg

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	validWebHookOptions = &WebHookOptions{
		TLSCertPath: "",
	}

	invalidWebHookOptions = &WebHookOptions{
		TLSCaCertPath: "",
	}
)

func TestWebHookOptionsValidation(t *testing.T) {
	var isValid bool
	var msg string

	// todo add caCert validation
	//isValid, msg = validWebHookOptions.valid()
	//assert.True(t, isValid, msg)

	isValid, msg = invalidWebHookOptions.valid()
	assert.True(t, !isValid, msg)
}
