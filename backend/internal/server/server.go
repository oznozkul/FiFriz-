package server

import (
	"context"
	"log"
	"strings"

	"fifriz/internal/ai"
	"fifriz/internal/ai/pipeline"
	"fifriz/internal/ai/tools"
	"fifriz/internal/approvals"
	"fifriz/internal/audit"
	"fifriz/internal/auth"
	"fifriz/internal/cache"
	"fifriz/internal/config"
	"fifriz/internal/dashboard"
	"fifriz/internal/documents"
	docvalidation "fifriz/internal/documents/validation"
	"fifriz/internal/documents/virusscan"
	"fifriz/internal/email"
	"fifriz/internal/embedding"
	"fifriz/internal/extensions"
	"fifriz/internal/folders"
	"fifriz/internal/health"
	"fifriz/internal/integrations"
	"fifriz/internal/integrations/plugins"
	"fifriz/internal/integrations/registry"
	"fifriz/internal/jobs"
	"fifriz/internal/listeners"
	"fifriz/internal/navigation"
	"fifriz/internal/notifications"
	"fifriz/internal/response"
	"fifriz/internal/search"
	"fifriz/internal/sharing"
	"fifriz/internal/smartcollections"
	"fifriz/internal/storage"
	"fifriz/internal/tags"
	"fifriz/internal/worker"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"gorm.io/gorm"
)

