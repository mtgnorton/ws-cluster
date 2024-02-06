package manager

import (
	"context"

	"github.com/mtgnorton/ws-cluster/core/client"
)

var DefaultManager = NewManager()

type Manager interface {
	Join(ctx context.Context, client client.Client)
	Remove(ctx context.Context, client client.Client)
	BindTag(ctx context.Context, client client.Client, tags ...string)
	UnbindTag(ctx context.Context, client client.Client, tags ...string)
	Clients(ctx context.Context, clientIDs ...string) []client.Client // 如果clientIDs为空，则返回所有client
	ClientsByUIDs(ctx context.Context, userIDs ...string) []client.Client
	ClientsByPIDs(ctx context.Context, projectIDs ...string) []client.Client
	ClientByTags(ctx context.Context, tags ...string) []client.Client
	ServersByPID(ctx context.Context, projectID string) []client.Client
	Admins(ctx context.Context) []client.Client
	Exist(ctx context.Context, clientID string) bool
	CheckExpired(ctx context.Context)
}
