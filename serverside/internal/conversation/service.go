package conversation

import (
	"context"
	"errors"
)

type Mode string

const (
	ModeAgent           Mode = "agent"
	ModeWaitingForAdmin Mode = "waiting_for_admin"
	ModeHuman           Mode = "human"
)

var ErrHumanMode = errors.New("conversation is in human mode")
var ErrWaitingForAdminMode = errors.New("conversation is waiting for admin")
var ErrRuntimeDisabled = errors.New("agent runtime disabled")
var ErrMessageLimitReached = errors.New("monthly message limit reached")

type Incoming struct{ ChannelID, ChannelType, ExternalMessageID, ExternalUserID, Text, Provider, Environment string }
type Conversation struct {
	ID, WorkspaceID, AgentID, ChannelID, Environment string
	Mode                                             Mode
}
type Message struct{ Sender, Content string }
type Result struct {
	MessageID, ConversationID, Content string
	HandoffRequested                   bool
	HandoffReason                      string
}
type RuntimeInput struct {
	WorkspaceID, AgentID, ConversationID, Message string
	History                                       []Message
}
type RuntimeOutput struct {
	Content          string
	HandoffRequested bool
	HandoffReason    string
}
type Repository interface {
	FindResult(context.Context, string, string) (*Result, error)
	ResolveOrCreate(context.Context, Incoming) (Conversation, error)
	SaveCustomer(context.Context, Conversation, Incoming) error
	RecentMessages(context.Context, string, int) ([]Message, error)
	SaveAgent(context.Context, Conversation, string) (Result, error)
	RequestHandoff(context.Context, Conversation, string) error
}
type Runtime interface {
	Generate(context.Context, RuntimeInput) (RuntimeOutput, error)
}
type Locker interface {
	WithLock(context.Context, string, func(context.Context) error) error
}

type HandoffNotifier interface {
	NotifyHandoff(ctx context.Context, channelID, workspaceID, conversationID, customerPhone, reason string) error
}

type Service struct {
	repo     Repository
	runtime  Runtime
	locks    Locker
	handoffs HandoffNotifier
}

func NewService(repo Repository, runtime Runtime, locks Locker) *Service {
	return &Service{repo: repo, runtime: runtime, locks: locks}
}

func (s *Service) SetHandoffNotifier(h HandoffNotifier) {
	s.handoffs = h
}

func (s *Service) Handle(ctx context.Context, in Incoming) (Result, error) {
	existing, err := s.repo.FindResult(ctx, in.ChannelID, in.ExternalMessageID)
	if err != nil {
		return Result{}, err
	}
	if existing != nil {
		return *existing, nil
	}
	conversation, err := s.repo.ResolveOrCreate(ctx, in)
	if err != nil {
		return Result{}, err
	}
	var result Result
	err = s.locks.WithLock(ctx, conversation.ID, func(lockCtx context.Context) error {
		if err := s.repo.SaveCustomer(lockCtx, conversation, in); err != nil {
			return err
		}
		if conversation.Mode == ModeHuman || conversation.Mode == ModeWaitingForAdmin {
			return ErrHumanMode
		}
		history, err := s.repo.RecentMessages(lockCtx, conversation.ID, 30)
		if err != nil {
			return err
		}
		out, err := s.runtime.Generate(lockCtx, RuntimeInput{WorkspaceID: conversation.WorkspaceID, AgentID: conversation.AgentID, ConversationID: conversation.ID, Message: in.Text, History: history})
		if err != nil {
			return err
		}
		result, err = s.repo.SaveAgent(lockCtx, conversation, out.Content)
		if err == nil {
			result.HandoffRequested = out.HandoffRequested
			result.HandoffReason = out.HandoffReason
		}
		if err == nil && out.HandoffRequested {
			_ = s.repo.RequestHandoff(lockCtx, conversation, out.HandoffReason)
			if s.handoffs != nil {
				_ = s.handoffs.NotifyHandoff(lockCtx, in.ChannelID, conversation.WorkspaceID, conversation.ID, in.ExternalUserID, out.HandoffReason)
			}
		}
		return err
	})
	return result, err
}
