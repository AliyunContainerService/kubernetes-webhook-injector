package openapi

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

var token = `{"access.key.id":"7657sVLqCXgvd/kgDoJKc3Ahun3KOnPOihW3AE4qKVTN9uD8H4j17dZbHC0KymCf",
              "access.key.secret":"xE+Zml1pt9u3WlbuO0YfEV9wTbsp7z1xvJ1RY0Q1IRxx7I2y+ntjDlfC7IHKVY9oUt7YNllsL0Yt8d0F4/CwHQ==",
              "security.token":"wNpYCLYSjY8EQk5h0oXp8rOW5JOZ2n1ApUKCrkj9Bl0ckJ0N+Fay0TLXPL1CNJm488+FIWlgT600NQ4xtE48x1vqppHUxQXASxa0hYh8Pw+MV7iqAPuK8YP+rr2d56yAuBo73f+BeM7/AErcKakJK8bZzKl5NdKIfnK5zeFLL0wFI8M42c+kjXTVPezuZ/8ao2XqMU/77n2KtxXBwDSiJ06tqLW0oqIr1qWfbU9KO+w84sYg1RbSGpyxD0EhWcqnJI96VKGcya3dNBXgWOwGF0WdTCy8ArzVX0bJG7zN2DmRwQ79RGuccOrXkhpNVp+3J46GuR7AuNMOLHz0Gs9GnMC7uG8mFL0dDl/jIu7x1Ir2DtriKezSFAo5BQgsoDLwWAkeGop6UGFu4iijO3vWBs9w6ix18pSbNtgVBz4WP4uyBtrZFlMuiJf9xT9ds+gp+HMFmfLMSt99dm0WbnFl8cBqK9uG/KyrUT8tAQFgTLJ8ZUK6H9Vreb6F2XhsDdZd5RSZf+4tBq6f9UCyIaW0BcY3zYazkmpaOYc4ML7ubuAqebN6vzu+9sKEAvjAtl9O6qHd7oNHGM47bbE9kEvk0XtOTzfdvfipvhd1bFKL7mIoNChkZpel4gtF6G3lERyTM15J2VYuWocdBswcVGrquM9dm8QrplmY8vBJT1amVzu60629926+IQWL/yHDBzGZiAFSPu/uXXvPHgVmg9RPEUWhfHwfe2sRNrTY7O5txx5ncLburbOoo1yBUgO9jmnZ",
              "expiration":"2030-08-31T03:26:37Z",
              "keyring":"BLDNDESNYTVFCDKN"}`

func TestExpiration(t *testing.T) {
	assert.True(t, cacheExpired())
	akInfo := Refresh(token)
	cachedAkInfo = &akInfo
	assert.False(t, cacheExpired())
}
