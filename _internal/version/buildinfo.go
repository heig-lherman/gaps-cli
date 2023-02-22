package version

import "fmt"

var (
	version = "dev"
	commit  = "none"
	arch    = "none"
	date    = "none"
)

type BuildInfo struct {
	Version string `json:"version,omitempty"`
	Commit  string `json:"commit,omitempty"`
	Arch    string `json:"arch,omitempty"`
	Date    string `json:"date,omitempty"`
}

func Get() BuildInfo {
	return BuildInfo{
		Version: version,
		Commit:  commit,
		Arch:    arch,
		Date:    date,
	}
}

func GetStr() string {
	return fmt.Sprintf("%#v", Get())
}
