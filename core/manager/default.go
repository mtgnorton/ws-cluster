package manager

import (
	"context"
	"time"

	"github.com/sasha-s/go-deadlock"

	"ws-cluster/core/client"
	"ws-cluster/shared/kit"
)

// Project 属于同一个项目的用户和服务端放在同一个Project中
type Project struct {
	pid      string
	uClients map[string][]client.Client // 连接的用户端 key:uid
	sClients map[string][]client.Client // 连接的服务端 key:uid
}

// manager 管理所有客户端
type manager struct {
	opts     Options
	clients  map[string]client.Client // key:cid value:client
	projects map[string]Project       // key:pid value:Project
	deadlock.RWMutex
}

func (m *manager) Join(ctx context.Context, c client.Client) {

	cid, uid, pid := c.GetIDs()
	m.Lock()
	defer m.Unlock()
	m.clients[cid] = c
	if _, ok := m.projects[pid]; !ok {
		m.projects[pid] = Project{
			pid:      pid,
			uClients: make(map[string][]client.Client),
			sClients: make(map[string][]client.Client),
		}
	}
	switch c.Type() {
	case client.CTypeServer:
		m.projects[pid].sClients[uid] = append(m.projects[pid].sClients[uid], c)
	case client.CTypeUser:
		m.projects[pid].uClients[uid] = append(m.projects[pid].uClients[uid], c)
	}
	m.opts.logger.Debugf(ctx, "manager-join c %s", c)
}

func (m *manager) Remove(ctx context.Context, c client.Client) {
	cid, uid, pid := c.GetIDs()

	m.Lock()
	defer m.Unlock()
	if _, ok := m.clients[cid]; !ok {
		m.opts.logger.Debugf(ctx, "manager-remove c %s not exist", c)
		return
	}
	delete(m.clients, cid)

	switch c.Type() {
	case client.CTypeServer:
		kit.SliceRangeRemoveElements(m.projects[pid].sClients[uid], func(c client.Client) bool {
			return c.GetCID() == cid
		})
	case client.CTypeUser:
		kit.SliceRangeRemoveElements(m.projects[pid].uClients[uid], func(c client.Client) bool {
			return c.GetCID() == cid
		})
	}
	c.Close()

	m.opts.logger.Debugf(ctx, "manager-remove c %s", c)
}

func (m *manager) Clients(ctx context.Context, clientIDs ...string) []client.Client {
	m.RLock()
	defer m.RUnlock()

	clients := make([]client.Client, 0)
	if len(clientIDs) == 0 {
		for _, c := range m.clients {
			clients = append(clients, c)
		}
		return clients
	}
	for _, id := range clientIDs {
		if _, ok := m.clients[id]; !ok {
			m.opts.logger.Debugf(ctx, "Clients client %s not exist", id)
			continue
		}
		clients = append(clients, m.clients[id])
	}
	return clients
}

func (m *manager) ClientsByUIDs(ctx context.Context, projectID string, userIDs ...string) []client.Client {
	m.RLock()
	defer m.RUnlock()
	var clients []client.Client
	for _, uid := range userIDs {
		clients = append(clients, m.projects[projectID].uClients[uid]...)
	}

	return clients
}

func (m *manager) ClientsByPIDs(ctx context.Context, projectIDs ...string) []client.Client {
	m.RLock()
	defer m.RUnlock()
	var clients []client.Client
	for _, pid := range projectIDs {
		for _, c := range m.projects[pid].uClients {
			clients = append(clients, c...)
		}
	}
	return clients
}

func (m *manager) ServersByPID(ctx context.Context, projectID string) []client.Client {
	m.RLock()
	defer m.RUnlock()
	var clients []client.Client
	for _, c := range m.projects[projectID].sClients {
		clients = append(clients, c...)
	}
	return clients
}
func (m *manager) Projects(ctx context.Context) []ProjectAllClients {
	m.RLock()
	defer m.RUnlock()

	projects := make([]ProjectAllClients, 0)
	for pid, project := range m.projects {
		ps := ProjectAllClients{
			PID:     pid,
			Clients: make([]client.Client, 0),
			Servers: make([]client.Client, 0),
		}
		for _, cc := range project.uClients {
			ps.Clients = append(ps.Clients, cc...)
		}
		for _, cc := range project.sClients {
			ps.Servers = append(ps.Servers, cc...)
		}
		projects = append(projects, ps)
	}
	return projects
}

func (m *manager) Exist(ctx context.Context, clientID string) bool {
	m.RLock()
	defer m.RUnlock()
	_, ok := m.clients[clientID]
	return ok
}

func (m *manager) inifiniteCheckExpired(ctx context.Context) {
	m.RLock()
	defer m.RUnlock()
	for _, c := range m.clients {
		if time.Now().Unix()-c.GetInteractTime() > 15 {
			m.opts.logger.Debugf(ctx, "checkExpired client %s expired", c)
			m.Remove(ctx, c)
		}
	}
}

func NewManager(opts ...Option) Manager {
	options := NewOptions(opts...)
	m := &manager{
		opts:     options,
		clients:  make(map[string]client.Client),
		projects: make(map[string]Project),
	}
	go m.inifiniteCheckExpired(context.Background())
	return m
}
