package jwtws

import (
	"testing"

	"github.com/mtgnorton/ws-cluster/config"
)

// pid 1 uid 1  user  eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJwaWQiOiIxIiwidWlkIjoiMSIsImNsaWVudF90eXBlIjowLCJleHAiOjIxNzk3NDQ0MDYsImlzcyI6IndzLWNsdXN0ZXIifQ.xExWBV4l4L8C6UoDc68DSanzjCQSnJOVJCxyEzxnymU
// pid 1 uid 2  user  eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJwaWQiOiIxIiwidWlkIjoiMiIsImNsaWVudF90eXBlIjowLCJleHAiOjIxNzk3NDQ0NTIsImlzcyI6IndzLWNsdXN0ZXIifQ.gL5gRohHHArvrNnpKyogfYG2jdOPAjVQ3SHH8QX_-XE
// pid 1 uid 99999 server  eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJwaWQiOiIxIiwidWlkIjoiOTk5OTkiLCJjbGllbnRfdHlwZSI6MSwiZXhwIjoyMTc5NzQ0NTM0LCJpc3MiOiJ3cy1jbHVzdGVyIn0.lDVixBSnT9nK7gsVvHsUtDk8qXA9BdwhX7Y2bqrrvtg
func Test_Jwt(t *testing.T) {
	defaultJwtWs := NewJwtWs(config.DefaultConfig)
	token, err := defaultJwtWs.Generate("1", "99999", 1)
	if err != nil {
		t.Error(err)
	}
	t.Log(token)
	c, err := defaultJwtWs.Parse(token)
	if err != nil {
		t.Error(err)
	}
	t.Logf("%+v", c)
	if c.PID != "1" {
		t.Error("pid error")
	}
}
