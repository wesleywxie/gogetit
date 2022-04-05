package config

import (
	"fmt"
	"github.com/spf13/viper"
)

type RunType string

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"

	ProjectName string = "gogetit"
	Socks5      string

	// UserAgent as the user agent for downloading task
	UserAgent string

	RunMode = ReleaseMode

	// DBLogMode 是否打印数据库日志
	DBLogMode = false

	DBHost string
	DBPort int
	DBUser string
	DBPass string
	DBName string
)

const (
	TestMode    RunType = "Test"
	ReleaseMode RunType = "Release"
)

func AppVersionInfo() (s string) {
	s = fmt.Sprintf("version %v, commit %v, built at %v", version, commit, date)
	return
}

// GetString get string config value by key
func GetString(key string) string {
	var value string
	if viper.IsSet(key) {
		value = viper.GetString(key)
	}

	return value
}

// GetInt get int config value by key
func GetInt(key string) int {
	var value int
	if viper.IsSet(key) {
		value = viper.GetInt(key)
	}

	return value
}
