package rx

import (
	"testing"
	"text/template"

	"github.com/stretchr/testify/assert"
)

func TestRxSyncer(t *testing.T) {
	temp := template.Must(template.New("sync-template").Parse("https://localhost:5005/"))
	syncer := NewRXSyncer(temp)
	syncInfo, err := syncer.GetUsersyncInfo("", "")
	assert.NoError(t, err)
	assert.Equal(t, "https://localhost:5005/", syncInfo.URL)
	assert.Equal(t, "redirect", syncInfo.Type)
	assert.EqualValues(t, 60, syncer.GDPRVendorID())
	assert.Equal(t, false, syncInfo.SupportCORS)
}
