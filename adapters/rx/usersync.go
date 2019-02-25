package rx

import (
	"text/template"

	"github.com/prebid/prebid-server/adapters"
	"github.com/prebid/prebid-server/usersync"
)

func NewRXSyncer(temp *template.Template) usersync.Usersyncer {
	return adapters.NewSyncer("rx", 60, temp, adapters.SyncTypeRedirect)
}
