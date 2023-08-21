package login

import (
	"testing"

	"github.com/ciiim/cloudborad/auth"
)

func Test_loginData_createIdentity(t *testing.T) {
	type fields struct {
		uid      uint64
		username string
		passwd   string
	}
	tests := []struct {
		name   string
		fields fields
		want   auth.IdentifyState
	}{
		{name: "case1", fields: fields{
			uid:      10,
			username: "123",
			passwd:   "123",
		},
			want: auth.Vaild,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &Login{
				Data: LoginData{
					uid:      tt.fields.uid,
					username: tt.fields.username,
					passwd:   tt.fields.passwd,
				},
			}
			got := d.createIdentity()
			uid, state, _ := got.Check()
			t.Log("uid:", uid)
			if state != auth.Vaild {
				t.Errorf("loginData.createIdentity() = %v, want %v", state, auth.Vaild)
			}
		})
	}
}
