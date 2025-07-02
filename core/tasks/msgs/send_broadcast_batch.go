package msgs

import (
	"context"
	"time"

	"github.com/nyaruka/mailroom/core/models"
	"github.com/nyaruka/mailroom/core/msgio"
	"github.com/nyaruka/mailroom/core/tasks"
	"github.com/nyaruka/mailroom/runtime"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const TypeSendBroadcastBatch = "send_broadcast_batch"

func init() {
	tasks.RegisterType(TypeSendBroadcastBatch, func() tasks.Task { return &SendBroadcastBatchTask{} })
}

// SendBroadcastTask is the task send broadcast batches
type SendBroadcastBatchTask struct {
	*models.BroadcastBatch
}

func (t *SendBroadcastBatchTask) Type() string {
	return TypeSendBroadcastBatch
}

// Timeout is the maximum amount of time the task can run for
func (t *SendBroadcastBatchTask) Timeout() time.Duration {
	return time.Minute * 60
}

func (t *SendBroadcastBatchTask) Perform(ctx context.Context, rt *runtime.Runtime, orgID models.OrgID) error {
	// always set our broadcast as sent if it is our last
	defer func() {
		if t.BroadcastBatch.IsLast && t.BroadcastBatch.BroadcastID != models.NilBroadcastID {
			err := models.MarkBroadcastSent(ctx, rt.DB, t.BroadcastBatch.BroadcastID)
			if err != nil {
				logrus.WithError(err).Error("error marking broadcast as sent")
			}
		}
	}()

	oa, err := models.GetOrgAssets(ctx, rt, t.BroadcastBatch.OrgID)
	if err != nil {
		return errors.Wrapf(err, "error getting org assets")
	}

	// create this batch of messages
	msgs, err := t.BroadcastBatch.CreateMessages(ctx, rt, oa)
	if err != nil {
		return errors.Wrapf(err, "error creating broadcast messages")
	}

	msgio.SendMessages(ctx, rt, rt.DB, nil, msgs)
	return nil
}
