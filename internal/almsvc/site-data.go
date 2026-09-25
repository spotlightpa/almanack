package almsvc

import "fmt"

func MessageForLoc(loc string) string {
	if msg := map[string]string{
		"data/berks-sidebar.json":        "Setting Berks County sidebar configuration",
		"data/statecollege-sidebar.json": "Setting State College sidebar configuration",
		"config/_default/params.json":    "Setting site parameters",
		"data/berks.json":                "Setting Berks County frontpage configuration",
		"data/editorsPicks.json":         "Setting homepage configuration",
		"data/sidebar.json":              "Setting sidebar configuration", // Obsolete
		"data/stateCollege.json":         "Setting State College frontpage configuration",
	}[loc]; msg != "" {
		return msg
	}
	return fmt.Sprintf("Updating %s", loc)
}
