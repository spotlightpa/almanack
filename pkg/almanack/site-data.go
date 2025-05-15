package almanack

import "fmt"

const (
	HomepageLoc     = "data/editorsPicks.json"
	SidebarLoc      = "data/sidebar.json"
	SiteParamsLoc   = "config/_default/params.json"
	StateCollegeLoc = "data/stateCollege.json"
	BerksLoc        = "data/berks.json"
)

var messageForLoc = map[string]string{
	HomepageLoc:     "Setting homepage configuration",
	SidebarLoc:      "Setting sidebar configuration",
	SiteParamsLoc:   "Setting site parameters",
	StateCollegeLoc: "Setting State College frontpage configuration",
	BerksLoc:        "Setting Berks County frontpage configuration",
}

func MessageForLoc(loc string) string {
	msg := messageForLoc[loc]
	if msg == "" {
		return fmt.Sprintf("Updating %s", loc)
	}
	return msg
}
