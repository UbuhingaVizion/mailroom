package models_test

import (
	"testing"

	"github.com/nyaruka/gocommon/urns"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TikTok identifies a contact by an app-scoped open_id and addresses a reply to a conversation. The
// conversation is kept in the URN query so a later message updates the route without changing the contact's
// identity, the same way the channel parameter is carried on other external URNs.
func TestTikTokURNChatIdentity(t *testing.T) {
	first, err := urns.NewURNFromParts("ext", "tt_abc123", "conversation=Y29udisxPT0", "")
	require.NoError(t, err)
	assert.Equal(t, "tt_abc123", first.Path())
	assert.Equal(t, "ext:tt_abc123", string(first.Identity()))

	second, err := urns.NewURNFromParts("ext", "tt_abc123", "conversation=bmV3", "")
	require.NoError(t, err)
	assert.Equal(t, first.Identity(), second.Identity())

	query, err := second.Query()
	require.NoError(t, err)
	assert.Equal(t, "bmV3", query.Get("conversation"))
}
