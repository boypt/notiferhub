package aria2rpc

import (
	"testing"
)

func TestAria2RPC_CallAria2Method(t *testing.T) {
	type args struct {
		method string
		args   []string
	}
	tests := []struct {
		name    string
		args    args
		want    *Aria2Resp
		wantErr bool
	}{
		// TODO: Add test cases.
		{"test", args{"aria2.getVersion", []string{}}, nil, false},
	}
	rpc := NewAria2RPC("ptpass", "http://127.0.0.1:2086/jsonrpc")
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := rpc.CallAria2Method(tt.args.method, tt.args.args)
			t.Logf("%v, %v", got, err)
			if (err != nil) != tt.wantErr {
				t.Errorf("Aria2RPC.CallAria2Method() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestAria2RPC_GetVersion(t *testing.T) {
	tests := []struct {
		name    string
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
		{"test", "1.31.0", false},
	}
	rpc := NewAria2RPC("ptpass", "http://127.0.0.1:2086/jsonrpc")
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := rpc.GetVersion()
			if (err != nil) != tt.wantErr {
				t.Errorf("Aria2RPC.GetVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Aria2RPC.GetVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAria2RPC_AddUris(t *testing.T) {
	type args struct {
		uris []string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{"test", args{[]string{"http://404domain.xz/123"}}, false},
	}
	rpc := NewAria2RPC("ptpass", "http://127.0.0.1:2086/jsonrpc")
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := rpc.AddUris(tt.args.uris); (err != nil) != tt.wantErr {
				t.Errorf("Aria2RPC.AddUris() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
