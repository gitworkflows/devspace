package jetbrains

import (
	"dev.khulnasoft.com/pkg/config"
	"dev.khulnasoft.com/pkg/ide"
	"dev.khulnasoft.com/log"
)

const (
	PycharmProductCode           = "PY"
	PycharmDownloadAmd64Template = "https://download.jetbrains.com/python/pycharm-professional-%s.tar.gz"
	PycharmDownloadArm64Template = "https://download.jetbrains.com/python/pycharm-professional-%s-aarch64.tar.gz"
)

var PyCharmOptions = ide.Options{
	VersionOption: {
		Name:        VersionOption,
		Description: "The version for the binary",
		Default:     "latest",
	},
	DownloadArm64Option: {
		Name:        DownloadArm64Option,
		Description: "The download url for the arm64 server binary",
	},
	DownloadAmd64Option: {
		Name:        DownloadAmd64Option,
		Description: "The download url for the amd64 server binary",
	},
}

func NewPyCharmServer(userName string, values map[string]config.OptionValue, log log.Logger) *GenericJetBrainsServer {
	amd64Download, arm64Download := getDownloadURLs(PyCharmOptions, values, PycharmProductCode, PycharmDownloadAmd64Template, PycharmDownloadArm64Template)
	return newGenericServer(userName, &GenericOptions{
		ID:            "pycharm",
		DisplayName:   "PyCharm",
		DownloadAmd64: amd64Download,
		DownloadArm64: arm64Download,
	}, log)
}
