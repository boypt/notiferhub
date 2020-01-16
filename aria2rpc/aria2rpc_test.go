package aria2rpc

import (
	"os"
	"testing"
)

var (
	rpc = NewAria2RPC(os.Getenv("ARIA2TOKEN"), os.Getenv("ARIA2RPC"))
)

func TestAria2RPC_GetVersion(t *testing.T) {
	tests := []struct {
		name    string
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
		{"test", "1.31.0", false},
	}
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
	if _, err := rpc.AddUri("http://404domain.xz/123", "123"); err != nil {
		t.Error(err)
	}
}

func TestAria2RPC_Tellstatus(t *testing.T) {
	if resp, err := rpc.TellStatus("13c1bf22c0705e64"); err != nil {
		t.Error(err)
	} else {
		t.Logf("%#v", resp)
	}
}
