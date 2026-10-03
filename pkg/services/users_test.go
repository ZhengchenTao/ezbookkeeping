package services

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/settings"
	"github.com/mayswind/ezbookkeeping/pkg/utils"
)

func useInternalAuthSettingForTest(t *testing.T, enableInternalAuth bool) *UserService {
	originalConfig := settings.Container.GetCurrentConfig()
	settings.SetCurrentConfig(&settings.Config{
		EnableInternalAuth: enableInternalAuth,
	})

	t.Cleanup(func() {
		settings.SetCurrentConfig(originalConfig)
	})

	return Users
}

func TestIsCurrentPasswordConfirmed_InternalAuthEnabled(t *testing.T) {
	users := useInternalAuthSettingForTest(t, true)
	user := &models.User{Salt: "salt", Password: utils.EncodePassword("123456", "salt")}
	userWithoutPassword := &models.User{Salt: "salt"}

	assert.True(t, users.IsCurrentPasswordConfirmed("123456", user))
	assert.False(t, users.IsCurrentPasswordConfirmed("654321", user))
	assert.False(t, users.IsCurrentPasswordConfirmed("", user))
	assert.False(t, users.IsCurrentPasswordConfirmed("", userWithoutPassword))
	assert.False(t, users.IsCurrentPasswordConfirmed("123456", userWithoutPassword))
}

func TestIsCurrentPasswordConfirmed_InternalAuthDisabled(t *testing.T) {
	users := useInternalAuthSettingForTest(t, false)
	user := &models.User{Salt: "salt", Password: utils.EncodePassword("123456", "salt")}
	userWithoutPassword := &models.User{Salt: "salt"}

	assert.True(t, users.IsCurrentPasswordConfirmed("", user))
	assert.True(t, users.IsCurrentPasswordConfirmed("654321", user))
	assert.True(t, users.IsCurrentPasswordConfirmed("", userWithoutPassword))
}
