package manager

import (
	"github.com/mtgnorton/ws-cluster/client"
)

var DefaultManager = NewManager()

type Manager interface {
	Join(client client.Client)
	Remove(client client.Client)
	Gets(clientIDs ...string) []client.Client
	GetByUIDs(userIDs ...string) []client.Client
	GetByPIDs(projectIDs ...string) []client.Client
	GetByTags(tags ...string) []client.Client
	All() []client.Client
	Exist(clientID string) bool
	CheckExpired()
}

type ToggleTag interface {
	BindTag(client client.Client, tags ...string)
	UnbindTag(client client.Client, tags ...string)
}
