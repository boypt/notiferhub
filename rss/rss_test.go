package rss

import (
	"log"
	"os"
	"testing"

	"github.com/spf13/viper"
)

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	// shutdown()
	os.Exit(code)
}

func setup() {
	viper.SetConfigName("cmddl")
	viper.AddConfigPath("/srv")
	viper.AddConfigPath("$HOME")
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalln("Error config file", err)
	}
}

func TestFindFromRSS(t *testing.T) {
	FindFromRSS()
}
