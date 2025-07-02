package ticket

import (
	"context"
	"net/http"

	"github.com/nyaruka/mailroom/core/models"
	"github.com/nyaruka/mailroom/runtime"
	"github.com/nyaruka/mailroom/web"
	"github.com/pkg/errors"
)

func init() {
	web.RegisterRoute(http.MethodPost, "/mr/ticket/change_topic", web.RequireAuthToken(web.JSONPayload(handleChangeTopic)))
}

type changeTopicRequest struct {
	bulkTicketRequest

	TopicID models.TopicID `json:"topic_id" validate:"required"`
}

// Changes the topic of the tickets with the given ids
//
//	{
//	  "org_id": 123,
//	  "user_id": 234,
//	  "ticket_ids": [1234, 2345],
//	  "topic_id": 345
//	}
func handleChangeTopic(ctx context.Context, rt *runtime.Runtime, r *changeTopicRequest) (any, int, error) {
	oa, err := models.GetOrgAssets(ctx, rt, r.OrgID)
	if err != nil {
		return nil, 0, errors.Wrap(err, "unable to load org assets")
	}

	tickets, err := models.LoadTickets(ctx, rt.DB, r.TicketIDs)
	if err != nil {
		return nil, 0, errors.Wrapf(err, "error loading tickets for org: %d", r.OrgID)
	}

	evts, err := models.TicketsChangeTopic(ctx, rt.DB, oa, r.UserID, tickets, r.TopicID)
	if err != nil {
		return nil, 0, errors.Wrap(err, "error changing topic of tickets")
	}

	return newBulkResponse(evts), http.StatusOK, nil
}
