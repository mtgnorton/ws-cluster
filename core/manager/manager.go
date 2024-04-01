package manager

import (
	"context"

	"github.com/mtgnorton/ws-cluster/core/client"
)

var DefaultManager = NewManager()

// ClientManager 客户端管理
// Clients... 相关方法是获取用户客户端
type Manager interface {
	Join(ctx context.Context, client client.Client)
	Remove(ctx context.Context, client client.Client)
	// 通过clientID获取用户客户端
	Clients(ctx context.Context, clientIDs ...string) []client.Client // 如果clientIDs为空，则返回所有client
	// 通过uid获取用户客户端
	ClientsByUIDs(ctx context.Context, projectID string, userIDs ...string) []client.Client

	// 通过pid获取用户客户端
	ClientsByPIDs(ctx context.Context, projectIDs ...string) []client.Client

	// 通过pid获取服务客户端
	ServersByPID(ctx context.Context, projectID string) []client.Client

	// 获取所有管理客户端
	Admins(ctx context.Context) []client.Client
	Exist(ctx context.Context, clientID string) bool
	CheckExpired(ctx context.Context)
}
