package tgbot

import (
	"os"
	"strconv"
	"testing"
)

func TestTGBot_SendMsg(t *testing.T) {
	type args struct {
		id   int64
		text string
	}

	chid, _ := strconv.ParseInt(os.Getenv("CHATID"), 10, 64)
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{"1", args{chid, "gotest text"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := NewTGBot(os.Getenv("BOTTOKEN"))
			if err := b.SendMsg(tt.args.id, tt.args.text, true); (err != nil) != tt.wantErr {
				t.Errorf("TGBot.SendMsg() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
