package main

import (
	"authbackend/generated/sqlc"
	"authbackend/internal/access"
	"authbackend/internal/adminmgmt"
	"authbackend/internal/agentconfig"
	"authbackend/internal/auth"
	"authbackend/internal/brain/obsidian"
	"authbackend/internal/channel"
	"authbackend/internal/config"
	"authbackend/internal/conversation"
	"authbackend/internal/dashboard"
	"authbackend/internal/database"
	"authbackend/internal/developer/apikey"
	"authbackend/internal/developer/publicapi"
	"authbackend/internal/developer/webhook"
	"authbackend/internal/handoff"
	"authbackend/internal/health"
	"authbackend/internal/hermes"
	"authbackend/internal/internalapi"
	"authbackend/internal/knowledge"
	mw "authbackend/internal/middleware"
	"authbackend/internal/onboarding"
	"authbackend/internal/storage"
	"authbackend/internal/template"
	"authbackend/internal/workspace"
	"authbackend/pkg/response"
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/joho/godotenv"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	_ = godotenv.Load()
	cfg, e := config.Load()
	if e != nil {
		slog.Error("configuration error", "error", e)
		os.Exit(1)
	}
	level := slog.LevelInfo
	if cfg.LogLevel == "debug" {
		level = slog.LevelDebug
	}
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level}))
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	pg, e := database.Postgres(ctx, cfg)
	if e != nil {
		log.Error("postgres connection failed", "error", e)
		os.Exit(1)
	}
	defer pg.Close()
	rd, e := database.Redis(cfg)
	if e != nil {
		log.Error("redis configuration failed", "error", e)
		os.Exit(1)
	}
	defer rd.Close()
	if e = pg.Ping(ctx); e != nil {
		log.Error("postgres ping failed", "error", e)
		os.Exit(1)
	}
	if e = rd.Ping(ctx).Err(); e != nil {
		log.Error("redis ping failed", "error", e)
		os.Exit(1)
	}

	resolver := access.NewResolver(pg)
	repo := auth.NewPostgresRepository(sqlc.New(pg))
	store := auth.NewRedisTokenStore(rd)
	svc := auth.NewService(repo, store, auth.NewDevelopmentEmailSender(log), auth.NewArgon2Hasher(auth.DefaultArgon2Params()), auth.NewJWTManager([]byte(cfg.JWTSecret), cfg.AccessTTL), cfg.RefreshTTL, cfg.ResetTTL)
	h := auth.NewHandler(svc)
	jwtManager := auth.NewJWTManager([]byte(cfg.JWTSecret), cfg.AccessTTL)
	workspaceService := workspace.NewService(workspace.NewPostgresRepository(pg), resolver, time.Now)
	workspaceHandler := workspace.NewHandler(workspaceService)
	agentRepository := agentconfig.NewPostgresRepository(pg)
	agentPersonalityService := agentconfig.NewService(agentRepository)
	agentSettingsService := agentconfig.NewSettingsService(agentRepository)
	agentConfigHandler := agentconfig.NewHandler(agentPersonalityService)
	settingsHandler := agentconfig.NewSettingsHandler(agentSettingsService)
	canonicalBusinessHandler := agentconfig.NewCanonicalBusinessHandler(agentSettingsService)
	brain := obsidian.NewFilesystem(cfg.VaultRoot)
	provider := hermes.NewCLIProvider(cfg.HermesBinary, cfg.HermesRoot, cfg.HermesTemplateProfile, nil)
	runtime := conversation.NewHermesRuntime(pg, provider, brain, brain)
	canonicalAgentHandler := agentconfig.NewCanonicalHandler(agentconfig.NewCanonicalService(agentRepository, agentPersonalityService, agentSettingsService, agentconfig.NewPostgresAgentTester(pg, runtime)))
	dashboardHandler := dashboard.NewHandler(dashboard.NewService(dashboard.NewPostgresRepository(pg)))
	objectStore, e := storage.NewMinIO(cfg.ObjectStorageEndpoint, cfg.ObjectStorageAccessKey, cfg.ObjectStorageSecretKey, cfg.ObjectStorageBucket, cfg.ObjectStorageUseSSL)
	if e != nil {
		log.Error("object storage configuration failed", "error", e)
		os.Exit(1)
	}
	knowledgeHandler := knowledge.NewHandler(knowledge.NewService(knowledge.NewPostgresRepository(pg), objectStore, 20<<20, 200))
	botClient := channel.NewHMACBotClient(cfg.WhatsAppServiceURL, cfg.InternalServiceSecret)
	channelHandler := channel.NewHandler(channel.NewService(channel.NewPostgresRepository(pg), botClient))
	apiKeyHandler := apikey.NewHandler(apikey.NewService(apikey.NewPostgresRepository(pg)))
	webhookHandler := webhook.NewHandler(webhook.NewService(webhook.NewPostgresRepository(pg), []byte(cfg.WebhookEncryptionKey)))

	// Admin Management, Human Handoff & Template Services
	adminRepo := adminmgmt.NewPostgresRepository(pg)
	adminService := adminmgmt.NewService(adminRepo)
	adminHandler := adminmgmt.NewHandler(adminService)

	handoffRepo := handoff.NewPostgresRepository(pg)
	handoffService := handoff.NewService(handoffRepo, adminService, botClient, log)
	handoffHandler := handoff.NewHandler(handoffService)

	templateRepo := template.NewPostgresRepository(pg)
	canonicalAgentService := agentconfig.NewCanonicalService(agentRepository, agentPersonalityService, agentSettingsService, agentconfig.NewPostgresAgentTester(pg, runtime))
	templateService := template.NewService(templateRepo, canonicalAgentService, agentSettingsService)
	templateHandler := template.NewHandler(templateService)

	onboardingRepo := onboarding.NewPostgresRepository(pg)
	onboardingService := onboarding.NewService(onboardingRepo, agentSettingsService, adminService, canonicalAgentService)
	onboardingHandler := onboarding.NewHandler(onboardingService)

	conversationService := conversation.NewService(conversation.NewPostgresRepository(pg), runtime, conversation.NewRedisLocker(rd, 30*time.Second))
	conversationService.SetHandoffNotifier(handoffService)

	inactivityWatcher := conversation.NewInactivityWatcher(pg, botClient, log)
	inactivityCtx, cancelInactivity := context.WithCancel(context.Background())
	defer cancelInactivity()
	inactivityWatcher.Start(inactivityCtx, 20*time.Second)
	publicAPIHandler := publicapi.NewHandler(publicapi.NewService(apikey.NewService(apikey.NewPostgresRepository(pg)), publicapi.NewPostgresRepository(pg), conversationService))
	healthHandler := health.NewHandler(pg.Ping, func(ctx context.Context) error { return rd.Ping(ctx).Err() })
	conversationDashboardHandler := conversation.NewDashboardHandler(conversation.NewDashboardService(conversation.NewDashboardPostgresRepository(pg)))
	internalHandler := internalapi.NewHandler(internalapi.NewService(internalapi.NewPostgresRepository(pg), conversationService))
	internalHandler.SetHandoffs(handoffService)
	app := fiber.New(fiber.Config{BodyLimit: cfg.BodyLimit, ReadTimeout: 60 * time.Second, WriteTimeout: 60 * time.Second, ErrorHandler: func(c *fiber.Ctx, e error) error {
		if fe, ok := e.(*fiber.Error); ok {
			return response.Fail(c, fe.Code, fe.Message, "HTTP_ERROR")
		}
		return response.Fail(c, 500, "Something went wrong", "INTERNAL_ERROR")
	}})
	app.Use(recover.New(), mw.RequestID, mw.Logger(log), mw.SecurityHeaders, compress.New(), cors.New(cors.Config{
		AllowOrigins:     join(cfg.CORSOrigins),
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization, X-Request-ID, X-API-Key, X-Workspace-ID, X-ChatSolv-Signature, X-ChatSolv-Timestamp",
		AllowMethods:     "GET,POST,PATCH,DELETE,PUT,OPTIONS,HEAD",
		AllowCredentials: true,
	}))
	app.Get("/health", healthHandler.Live)
	app.Get("/ready", healthHandler.Ready)
	app.Get("/health/live", healthHandler.Live)
	app.Get("/health/ready", healthHandler.Ready)

	// Public Auth endpoints
	a := app.Group("/api/v1/auth", mw.RateLimit(rd, cfg.RateLimit, cfg.RateWindow))
	a.Post("/register", h.Register)
	a.Post("/login", h.Login)
	a.Post("/refresh", h.Refresh)
	a.Post("/forgot-password", h.Forgot)
	a.Post("/reset-password", h.Reset)

	// Payment callbacks may activate a subscription, so they must never trust an
	// unauthenticated browser request. The payment adapter signs the raw callback
	// body with the internal secret until a provider-specific verifier is wired.
	app.Post("/api/v1/payments/webhook", mw.InternalHMAC(cfg.InternalServiceSecret, 5*time.Minute, time.Now), workspaceHandler.PaymentWebhook)

	// Public API (Protected via API Key and plan entitlement check)
	publicAPI := app.Group("/api/v1/agent-sessions", mw.RateLimit(rd, cfg.RateLimit, cfg.RateWindow))
	publicAPI.Post("/", publicAPIHandler.CreateSession)
	publicAPI.Post("/:id/messages", publicAPIHandler.SendMessage)
	publicAPI.Post("/:id/messages/stream", publicAPIHandler.StreamMessage)

	// Dashboard Authenticated API (Common user & workspace access)
	dashboardAPI := app.Group("/api/v1", auth.RequireAccessToken(jwtManager))
	dashboardAPI.Get("/me", dashboardHandler.Me)
	dashboardAPI.Get("/dashboard", dashboardHandler.Overview)
	dashboardAPI.Get("/workspace", workspaceHandler.CanonicalGet)
	dashboardAPI.Patch("/workspace", workspaceHandler.CanonicalUpdate)

	// AI Agent Templates catalog is accessible for all workspace members
	dashboardAPI.Get("/agent/templates", templateHandler.List)

	// Billing / Subscription / Checkout endpoints (always accessible for authenticated user)
	dashboardAPI.Post("/workspaces/:workspaceID/checkout", workspaceHandler.CreateCheckout)
	dashboardAPI.Get("/workspaces/:workspaceID/subscription", workspaceHandler.Subscription)

	// Gated Product Features (Require Active Subscription OR Developer Platform Role)
	gatedAPI := dashboardAPI.Group("", resolver.RequireActiveSubscription())

	// Onboarding endpoints
	gatedAPI.Get("/onboarding", onboardingHandler.Get)
	gatedAPI.Put("/onboarding", onboardingHandler.SaveProgress)
	gatedAPI.Post("/onboarding/complete", onboardingHandler.Complete)

	// Admin team management endpoints
	gatedAPI.Get("/admins", adminHandler.List)
	gatedAPI.Post("/admins", adminHandler.Create)
	gatedAPI.Patch("/admins/:id", adminHandler.Update)
	gatedAPI.Delete("/admins/:id", adminHandler.Delete)

	// Human handoff endpoints
	gatedAPI.Get("/handoffs", handoffHandler.List)
	gatedAPI.Get("/conversations/:id/events", handoffHandler.ListEvents)
	gatedAPI.Post("/conversations/:id/takeover", handoffHandler.Takeover)
	gatedAPI.Post("/conversations/:id/return-to-ai", handoffHandler.ReturnToAI)

	// AI Agent Templates endpoints
	gatedAPI.Get("/agent/templates", templateHandler.List)
	gatedAPI.Post("/agent/templates/:id/apply", templateHandler.Apply)

	gatedAPI.Get("/agent", canonicalAgentHandler.Get)
	gatedAPI.Patch("/agent", canonicalAgentHandler.Update)
	gatedAPI.Get("/agent/profile", canonicalAgentHandler.GetProfile)
	gatedAPI.Patch("/agent/profile", canonicalAgentHandler.UpdateProfile)
	gatedAPI.Get("/agent/personality", canonicalAgentHandler.GetPersonality)
	gatedAPI.Patch("/agent/personality", canonicalAgentHandler.UpdatePersonality)
	gatedAPI.Post("/agent/test", canonicalAgentHandler.Test)
	gatedAPI.Post("/agent/generate-setup", canonicalAgentHandler.GenerateSetup)
	gatedAPI.Get("/business", canonicalBusinessHandler.Get)
	gatedAPI.Patch("/business", canonicalBusinessHandler.Update)
	gatedAPI.Get("/channels", channelHandler.List)
	gatedAPI.Get("/channels/:id/profile", channelHandler.GetProfile)
	gatedAPI.Post("/channels/whatsapp/connect", channelHandler.ConnectWhatsApp)
	gatedAPI.Post("/channels/:id/restart", channelHandler.Restart)
	gatedAPI.Patch("/channels/:id/status", channelHandler.ToggleStatus)
	gatedAPI.Delete("/channels/:id", channelHandler.Delete)
	gatedAPI.Get("/api-keys", apiKeyHandler.List)
	gatedAPI.Post("/api-keys", apiKeyHandler.Create)
	gatedAPI.Delete("/api-keys/:id", apiKeyHandler.Delete)
	gatedAPI.Get("/webhooks", webhookHandler.List)
	gatedAPI.Post("/webhooks", webhookHandler.Create)
	gatedAPI.Patch("/webhooks/:id", webhookHandler.Update)
	gatedAPI.Delete("/webhooks/:id", webhookHandler.Delete)
	gatedAPI.Get("/conversations", conversationDashboardHandler.List)
	gatedAPI.Get("/conversations/:id", conversationDashboardHandler.Get)
	gatedAPI.Get("/conversations/:id/messages", conversationDashboardHandler.Messages)
	gatedAPI.Patch("/conversations/:id/mode", conversationDashboardHandler.SetMode)

	// Conversation attachments contain customer data and must never be served to
	// anonymous visitors. Filenames are also randomized by the WhatsApp service.
	app.Use("/api/v1/media", auth.RequireAccessToken(jwtManager))
	app.Static("/api/v1/media", "/tmp/chatsolv-media", fiber.Static{
		Compress:  true,
		ByteRange: true,
		Browse:    false,
	})

	internalAPI := app.Group("/internal/v1", mw.InternalHMAC(cfg.InternalServiceSecret, 5*time.Minute, time.Now))
	internalAPI.Post("/channels/events", internalHandler.ChannelEvent)
	internalAPI.Post("/channels/status", internalHandler.ChannelStatus)
	internalAPI.Post("/messages/incoming", internalHandler.Incoming)
	internalAPI.Post("/agents/:agentID/respond", internalHandler.Respond)
	internalAPI.Get("/agents/:agentID/health", internalHandler.Health)

	w := app.Group("/api/v1/workspaces", auth.RequireAccessToken(jwtManager))
	w.Post("/", workspaceHandler.Create)
	w.Get("/:workspaceID", workspaceHandler.Get)
	w.Patch("/:workspaceID", workspaceHandler.Update)
	w.Get("/:workspaceID/subscription", workspaceHandler.Subscription)
	w.Post("/:workspaceID/checkout", workspaceHandler.CreateCheckout)

	agents := app.Group("/api/v1/agents", auth.RequireAccessToken(jwtManager), resolver.RequireActiveSubscription())
	agents.Get("/:agentID/personality", agentConfigHandler.GetPersonality)
	agents.Patch("/:agentID/personality", agentConfigHandler.UpdatePersonality)
	agents.Get("/:agentID/profile", settingsHandler.GetAgentProfile)
	agents.Patch("/:agentID/profile", settingsHandler.UpdateAgentProfile)

	settings := app.Group("/api/v1/settings/workspaces", auth.RequireAccessToken(jwtManager), resolver.RequireActiveSubscription())
	settings.Get("/:workspaceID/business", settingsHandler.GetBusiness)
	settings.Patch("/:workspaceID/business", settingsHandler.UpdateBusiness)
	settings.Get("/:workspaceID/policies", settingsHandler.GetPolicies)
	settings.Patch("/:workspaceID/policies", settingsHandler.UpdatePolicies)

	knowledgeAPI := app.Group("/api/v1/knowledge", auth.RequireAccessToken(jwtManager), resolver.RequireActiveSubscription())
	knowledgeAPI.Get("/", knowledgeHandler.List)
	knowledgeAPI.Get("/:id", knowledgeHandler.Get)
	knowledgeAPI.Patch("/:id", knowledgeHandler.Update)
	knowledgeAPI.Delete("/:id", knowledgeHandler.Delete)
	knowledgeAPI.Post("/:id/retry", knowledgeHandler.Retry)
	knowledgeAPI.Post("/documents", mw.RateLimit(rd, cfg.RateLimit, cfg.RateWindow), knowledgeHandler.UploadDocument)
	knowledgeAPI.Post("/text", mw.RateLimit(rd, cfg.RateLimit, cfg.RateWindow), knowledgeHandler.CreateText)
	knowledgeAPI.Post("/faqs", mw.RateLimit(rd, cfg.RateLimit, cfg.RateWindow), knowledgeHandler.CreateFAQ)

	errCh := make(chan error, 1)
	go func() { errCh <- app.Listen(":" + cfg.Port) }()
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	select {
	case e = <-errCh:
		if e != nil {
			log.Error("server stopped", "error", e)
		}
	case <-sig:
		log.Info("shutdown requested")
	}
	if e = app.ShutdownWithTimeout(cfg.ShutdownTimeout); e != nil {
		log.Error("shutdown failed", "error", e)
	}
}

func join(s []string) string {
	r := ""
	for i, v := range s {
		if i > 0 {
			r += ","
		}
		r += v
	}
	return r
}
