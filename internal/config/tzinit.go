package config

import (
	"os"
)

func init() {
	_ = os.Setenv("TZ", "UTC")
}
