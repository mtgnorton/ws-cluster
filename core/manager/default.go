package manager

import (
	"context"
	"time"

	"github.com/sasha-s/go-deadlock"

	"github.com/mtgnorton/ws-cluster/core/client"
)

type manager struct {
	opts          Options
	clients       map[string]client.Client // key:clientID value:client
	uClients      map[string][]string      // key:uid value:clientID
	pClients      map[string][]string      // key:pid value:clientID
	tagClients    map[string][]string      // key:tag value:clientID
	serverClients map[string][]string      // key:pid value:clientID
	adminClients  []string                 // value:clientID
	deadlock.RWMutex
}

func (m *manager) Join(ctx context.Context, c client.Client) {

	id, uid, pid := c.GetIDs()
	m.Lock()
	defer m.Unlock()
	m.clients[id] = c
	switch c.Type() {
	case client.CTypeAdmin:
		m.adminClients = append(m.adminClients, id)
	case client.CTypeServer:
		m.serverClients[pid] = append(m.serverClients[pid], id)
	case client.CTypeUser:
		m.uClients[uid] = append(m.uClients[uid], id)
		m.pClients[pid] = append(m.pClients[pid], id)
	}
	m.opts.logger.Debugf(ctx, "manager-join c %s", c)
}

func (m *manager) Remove(ctx context.Context, c client.Client) {
	id, uid, pid := c.GetIDs()
	c.Close()

	m.Lock()
	defer m.Unlock()
	delete(m.clients, id)

	switch c.Type() {
	case client.CTypeAdmin:
		for i, cid := range m.adminClients {
			if cid == id {
				m.adminClients = append(m.adminClients[:i], m.adminClients[i+1:]...)
			}
		}
	case client.CTypeServer:
		for i, cid := range m.serverClients[pid] {
			if cid == id {
				m.serverClients[pid] = append(m.serverClients[pid][:i], m.serverClients[pid][i+1:]...)
			}
		}
	case client.CTypeUser:
		for i, cid := range m.uClients[uid] {
			if cid == id {
				m.uClients[uid] = append(m.uClients[uid][:i], m.uClients[uid][i+1:]...)
			}
		}
		for i, cid := range m.pClients[pid] {
			if cid == id {
				m.pClients[pid] = append(m.pClients[pid][:i], m.pClients[pid][i+1:]...)
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

func (m *manager) ClientsByUIDs(ctx context.Context, userIDs ...string) []client.Client {
	m.RLock()
	defer m.RUnlock()
	var clients []client.Client
	for _, userID := range userIDs {
		for _, id := range m.uClients[userID] {
			clients = append(clients, m.clients[id])
		}
	}

	return clients
}

func (m *manager) ClientsByPIDs(ctx context.Context, projectIDs ...string) []client.Client {
	m.RLock()
	defer m.RUnlock()
	var clients []client.Client
	for _, pid := range projectIDs {
		for _, id := range m.pClients[pid] {
			clients = append(clients, m.clients[id])
		}
	}
	return clients
}

func (m *manager) ClientByTags(ctx context.Context, tags ...string) []client.Client {
	m.RLock()
	defer m.RUnlock()
	var clients []client.Client
	for _, tag := range tags {
		for _, id := range m.tagClients[tag] {
			clients = append(clients, m.clients[id])
		}
	}
	return clients
}

func (m *manager) ServersByPID(ctx context.Context, projectID string) []client.Client {
	m.RLock()
	defer m.RUnlock()
	var clients []client.Client
	for pid, ids := range m.serverClients {
		if pid == projectID {
			for _, id := range ids {
				clients = append(clients, m.clients[id])
			}
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
		if time.Now().Unix()-c.GetReplyTime() > 60 {
			m.opts.logger.Debugf(ctx, "checkExpired client %s expired", c)
			m.Remove(ctx, c)
		}
	}
}

func (m *manager) BindTag(ctx context.Context, client client.Client, tags ...string) {
	m.Lock()
	defer m.Unlock()
	id, _, _ := client.GetIDs()
	for _, tag := range tags {
		m.tagClients[tag] = append(m.tagClients[tag], id)
	}
	m.opts.logger.Debugf(ctx, "manager-BindTag  %s to client %s", tags, client)
}
func (m *manager) UnbindTag(ctx context.Context, client client.Client, tags ...string) {
	m.Lock()
	defer m.Unlock()
	id, _, _ := client.GetIDs()
	for _, tag := range tags {
		for i, cid := range m.tagClients[tag] {
			if cid == id {
				m.tagClients[tag] = append(m.tagClients[tag][:i], m.tagClients[tag][i+1:]...)
			}
		}
	}
	m.opts.logger.Debugf(ctx, "manager-UnbindTag  %s from client %s", tags, client)
}

func NewManager(opts ...Option) Manager {
	options := NewOptions(opts...)
	return &manager{
		opts:          options,
		clients:       make(map[string]client.Client),
		uClients:      make(map[string][]string),
		pClients:      make(map[string][]string),
		tagClients:    make(map[string][]string),
		serverClients: make(map[string][]string),
	}
}
