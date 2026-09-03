package whatsapp

import (
	"context"
	"fmt"
	"sync"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	waLog "go.mau.fi/whatsmeow/util/log"
	"google.golang.org/protobuf/proto"
)

// Session holds a single active WhatsApp connection for one channel.
type Session struct {
	ChannelID string
	client    *whatsmeow.Client
	store     *sqlstore.Container
	cancel    context.CancelFunc
	closeOnce sync.Once
}

// ConnectResult is returned by Session.connect.
type ConnectResult struct {
	SessionID   string // always == ChannelID
	Status      string // "waiting_pairing" | "connected"
	QR          string // non-empty only when status == "waiting_pairing" and no phone
	PairingCode string // pairing code (XXXX-XXXX) when phone number is provided
}

// newSession creates a whatsmeow client backed by the given store container
// and registers eventFn as its event handler.
func newSession(ctx context.Context, channelID, phoneNumber, dbRoot string, eventFn func(interface{})) (*Session, ConnectResult, error) {
	container, err := openStore(ctx, dbRoot, channelID)
	if err != nil {
		return nil, ConnectResult{}, err
	}

	device, err := container.GetFirstDevice(ctx)
	if err != nil {
		_ = container.Close()
		return nil, ConnectResult{}, fmt.Errorf("get device: %w", err)
	}

	client := whatsmeow.NewClient(device, waLog.Noop)
	client.AddEventHandler(eventFn)

	lifecycleCtx, cancel := detachedSessionContext(ctx)
	sess := &Session{ChannelID: channelID, client: client, store: container, cancel: cancel}

	if client.Store.ID == nil {
		// No paired device yet: open QR or pair with phone number.
		result, err := sess.connectNew(ctx, lifecycleCtx, phoneNumber)
		if err != nil {
			sess.Disconnect()
			return nil, ConnectResult{}, err
		}
		return sess, result, nil
	}

	// Already paired: just reconnect.
	if err = client.Connect(); err != nil {
		sess.Disconnect()
		return nil, ConnectResult{}, fmt.Errorf("reconnect: %w", err)
	}
	return sess, ConnectResult{SessionID: channelID, Status: "connected"}, nil
}

// connectNew opens a QR channel, starts the connection, and waits for the
// first QR code or requests a phone pairing code.
func (s *Session) connectNew(requestCtx, lifecycleCtx context.Context, phoneNumber string) (ConnectResult, error) {
	qrChan, err := s.client.GetQRChannel(lifecycleCtx)
	if err != nil {
		return ConnectResult{}, fmt.Errorf("get qr channel: %w", err)
	}
	if err = s.client.Connect(); err != nil {
		return ConnectResult{}, fmt.Errorf("client connect: %w", err)
	}

	select {
	case qrEvt := <-qrChan:
		if qrEvt.Event == "code" {
			if phoneNumber != "" {
				code, pairErr := s.client.PairPhone(requestCtx, phoneNumber, true, whatsmeow.PairClientChrome, "Chrome (Linux)")
				if pairErr != nil {
					return ConnectResult{}, fmt.Errorf("pair phone: %w", pairErr)
				}
				return ConnectResult{
					SessionID:   s.ChannelID,
					Status:      "waiting_pairing",
					QR:          qrEvt.Code,
					PairingCode: code,
				}, nil
			}
			return ConnectResult{SessionID: s.ChannelID, Status: "waiting_pairing", QR: qrEvt.Code}, nil
		}
		return ConnectResult{SessionID: s.ChannelID, Status: "waiting_pairing"}, nil
	case <-requestCtx.Done():
		return ConnectResult{}, requestCtx.Err()
	}
}

// detachedSessionContext preserves request values without tying the WhatsApp
// socket lifecycle to the short-lived HTTP request that created it.
func detachedSessionContext(requestCtx context.Context) (context.Context, context.CancelFunc) {
	return context.WithCancel(context.WithoutCancel(requestCtx))
}

// Disconnect tears down the WhatsApp connection for this session.
func (s *Session) Disconnect() {
	s.closeOnce.Do(func() {
		s.cancel()
		s.client.Disconnect()
		_ = s.store.Close()
	})
}

// IsConnected reports whether the underlying WhatsApp client is connected.
func (s *Session) IsConnected() bool {
	return s.client.IsConnected()
}

