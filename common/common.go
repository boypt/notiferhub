package common

import (
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// Must ...
func Must(err error) {
	if err != nil {
		panic(err)
	}
}

// Must2 ...
func Must2(v interface{}, err error) interface{} {
	Must(err)
	return v
}

// Error2 ...
func Error2(v interface{}, err error) error {
	return err
}

func HumaneSize(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB",
		float64(b)/float64(div), "KMGTPE"[exp])
}

func KitchenDuration(d time.Duration) string {
	d = d.Round(time.Second)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second
	return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
}

func ExecCommand(cmd string, args ...string) (string, error) {
	c := exec.Command(cmd, args...)
	data, err := c.CombinedOutput()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}
