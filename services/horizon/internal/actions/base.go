package actions

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"net/http"

	horizonContext "github.com/kinecosystem/go/services/horizon/internal/context"
	"github.com/kinecosystem/go/services/horizon/internal/render"
	hProblem "github.com/kinecosystem/go/services/horizon/internal/render/problem"
	"github.com/kinecosystem/go/services/horizon/internal/render/sse"
	"github.com/kinecosystem/go/support/errors"
	"github.com/kinecosystem/go/support/log"
	"github.com/kinecosystem/go/support/render/problem"
)

// Base is a helper struct you can use as part of a custom action via
// composition.
//
// TODO: example usage
type Base struct {
	W   http.ResponseWriter
	R   *http.Request
	Err error

	appCtx  context.Context
	isSetup bool
}

// Prepare established the common attributes that get used in nearly every
// action.  "Child" actions may override this method to extend action, but it
// is advised you also call this implementation to maintain behavior.
func (base *Base) Prepare(w http.ResponseWriter, r *http.Request, appCtx context.Context) {
	base.W = w
	base.R = r
	base.appCtx = appCtx
}

// Execute trigger content negotiation and the actual execution of one of the
// action's handlers.
func (base *Base) Execute(action interface{}) {
	ctx := base.R.Context()
	contentType := render.Negotiate(base.R)

	switch contentType {
	case render.MimeHal, render.MimeJSON:
		action, ok := action.(JSON)
		if !ok {
			goto NotAcceptable
		}

		action.JSON()

		if base.Err != nil {
			problem.Render(ctx, base.W, base.Err)
			return
		}

	case render.MimeEventStream:
<<<<<<< HEAD
		var notification chan interface{}

		action, ok := action.(SSE)
		if !ok {
=======
		switch action.(type) {
		case SSE, SingleObjectStreamer:
		default:
>>>>>>> horizon-v0.16.0
			goto NotAcceptable
		}

		// Subscribe this handler to the topic if the SSE request is related to a specific topic (tx_id, account_id, etc.).
		// This causes action.SSE to only be triggered by this topic. Unsubscribe when done.
		topic := action.GetTopic()
		if topic != "" {
			notification = sse.Subscribe(topic)
			defer sse.Unsubscribe(notification, topic)
		}

		stream := sse.NewStream(ctx, base.W)

		var oldHash [32]byte
		for {
			// Rate limit the request if it's a call to stream since it queries the DB every second. See
			// https://github.com/kinecosystem/go/issues/715 for more details.
			app := base.R.Context().Value(&horizonContext.AppContextKey)
			rateLimiter := app.(RateLimiterProvider).GetRateLimiter()
			if rateLimiter != nil {
				limited, _, err := rateLimiter.RateLimiter.RateLimit(rateLimiter.VaryBy.Key(base.R), 1)
				if err != nil {
					log.Ctx(ctx).Error(errors.Wrap(err, "RateLimiter error"))
					stream.Err(errors.New("Unexpected stream error"))
					return
				}
				if limited {
					stream.Err(errors.New("rate limit exceeded"))
					return
				}
			}

			switch ac := action.(type) {
			case SSE:
				ac.SSE(stream)

			case SingleObjectStreamer:
				newEvent := ac.LoadEvent()
				if base.Err != nil {
					break
				}
				resource, err := json.Marshal(newEvent.Data)
				if err != nil {
					log.Ctx(ctx).Error(errors.Wrap(err, "unable to marshal next action resource"))
					stream.Err(errors.New("Unexpected stream error"))
					return
				}

				nextHash := sha256.Sum256(resource)
				if bytes.Equal(nextHash[:], oldHash[:]) {
					break
				}

				oldHash = nextHash
				stream.SetLimit(10)
				stream.Send(newEvent)
			}
			// TODO: better error handling. We should probably handle the error immediately in the error case above
			// instead of breaking out from the switch statement.
			if base.Err != nil {
				// If we haven't sent an event, we should simply return the normal HTTP
				// error because it means that we haven't sent the preamble.
				if stream.SentCount() == 0 {
					problem.Render(ctx, base.W, base.Err)
					return
				}

				if errors.Cause(base.Err) == sql.ErrNoRows {
					base.Err = errors.New("Object not found")
				} else {
					log.Ctx(ctx).Error(base.Err)
					base.Err = errors.New("Unexpected stream error")
				}

				// Send errors through the stream and then close the stream.
				stream.Err(base.Err)
			}

			// Manually send the preamble in case there are no data events in SSE to trigger a stream.Send call.
			// This method is called every iteration of the loop, but is protected by a sync.Once variable so it's
			// only executed once.
			stream.Init()

			if stream.IsDone() {
				return
			}

			select {
			case <-notification:
				// No-op, continue onto the next iteration.
				continue
			case <-ctx.Done():
				// Close stream and exit.
				stream.Done()
				return
			case <-base.appCtx.Done():
				// Close stream and exit.
				stream.Done()
				return
			}
		}
	case render.MimeRaw:
		action, ok := action.(Raw)
		if !ok {
			goto NotAcceptable
		}

		action.Raw()

		if base.Err != nil {
			problem.Render(ctx, base.W, base.Err)
			return
		}
	default:
		goto NotAcceptable
	}
	return

NotAcceptable:
	problem.Render(ctx, base.W, hProblem.NotAcceptable)
	return
}

// Do executes the provided func iff there is no current error for the action.
// Provides a nicer way to invoke a set of steps that each may set `action.Err`
// during execution
func (base *Base) Do(fns ...func()) {
	for _, fn := range fns {
		if base.Err != nil {
			return
		}

		fn()
	}
}

// Setup runs the provided funcs if and only if no call to Setup() has been
// made previously on this action.
func (base *Base) Setup(fns ...func()) {
	if base.isSetup {
		return
	}
	base.Do(fns...)
	base.isSetup = true
}
