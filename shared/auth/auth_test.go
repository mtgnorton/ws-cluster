package auth

import (
	"fmt"
	"testing"
)

func Test_Encode(t *testing.T) {
	r := MustEncode(&UserData{
		PID:        "66",
		UID:        "22222",
		ClientType: 0,
	})
	fmt.Println(r)
	r = "xhgZjcBLZ25e_sZO_71JmQ"
	user, err := Decode(r)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(user)
}
