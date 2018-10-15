package resumes

import (
	"encoding/json"

	"github.com/nyaruka/goflow/flows"
	"github.com/nyaruka/goflow/flows/events"
	"github.com/nyaruka/goflow/utils"
)

func init() {
	RegisterType(TypeWaitTimeout, ReadWaitTimeoutResume)
}

// TypeWaitTimeout is the type for resuming a session when a wait has timed out
const TypeWaitTimeout string = "wait_timeout"

// WaitTimeoutResume is used when a session is resumed because a wait has timed out
//
//   {
//     "type": "wait_timeout",
//     "contact": {
//       "uuid": "9f7ede93-4b16-4692-80ad-b7dc54a1cd81",
//       "name": "Bob",
//       "created_on": "2018-01-01T12:00:00.000000Z",
//       "language": "fra",
//       "fields": {"gender": {"text": "Male"}},
//       "groups": []
//     },
//     "resumed_on": "2000-01-01T00:00:00.000000000-00:00"
//   }
//
// @resume wait_timeout
type WaitTimeoutResume struct {
	baseResume
}

// NewWaitTimeoutResume creates a new timeout resume with the passed in values
func NewWaitTimeoutResume(env utils.Environment, contact *flows.Contact) *WaitTimeoutResume {
	return &WaitTimeoutResume{
		baseResume: newBaseResume(TypeWaitTimeout, env, contact),
	}
}

// Apply applies our state changes and saves any events to the run
func (r *WaitTimeoutResume) Apply(run flows.FlowRun, step flows.Step) error {
	// clear the last input
	run.Session().SetInput(nil)
	run.LogEvent(step, events.NewWaitTimedOutEvent())

	return r.baseResume.Apply(run, step)
}

var _ flows.Resume = (*WaitTimeoutResume)(nil)

//------------------------------------------------------------------------------------------
// JSON Encoding / Decoding
//------------------------------------------------------------------------------------------

// ReadWaitTimeoutResume reads a timeout resume
func ReadWaitTimeoutResume(session flows.Session, data json.RawMessage) (flows.Resume, error) {
	e := &baseResumeEnvelope{}
	if err := utils.UnmarshalAndValidate(data, e); err != nil {
		return nil, err
	}

	r := &WaitTimeoutResume{}

	if err := r.unmarshal(session, e); err != nil {
		return nil, err
	}

	return r, nil
}

// MarshalJSON marshals this resume into JSON
func (r *WaitTimeoutResume) MarshalJSON() ([]byte, error) {
	e := &baseResumeEnvelope{}

	if err := r.marshal(e); err != nil {
		return nil, err
	}

	return json.Marshal(e)
}
