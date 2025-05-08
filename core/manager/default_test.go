package manager

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/mtgnorton/ws-cluster/core/client"
	"github.com/mtgnorton/ws-cluster/shared/kit"

	"golang.org/x/exp/rand"
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

func (m *mockClient) GetCID() string {
	return m.id
}

func (m *mockClient) GetUID() string {
	return m.uid
}

func (m *mockClient) GetPID() string {
	return m.pid
}

func TestBatchRemoveAndQuery(t *testing.T) {
	var (
		ctx     = context.Background()
		m       = NewManager()
		clients = make([]client.Client, 10000)
	)

	// 创建一万个模拟客户端并加入管理器
	for i := 0; i < 10000; i++ {
		clients[i] = &mockClient{
			id:  fmt.Sprintf("client-%d", i),
			uid: fmt.Sprintf("user-%d", i),
			pid: "test-project",
		}
		m.Join(ctx, clients[i])
	}

	var wg sync.WaitGroup
	wg.Add(2)

	// 协程1: 批量删除客户端
	go func() {
		defer wg.Done()
		var wg2 sync.WaitGroup
		wg2.Add(100)
		for i := 0; i < 100; i++ { // 启动100个协程
			go func(start int) {
				defer wg2.Done()
				// 直接计算每个协程需要处理的起始和结束索引
				begin := start * 100
				end := begin + 100
				if end > 10000 {
					end = 10000
				}
				// 移除指定范围内的客户端
				for j := begin; j < end; j++ {
					m.Remove(ctx, clients[j])
				}
			}(i)
		}
		wg2.Wait()
	}()

	// 协程2: 持续查询客户端
	go func() {
		defer wg.Done()
		timeConsume := kit.ConsumeTimeStaistics("manager-Clients")
		for i := 0; i < 100; i++ { // 查询100次
			clientID := fmt.Sprintf("client-%d", rand.Intn(10000))
			m.Clients(ctx, clientID)
			t.Log(timeConsume(fmt.Sprintf("manager-Clients-%d", i)))
		}
	}()

	wg.Wait()
	fmt.Println(111)
}
