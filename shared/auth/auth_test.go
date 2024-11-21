package auth

import (
	"fmt"
	"testing"
)

func Test_Encode(t *testing.T) {
	r := MustEncode(&UserData{
		PID:        "578",
		UID:        "1996",
		ClientType: 32832,
	})
	fmt.Println(r)
	fmt.Println(Decode(r))
}

func Test_Decode(t *testing.T) {
	r := "1RsNiOc8sqaIRJ0j95p-aVObzarvZOFQPc-kbbOWXX0="
	//fmt.Println(r[44])
	user, err := Decode(r)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(user)
}
