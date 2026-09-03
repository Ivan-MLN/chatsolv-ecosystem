package conversation

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseRelayCommand(t *testing.T) {
	// Test #ACC
	cmd := ParseRelayCommand("#ACC 12345678", false)
	require.Equal(t, RelayCmdAccept, cmd.Type)
	require.Equal(t, "12345678", cmd.ConversationID)

	cmd = ParseRelayCommand("#acc #CNV-abcd1234", false)
	require.Equal(t, RelayCmdAccept, cmd.Type)
	require.Equal(t, "abcd1234", cmd.ConversationID)

	// Test #DONE & #CLOSE
	cmd = ParseRelayCommand("#done", true)
	require.Equal(t, RelayCmdDone, cmd.Type)

	cmd = ParseRelayCommand("#CLOSE", true)
	require.Equal(t, RelayCmdDone, cmd.Type)

	cmd = ParseRelayCommand("#selesai", true)
	require.Equal(t, RelayCmdDone, cmd.Type)

	// Test direct message during active relay session (tanpa prefix #)
	cmd = ParseRelayCommand("Halo kak, barang ready stock ya!", true)
	require.Equal(t, RelayCmdRelay, cmd.Type)
	require.Equal(t, "Halo kak, barang ready stock ya!", cmd.Text)

	// Test with # prefix during active session
	cmd = ParseRelayCommand("# Halo kak, ada diskon 10%", true)
	require.Equal(t, RelayCmdRelay, cmd.Type)
	require.Equal(t, "Halo kak, ada diskon 10%", cmd.Text)

	// Test regular message without active session
	cmd = ParseRelayCommand("Halo kak, saya mau tanya harga", false)
	require.Equal(t, RelayCmdNone, cmd.Type)
}

func TestFormatEscalationBroadcast(t *testing.T) {
	text := FormatEscalationBroadcast("6281234567890", "98bb6c6d-1234", "Bisa kirim hari ini?")
	require.Contains(t, text, "⚠️ [ESKALASI PERCAKAPAN]")
	require.Contains(t, text, "6281234567890")
	require.Contains(t, text, "#CNV-98bb6c6d")
	require.Contains(t, text, "Bisa kirim hari ini?")
	require.Contains(t, text, "#ACC 98bb6c6d")

	fwd := FormatForwardToAdmin("6281234567890", "Halo saya kirim bukti transfer ya")
	require.Contains(t, fwd, "📩 [PESAN DARI PELANGGAN: 6281234567890]")
	require.Contains(t, fwd, "Halo saya kirim bukti transfer ya")
	require.Contains(t, fwd, "Ketik balasan Anda langsung")
}
