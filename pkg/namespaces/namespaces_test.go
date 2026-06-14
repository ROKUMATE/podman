package namespaces

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUsernsMode(t *testing.T) {
	tests := []struct {
		mode        UsernsMode
		isHost      bool
		isKeepID    bool
		isNoMap     bool
		isAuto      bool
		isDefault   bool
		isPrivate   bool
		isNS        bool
		isContainer bool
		container   string
		valid       bool
	}{
		{mode: "", isDefault: true, isPrivate: true, valid: true},
		// "default" is the default *value* but is not in Valid()'s allowed list.
		{mode: "default", isDefault: true, isPrivate: true, valid: false},
		{mode: "host", isHost: true, valid: true},
		{mode: "private", isPrivate: true, valid: true},
		{mode: "keep-id", isKeepID: true, isPrivate: true, valid: true},
		{mode: "keep-id:uid=1000", isKeepID: true, isPrivate: true, valid: true},
		{mode: "nomap", isNoMap: true, isPrivate: true, valid: true},
		{mode: "auto", isAuto: true, isPrivate: true, valid: true},
		{mode: "auto:size=1000", isAuto: true, isPrivate: true, valid: true},
		{mode: "ns:/run/userns/x", isNS: true, isPrivate: true, valid: true},
		{mode: "container:ctr1", isContainer: true, container: "ctr1", valid: true},
		{mode: "container", isPrivate: true, valid: false},    // no name
		{mode: "container:", isContainer: true, valid: false}, // empty name
		{mode: "bogus", isPrivate: true, valid: false},
	}
	for _, tt := range tests {
		t.Run(string(tt.mode), func(t *testing.T) {
			assert.Equal(t, tt.isHost, tt.mode.IsHost(), "IsHost")
			assert.Equal(t, tt.isKeepID, tt.mode.IsKeepID(), "IsKeepID")
			assert.Equal(t, tt.isNoMap, tt.mode.IsNoMap(), "IsNoMap")
			assert.Equal(t, tt.isAuto, tt.mode.IsAuto(), "IsAuto")
			assert.Equal(t, tt.isDefault, tt.mode.IsDefaultValue(), "IsDefaultValue")
			assert.Equal(t, tt.isPrivate, tt.mode.IsPrivate(), "IsPrivate")
			assert.Equal(t, tt.isNS, tt.mode.IsNS(), "IsNS")
			assert.Equal(t, tt.isContainer, tt.mode.IsContainer(), "IsContainer")
			assert.Equal(t, tt.container, tt.mode.Container(), "Container")
			assert.Equal(t, tt.valid, tt.mode.Valid(), "Valid")
		})
	}
}

func TestUsernsModeNS(t *testing.T) {
	assert.True(t, UsernsMode("ns:/run/userns/x").IsNS())
	assert.Equal(t, "/run/userns/x", UsernsMode("ns:/run/userns/x").NS())
}

func TestUsernsModeGetKeepIDOptions(t *testing.T) {
	t.Run("wrong mode errors", func(t *testing.T) {
		_, err := UsernsMode("host").GetKeepIDOptions()
		assert.Error(t, err)
	})

	t.Run("keep-id without options", func(t *testing.T) {
		opts, err := UsernsMode("keep-id").GetKeepIDOptions()
		require.NoError(t, err)
		assert.Nil(t, opts.UID)
		assert.Nil(t, opts.GID)
		assert.Nil(t, opts.MaxSize)
	})

	t.Run("keep-id with uid, gid and size", func(t *testing.T) {
		opts, err := UsernsMode("keep-id:uid=1000,gid=2000,size=65536").GetKeepIDOptions()
		require.NoError(t, err)
		require.NotNil(t, opts.UID)
		require.NotNil(t, opts.GID)
		require.NotNil(t, opts.MaxSize)
		assert.Equal(t, uint32(1000), *opts.UID)
		assert.Equal(t, uint32(2000), *opts.GID)
		assert.Equal(t, uint32(65536), *opts.MaxSize)
	})

	t.Run("non-numeric value errors", func(t *testing.T) {
		_, err := UsernsMode("keep-id:uid=abc").GetKeepIDOptions()
		assert.Error(t, err)
	})

	t.Run("option without a value errors", func(t *testing.T) {
		_, err := UsernsMode("keep-id:uid").GetKeepIDOptions()
		assert.Error(t, err)
	})

	t.Run("unknown option errors", func(t *testing.T) {
		_, err := UsernsMode("keep-id:bogus=1").GetKeepIDOptions()
		assert.Error(t, err)
	})
}

func TestNetworkMode(t *testing.T) {
	tests := []struct {
		mode        NetworkMode
		isNone      bool
		isHost      bool
		isDefault   bool
		isBridge    bool
		isPasta     bool
		isPod       bool
		isPrivate   bool
		isContainer bool
		container   string
	}{
		{mode: "none", isNone: true, isPrivate: true},
		{mode: "host", isHost: true},
		{mode: "default", isDefault: true, isPrivate: true},
		{mode: "bridge", isBridge: true, isPrivate: true},
		{mode: "pasta", isPasta: true, isPrivate: true},
		{mode: "pasta:-T,5", isPasta: true, isPrivate: true},
		{mode: "pod", isPod: true, isPrivate: true},
		{mode: "container:web", isContainer: true, container: "web"},
		{mode: "mynet", isPrivate: true},
	}
	for _, tt := range tests {
		t.Run(string(tt.mode), func(t *testing.T) {
			assert.Equal(t, tt.isNone, tt.mode.IsNone(), "IsNone")
			assert.Equal(t, tt.isHost, tt.mode.IsHost(), "IsHost")
			assert.Equal(t, tt.isDefault, tt.mode.IsDefault(), "IsDefault")
			assert.Equal(t, tt.isBridge, tt.mode.IsBridge(), "IsBridge")
			assert.Equal(t, tt.isPasta, tt.mode.IsPasta(), "IsPasta")
			assert.Equal(t, tt.isPod, tt.mode.IsPod(), "IsPod")
			assert.Equal(t, tt.isPrivate, tt.mode.IsPrivate(), "IsPrivate")
			assert.Equal(t, tt.isContainer, tt.mode.IsContainer(), "IsContainer")
			assert.Equal(t, tt.container, tt.mode.Container(), "Container")
		})
	}
}

func TestNetworkModeUserDefined(t *testing.T) {
	assert.True(t, NetworkMode("mynet").IsUserDefined())
	assert.Equal(t, "mynet", NetworkMode("mynet").UserDefined())

	for _, builtin := range []NetworkMode{"bridge", "host", "none", "default", "container:web"} {
		assert.False(t, builtin.IsUserDefined(), "%s should not be user-defined", builtin)
		assert.Empty(t, builtin.UserDefined(), "%s UserDefined", builtin)
	}
}

func TestNetworkModeNS(t *testing.T) {
	assert.True(t, NetworkMode("ns:/run/netns/x").IsNS())
	assert.Equal(t, "/run/netns/x", NetworkMode("ns:/run/netns/x").NS())
}
