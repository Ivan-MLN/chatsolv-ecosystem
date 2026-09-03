package main

import (
	"authbackend/internal/agentconfig"
	"authbackend/internal/brain/obsidian"
	"authbackend/internal/config"
	"authbackend/internal/database"
	"authbackend/internal/hermes"
	"authbackend/internal/jobs"
	"authbackend/internal/knowledge"
	"authbackend/internal/provisioning"
	"authbackend/internal/storage"
	"context"
	"github.com/joho/godotenv"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	_ = godotenv.Load()
	cfg, err := config.Load()
	if err != nil {
		slog.Error("configuration error", "error", err)
		os.Exit(1)
	}
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	connectCtx, connectCancel := context.WithTimeout(ctx, 10*time.Second)
	defer connectCancel()
	pool, err := database.Postgres(connectCtx, cfg)
	if err != nil {
		log.Error("postgres connection failed", "error", err)
		os.Exit(1)
	}
	defer pool.Close()
	brain := obsidian.NewFilesystem(cfg.VaultRoot)
	provider := hermes.NewCLIProvider(cfg.HermesBinary, cfg.HermesRoot, cfg.HermesTemplateProfile, nil)
	objects, err := storage.NewMinIO(cfg.ObjectStorageEndpoint, cfg.ObjectStorageAccessKey, cfg.ObjectStorageSecretKey, cfg.ObjectStorageBucket, cfg.ObjectStorageUseSSL)
	if err != nil {
		log.Error("object storage configuration failed", "error", err)
		os.Exit(1)
	}
	if err = objects.EnsureBucket(connectCtx); err != nil {
		log.Error("object storage readiness failed", "error", err)
		os.Exit(1)
	}
	provisioner := provisioning.NewService(provisioning.NewPostgresRepository(pool), brain, provider)
	syncer := agentconfig.NewSyncService(pool, brain, provider)
	ingestion := knowledge.NewIngestion(pool, objects, brain)
	worker := jobs.NewWorker(jobs.NewPostgresQueue(pool), map[string]jobs.Handler{"workspace.provision": provisioner.Provision, "agent.sync": syncer.Sync, "knowledge.ingest": ingestion.Ingest, "knowledge.delete": ingestion.Delete}, cfg.JobMaxAttempts, time.Now)
	log.Info("worker started", "poll_interval", cfg.JobPollInterval)
	if err = worker.Run(ctx, cfg.JobPollInterval); err != nil && err != context.Canceled {
		log.Error("worker stopped", "error", err)
		os.Exit(1)
	}
	log.Info("worker stopped")
}