func New(cfg *config.Config, db *gorm.DB) (*fiber.App, *worker.Worker, *integrations.Service) {
	app := fiber.New(fiber.Config{
		AppName:      "FiFriz API",
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
		IdleTimeout:  cfg.HTTP.IdleTimeout,
		ReadBufferSize: cfg.HTTP.ReadBufferSize,
		ErrorHandler: errorHandler,
	})

	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: strings.Join(cfg.CORS.AllowedOrigins, ","),
		AllowMethods: "GET,POST,PUT,PATCH,DELETE,OPTIONS",
		AllowHeaders: "Origin,Content-Type,Accept,Authorization,X-Share-Access-Token",
	}))

	health.RegisterRoutes(app, db)

	api := app.Group("/api")
	v1 := api.Group("/v1")

	authRepository := auth.NewRepository(db)
	auditRepository := audit.NewRepository(db)
	auditService := audit.NewService(auditRepository)
	cacheService := cache.NewService(cache.NewRedisClient(cfg.Redis.URL))
	notificationRepository := notifications.NewRepository(db)
	notificationService := notifications.NewService(notificationRepository)
	jobQueue := jobs.NewQueue(cfg.Redis.URL, cfg.Redis.QueueKey)
	extensionRegistry := extensions.NewRegistry()
	eventRuntime := listeners.Bootstrap(listeners.Dependencies{
		Audit:         auditService,
		Notifications: notificationService,
		Cache:         cacheService,
		Queue:         jobQueue,
		Extensions:    extensionRegistry,
	})
	publisher := eventRuntime.Publisher

	authService := auth.NewService(authRepository, auth.JWTConfig{
		Secret:   cfg.JWT.Secret,
		Issuer:   cfg.JWT.Issuer,
		Duration: cfg.JWT.Duration,
	}, publisher)
	authService.SetPasswordResetConfig(auth.PasswordResetConfig{
		FrontendURL: cfg.PasswordReset.FrontendURL,
		TokenTTL:    cfg.PasswordReset.TokenTTL,
		AppEnv:      cfg.App.Env,
	})
	authService.SetEmailSender(email.NewConsoleSender(cfg.App.Env != "production"))
	authHandler := auth.NewHandler(authService)
	auth.RegisterRoutes(v1, authHandler)

	documentRepository := documents.NewRepository(db)
	tagRepository := tags.NewRepository(db)
	tagService := tags.NewService(tagRepository, publisher, cacheService)
	tagHandler := tags.NewHandler(tagService)
	embeddingProvider := embedding.NewProvider(embedding.Config{
		Provider: cfg.Embedding.Provider,
		Ollama: embedding.OllamaConfig{
			BaseURL: cfg.Embedding.Ollama.BaseURL,
			Model:   cfg.Embedding.Ollama.Model,
		},
	})
	if embeddingProvider == nil {
		log.Printf(
			"embedding provider disabled: configured_provider=%q ollama_base_url=%s ollama_model=%s",
			cfg.Embedding.Provider,
			cfg.Embedding.Ollama.BaseURL,
			cfg.Embedding.Ollama.Model,
		)
	} else {
		log.Printf(
			"embedding provider ready: provider=%s model=%s base_url=%s",
			embeddingProvider.ProviderName(),
			embeddingProvider.ModelName(),
			cfg.Embedding.Ollama.BaseURL,
		)
	}

	chatProvider := ai.NewChatProvider(cfg.AI)
	if chatProvider == nil {
		log.Printf(
			"ai chat provider disabled: configured_provider=%q ollama_chat_model=%s",
			cfg.AI.Provider,
			cfg.AI.Ollama.Model,
		)
	} else {
		log.Printf("ai chat provider ready: provider=%s", chatProvider.Name())
	}

	sharingRepository := sharing.NewRepository(db)
	shareAccess := documents.NewShareAccessService(documentRepository)
	storageManager, err := storage.ManagerFromConfig(cfg.Storage)
	if err != nil {
		log.Fatalf("storage manager init failed: %v", err)
	}
	log.Printf("storage provider ready: provider=%s bucket=%s", storageManager.DefaultProviderName(), storageManager.BucketName())
	documentService := documents.NewService(documentRepository, storageManager, embeddingProvider, publisher, tagService, shareAccess, jobQueue, cacheService)
	documentService.SetSummaryChatProvider(NewSummaryChatAdapter(chatProvider), cfg.AI.Ollama.Model)
	documentService.SetAuditRecorder(auditService)
	if policies, err := documentRepository.ListTenantDocumentPolicies(context.Background()); err == nil {
		overrides := make([]docvalidation.TenantPolicyOverride, 0, len(policies))
		for _, policy := range policies {
			overrides = append(overrides, docvalidation.TenantPolicyOverride{
				TenantID:         policy.TenantID,
				MaxFileSizeBytes: policy.MaxFileSizeBytes,
				AllowedMIMETypes: documents.ParseTenantPolicyMIMEList(policy.AllowedMimeTypes),
			})
		}
		documentService.SetEnterpriseConfig(documents.EnterpriseConfig{
			PolicyResolver: docvalidation.NewPolicyResolver(docvalidation.Policy{
				MaxFileSizeBytes: cfg.Documents.MaxFileSizeBytes,
				AllowedMIMETypes: docvalidation.ParseMIMEAllowlist(cfg.Documents.AllowedMIMETypes),
			}, overrides),
			VirusScanner:       virusscan.NoOpScanner{},
			DownloadMaxRetries: cfg.Documents.StorageDownloadRetries,
			DownloadTimeout:    cfg.Documents.StorageDownloadTimeout,
		})
	}

	pipelineRegistry := pipeline.NewRegistry()
	pipelineExecutor := documents.NewPipelineExecutor(documentService)
	if err := pipeline.RegisterDefaultSteps(pipelineRegistry, pipelineExecutor); err != nil {
		log.Fatalf("register ai pipeline steps: %v", err)
	}
	pipelineJobRepository := pipeline.NewJobRepository(db)
	pipelineRunner := pipeline.NewRunner(pipeline.RunnerConfig{
		Registry: pipelineRegistry,
		Tracker:  pipelineJobRepository,
		Enqueuer: pipeline.NewQueueEnqueuer(jobQueue),
		Events:   pipeline.NewEventPublisher(publisher),
		Retry:    pipeline.DefaultRetryPolicy(),
	})
	documentService.SetPipelineRunner(pipelineRunner)
	log.Printf("ai processing pipeline ready: steps=%d", len(pipelineRegistry.StepNames()))

	documentHandler := documents.NewHandler(documentService)
	integrationRepository := integrations.NewRepository(db)
	integrationImporter := integrations.NewImportService(integrationRepository, documentRepository, storageManager, jobQueue)
	integrationImporter.SetDocumentScheduler(documentService)
	googleOAuth := integrations.GoogleOAuthConfig{
		ClientID:     cfg.Google.ClientID,
		ClientSecret: cfg.Google.ClientSecret,
		RedirectURL:  cfg.Google.RedirectURL,
		Scopes:       cfg.Google.Scopes,
	}
	googleDriveOAuth := integrations.GoogleDriveOAuthConfig{
		ClientID:     cfg.GoogleDrive.ClientID,
		ClientSecret: cfg.GoogleDrive.ClientSecret,
		RedirectURL:  cfg.GoogleDrive.RedirectURL,
		Scopes:       cfg.GoogleDrive.Scopes,
	}
	microsoftOAuth := integrations.MicrosoftOAuthConfig{
		ClientID:     cfg.Microsoft.ClientID,
		ClientSecret: cfg.Microsoft.ClientSecret,
		Tenant:       cfg.Microsoft.Tenant,
		RedirectURL:  cfg.Microsoft.RedirectURL,
		Scopes:       cfg.Microsoft.Scopes,
	}
	oneDriveOAuth := integrations.OneDriveOAuthConfig{
		ClientID:     cfg.MicrosoftOneDrive.ClientID,
		ClientSecret: cfg.MicrosoftOneDrive.ClientSecret,
		Tenant:       cfg.MicrosoftOneDrive.Tenant,
		RedirectURL:  cfg.MicrosoftOneDrive.RedirectURL,
		Scopes:       cfg.MicrosoftOneDrive.Scopes,
	}
	providerRegistry := registry.New()
	integrationCredentials := integrations.NewCredentialService()
	if err := plugins.RegisterAll(providerRegistry, plugins.Dependencies{
		Repository:  integrationRepository,
		Importer:    integrationImporter,
		Credentials: integrationCredentials,
		Google:      googleOAuth,
		GoogleDrive: googleDriveOAuth,
		Microsoft:   microsoftOAuth,
		OneDrive:    oneDriveOAuth,
		StateSecret: cfg.JWT.Secret,
	}); err != nil {
		log.Fatalf("register integration providers: %v", err)
	}
	integrationService := integrations.NewService(
		integrationRepository,
		integrationImporter,
		providerRegistry,
		integrationCredentials,
		publisher,
		notificationService,
		jobQueue,
		cfg.JWT.Secret,
		cfg.Public.AppURL,
		cfg.Integrations.SyncInterval,
	)
	integrationHandler := integrations.NewHandler(integrationService)
	documentWorker := worker.New(jobQueue, documentService, integrationService, documentService, documentService)
	documentService.RecoverStuckProcessing(context.Background())
	approvalRepository := approvals.NewRepository(db)
	approvalService := approvals.NewService(approvalRepository, documentRepository, authRepository, shareAccess, publisher)
	approvalHandler := approvals.NewHandler(approvalService)
	sharingService := sharing.NewService(
		sharingRepository,
		documentRepository,
		authRepository,
		shareAccess,
		storageManager,
		cfg.Public.AppURL,
		cfg.Public.APIURL,
		cfg.JWT.Secret,
		publisher,
		notificationService,
	)
	sharingService.SetAuditRecorder(auditService)
	sharingService.SetDownloadConfig(cfg.Documents.StorageDownloadRetries, cfg.Documents.StorageDownloadTimeout)
	sharingHandler := sharing.NewHandler(sharingService)
	protectedV1 := v1.Group("", authHandler.Middleware())
	protectedV1.Use(requestMetaMiddleware())
	auth.RegisterTenantRoutes(protectedV1, authHandler)
	auth.RegisterProfileRoutes(protectedV1, authHandler)
	sharing.RegisterProtectedRoutes(protectedV1, sharingHandler)
	sharing.RegisterPublicRoutes(v1, sharingHandler)
	integrations.RegisterPublicRoutes(v1, integrationHandler)
	tags.RegisterRoutes(protectedV1, tagHandler)
	approvals.RegisterRoutes(protectedV1, approvalHandler)
	integrations.RegisterRoutes(protectedV1, integrationHandler)
	documents.RegisterRoutes(protectedV1, documentHandler)
	documents.RegisterImportantDateRoutes(protectedV1, documentHandler)
	documents.RegisterRelationshipRoutes(protectedV1, documentHandler)
	folderRepository := folders.NewRepository(db)
	folderService := folders.NewService(folderRepository, publisher, authRepository, cacheService)
	authService.SetPersonalTenantBootstrapper(folderService)
	folderHandler := folders.NewHandler(folderService, documentService)
	folders.RegisterRoutes(protectedV1, folderHandler)
	auditHandler := audit.NewHandler(auditService)
	audit.RegisterRoutes(protectedV1, auditHandler)
	dashboardRepository := dashboard.NewRepository(db)
	dashboardService := dashboard.NewService(dashboardRepository, cacheService)
	dashboardHandler := dashboard.NewHandler(dashboardService)
	dashboard.RegisterRoutes(protectedV1, dashboardHandler)
	notificationHandler := notifications.NewHandler(notificationService)
	notifications.RegisterRoutes(protectedV1, notificationHandler)
	smartCollectionRepository := smartcollections.NewRepository(db)
	smartCollectionService := smartcollections.NewService(smartCollectionRepository, documentService, publisher)
	smartCollectionHandler := smartcollections.NewHandler(smartCollectionService)
	smartcollections.RegisterRoutes(protectedV1, smartCollectionHandler)

	ragService := ai.NewRAGService(ai.RAGServiceConfig{
		Documents:   documentRepository,
		Collections: ai.NewSmartCollectionAdapter(smartCollectionRepository),
		Embeddings:  embeddingProvider,
		Chat:        chatProvider,
		ShareAccess: shareAccess,
	})
	aiSessionRepository := ai.NewSessionRepository(db)
	aiService := ai.NewService(ai.ServiceConfig{
		RAG:      ragService,
		Sessions: aiSessionRepository,
		Audit:    auditService,
		Folders:  folderRepository,
		Tags:     tagRepository,
	})
	workspaceInsightsRepository := ai.NewWorkspaceInsightsRepository(db)
	workspaceInsightsService := ai.NewWorkspaceInsightsService(workspaceInsightsRepository, approvalRepository, cacheService)
	toolRegistry := tools.NewRegistry()
	tools.RegisterBuiltInTools(toolRegistry, tools.Dependencies{
		Documents:      documentService,
		DocumentReader: documentRepository,
		Tags:           tagService,
		Approvals:      approvalService,
		Access:         shareAccess,
	})
	toolExecutor := tools.NewExecutor(toolRegistry, auditService)
	actionSuggestionRepository := ai.NewActionSuggestionRepository(db)
	actionSuggestionService := ai.NewActionSuggestionService(
		actionSuggestionRepository,
		documentRepository,
		folderRepository,
		tagRepository,
		approvalRepository,
		shareAccess,
		auditService,
		toolExecutor,
	)
	agentChatService := ai.NewAgentChatService(ai.AgentChatServiceConfig{
		RAG:          ragService,
		Sessions:     aiSessionRepository,
		Chat:         chatProvider,
		ToolRegistry: toolRegistry,
		Suggestions:  actionSuggestionRepository,
		Documents:    documentRepository,
		Folders:      folderRepository,
		Tags:         tagRepository,
		Approvals:    approvalRepository,
		ShareAccess:  shareAccess,
		Audit:        auditService,
	})
	aiHandler := ai.NewHandler(aiService, workspaceInsightsService, actionSuggestionService, agentChatService)
	ai.RegisterRoutes(protectedV1, aiHandler)
	ai.RegisterDocumentRoutes(protectedV1, aiHandler)

	searchRepository := search.NewRepository(db)
	searchService := search.NewService(
		searchRepository,
		documentService,
		folderRepository,
		tagRepository,
		authRepository,
		integrationRepository,
		sharingService,
		cacheService,
	)
	searchHandler := search.NewHandler(searchService)
	search.RegisterRoutes(protectedV1, searchHandler)

	navigationRepository := navigation.NewRepository(db)
	recentStore := navigation.NewRecentStore(cacheService)
	listeners.BootstrapNavigation(eventRuntime.Bus, eventRuntime.Invalidator, recentStore)
	navigationService := navigation.NewService(navigationRepository, recentStore, cacheService, eventRuntime.Invalidator, publisher)
	navigationHandler := navigation.NewHandler(navigationService)
	navigation.RegisterRoutes(protectedV1, navigationHandler)

	storageMigrationService := documents.NewStorageMigrationService(storageManager, documentRepository, auditService, cfg.App.Env)
	documents.RegisterStorageAdminRoutes(protectedV1, documents.NewStorageMigrationHandler(storageMigrationService))

	return app, documentWorker, integrationService
}

func requestMetaMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.SetUserContext(audit.WithRequestMeta(c.UserContext(), audit.RequestMeta{
			IPAddress: c.IP(),
			UserAgent: c.Get("User-Agent"),
		}))
		return c.Next()
	}
}

func errorHandler(c *fiber.Ctx, err error) error {
	status := fiber.StatusInternalServerError
	code := "internal_error"

	if fiberErr, ok := err.(*fiber.Error); ok {
		status = fiberErr.Code
		code = strings.ToLower(strings.ReplaceAll(fiberErr.Message, " ", "_"))
	}

	return c.Status(status).JSON(response.Fail(code, err.Error()))
}
