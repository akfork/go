package horizon

import (
	"github.com/kinecosystem/go/services/horizon/internal/db2/core"
	"github.com/kinecosystem/go/services/horizon/internal/resource"
	"github.com/kinecosystem/go/support/render/hal"
	"github.com/kinecosystem/go/services/horizon/internal/render/sse"
)

// This file contains the actions:
//
// AccountShowActionBasic: details for single account (from stellar-core)

// AccountShowActionBasic renders a account summary found by its address.
type AccountShowActionBasic struct {
	Action
	Address        string
	CoreRecord     core.Account
	Resource       resource.Account
}

// JSON is a method for actions.JSON
func (action *AccountShowActionBasic) JSON() {
	action.Do(
		action.loadParams,
		action.loadRecord,
		action.loadResource,
		func() {
			hal.Render(action.W, action.Resource)
		},
	)
}

// SSE is a method for actions.SSE
func (action *AccountShowActionBasic) SSE(stream sse.Stream) {
	action.Do(
		action.loadParams,
		action.loadRecord,
		action.loadResource,
		func() {
			stream.SetLimit(10)
			stream.Send(sse.Event{Data: action.Resource})
		},
	)
}

// GetTopic is a method for actions.SSE
func (action *AccountShowActionBasic) GetTopic() string {
	return action.GetString("id")
}

func (action *AccountShowActionBasic) loadParams() {
	action.Address = action.GetString("id")
}

func (action *AccountShowActionBasic) loadRecord() {
	action.Err = action.CoreQ().
		AccountByAddress(&action.CoreRecord, action.Address)
}

func (action *AccountShowActionBasic) loadResource() {
	action.Err = action.Resource.PopulateBasic(
		action.Ctx,
		action.CoreRecord,
	)
}
