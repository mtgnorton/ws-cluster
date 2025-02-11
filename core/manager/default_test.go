package manager

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"ws-cluster/core/client"
)

// TestConcurrentAccess 测试并发访问时的锁竞争情况
func TestConcurrentAccess(t *testing.T) {

	m := NewManager()
	ctx := context.Background()

	// 创建多个项目的客户端
	const numProjects = 10
	const numClientsPerProject = 100
	clients := make([][]client.Client, numProjects)
	for i := 0; i < numProjects; i++ {
		clients[i] = make([]client.Client, numClientsPerProject)
		for j := 0; j < numClientsPerProject; j++ {
			c := &mockClient{
				id:  fmt.Sprintf("client_%d_%d", i, j),
				uid: fmt.Sprintf("user_%d_%d", i, j),
				pid: fmt.Sprintf("project_%d", i),
			}
			clients[i][j] = c
		}
	}

	var wg sync.WaitGroup

	// 并发添加客户端
	wg.Add(numProjects * numClientsPerProject)
	for i := 0; i < numProjects; i++ {
		for j := 0; j < numClientsPerProject; j++ {
			go func(projectIdx, clientIdx int) {
				defer wg.Done()
				time.Sleep(time.Duration(projectIdx*100) * time.Millisecond) // 增加时间间隔
				m.Join(ctx, clients[projectIdx][clientIdx])
			}(i, j)
		}
	}
	wg.Wait()

	// 并发读取客户端
	wg.Add(numProjects * 10) // 每个项目10个并发读取
	for i := 0; i < numProjects; i++ {
		for j := 0; j < 10; j++ {
			go func(projectIdx int) {
				defer wg.Done()
				time.Sleep(time.Second) // 持续读取1秒
				m.ClientsByPIDs(ctx, fmt.Sprintf("project_%d", projectIdx))
			}(i)
		}
	}
	wg.Wait()

	// 并发删除客户端
	wg.Add(numProjects * numClientsPerProject)
	for i := 0; i < numProjects; i++ {
		for j := 0; j < numClientsPerProject; j++ {
			go func(projectIdx, clientIdx int) {
				defer wg.Done()
				time.Sleep(time.Duration(projectIdx*100) * time.Millisecond) // 增加时间间隔
				m.Remove(ctx, clients[projectIdx][clientIdx])
			}(i, j)
		}
	}
	wg.Wait()
}

// mockClient 用于测试的模拟客户端
type mockClient struct {
	id  string
	uid string
	pid string
}

func (m *mockClient) Init(opts ...client.Option) {
}

func (m *mockClient) Options() client.Options {
	return client.Options{}
}

func (m *mockClient) Send(ctx context.Context, message interface{}) {
}

func (m *mockClient) Status() client.Status {
	return client.Status(0)
}

func (m *mockClient) UpdateInteractTime() {
}

func (m *mockClient) GetIDs() (string, string, string) {
	return m.id, m.uid, m.pid
}

func (m *mockClient) Type() client.CType {
	return client.CTypeUser
}

func (m *mockClient) GetInteractTime() int64 {
	return time.Now().Unix()
}

func (m *mockClient) Close() {}

func (m *mockClient) String() string {
	return m.id
}
