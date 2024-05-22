package manager

import (
	"context"

	"github.com/mtgnorton/ws-cluster/core/client"
)

var DefaultManager = NewManager()

type ProjectServerClients struct {
	PID     string
	Servers []client.Client
	Clients []client.Client
}

// ClientManager 客户端管理
// Clients... 相关方法是获取用户客户端
type Manager interface {
	Join(ctx context.Context, client client.Client)
	Remove(ctx context.Context, client client.Client)
	//Clients 通过clientID获取用户客户端
	Clients(ctx context.Context, clientIDs ...string) []client.Client // 如果clientIDs为空，则返回所有client
	// ClientsByUIDs 通过uid获取用户客户端
	ClientsByUIDs(ctx context.Context, projectID string, userIDs ...string) []client.Client

	//ClientsByPIDs 通过pid获取用户客户端
	ClientsByPIDs(ctx context.Context, projectIDs ...string) []client.Client

	//ServersByPID 通过pid获取服务客户端
	ServersByPID(ctx context.Context, projectID string) []client.Client

	//Projects 获取所有项目的服务客户端和用户客户端
	Projects(ctx context.Context) []ProjectServerClients
	//Admins 获取所有管理客户端
	Admins(ctx context.Context) []client.Client
	Exist(ctx context.Context, clientID string) bool
	CheckExpired(ctx context.Context)
}
