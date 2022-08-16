package tgbot

import (
	"testing"

	"github.com/spf13/viper"
)

func TestTGBot_SendMsg(t *testing.T) {
	type args struct {
		id   string
		text string
	}

	chid := viper.GetString("CHATID")
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
			b := NewTGBot(viper.GetString("BOTTOKEN"))
			if _, err := b.SendMsg(tt.args.id, tt.args.text, true); (err != nil) != tt.wantErr {
				t.Errorf("TGBot.SendMsg() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
