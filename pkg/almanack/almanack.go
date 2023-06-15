package almanack

import (
	_ "embed"
	"strings"

	"github.com/carlmjohnson/versioninfo"
)

var (
	BuildVersion = versioninfo.Revision
	//go:embed deploy-url.txt
	deployURL string
	DeployURL = strings.TrimSpace(deployURL)
)
