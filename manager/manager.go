package manager

import (
	"github.com/mtgnorton/ws-cluster/client"
	"sync"
	"time"
)

var DefaultManager = NewManager()

type Manager interface {
	Join(client client.Client)
	Remove(client client.Client)
	Get(clientID string) client.Client
	GetByUID(uid string) []client.Client
	GetByPID(pid string) []client.Client
	All() []client.Client
	Exist(clientID string) bool
	CheckExpired()
}

type manager struct {
	opts     Options
	clients  map[string]client.Client
	uclients map[string][]string
	pclients map[string][]string
	sync.RWMutex
}

func (m *manager) Join(client client.Client) {
	id, uid, pid := client.GetIDs()

	m.Lock()
	defer m.Unlock()
	m.clients[id] = client
	m.uclients[uid] = append(m.uclients[uid], id)
	m.pclients[pid] = append(m.pclients[pid], id)
}

func (m *manager) Remove(client client.Client) {
	id, uid, pid := client.GetIDs()
	client.Close()
	m.Lock()
	defer m.Unlock()
	delete(m.clients, id)
	for i, cid := range m.uclients[uid] {
		if cid == id {
			m.uclients[uid] = append(m.uclients[uid][:i], m.uclients[uid][i+1:]...)
		}
	}
	for i, cid := range m.pclients[pid] {
		if cid == id {
			m.pclients[pid] = append(m.pclients[pid][:i], m.pclients[pid][i+1:]...)
		}
	}
}

func (m *manager) Get(clientID string) client.Client {
	m.RLock()
	defer m.RUnlock()
	return m.clients[clientID]
}

func (m *manager) GetByUID(uid string) []client.Client {
	m.RLock()
	defer m.RUnlock()
	var clients []client.Client
	for _, id := range m.uclients[uid] {
		clients = append(clients, m.clients[id])
	}
	return clients
}

func (m *manager) GetByPID(pid string) []client.Client {
	m.RLock()
	defer m.RUnlock()
	var clients []client.Client
	for _, id := range m.pclients[pid] {
		clients = append(clients, m.clients[id])
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

func NewManager(opts ...Option) Manager {
	options := Options{}
	for _, o := range opts {
		o(&options)
	}
	return &manager{
		opts:     options,
		clients:  make(map[string]client.Client),
		uclients: make(map[string][]string),
		pclients: make(map[string][]string),
	}
}
