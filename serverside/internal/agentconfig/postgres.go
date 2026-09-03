package agentconfig

import (
	"authbackend/generated/sqlc"
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct{ pool *pgxpool.Pool }

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository { return &PostgresRepository{pool} }
func (r *PostgresRepository) ResolveDefaultAgent(ctx context.Context, userID, workspaceID string) (Agent, string, error) {
	uid, err := dbUUID(userID)
	if err != nil {
		return Agent{}, "", err
	}
	wid, err := dbUUID(workspaceID)
	if err != nil {
		return Agent{}, "", err
	}
	row, err := sqlc.New(r.pool).GetDefaultAgentForWorkspaceMember(ctx, sqlc.GetDefaultAgentForWorkspaceMemberParams{WorkspaceID: wid, UserID: uid})
	if err != nil {
		return Agent{}, "", err
	}
	return Agent{ID: idString(row.ID), WorkspaceID: idString(row.WorkspaceID), Name: row.Name, Status: row.Status, Provider: row.Provider, ConfigVersion: row.ConfigVersion, SyncedConfigVersion: row.SyncedConfigVersion}, row.Role, nil
}
func (r *PostgresRepository) UpdateAgent(ctx context.Context, agent Agent) (Agent, error) {
	id, err := dbUUID(agent.ID)
	if err != nil {
		return Agent{}, err
	}
	row, err := sqlc.New(r.pool).UpdateAgentName(ctx, sqlc.UpdateAgentNameParams{ID: id, Name: agent.Name})
	if err != nil {
		return Agent{}, err
	}
	return Agent{ID: idString(row.ID), WorkspaceID: idString(row.WorkspaceID), Name: row.Name, Status: row.Status, Provider: row.Provider, ConfigVersion: row.ConfigVersion, SyncedConfigVersion: row.SyncedConfigVersion}, nil
}
func (r *PostgresRepository) Authorize(ctx context.Context, userID, agentID string) (string, error) {
	uid, err := dbUUID(userID)
	if err != nil {
		return "", err
	}
	aid, err := dbUUID(agentID)
	if err != nil {
		return "", err
	}
	row, err := sqlc.New(r.pool).AuthorizeAgentMember(ctx, sqlc.AuthorizeAgentMemberParams{ID: aid, UserID: uid})
	return row.Role, err
}
func (r *PostgresRepository) SavePersonality(ctx context.Context, p Personality) (int64, error) {
	aid, err := dbUUID(p.AgentID)
	if err != nil {
		return 0, err
	}
	row, err := sqlc.New(r.pool).GetAgentSyncResource(ctx, aid)
	if err != nil {
		return 0, err
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	q := sqlc.New(tx)
	id, _ := dbUUID(uuid.NewString())
	behavior, _ := json.Marshal(p.BehaviorRules)
	escalation, _ := json.Marshal(p.EscalationRules)
	forbidden, _ := json.Marshal(p.ForbiddenTopics)
	_, err = q.UpsertPersonality(ctx, sqlc.UpsertPersonalityParams{ID: id, WorkspaceID: row.WorkspaceID, AgentID: aid, BotName: p.BotName, Role: p.Role, Tone: p.Tone, CommunicationStyle: p.CommunicationStyle, PrimaryLanguage: p.PrimaryLanguage, ResponseLength: p.ResponseLength, EmojiUsage: p.EmojiUsage, GreetingStyle: p.GreetingStyle, ClosingStyle: p.ClosingStyle, CustomInstructions: p.CustomInstructions, BehaviorRules: behavior, EscalationRules: escalation, ForbiddenTopics: forbidden, FallbackBehavior: p.FallbackBehavior})
	if err != nil {
		return 0, err
	}
	version, err := q.IncrementAgentConfigVersion(ctx, aid)
	if err != nil {
		return 0, err
	}
	eventID, _ := dbUUID(uuid.NewString())
	if err = q.CreateAgentSyncEvent(ctx, sqlc.CreateAgentSyncEventParams{ID: eventID, WorkspaceID: row.WorkspaceID, AggregateID: aid, Payload: []byte(`{}`)}); err != nil {
		return 0, err
	}
	return version, tx.Commit(ctx)
}
func (r *PostgresRepository) GetPersonality(ctx context.Context, agentID string) (Personality, error) {
	aid, err := dbUUID(agentID)
	if err != nil {
		return Personality{}, err
	}
	row, err := sqlc.New(r.pool).GetPersonality(ctx, aid)
	if err != nil {
		return Personality{}, err
	}
	p := Personality{WorkspaceID: idString(row.WorkspaceID), AgentID: idString(row.AgentID), BotName: row.BotName, Role: row.Role, Tone: row.Tone, CommunicationStyle: row.CommunicationStyle, PrimaryLanguage: row.PrimaryLanguage, ResponseLength: row.ResponseLength, EmojiUsage: row.EmojiUsage, GreetingStyle: row.GreetingStyle, ClosingStyle: row.ClosingStyle, CustomInstructions: row.CustomInstructions, FallbackBehavior: row.FallbackBehavior}
	_ = json.Unmarshal(row.BehaviorRules, &p.BehaviorRules)
	_ = json.Unmarshal(row.EscalationRules, &p.EscalationRules)
	_ = json.Unmarshal(row.ForbiddenTopics, &p.ForbiddenTopics)
	return p, nil
}
func dbUUID(value string) (pgtype.UUID, error) {
	id, err := uuid.Parse(value)
	return pgtype.UUID{Bytes: id, Valid: err == nil}, err
}
func idString(value pgtype.UUID) string { return uuid.UUID(value.Bytes).String() }

func (r *PostgresRepository) AuthorizeWorkspace(ctx context.Context, userID, workspaceID string) (string, error) {
	uid, err := dbUUID(userID)
	if err != nil {
		return "", err
	}
	wid, err := dbUUID(workspaceID)
	if err != nil {
		return "", err
	}
	return sqlc.New(r.pool).AuthorizeWorkspaceMember(ctx, sqlc.AuthorizeWorkspaceMemberParams{WorkspaceID: wid, UserID: uid})
}

func (r *PostgresRepository) SaveAgentProfile(ctx context.Context, p AgentProfile) (int64, error) {
	aid, err := dbUUID(p.AgentID)
	if err != nil {
		return 0, err
	}
	resource, err := sqlc.New(r.pool).GetAgentSyncResource(ctx, aid)
	if err != nil {
		return 0, err
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	q := sqlc.New(tx)
	id, _ := dbUUID(uuid.NewString())
	_, err = q.UpsertAgentProfile(ctx, sqlc.UpsertAgentProfileParams{ID: id, WorkspaceID: resource.WorkspaceID, AgentID: aid, DisplayName: p.DisplayName, AvatarObjectKey: textValue(p.AvatarObjectKey), Description: p.Description, GreetingMessage: p.GreetingMessage, AwayMessage: p.AwayMessage, FallbackMessage: p.FallbackMessage, Language: p.Language})
	if err != nil {
		return 0, err
	}
	version, err := q.IncrementAgentConfigVersion(ctx, aid)
	if err != nil {
		return 0, err
	}
	eventID, _ := dbUUID(uuid.NewString())
	err = q.CreateAgentSyncEvent(ctx, sqlc.CreateAgentSyncEventParams{ID: eventID, WorkspaceID: resource.WorkspaceID, AggregateID: aid, Payload: []byte(`{"resource":"profile"}`)})
	if err != nil {
		return 0, err
	}
	return version, tx.Commit(ctx)
}

func (r *PostgresRepository) GetAgentProfile(ctx context.Context, agentID string) (AgentProfile, error) {
	aid, err := dbUUID(agentID)
	if err != nil {
		return AgentProfile{}, err
	}
	row, err := sqlc.New(r.pool).GetAgentProfile(ctx, aid)
	if err != nil {
		return AgentProfile{}, err
	}
	return AgentProfile{WorkspaceID: idString(row.WorkspaceID), AgentID: idString(row.AgentID), DisplayName: row.DisplayName, AvatarObjectKey: row.AvatarObjectKey.String, Description: row.Description, GreetingMessage: row.GreetingMessage, AwayMessage: row.AwayMessage, FallbackMessage: row.FallbackMessage, Language: row.Language}, nil
}

func (r *PostgresRepository) SaveBusinessProfile(ctx context.Context, p BusinessProfile) (int64, error) {
	wid, err := dbUUID(p.WorkspaceID)
	if err != nil {
		return 0, err
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	q := sqlc.New(tx)
	hours, _ := json.Marshal(p.BusinessHours)
	values, _ := json.Marshal(p.CompanyValues)
	useCases, _ := json.Marshal(p.PrimaryUseCases)
	var handoffRules []byte
	if p.HandoffRules != nil {
		handoffRules, _ = json.Marshal(p.HandoffRules)
	} else {
		handoffRules = []byte(`{"customer_request": true, "low_confidence": true, "serious_complaint": true, "refund": true, "payment_issue": true, "timeout_minutes": 2, "rotation_system": "round_robin"}`)
	}
	var opHours []byte
	if p.OperatingHours != nil {
		opHours, _ = json.Marshal(p.OperatingHours)
	} else {
		opHours = []byte(`{}`)
	}
	id, _ := dbUUID(uuid.NewString())
	bType := p.BusinessType
	if bType == "" {
		bType = "products_and_services"
	}
	commStyle := p.CommunicationStyle
	if commStyle == "" {
		commStyle = "friendly_professional"
	}
	_, err = q.UpsertBusinessProfile(ctx, sqlc.UpsertBusinessProfileParams{
		ID:                  id,
		WorkspaceID:         wid,
		BusinessName:        p.BusinessName,
		Industry:            p.Industry,
		BusinessDescription: p.BusinessDescription,
		Website:             textValue(p.Website),
		Email:               textValue(p.Email),
		Phone:               textValue(p.Phone),
		Address:             p.Address,
		BusinessHours:       hours,
		Timezone:            p.Timezone,
		BrandVoice:          p.BrandVoice,
		CompanyValues:       values,
		BusinessType:        bType,
		TargetCustomer:      p.TargetCustomer,
		ProductsServices:    p.ProductsServices,
		CommunicationStyle:  commStyle,
		PrimaryUseCases:     useCases,
		HandoffRules:        handoffRules,
		OperatingHours:      opHours,
	})
	if err != nil {
		return 0, err
	}
	version, err := q.IncrementWorkspaceAgentConfig(ctx, wid)
	if err != nil {
		return 0, err
	}
	eventID, _ := dbUUID(uuid.NewString())
	err = q.QueueAgentSyncForWorkspace(ctx, sqlc.QueueAgentSyncForWorkspaceParams{ID: eventID, WorkspaceID: wid, Payload: []byte(`{"resource":"business"}`)})
	if err != nil {
		return 0, err
	}
	return version, tx.Commit(ctx)
}

func (r *PostgresRepository) GetBusinessProfile(ctx context.Context, workspaceID string) (BusinessProfile, error) {
	wid, err := dbUUID(workspaceID)
	if err != nil {
		return BusinessProfile{}, err
	}
	row, err := sqlc.New(r.pool).GetBusinessProfile(ctx, wid)
	if err != nil {
		return BusinessProfile{}, err
	}
	p := BusinessProfile{
		WorkspaceID:         workspaceID,
		BusinessName:        row.BusinessName,
		Industry:            row.Industry,
		BusinessType:        row.BusinessType,
		BusinessDescription: row.BusinessDescription,
		TargetCustomer:      row.TargetCustomer,
		ProductsServices:    row.ProductsServices,
		CommunicationStyle:  row.CommunicationStyle,
		Website:             row.Website.String,
		Email:               row.Email.String,
		Phone:               row.Phone.String,
		Address:             row.Address,
		Timezone:            row.Timezone,
		BrandVoice:          row.BrandVoice,
	}
	_ = json.Unmarshal(row.BusinessHours, &p.BusinessHours)
	_ = json.Unmarshal(row.CompanyValues, &p.CompanyValues)
	_ = json.Unmarshal(row.PrimaryUseCases, &p.PrimaryUseCases)
	if len(row.HandoffRules) > 0 {
		var hr HandoffRulesConfig
		if json.Unmarshal(row.HandoffRules, &hr) == nil {
			p.HandoffRules = &hr
		}
	}
	if len(row.OperatingHours) > 0 {
		_ = json.Unmarshal(row.OperatingHours, &p.OperatingHours)
	}
	return p, nil
}

func (r *PostgresRepository) SaveBusinessPolicies(ctx context.Context, p BusinessPolicies) (int64, error) {
	wid, err := dbUUID(p.WorkspaceID)
	if err != nil {
		return 0, err
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	q := sqlc.New(tx)
	id, _ := dbUUID(uuid.NewString())
	_, err = q.UpsertBusinessPolicies(ctx, sqlc.UpsertBusinessPoliciesParams{ID: id, WorkspaceID: wid, ShippingPolicy: p.ShippingPolicy, RefundPolicy: p.RefundPolicy, ReturnPolicy: p.ReturnPolicy, WarrantyPolicy: p.WarrantyPolicy, PaymentPolicy: p.PaymentPolicy, ComplaintPolicy: p.ComplaintPolicy})
	if err != nil {
		return 0, err
	}
	version, err := q.IncrementWorkspaceAgentConfig(ctx, wid)
	if err != nil {
		return 0, err
	}
	eventID, _ := dbUUID(uuid.NewString())
	err = q.QueueAgentSyncForWorkspace(ctx, sqlc.QueueAgentSyncForWorkspaceParams{ID: eventID, WorkspaceID: wid, Payload: []byte(`{"resource":"policies"}`)})
	if err != nil {
		return 0, err
	}
	return version, tx.Commit(ctx)
}

func (r *PostgresRepository) GetBusinessPolicies(ctx context.Context, workspaceID string) (BusinessPolicies, error) {
	wid, err := dbUUID(workspaceID)
	if err != nil {
		return BusinessPolicies{}, err
	}
	row, err := sqlc.New(r.pool).GetBusinessPolicies(ctx, wid)
	if err != nil {
		return BusinessPolicies{}, err
	}
	return BusinessPolicies{WorkspaceID: workspaceID, ShippingPolicy: row.ShippingPolicy, RefundPolicy: row.RefundPolicy, ReturnPolicy: row.ReturnPolicy, WarrantyPolicy: row.WarrantyPolicy, PaymentPolicy: row.PaymentPolicy, ComplaintPolicy: row.ComplaintPolicy}, nil
}
func textValue(value string) pgtype.Text { return pgtype.Text{String: value, Valid: value != ""} }
