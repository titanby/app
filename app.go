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

	flVersion    = flag.Bool("v", false, "-v = show app version")
	flConfigFile = flag.String("c", "", "-c = set config file (default: <AppName>.json)")
	flDebugMode  = flag.Bool("d", false, "-d = set debug mode to true")
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func setVersion(appVersion, buildVersion string) {
	appInfo.AppVersion = appVersion
	appInfo.BuildVersion = buildVersion
}

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	flag.Parse()
	if *flDebugMode {
		SetDebugMode()
	} else {
		SetInfoMode()
	}
}

func checkFlagVersion() {
	if *flVersion {
		println(Name(), Version())
		os.Exit(0)
	}
}

func InitBasic(appVersion, buildVersion string) {
	setVersion(appVersion, buildVersion)
	checkFlagVersion()
}

func Init(appVersion, buildVersion string) {
	setVersion(appVersion, buildVersion)
	checkFlagVersion()

	LogWith(
		"Name", Name(),
		"Version", Version(),
		"DebugMode", DebugMode(),
	).Info("Start application")

	configFile := ""
	configFile = *flConfigFile
	if configFile == "" {
		configFile = Name() + ".json"
	}

	if err := readConfigFile(configFile); err != nil {
		LogWith(
			"Error", err,
		).Fatal("Read config file")
	}
}

// DebugMode - returns true if flag 'd' exists in cmd params
func DebugMode() bool {
	return appInfo.DebugMode
}

func SetDebugMode() {
	appInfo.DebugMode = true
	setLogLevel(levelDebug)
}

func SetInfoMode() {
	appInfo.DebugMode = false
	setLogLevel(levelInfo)
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
