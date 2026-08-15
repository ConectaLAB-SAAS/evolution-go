package event_types

import "strings"

const (
	ALL           = "ALL"
	MESSAGE       = "MESSAGE"
	SEND_MESSAGE  = "SEND_MESSAGE"
	READ_RECEIPT  = "READ_RECEIPT"
	PRESENCE      = "PRESENCE"
	HISTORY_SYNC  = "HISTORY_SYNC"
	CHAT_PRESENCE = "CHAT_PRESENCE"
	CALL          = "CALL"
	CONNECTION    = "CONNECTION"
	LABEL         = "LABEL"
	CONTACT       = "CONTACT"
	GROUP         = "GROUP"
	NEWSLETTER    = "NEWSLETTER"
	QRCODE        = "QRCODE"
	BUTTON_CLICK  = "BUTTON_CLICK"
	PICTURE       = "PICTURE"
	USER_ABOUT    = "USER_ABOUT"
)

var AllEventTypes = []string{
	MESSAGE,
	SEND_MESSAGE,
	READ_RECEIPT,
	PRESENCE,
	HISTORY_SYNC,
	CHAT_PRESENCE,
	CALL,
	CONNECTION,
	LABEL,
	CONTACT,
	GROUP,
	NEWSLETTER,
	QRCODE,
	BUTTON_CLICK,
	PICTURE,
	USER_ABOUT,
}

var validEventTypes = map[string]bool{
	ALL:           true,
	MESSAGE:       true,
	SEND_MESSAGE:  true,
	READ_RECEIPT:  true,
	PRESENCE:      true,
	HISTORY_SYNC:  true,
	CHAT_PRESENCE: true,
	CALL:          true,
	CONNECTION:    true,
	LABEL:         true,
	CONTACT:       true,
	GROUP:         true,
	NEWSLETTER:    true,
	QRCODE:        true,
	BUTTON_CLICK:  true,
	PICTURE:       true,
	USER_ABOUT:    true,
}

func IsEventType(eventType string) bool {
	return validEventTypes[eventType]
}

// ParseSubscribedEvents normalizes persisted subscriptions and keeps MESSAGE
// as the safe default when the stored value is empty or invalid.
func ParseSubscribedEvents(events string) []string {
	subscriptions := make([]string, 0)
	seen := make(map[string]struct{})

	for _, event := range strings.Split(events, ",") {
		event = strings.TrimSpace(event)
		if !IsEventType(event) {
			continue
		}
		if _, exists := seen[event]; exists {
			continue
		}

		seen[event] = struct{}{}
		subscriptions = append(subscriptions, event)
	}

	if len(subscriptions) == 0 {
		return []string{MESSAGE}
	}

	return subscriptions
}
