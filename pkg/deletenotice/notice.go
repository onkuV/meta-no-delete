// Package deletenotice builds the bridged "deletion attempted" notice used by
// this fork instead of redacting messages that were deleted on the network side.
//
// It lives in its own package so both the Messenger connector (pkg/connector)
// and the Instagram connector (pkg/igconnector) emit an identical notice, and
// so the fork carries as little code as possible inside files that upstream
// also modifies.
package deletenotice

import (
	"fmt"

	"maunium.net/go/mautrix/bridgev2"
	"maunium.net/go/mautrix/bridgev2/networkid"
	"maunium.net/go/mautrix/event"
)

// Make builds the deletion notice for the given message ID. replyTo may be
// empty when the network ID of the deleted message cannot be resolved, in
// which case the notice is sent without a reply.
func Make(deletedID string, replyTo networkid.MessageID) *bridgev2.ConvertedMessage {
	msg := &bridgev2.ConvertedMessage{
		Parts: []*bridgev2.ConvertedMessagePart{{
			Type: event.EventMessage,
			Content: &event.MessageEventContent{
				MsgType: event.MsgText,
				Body:    fmt.Sprintf("🚮 Message deletion attempted (ID: %s)", deletedID),
			},
		}},
	}
	if replyTo != "" {
		msg.ReplyTo = &networkid.MessageOptionalPartID{MessageID: replyTo}
	}
	return msg
}
