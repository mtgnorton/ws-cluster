package manager

import (
	"context"
	"time"

	"github.com/sasha-s/go-deadlock"

	"ws-cluster/core/client"
)

// Project 属于同一个项目的用户和服务端放在同一个PClients中
type Project struct {
	pid      string
	uClients map[string][]string // 连接的用户端 key:uid value:[]clientID
	sClients map[string][]string // 连接的服务端 key:uid value:[]clientID
}

// manager 管理所有客户端
type manager struct {
	opts     Options
	clients  map[string]client.Client // key:clientID value:client
	projects map[string]Project       // key:pid value:PUClient
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
			uClients: make(map[string][]string),
			sClients: make(map[string][]string),
		}
	}
	switch c.Type() {
	case client.CTypeServer:
		m.projects[pid].sClients[uid] = append(m.projects[pid].sClients[uid], cid)
	case client.CTypeUser:
		m.projects[pid].uClients[uid] = append(m.projects[pid].uClients[uid], cid)
	}
	m.opts.logger.Debugf(ctx, "manager-join c %s", c)
}

func (m *manager) Remove(ctx context.Context, c client.Client) {
	cid, uid, pid := c.GetIDs()
	c.Close()

	m.Lock()
	defer m.Unlock()
	delete(m.clients, cid)

	switch c.Type() {
	case client.CTypeServer:
		for i, tempCid := range m.projects[pid].sClients[uid] {
			if tempCid == cid {
				m.projects[pid].sClients[uid] = append(m.projects[pid].sClients[uid][:i], m.projects[pid].sClients[uid][i+1:]...)
			}
		}
	case client.CTypeUser:
		for i, tempCid := range m.projects[pid].uClients[uid] {
			if tempCid == cid {
				m.projects[pid].uClients[uid] = append(m.projects[pid].uClients[uid][:i], m.projects[pid].uClients[uid][i+1:]...)
			}
		}
	}
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
		for _, cid := range m.projects[projectID].uClients[uid] {
			if _, ok := m.clients[cid]; !ok {
				m.opts.logger.Debugf(ctx, "ClientsByUIDs client %s not exist", cid)
				continue
			}
			clients = append(clients, m.clients[cid])
		}
	}

	return clients
}

func (m *manager) ClientsByPIDs(ctx context.Context, projectIDs ...string) []client.Client {
	m.RLock()
	defer m.RUnlock()
	var clients []client.Client
	for _, pid := range projectIDs {
		for _, ids := range m.projects[pid].uClients {
			for _, id := range ids {
				if _, ok := m.clients[id]; !ok {
					m.opts.logger.Debugf(ctx, "ClientsByPIDs client %s not exist", id)
					continue
				}
				clients = append(clients, m.clients[id])
			}
		}
	}
	return clients
}

func (m *manager) ServersByPID(ctx context.Context, projectID string) []client.Client {
	m.RLock()
	defer m.RUnlock()
	var clients []client.Client
	for _, ids := range m.projects[projectID].sClients {
		for _, id := range ids {
			if _, ok := m.clients[id]; !ok {
				continue
			}
			clients = append(clients, m.clients[id])
		}
	}
	return clients
}
func (m *manager) Projects(ctx context.Context) []ProjectAllClients {
	m.RLock()
	defer m.RUnlock()

	projects := make([]ProjectAllClients, 0)
	for pid, project := range m.projects {
		ps := ProjectAllClients{
			PID: pid,
		}
		for _, ids := range project.uClients {
			for _, id := range ids {
				if _, ok := m.clients[id]; !ok {
					continue
				}
				ps.Clients = append(ps.Clients, m.clients[id])
			}
		}
		for _, ids := range project.sClients {
			for _, id := range ids {
				if _, ok := m.clients[id]; !ok {
					continue
				}
				ps.Servers = append(ps.Servers, m.clients[id])
			}
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

func (m *manager) checkExpired(ctx context.Context) {
	m.RLock()
	defer m.RUnlock()
	for _, c := range m.clients {
		if time.Now().Unix()-c.GetInteractTime() > 60 {
			m.opts.logger.Debugf(ctx, "checkExpired client %s expired", c)
			m.Remove(ctx, c)
		}
	}
}

func NewManager(opts ...Option) Manager {
	options := NewOptions(opts...)
	return &manager{
		opts:     options,
		clients:  make(map[string]client.Client),
		projects: make(map[string]Project),
	}
}
