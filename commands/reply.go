package commands

import "github.com/jrevanaldi-ai/gowa/proto/waE2E"

// createSimpleReply membuat Message dengan reply sederhana (tanpa ExternalAdReply)
func createSimpleReply(text string, replyToMsgID string, senderJID string) *waE2E.Message {
	// RemoteJID untuk reply
	remoteJID := senderJID
	
	return &waE2E.Message{
		ExtendedTextMessage: &waE2E.ExtendedTextMessage{
			Text: &text,
			ContextInfo: &waE2E.ContextInfo{
				StanzaID:    &replyToMsgID,
				Participant: &senderJID,
				RemoteJID:   &remoteJID,
				// Clear other ContextInfo fields
				Expiration:              nil,
				EphemeralSettingTimestamp: nil,
				ExternalAdReply:         nil,
				ForwardingScore:         nil,
				IsForwarded:             nil,
			},
		},
	}
}
