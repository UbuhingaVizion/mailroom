package ticket

import (
	"testing"
	"time"

	_ "github.com/nyaruka/mailroom/services/tickets/intern"
	_ "github.com/nyaruka/mailroom/services/tickets/mailgun"
	_ "github.com/nyaruka/mailroom/services/tickets/zendesk"
	"github.com/nyaruka/mailroom/testsuite"
	"github.com/nyaruka/mailroom/testsuite/testdata"
)

func TestTicketAssign(t *testing.T) {
	ctx, rt := testsuite.Runtime()

	defer testsuite.Reset(testsuite.ResetData)

	testdata.InsertOpenTicket(rt, testdata.Org1, testdata.Cathy, testdata.Internal, testdata.DefaultTopic, "Have you seen my cookies?", "17", time.Now(), testdata.Admin)
	testdata.InsertOpenTicket(rt, testdata.Org1, testdata.Cathy, testdata.Internal, testdata.DefaultTopic, "Have you seen my cookies?", "21", time.Now(), testdata.Agent)
	testdata.InsertClosedTicket(rt, testdata.Org1, testdata.Cathy, testdata.Internal, testdata.DefaultTopic, "Have you seen my cookies?", "34", nil)
	testdata.InsertClosedTicket(rt, testdata.Org1, testdata.Bob, testdata.Internal, testdata.DefaultTopic, "", "", nil)

	testsuite.RunWebTests(t, ctx, rt, "testdata/assign.json", nil)
}

func TestTicketAddNote(t *testing.T) {
	ctx, rt := testsuite.Runtime()

	defer testsuite.Reset(testsuite.ResetData)

	testdata.InsertOpenTicket(rt, testdata.Org1, testdata.Cathy, testdata.Internal, testdata.DefaultTopic, "Have you seen my cookies?", "17", time.Now(), testdata.Admin)
	testdata.InsertOpenTicket(rt, testdata.Org1, testdata.Cathy, testdata.Internal, testdata.DefaultTopic, "Have you seen my cookies?", "21", time.Now(), testdata.Agent)
	testdata.InsertClosedTicket(rt, testdata.Org1, testdata.Cathy, testdata.Internal, testdata.DefaultTopic, "Have you seen my cookies?", "34", nil)

	testsuite.RunWebTests(t, ctx, rt, "testdata/add_note.json", nil)
}

func TestTicketChangeTopic(t *testing.T) {
	ctx, rt := testsuite.Runtime()

	defer testsuite.Reset(testsuite.ResetData)

	testdata.InsertOpenTicket(rt, testdata.Org1, testdata.Cathy, testdata.Internal, testdata.DefaultTopic, "Have you seen my cookies?", "17", time.Now(), testdata.Admin)
	testdata.InsertOpenTicket(rt, testdata.Org1, testdata.Cathy, testdata.Internal, testdata.SupportTopic, "Have you seen my cookies?", "21", time.Now(), testdata.Agent)
	testdata.InsertClosedTicket(rt, testdata.Org1, testdata.Cathy, testdata.Internal, testdata.SalesTopic, "Have you seen my cookies?", "34", nil)

	testsuite.RunWebTests(t, ctx, rt, "testdata/change_topic.json", nil)
}

func TestTicketClose(t *testing.T) {
	ctx, rt := testsuite.Runtime()

	defer testsuite.Reset(testsuite.ResetData)

	// create 2 open tickets and 1 closed one for Cathy across two different ticketers
	testdata.InsertOpenTicket(rt, testdata.Org1, testdata.Cathy, testdata.Mailgun, testdata.DefaultTopic, "Have you seen my cookies?", "17", time.Now(), testdata.Admin)
	testdata.InsertOpenTicket(rt, testdata.Org1, testdata.Cathy, testdata.Zendesk, testdata.DefaultTopic, "Have you seen my cookies?", "21", time.Now(), nil)
	testdata.InsertClosedTicket(rt, testdata.Org1, testdata.Cathy, testdata.Zendesk, testdata.DefaultTopic, "Have you seen my cookies?", "34", testdata.Editor)
	testdata.InsertOpenTicket(rt, testdata.Org1, testdata.Cathy, testdata.Zendesk, testdata.DefaultTopic, "Have you seen my cookies?", "21", time.Now(), nil)

	testsuite.RunWebTests(t, ctx, rt, "testdata/close.json", nil)
}

func TestTicketReopen(t *testing.T) {
	ctx, rt := testsuite.Runtime()

	defer testsuite.Reset(testsuite.ResetData | testsuite.ResetRedis)

	// we should be able to reopen ticket #1 because Cathy has no other tickets open
	testdata.InsertClosedTicket(rt, testdata.Org1, testdata.Cathy, testdata.Mailgun, testdata.DefaultTopic, "Have you seen my cookies?", "17", testdata.Admin)

	// but then we won't be able to open ticket #2
	testdata.InsertClosedTicket(rt, testdata.Org1, testdata.Cathy, testdata.Zendesk, testdata.DefaultTopic, "Have you seen my cookies?", "21", nil)

	testdata.InsertClosedTicket(rt, testdata.Org1, testdata.Bob, testdata.Zendesk, testdata.DefaultTopic, "Have you seen my cookies?", "27", testdata.Editor)
	testdata.InsertClosedTicket(rt, testdata.Org1, testdata.Alexandria, testdata.Internal, testdata.DefaultTopic, "Have you seen my cookies?", "", testdata.Editor)

	testsuite.RunWebTests(t, ctx, rt, "testdata/reopen.json", nil)
}
