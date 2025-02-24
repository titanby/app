package app

import (
	"flag"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

var (
	appInfo = struct {
		DebugMode    bool
		AppVersion   string
		BuildVersion string
	}{}
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func Init(appVersion, buildVersion string) error {
	appInfo.AppVersion = appVersion
	appInfo.BuildVersion = buildVersion

	flVersion := flag.Bool("v", false, "-v = show version")
	flConfigFile := flag.String("c", "", "-c = set config file (default: <AppName>.json)")
	flDebugMode := flag.Bool("d", false, "-d = set debug mode to true")

	flag.Parse()

	if *flVersion {
		println(Name(), Version())
		os.Exit(0)
	}

	appInfo.DebugMode = *flDebugMode

	configFile := ""
	configFile = *flConfigFile
	if configFile == "" {
		configFile = Name() + ".json"
	}

	return readConfigFile(configFile)
}

// DebugMode - returns true if flag 'd' exists in cmd params
func DebugMode() bool {
	return appInfo.DebugMode
}

// Name - returns application name as filename without extension
func Name() string {
	return strings.Replace(filepath.Base(os.Args[0]), filepath.Ext(os.Args[0]), "", -1)
}

// Version - returns application version
func Version() string {
	res := appInfo.AppVersion
	if appInfo.AppVersion != "" && appInfo.BuildVersion != "" {
		res += "_"
	}
	res += appInfo.BuildVersion
	return res
}
