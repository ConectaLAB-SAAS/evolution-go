package whatsmeow_service

import (
	"bytes"
	"testing"

	"go.mau.fi/whatsmeow/proto/waCompanionReg"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	"google.golang.org/protobuf/proto"
)

func TestApplyWAVersionUpdatesPairingAndReconnectHandshake(t *testing.T) {
	preserveWAVersion(t)
	version := clientVersion{Major: 2, Minor: 3000, Patch: 1052889962}
	want := store.WAVersionContainer{2, 3000, 1052889962}
	applyWAVersion(version)

	container := &sqlstore.Container{}
	device := container.NewDevice()
	payload := device.GetClientPayload()
	appVersion := payload.GetUserAgent().GetAppVersion()
	got := store.WAVersionContainer{appVersion.GetPrimary(), appVersion.GetSecondary(), appVersion.GetTertiary()}
	if got != want {
		t.Fatalf("pairing handshake version = %s, want %s", got, want)
	}
	hash := want.Hash()
	if !bytes.Equal(payload.GetDevicePairingData().GetBuildHash(), hash[:]) {
		t.Fatal("pairing build hash must match the advertised handshake version")
	}
	props := &waCompanionReg.DeviceProps{}
	if err := proto.Unmarshal(payload.GetDevicePairingData().GetDeviceProps(), props); err != nil {
		t.Fatal(err)
	}
	if props.GetVersion().GetTertiary() != want[2] {
		t.Fatal("companion device version must preserve the configured version")
	}

	device.ID = &types.JID{User: "15555550123", Device: 1, Server: types.DefaultUserServer}
	appVersion = device.GetClientPayload().GetUserAgent().GetAppVersion()
	got = store.WAVersionContainer{appVersion.GetPrimary(), appVersion.GetSecondary(), appVersion.GetTertiary()}
	if got != want {
		t.Fatalf("reconnect handshake version = %s, want %s", got, want)
	}
}

func TestApplyWAVersionPreservesFallbackWhenVersionIsUnavailable(t *testing.T) {
	preserveWAVersion(t)
	want := store.GetWAVersion()
	props := proto.Clone(store.DeviceProps)
	for _, version := range []clientVersion{{}, {Major: 2, Minor: 3000}, {Major: 2, Patch: 1}} {
		applyWAVersion(version)
		if got := store.GetWAVersion(); got != want {
			t.Fatalf("unavailable version replaced fallback %s with %s", want, got)
		}
		if !proto.Equal(store.DeviceProps, props) {
			t.Fatal("unavailable version must preserve companion device properties")
		}
	}
}

func preserveWAVersion(t *testing.T) {
	t.Helper()
	version := store.GetWAVersion()
	props := proto.Clone(store.DeviceProps).(*waCompanionReg.DeviceProps)
	t.Cleanup(func() {
		store.SetWAVersion(version)
		store.DeviceProps = props
	})
}
