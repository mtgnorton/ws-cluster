package handler

import (
	"github.com/mtgnorton/ws-cluster/core/client"
)

func intersect(s1, s2 []client.Client) (c []client.Client) {
	if len(s1) == 0 || len(s2) == 0 {
		return
	}
	m := make(map[string]struct{})

	for _, c1 := range s1 {
		c1ID, _, _ := c1.GetIDs()
		m[c1ID] = struct{}{}
	}
	for _, c2 := range s2 {
		c2ID, _, _ := c2.GetIDs()
		if _, ok := m[c2ID]; ok {
			c = append(c, c2)
		}
	}
	return
}
