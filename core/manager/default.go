package manager

import (
	"context"
	"time"

	"github.com/sasha-s/go-deadlock"

	"github.com/mtgnorton/ws-cluster/core/client"
)

// PClients 属于同一个项目的用户和服务端放在同一个PClients中
type PClients struct {
	pid      string
	uClients map[string][]string // 连接的用户端 key:uid value:[]clientID
	sClients map[string][]string // 连接的服务端 key:uid value:[]clientID
}
type manager struct {
	opts         Options
	clients      map[string]client.Client // key:clientID value:client
	pClients     map[string]PClients      // key:pid value:PUClient
	adminClients []string                 // value:clientID
	deadlock.RWMutex
}

func (m *manager) Join(ctx context.Context, c client.Client) {

	cid, uid, pid := c.GetIDs()
	m.Lock()
	defer m.Unlock()
	m.clients[cid] = c
	switch c.Type() {
	case client.CTypeAdmin:
		m.adminClients = append(m.adminClients, cid)
	case client.CTypeServer:
		if _, ok := m.pClients[pid]; !ok {
			m.pClients[pid] = PClients{
				pid:      pid,
				uClients: make(map[string][]string),
				sClients: make(map[string][]string),
			}
		}
		m.pClients[pid].sClients[uid] = append(m.pClients[pid].sClients[uid], cid)
	case client.CTypeUser:
		if _, ok := m.pClients[pid]; !ok {
			m.pClients[pid] = PClients{
				pid:      pid,
				uClients: make(map[string][]string),
				sClients: make(map[string][]string),
			}
		}
		m.pClients[pid].uClients[uid] = append(m.pClients[pid].uClients[uid], cid)
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
	case client.CTypeAdmin:
		for i, tempCid := range m.adminClients {
			if tempCid == cid {
				m.adminClients = append(m.adminClients[:i], m.adminClients[i+1:]...)
			}
		}
	case client.CTypeServer:
		for i, tempCid := range m.pClients[pid].sClients[uid] {
			if tempCid == cid {
				m.pClients[pid].sClients[uid] = append(m.pClients[pid].sClients[uid][:i], m.pClients[pid].sClients[uid][i+1:]...)
			}
		}
	case client.CTypeUser:
		for i, tempCid := range m.pClients[pid].uClients[uid] {
			if tempCid == cid {
				m.pClients[pid].uClients[uid] = append(m.pClients[pid].uClients[uid][:i], m.pClients[pid].uClients[uid][i+1:]...)
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
		clients = append(clients, m.clients[id])
	}
	return clients
}

func (m *manager) ClientsByUIDs(ctx context.Context, projectID string, userIDs ...string) []client.Client {
	m.RLock()
	defer m.RUnlock()
	var clients []client.Client
	for _, uid := range userIDs {
		for _, cid := range m.pClients[projectID].uClients[uid] {
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
		for _, ids := range m.pClients[pid].uClients {
			for _, id := range ids {
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
	for _, ids := range m.pClients[projectID].sClients {
		for _, id := range ids {
			clients = append(clients, m.clients[id])
		}
	}
	return clients
}

func (m *manager) Admins(ctx context.Context) []client.Client {
	m.RLock()
	defer m.RUnlock()
	var clients []client.Client
	for _, id := range m.adminClients {
		clients = append(clients, m.clients[id])
	}
	return clients
}

func (m *manager) Exist(ctx context.Context, clientID string) bool {
	m.RLock()
	defer m.RUnlock()
	_, ok := m.clients[clientID]
	return ok
}

func (m *manager) CheckExpired(ctx context.Context) {
	m.checkExpired(ctx)
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
		opts:         options,
		clients:      make(map[string]client.Client),
		pClients:     make(map[string]PClients),
		adminClients: make([]string, 0),
	}
}
