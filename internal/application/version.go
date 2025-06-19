package application

import (
	"fmt"
	"runtime/debug"
	"time"
)

func getKey(info *debug.BuildInfo, key string) string {
	if info == nil {
		return ""
	}
	for _, iter := range info.Settings {
		if iter.Key == key {
			return iter.Value
		}
	}
	return ""
}

const shaLen = 7

func getVersion() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return ""
	}

	fmt.Println("Go version:", info.GoVersion)
	fmt.Println("sum", info.Main.Sum)
	fmt.Println("path", info.Main.Path)
	fmt.Println("version", info.Main.Version)

	commit := "unknown"
	commitDate := time.Now()

	for _, kv := range info.Settings {
		switch kv.Key {
		case "vcs.revision":
			commit = kv.Value
		case "vcs.time":
			commitDate, _ = time.Parse(time.RFC3339, kv.Value)
		}
	}

	if len(commit) > 7 {
		commit = commit[:7]
	}

	version := fmt.Sprintf("%s-%s", commitDate.Format("2006-01-02-150405"), commit)

	return version
}
