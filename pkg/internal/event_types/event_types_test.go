package event_types

import (
	"reflect"
	"testing"
)

func TestParseSubscribedEvents(t *testing.T) {
	tests := []struct {
		name   string
		events string
		want   []string
	}{
		{
			name:   "empty defaults to message",
			events: "",
			want:   []string{MESSAGE},
		},
		{
			name:   "whitespace and invalid values default to message",
			events: " ,UNKNOWN, ",
			want:   []string{MESSAGE},
		},
		{
			name:   "valid values are trimmed and deduplicated",
			events: " MESSAGE, CONNECTION,MESSAGE,QRCODE ",
			want:   []string{MESSAGE, CONNECTION, QRCODE},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := ParseSubscribedEvents(test.events)
			if !reflect.DeepEqual(got, test.want) {
				t.Fatalf("ParseSubscribedEvents(%q) = %#v, want %#v", test.events, got, test.want)
			}
		})
	}
}
