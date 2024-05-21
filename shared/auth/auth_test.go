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
	r = "hhV4m8A27TRcSAAxcws5YA"
	user, err := Decode(r)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(user)
}
