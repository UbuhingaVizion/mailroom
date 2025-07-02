package schedules

import (
	"testing"

	"github.com/nyaruka/gocommon/dbutil/assertdb"
	"github.com/nyaruka/goflow/envs"
	"github.com/nyaruka/mailroom/core/models"
	"github.com/nyaruka/mailroom/core/queue"
	"github.com/nyaruka/mailroom/testsuite"
	"github.com/nyaruka/mailroom/testsuite/testdata"
	"github.com/stretchr/testify/assert"
)

func TestCheckSchedules(t *testing.T) {
	ctx, rt := testsuite.Runtime()
	rc := rt.RP.Get()
	defer rc.Close()

	defer testsuite.Reset(testsuite.ResetData | testsuite.ResetRedis)

	// add a schedule and tie a broadcast to it
	var s1 models.ScheduleID
	err := rt.DB.Get(
		&s1,
		`INSERT INTO schedules_schedule(is_active, repeat_period, created_on, modified_on, next_fire, created_by_id, modified_by_id, org_id)
			VALUES(TRUE, 'O', NOW(), NOW(), NOW()- INTERVAL '1 DAY', 1, 1, $1) RETURNING id`,
		testdata.Org1.ID,
	)
	assert.NoError(t, err)

	b1 := testdata.InsertBroadcast(rt, testdata.Org1, "eng", map[envs.Language]string{"eng": "Test message", "fra": "Un Message"}, s1,
		[]*testdata.Contact{testdata.Cathy, testdata.George}, []*testdata.Group{testdata.DoctorsGroup},
	)

	// add another and tie a trigger to it
	var s2 models.ScheduleID
	err = rt.DB.Get(
		&s2,
		`INSERT INTO schedules_schedule(is_active, repeat_period, created_on, modified_on, next_fire, created_by_id, modified_by_id, org_id)
			VALUES(TRUE, 'O', NOW(), NOW(), NOW()- INTERVAL '2 DAY', 1, 1, $1) RETURNING id`,
		testdata.Org1.ID,
	)
	assert.NoError(t, err)
	var t1 models.TriggerID
	err = rt.DB.Get(
		&t1,
		`INSERT INTO triggers_trigger(is_active, created_on, modified_on, is_archived, trigger_type, created_by_id, modified_by_id, org_id, flow_id, schedule_id)
			VALUES(TRUE, NOW(), NOW(), FALSE, 'S', 1, 1, $1, $2, $3) RETURNING id`,
		testdata.Org1.ID, testdata.Favorites.ID, s2,
	)
	assert.NoError(t, err)

	// add a few contacts to the trigger
	rt.DB.MustExec(`INSERT INTO triggers_trigger_contacts(trigger_id, contact_id) VALUES($1, $2),($1, $3)`, t1, testdata.Cathy.ID, testdata.George.ID)

	// and a group
	rt.DB.MustExec(`INSERT INTO triggers_trigger_groups(trigger_id, contactgroup_id) VALUES($1, $2)`, t1, testdata.DoctorsGroup.ID)

	var s3 models.ScheduleID
	err = rt.DB.Get(
		&s3,
		`INSERT INTO schedules_schedule(is_active, repeat_period, created_on, modified_on, next_fire, created_by_id, modified_by_id, org_id)
			VALUES(TRUE, 'O', NOW(), NOW(), NOW()- INTERVAL '3 DAY', 1, 1, $1) RETURNING id`,
		testdata.Org1.ID,
	)
	assert.NoError(t, err)

	// run our task
	err = checkSchedules(ctx, rt)
	assert.NoError(t, err)

	// should have one flow start added to our DB ready to go
	assertdb.Query(t, rt.DB, `SELECT count(*) FROM flows_flowstart WHERE flow_id = $1 AND start_type = 'T' AND status = 'P'`, testdata.Favorites.ID).Returns(1)

	// with the right count of groups and contacts
	assertdb.Query(t, rt.DB, `SELECT count(*) from flows_flowstart_contacts WHERE flowstart_id = 1`).Returns(2)
	assertdb.Query(t, rt.DB, `SELECT count(*) from flows_flowstart_groups WHERE flowstart_id = 1`).Returns(1)

	// and one broadcast as well
	assertdb.Query(t, rt.DB, `SELECT count(*) FROM msgs_broadcast WHERE org_id = $1 
		AND parent_id = $2 
		AND translations -> 'eng' ->> 'text' = 'Test message'
		AND translations -> 'fra' ->> 'text' = 'Un Message'
		AND status = 'Q' 
		AND base_language = 'eng'`, testdata.Org1.ID, b1).Returns(1)

	// with the right count of contacts and groups
	assertdb.Query(t, rt.DB, `SELECT count(*) from msgs_broadcast_contacts WHERE broadcast_id = 2`).Returns(2)
	assertdb.Query(t, rt.DB, `SELECT count(*) from msgs_broadcast_groups WHERE broadcast_id = 2`).Returns(1)

	// we shouldn't have any pending schedules since there were all one time fires, but all should have last fire
	assertdb.Query(t, rt.DB, `SELECT count(*) FROM schedules_schedule WHERE next_fire IS NULL and last_fire < NOW();`).Returns(3)

	// check the tasks created
	task, err := queue.PopNextTask(rc, queue.BatchQueue)

	// first should be the flow start
	assert.NoError(t, err)
	assert.NotNil(t, task)
	assert.Equal(t, "start_flow", task.Type)

	// then the broadacast
	task, err = queue.PopNextTask(rc, queue.BatchQueue)
	assert.NoError(t, err)
	assert.NotNil(t, task)
	assert.Equal(t, "send_broadcast", task.Type)

	// nothing more
	task, err = queue.PopNextTask(rc, queue.BatchQueue)
	assert.NoError(t, err)
	assert.Nil(t, task)
}
