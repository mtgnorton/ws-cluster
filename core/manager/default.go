package manager

import (
	"github.com/mtgnorton/ws-cluster/core/client"
	"sync"
	"time"
)

type manager struct {
	opts       Options
	clients    map[string]client.Client
	uClients   map[string][]string
	pClients   map[string][]string
	tagClients map[string][]string
	sync.RWMutex
}

func (m *manager) Join(client client.Client) {

	id, uid, pid := client.GetIDs()
	m.Lock()
	defer m.Unlock()
	m.clients[id] = client
	m.uClients[uid] = append(m.uClients[uid], id)
	m.pClients[pid] = append(m.pClients[pid], id)
	m.opts.logger.Debugf("manager-join client %s", client)
}

func (m *manager) Remove(client client.Client) {
	id, uid, pid := client.GetIDs()
	client.Close()
	m.Lock()
	defer m.Unlock()
	delete(m.clients, id)
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
	m.opts.logger.Debugf("manager-remove client %s", client)
}

func (m *manager) Gets(clientIDs ...string) []client.Client {
	m.RLock()
	defer m.RUnlock()
	clients := make([]client.Client, 0)
	for _, id := range clientIDs {
		clients = append(clients, m.clients[id])
	}
	return clients
}

func (m *manager) GetByUIDs(userIDs ...string) []client.Client {
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

func (m *manager) GetByPIDs(projectIDs ...string) []client.Client {
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

func (m *manager) GetByTags(tags ...string) []client.Client {
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

func (m *manager) All() []client.Client {
	m.RLock()
	defer m.RUnlock()
	var clients []client.Client
	for _, c := range m.clients {
		clients = append(clients, c)
	}
	return clients
}

func (m *manager) Exist(clientID string) bool {
	m.RLock()
	defer m.RUnlock()
	_, ok := m.clients[clientID]
	return ok
}

func (m *manager) CheckExpired() {
	m.checkExpired()
}

func (m *manager) checkExpired() {
	m.RLock()
	defer m.RUnlock()
	for _, c := range m.clients {
		if time.Now().Unix()-c.GetReplyTime() > 60 {
			m.opts.logger.Debugf("checkExpired client %s expired", c)
			m.Remove(c)
		}
	}
}

func (m *manager) BindTag(client client.Client, tags ...string) {
	m.Lock()
	defer m.Unlock()
	id, _, _ := client.GetIDs()
	for _, tag := range tags {
		m.tagClients[tag] = append(m.tagClients[tag], id)
	}
	m.opts.logger.Debugf("manager-BindTag  %s to client %s", tags, client)
}
func (m *manager) UnbindTag(client client.Client, tags ...string) {
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
	m.opts.logger.Debugf("manager-UnbindTag  %s from client %s", tags, client)
}

func NewManager(opts ...Option) Manager {
	options := NewOptions(opts...)
	return &manager{
		opts:       options,
		clients:    make(map[string]client.Client),
		uClients:   make(map[string][]string),
		pClients:   make(map[string][]string),
		tagClients: make(map[string][]string),
	}
}