// WhatsAppProfile holds full WhatsApp account and business profile information.
type WhatsAppProfile struct {
	PushName        string   `json:"push_name"`
	Phone           string   `json:"phone"`
	JID             string   `json:"jid"`
	LID             string   `json:"lid,omitempty"`
	Status          string   `json:"status,omitempty"`
	ProfilePicture  string   `json:"profile_picture,omitempty"`
	IsBusiness      bool     `json:"is_business"`
	Description     string   `json:"description,omitempty"`
	Address         string   `json:"address,omitempty"`
	Email           string   `json:"email,omitempty"`
	Categories      []string `json:"categories,omitempty"`
	BusinessHoursTZ string   `json:"business_hours_tz,omitempty"`
}

// GetProfile fetches the connected user's profile and business metadata from WhatsApp.
func (s *Session) GetProfile(ctx context.Context) (WhatsAppProfile, error) {
	if s.client == nil || s.client.Store == nil || s.client.Store.ID == nil {
		return WhatsAppProfile{}, fmt.Errorf("session not ready")
	}

	userJID := *s.client.Store.ID
	profile := WhatsAppProfile{
		PushName: s.client.Store.PushName,
		Phone:    userJID.User,
		JID:      userJID.String(),
	}
	if !s.client.Store.LID.IsEmpty() {
		profile.LID = s.client.Store.LID.String()
	}

	// 1. Get Profile Picture
	if pic, err := s.client.GetProfilePictureInfo(ctx, userJID, &whatsmeow.GetProfilePictureParams{Preview: false}); err == nil && pic != nil {
		profile.ProfilePicture = pic.URL
	}

	// 2. Get Basic User Info (Status / About)
	if uInfoMap, err := s.client.GetUserInfo(ctx, []types.JID{userJID}); err == nil {
		if uInfo, ok := uInfoMap[userJID]; ok {
			profile.Status = uInfo.Status
			if profile.PushName == "" && uInfo.VerifiedName != nil && uInfo.VerifiedName.Details != nil {
				profile.PushName = uInfo.VerifiedName.Details.GetVerifiedName()
			}
		}
	}

	// 3. Try Get Business Profile
	if bProfile, err := s.client.GetBusinessProfile(ctx, userJID); err == nil && bProfile != nil {
		profile.IsBusiness = true
		profile.Address = bProfile.Address
		profile.Email = bProfile.Email
		profile.BusinessHoursTZ = bProfile.BusinessHoursTimeZone
		if desc, ok := bProfile.ProfileOptions["description"]; ok {
			profile.Description = desc
		}
		for _, cat := range bProfile.Categories {
			profile.Categories = append(profile.Categories, cat.Name)
		}
	}

	return profile, nil
}
func (s *Session) ResolvePhoneNumber(ctx context.Context, jid types.JID) types.JID {
	if jid.Server == types.DefaultUserServer && jid.User != "" {
		return jid
	}
	if s.client != nil && s.client.Store != nil {
		if alt, err := s.client.Store.GetAltJID(ctx, jid); err == nil && !alt.IsEmpty() && alt.Server == types.DefaultUserServer {
			return alt
		}
		if s.client.Store.LIDs != nil {
			if pn, err := s.client.Store.LIDs.GetPNForLID(ctx, jid); err == nil && !pn.IsEmpty() {
				return pn
			}
		}
	}
	return jid
}
func (s *Session) DownloadAttachment(ctx context.Context, msg whatsmeow.DownloadableMessage) ([]byte, error) {
	return s.client.Download(ctx, msg)
}
func (s *Session) SendPresence(ctx context.Context, jid types.JID, state types.ChatPresence) error {
	return s.client.SendChatPresence(ctx, jid, state, types.ChatPresenceMediaText)
}

func (s *Session) SendText(ctx context.Context, jid types.JID, text string, quotedMsgID string, quotedParticipant string, quotedMsg *waE2E.Message) error {
	if quotedMsgID != "" {
		contextInfo := &waE2E.ContextInfo{
			StanzaID:      proto.String(quotedMsgID),
			Participant:   proto.String(quotedParticipant),
			QuotedMessage: quotedMsg,
		}
		_, err := s.client.SendMessage(ctx, jid, &waE2E.Message{
			ExtendedTextMessage: &waE2E.ExtendedTextMessage{
				Text:        proto.String(text),
				ContextInfo: contextInfo,
			},
		})
		return err
	}
	_, err := s.client.SendMessage(ctx, jid, &waE2E.Message{Conversation: proto.String(text)})
	return err
}
