package campaigns

import (
	"testing"
	"time"

	_ "github.com/nyaruka/mailroom/core/handlers"
	"github.com/nyaruka/mailroom/core/models"
	"github.com/nyaruka/mailroom/core/queue"
	"github.com/nyaruka/mailroom/core/tasks"
	"github.com/nyaruka/mailroom/testsuite"
	"github.com/nyaruka/mailroom/testsuite/testdata"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCampaigns(t *testing.T) {
	testsuite.Reset()
	ctx := testsuite.CTX()

	rt := testsuite.RT()
	rc := testsuite.RC()
	defer rc.Close()

	// let's create a campaign event fire for one of our contacts (for now this is totally hacked, they aren't in the group and
	// their relative to date isn't relative, but this still tests execution)
	rt.DB.MustExec(`INSERT INTO campaigns_eventfire(scheduled, contact_id, event_id) VALUES (NOW(), $1, $3), (NOW(), $2, $3);`, testdata.Cathy.ID, testdata.George.ID, testdata.RemindersEvent1.ID)
	time.Sleep(10 * time.Millisecond)

	// schedule our campaign to be started
	err := fireCampaignEvents(ctx, rt.DB, rt.RP, campaignsLock, "lock")
	assert.NoError(t, err)

	// then actually work on the event
	task, err := queue.PopNextTask(rc, queue.BatchQueue)
	assert.NoError(t, err)
	assert.NotNil(t, task)

	typedTask, err := tasks.ReadTask(task.Type, task.Task)
	require.NoError(t, err)

	// work on that task
	err = typedTask.Perform(ctx, rt, models.OrgID(task.OrgID))
	assert.NoError(t, err)

	// should now have a flow run for that contact and flow
	testsuite.AssertQueryCount(t, rt.DB, `SELECT COUNT(*) from flows_flowrun WHERE contact_id = $1 AND flow_id = $2;`, []interface{}{testdata.Cathy.ID, testdata.Favorites.ID}, 1)
	testsuite.AssertQueryCount(t, rt.DB, `SELECT COUNT(*) from flows_flowrun WHERE contact_id = $1 AND flow_id = $2;`, []interface{}{testdata.George.ID, testdata.Favorites.ID}, 1)
}

func TestIVRCampaigns(t *testing.T) {
	testsuite.Reset()
	ctx := testsuite.CTX()
	rt := testsuite.RT()
	db := rt.DB
	rc := testsuite.RC()
	defer rc.Close()

	// let's create a campaign event fire for one of our contacts (for now this is totally hacked, they aren't in the group and
	// their relative to date isn't relative, but this still tests execution)
	rt.DB.MustExec(`UPDATE campaigns_campaignevent SET flow_id = $1 WHERE id = $2`, testdata.IVRFlow.ID, testdata.RemindersEvent1.ID)
	rt.DB.MustExec(`INSERT INTO campaigns_eventfire(scheduled, contact_id, event_id) VALUES (NOW(), $1, $3), (NOW(), $2, $3);`, testdata.Cathy.ID, testdata.George.ID, testdata.RemindersEvent1.ID)
	time.Sleep(10 * time.Millisecond)

	// schedule our campaign to be started
	err := fireCampaignEvents(ctx, rt.DB, rt.RP, campaignsLock, "lock")
	assert.NoError(t, err)

	// then actually work on the event
	task, err := queue.PopNextTask(rc, queue.BatchQueue)
	assert.NoError(t, err)
	assert.NotNil(t, task)

	typedTask, err := tasks.ReadTask(task.Type, task.Task)
	require.NoError(t, err)

	// work on that task
	err = typedTask.Perform(ctx, rt, models.OrgID(task.OrgID))
	assert.NoError(t, err)

	// should now have a flow start created
	testsuite.AssertQueryCount(t, db, `SELECT COUNT(*) from flows_flowstart WHERE flow_id = $1 AND start_type = 'T' AND status = 'P';`, []interface{}{testdata.IVRFlow.ID}, 1)
	testsuite.AssertQueryCount(t, db, `SELECT COUNT(*) from flows_flowstart_contacts WHERE contact_id = $1 AND flowstart_id = 1;`, []interface{}{testdata.Cathy.ID}, 1)
	testsuite.AssertQueryCount(t, db, `SELECT COUNT(*) from flows_flowstart_contacts WHERE contact_id = $1 AND flowstart_id = 1;`, []interface{}{testdata.George.ID}, 1)

	// event should be marked as fired
	testsuite.AssertQueryCount(t, db, `SELECT COUNT(*) from campaigns_eventfire WHERE event_id = $1 AND fired IS NOT NULL;`, []interface{}{testdata.RemindersEvent1.ID}, 2)

	// pop our next task, should be the start
	task, err = queue.PopNextTask(rc, queue.BatchQueue)
	assert.NoError(t, err)
	assert.NotNil(t, task)

	assert.Equal(t, task.Type, queue.StartIVRFlowBatch)
}
