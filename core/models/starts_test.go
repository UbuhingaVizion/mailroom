package models_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/nyaruka/goflow/flows"
	"github.com/nyaruka/mailroom/core/models"
	"github.com/nyaruka/mailroom/testsuite"
	"github.com/nyaruka/mailroom/testsuite/testdata"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStarts(t *testing.T) {
	ctx := testsuite.CTX()
	db := testsuite.DB()

	startID := testdata.InsertFlowStart(t, db, testdata.Org1, testdata.SingleMessage, []*testdata.Contact{testdata.Cathy, testdata.Bob})

	startJSON := []byte(fmt.Sprintf(`{
		"start_id": %d,
		"start_type": "M",
		"org_id": %d,
		"created_by": "rowan@nyaruka.com",
		"flow_id": %d,
		"flow_type": "M",
		"contact_ids": [%d, %d],
		"group_ids": [6789],
		"urns": ["tel:+12025550199"],
		"query": null,
		"restart_participants": true,
		"include_active": true,
		"parent_summary": {"uuid": "b65b1a22-db6d-4f5a-9b3d-7302368a82e6"},
		"session_history": {"parent_uuid": "532a3899-492f-4ffe-aed7-e75ad524efab", "ancestors": 3, "ancestors_since_input": 1},
		"extra": {"foo": "bar"}
	}`, startID, testdata.Org1.ID, testdata.SingleMessage.ID, testdata.Cathy.ID, testdata.Bob.ID))

	start := &models.FlowStart{}
	err := json.Unmarshal(startJSON, start)

	require.NoError(t, err)
	assert.Equal(t, startID, start.ID())
	assert.Equal(t, testdata.Org1.ID, start.OrgID())
	assert.Equal(t, testdata.SingleMessage.ID, start.FlowID())
	assert.Equal(t, models.FlowTypeMessaging, start.FlowType())
	assert.Equal(t, "", start.Query())
	assert.Equal(t, models.DoRestartParticipants, start.RestartParticipants())
	assert.Equal(t, models.DoIncludeActive, start.IncludeActive())

	assert.Equal(t, json.RawMessage(`{"uuid": "b65b1a22-db6d-4f5a-9b3d-7302368a82e6"}`), start.ParentSummary())
	assert.Equal(t, json.RawMessage(`{"parent_uuid": "532a3899-492f-4ffe-aed7-e75ad524efab", "ancestors": 3, "ancestors_since_input": 1}`), start.SessionHistory())
	assert.Equal(t, json.RawMessage(`{"foo": "bar"}`), start.Extra())

	err = models.MarkStartStarted(ctx, db, startID, 2, []models.ContactID{testdata.George.ID})
	require.NoError(t, err)

	testsuite.AssertQueryCount(t, db, `SELECT count(*) FROM flows_flowstart WHERE id = $1 AND status = 'S' AND contact_count = 2`, []interface{}{startID}, 1)
	testsuite.AssertQueryCount(t, db, `SELECT count(*) FROM flows_flowstart_contacts WHERE flowstart_id = $1`, []interface{}{startID}, 3)

	batch := start.CreateBatch([]models.ContactID{testdata.Cathy.ID, testdata.Bob.ID}, false, 3)
	assert.Equal(t, startID, batch.StartID())
	assert.Equal(t, models.StartTypeManual, batch.StartType())
	assert.Equal(t, testdata.SingleMessage.ID, batch.FlowID())
	assert.Equal(t, []models.ContactID{testdata.Cathy.ID, testdata.Bob.ID}, batch.ContactIDs())
	assert.Equal(t, models.DoRestartParticipants, batch.RestartParticipants())
	assert.Equal(t, models.DoIncludeActive, batch.IncludeActive())
	assert.Equal(t, "rowan@nyaruka.com", batch.CreatedBy())
	assert.False(t, batch.IsLast())
	assert.Equal(t, 3, batch.TotalContacts())

	assert.Equal(t, json.RawMessage(`{"uuid": "b65b1a22-db6d-4f5a-9b3d-7302368a82e6"}`), batch.ParentSummary())
	assert.Equal(t, json.RawMessage(`{"parent_uuid": "532a3899-492f-4ffe-aed7-e75ad524efab", "ancestors": 3, "ancestors_since_input": 1}`), batch.SessionHistory())
	assert.Equal(t, json.RawMessage(`{"foo": "bar"}`), batch.Extra())

	history, err := models.ReadSessionHistory(batch.SessionHistory())
	assert.NoError(t, err)
	assert.Equal(t, flows.SessionUUID("532a3899-492f-4ffe-aed7-e75ad524efab"), history.ParentUUID)

	_, err = models.ReadSessionHistory([]byte(`{`))
	assert.EqualError(t, err, "unexpected end of JSON input")

	err = models.MarkStartComplete(ctx, db, startID)
	require.NoError(t, err)

	testsuite.AssertQueryCount(t, db, `SELECT count(*) FROM flows_flowstart WHERE id = $1 AND status = 'C'`, []interface{}{startID}, 1)
}
