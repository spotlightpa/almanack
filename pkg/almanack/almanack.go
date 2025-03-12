package almanack

import (
	_ "embed"
	"strings"

	"github.com/earthboundkid/versioninfo/v2"
)

var (
	BuildVersion = versioninfo.Revision
	//go:embed deploy-url.txt
	deployURL string
	DeployURL = strings.TrimSpace(deployURL)
)
