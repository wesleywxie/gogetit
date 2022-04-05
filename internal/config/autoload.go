package config

import (
	"flag"
	"fmt"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func init() {
	if isInTests() {
		// 测试环境
		RunMode = TestMode
		return
	}

	workDirFlag := flag.String("d", "./", "work directory of gogetid")
	configFile := flag.String("c", "", "config file of gogetid")
	printVersionFlag := flag.Bool("v", false, "prints gogetid version")

	testing.Init()
	flag.Parse()

	if *printVersionFlag {
		// print version
		fmt.Printf(AppVersionInfo())
		os.Exit(0)
	}

	workDir := filepath.Clean(*workDirFlag)

	if *configFile != "" {
		viper.SetConfigFile(*configFile)
	} else {
		viper.SetConfigFile(filepath.Join(workDir, "config.yml"))
	}

	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {             // Handle errors reading the config file
		panic(fmt.Errorf("fatal error on reading config file: %s", err))
	}

	Socks5 = viper.GetString("socks5")

	if viper.IsSet("log.db_log") {
		DBLogMode = viper.GetBool("log.db_log")
	}

	if viper.IsSet("dl.user_agent") {
		UserAgent = viper.GetString("dl.user_agent")
	}
}

func isInTests() bool {
	for _, arg := range os.Args {
		if strings.HasPrefix(arg, "-test") {
			return true
		}
	}
	return false
}
