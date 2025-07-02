package mailgun

import (
	"testing"
	"time"

	"github.com/nyaruka/mailroom/testsuite"
	"github.com/nyaruka/mailroom/testsuite/testdata"
)

func TestReceive(t *testing.T) {
	ctx, rt := testsuite.Runtime()

	defer testsuite.Reset(testsuite.ResetData | testsuite.ResetStorage)

	// create a mailgun ticket for Cathy
	ticket := testdata.InsertOpenTicket(rt, testdata.Org1, testdata.Cathy, testdata.Mailgun, testdata.DefaultTopic, "Have you seen my cookies?", "", time.Now(), nil)

	testsuite.RunWebTests(t, ctx, rt, "testdata/receive.json", map[string]string{"cathy_ticket_uuid": string(ticket.UUID)})
}
