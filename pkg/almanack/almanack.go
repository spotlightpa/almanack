package almanack

import (
	_ "embed"
	"strings"
)

var (
	//go:embed build-version.txt
	buildVersion string
	BuildVersion = strings.TrimSpace(buildVersion)
	//go:embed deploy-url.txt
	deployURL string
	DeployURL = strings.TrimSpace(deployURL)
)
