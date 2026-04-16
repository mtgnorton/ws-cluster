package manager

import (
	"context"
	"sync"
	"time"

	"github.com/mtgnorton/ws-cluster/core/client"
)

// Project 属于同一个项目的用户和服务端放在同一个Project中
type Project struct {
	pid      string
	uClients map[string]map[string]client.Client // 连接的用户端 key:uid->cid->client
	sClients map[string]map[string]client.Client // 连接的服务端 key:uid->cid->client
}

// manager 管理所有客户端
type manager struct {
	opts     Options
	clients  map[string]client.Client // key:cid value:client
	projects map[string]Project       // key:pid value:Project
	sync.RWMutex
}

func (m *manager) Join(ctx context.Context, c client.Client) {
	cid, uid, pid := c.GetIDs()
	m.Lock()
	defer m.Unlock()
	m.clients[cid] = c
	if _, ok := m.projects[pid]; !ok {
		m.projects[pid] = Project{
			pid:      pid,
			uClients: make(map[string]map[string]client.Client),
			sClients: make(map[string]map[string]client.Client),
		}
	}
	switch c.Type() {
	case client.CTypeServer:
		if _, ok := m.projects[pid].sClients[uid]; !ok {
			m.projects[pid].sClients[uid] = make(map[string]client.Client)
		}
		m.projects[pid].sClients[uid][cid] = c
	case client.CTypeUser:
		if _, ok := m.projects[pid].uClients[uid]; !ok {
			m.projects[pid].uClients[uid] = make(map[string]client.Client)
		}
		m.projects[pid].uClients[uid][cid] = c
	}
	m.opts.logger.Debugf(ctx, "manager-join c %s", c)
}

func (m *manager) Remove(ctx context.Context, c client.Client) {
	cid, uid, pid := c.GetIDs()

	m.Lock()
	if _, ok := m.clients[cid]; !ok {
		m.Unlock()
		m.opts.logger.Debugf(ctx, "manager-remove c %s not exist", c)
		return
	}
	delete(m.clients, cid)
	if project, ok := m.projects[pid]; ok {
		switch c.Type() {
		case client.CTypeServer:
			if serverClients, ok := project.sClients[uid]; ok {
				delete(serverClients, cid)
				if len(serverClients) == 0 {
					delete(project.sClients, uid)
				}
			}
		case client.CTypeUser:
			if userClients, ok := project.uClients[uid]; ok {
				delete(userClients, cid)
				if len(userClients) == 0 {
					delete(project.uClients, uid)
				}
			}
		}
		if len(project.uClients) == 0 && len(project.sClients) == 0 {
			delete(m.projects, pid)
		}
	}
	m.Unlock()

	c.Close()

	m.opts.logger.Debugf(ctx, "manager-remove c %s", c)
}

func (m *manager) Clients(ctx context.Context, clientIDs ...string) []client.Client {
	m.RLock()
	defer m.RUnlock()

	clients := make([]client.Client, 0, len(clientIDs))
	if len(clientIDs) == 0 {
		clients = make([]client.Client, 0, len(m.clients))
		for _, c := range m.clients {
			clients = append(clients, c)
		}
		return clients
	}
	for _, id := range clientIDs {
		if c, ok := m.clients[id]; ok {
			clients = append(clients, c)
		}
	}
	return clients
}

func (m *manager) ClientsByUIDs(ctx context.Context, projectID string, userIDs ...string) []client.Client {
	m.RLock()
	defer m.RUnlock()
	var clients []client.Client
	for _, uid := range userIDs {
		for _, c := range m.projects[projectID].uClients[uid] {
			clients = append(clients, c)
		}
	}
	return clients
}

func (m *manager) ClientsByPIDs(ctx context.Context, projectIDs ...string) []client.Client {
	m.RLock()
	defer m.RUnlock()
	var clients []client.Client
	for _, pid := range projectIDs {
		for _, userClients := range m.projects[pid].uClients {
			for _, c := range userClients {
				clients = append(clients, c)
			}
		}
	}
	return clients
}

func (m *manager) ServersByPID(ctx context.Context, projectID string) []client.Client {
	m.RLock()
	defer m.RUnlock()
	var clients []client.Client
	for _, serverClients := range m.projects[projectID].sClients {
		for _, c := range serverClients {
			clients = append(clients, c)
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
			PID:     pid,
			Clients: make([]client.Client, 0),
			Servers: make([]client.Client, 0),
		}
		for _, userClients := range project.uClients {
			for _, c := range userClients {
				ps.Clients = append(ps.Clients, c)
			}
		}
		for _, serverClients := range project.sClients {
			for _, c := range serverClients {
				ps.Servers = append(ps.Servers, c)
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

func (m *manager) infiniteCheckExpired(ctx context.Context) {
	ticker := time.NewTicker(time.Second * 10)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			expiredClients := make([]client.Client, 0)
			m.RLock()
			for _, c := range m.clients {
				if c.Type() != client.CTypeUser {
					continue
				}
				if time.Now().Unix()-c.GetInteractTime() > 15 {
					expiredClients = append(expiredClients, c)
				}
			}
			m.RUnlock()
			for _, expiredClient := range expiredClients {
				m.opts.logger.Debugf(ctx, "checkExpired client %s expired", expiredClient)
				expiredClient.Close()
			}
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
	go m.infiniteCheckExpired(options.ctx)
	return m
}
