package testdata

import (
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/nyaruka/goflow/assets"
	"github.com/nyaruka/goflow/flows"
	"github.com/nyaruka/mailroom/core/models"

	"github.com/stretchr/testify/require"
)

type Ticketer struct {
	ID   models.TicketerID
	UUID assets.TicketerUUID
}

// InsertOpenTicket inserts an open ticket
func InsertOpenTicket(t *testing.T, db *sqlx.DB, org *Org, contact *Contact, ticketer *Ticketer, uuid flows.TicketUUID, subject, body, externalID string) models.TicketID {
	var id models.TicketID
	err := db.Get(&id,
		`INSERT INTO tickets_ticket(uuid, org_id, contact_id, ticketer_id, status, subject, body, external_id, opened_on, modified_on)
		VALUES($1, $2, $3, $4, 'O', $5, $6, $7, NOW(), NOW()) RETURNING id`, uuid, org.ID, contact.ID, ticketer.ID, subject, body, externalID,
	)
	require.NoError(t, err)
	return id
}

// InsertClosedTicket inserts a closed ticket
func InsertClosedTicket(t *testing.T, db *sqlx.DB, org *Org, contact *Contact, ticketer *Ticketer, uuid flows.TicketUUID, subject, body, externalID string) models.TicketID {
	var id models.TicketID
	err := db.Get(&id,
		`INSERT INTO tickets_ticket(uuid, org_id, contact_id, ticketer_id, status, subject, body, external_id, opened_on, modified_on, closed_on)
		VALUES($1, $2, $3, $4, 'C', $5, $6, $7, NOW(), NOW(), NOW()) RETURNING id`, uuid, org.ID, contact.ID, ticketer.ID, subject, body, externalID,
	)
	require.NoError(t, err)
	return id
}
