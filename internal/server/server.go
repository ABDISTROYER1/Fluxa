package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/fluxa/fluxa/internal/anchor"
	"github.com/fluxa/fluxa/internal/apikey"
	"github.com/fluxa/fluxa/internal/audit"
	"github.com/fluxa/fluxa/internal/auth"
	"github.com/fluxa/fluxa/internal/batch"
	"github.com/fluxa/fluxa/internal/beneficiary"
	"github.com/fluxa/fluxa/internal/claimable"
	"github.com/fluxa/fluxa/internal/compliance"
	"github.com/fluxa/fluxa/internal/domain"
	"github.com/fluxa/fluxa/internal/fees"
	"github.com/fluxa/fluxa/internal/fiat"
	"github.com/fluxa/fluxa/internal/fx"
	fluxahealth "github.com/fluxa/fluxa/internal/health"
	"github.com/fluxa/fluxa/internal/org"
	"github.com/fluxa/fluxa/internal/paymentlink"
	"github.com/fluxa/fluxa/internal/postgres"
	"github.com/fluxa/fluxa/internal/reconcile"
	"github.com/fluxa/fluxa/internal/refund"
	"github.com/fluxa/fluxa/internal/schedule"
	"github.com/fluxa/fluxa/internal/server/idempotency"
	"github.com/fluxa/fluxa/internal/status"
	"github.com/fluxa/fluxa/internal/transfer"
	"github.com/fluxa/fluxa/internal/treasury"
	"github.com/fluxa/fluxa/internal/wallet"
	"github.com/fluxa/fluxa/internal/webhook"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Server struct {
	router *chi.Mux
	http   *http.Server
}

func New(
	authHandler *auth.Handler,
	orgHandler *org.Handler,
	walletHandler *wallet.Handler,
	transferHandler *transfer.Handler,
	fxHandler *fx.Handler,
	fiatHandler *fiat.Handler,
	anchorFiatHandler *fiat.AnchorHandler,
	anchorHandler *anchor.Handler,
	feeHandler *fees.Handler,
	reconcileHandler *reconcile.Handler,
	apikeyHandler *apikey.Handler,
	apiKeyRepo *postgres.APIKeyRepo,
	webhookHandler *webhook.Handler,
	batchHandler *batch.Handler,
	scheduleHandler *schedule.Handler,
	treasuryHandler *treasury.Handler,
	claimableHandler *claimable.Handler,
	statusHandler *status.Handler,
	complianceHandler *compliance.Handler,
	auditHandler *audit.Handler,
	usageHandler *UsageHandler,
	idempotencyHandler *idempotency.Handler,
	jwtSecret []byte,
	port string,
	healthChecks map[string]DependencyCheck,
	membershipValidator MembershipValidator,
	corsOrigins []string,
	options ...interface{},
) *Server {
	r := chi.NewRouter()

	rateCfg := DefaultAuthRateLimitConfig()
	var beneficiaryHandler *beneficiary.Handler
	var paymentLinkHandler *paymentlink.Handler
	var refundHandler *refund.Handler
	for _, option := range options {
		switch value := option.(type) {
		case AuthRateLimitConfig:
			rateCfg = value
		case *beneficiary.Handler:
			beneficiaryHandler = value
		case *paymentlink.Handler:
			paymentLinkHandler = value
		case *refund.Handler:
			refundHandler = value
		}
	}
	authLimiter := NewAuthRateLimiter(rateCfg)

	r.Use(middleware.RealIP)
	r.Use(requestID)
	r.Use(logger)
	r.Use(recoverer)
	r.Use(CORS(corsOrigins))
	r.Use(MaxBodySize(1 << 20))
	r.Use(MetricsMiddleware)

	componentProbes := make(map[string]fluxahealth.Probe, len(healthChecks))
	for name, check := range healthChecks {
		probe := check
		componentProbes[name] = func(ctx context.Context) (interface{}, error) { return nil, probe(ctx) }
	}
	healthService := fluxahealth.New(componentProbes)
	r.Get("/health", healthService.Handler())
	r.Get("/health/ready", healthService.ReadyHandler())
	r.Get("/health/live", fluxahealth.LiveHandler())
	r.Get("/metrics", MetricsHandler)

	// Platform status is intentionally public and must remain outside tenant
	// authentication.
	if statusHandler != nil {
		statusHandler.RegisterRoutes(r)
	}

	r.Route("/v1", func(r chi.Router) {
		// Unauthenticated public endpoints
		if paymentLinkHandler != nil {
			r.Route("/public/payment-links", paymentLinkHandler.PublicRoutes())
		}
		r.Route("/auth", func(r chi.Router) {
			r.With(authLimiter.Limit(ExtractEmail)).Post("/register", authHandler.Register)
			r.With(authLimiter.Limit(ExtractEmail)).Post("/login", authHandler.Login)
			r.Post("/refresh", authHandler.Refresh)
		})
		r.With(authLimiter.Limit(ExtractInviteToken)).Post("/org/invites/accept", orgHandler.AcceptInvite)
		// Registered as a direct path because the authenticated group mounts /webhooks
		r.With(webhook.VerifyRateLimit()).Post("/webhooks/verify", webhookHandler.VerifySignature)

		// Authenticated endpoints
		r.Group(func(r chi.Router) {
			r.Use(AuthMiddleware(apiKeyRepo, jwtSecret, membershipValidator))
			r.Use(RateLimit(100, 200))

			// API Keys (Owner & Admin only for creation, expiry update, rotation & revocation)
			r.Route("/keys", func(r chi.Router) {
				r.With(RequireRole(domain.RoleOwner, domain.RoleAdmin), RequireScope(domain.ScopeKeysWrite)).Post("/", apikeyHandler.Create)
				r.With(RequireScope(domain.ScopeKeysRead)).Get("/", apikeyHandler.List)
				r.With(RequireRole(domain.RoleOwner, domain.RoleAdmin), RequireScope(domain.ScopeKeysWrite)).Delete("/{id}", apikeyHandler.Revoke)
				r.With(RequireRole(domain.RoleOwner, domain.RoleAdmin), RequireScope(domain.ScopeKeysWrite)).Patch("/{id}/expiry", apikeyHandler.UpdateExpiry)
				r.With(RequireRole(domain.RoleOwner, domain.RoleAdmin), RequireScope(domain.ScopeKeysWrite)).Post("/{id}/rotate", apikeyHandler.Rotate)
			})

			// Audit Log (Tenant-visible append-only audit log)
			if auditHandler != nil {
				r.Route("/audit", func(r chi.Router) {
					r.Use(RequireScope(domain.ScopeAuditRead))
					r.Get("/", auditHandler.List)
					r.Get("/export", auditHandler.Export)
				})
			}

			// Usage Introspection
			if usageHandler != nil {
				r.Get("/usage", usageHandler.GetUsage)
			}

			// Idempotency Key Inspection
			if idempotencyHandler != nil {
				r.Route("/idempotency", idempotencyHandler.Routes())
			}

			// Org Member Management (Owner & Admin for invite, role update, remove)
			r.Route("/org", func(r chi.Router) {
				r.With(RequireRole(domain.RoleOwner, domain.RoleAdmin)).Post("/members/invite", orgHandler.InviteMember)
				r.Get("/members", orgHandler.ListMembers)
				r.With(RequireRole(domain.RoleOwner, domain.RoleAdmin)).Patch("/members/{userId}", orgHandler.UpdateRole)
				r.With(RequireRole(domain.RoleOwner, domain.RoleAdmin)).Delete("/members/{userId}", orgHandler.RemoveMember)
			})

			// Webhooks (Owner & Admin for management, viewer/dev read)
			r.Route("/webhooks", func(r chi.Router) {
				r.With(RequireRole(domain.RoleOwner, domain.RoleAdmin), RequireScope(domain.ScopeWebhooksWrite)).Post("/", webhookHandler.RegisterEndpoint)
				r.With(RequireScope(domain.ScopeWebhooksRead)).Get("/", webhookHandler.ListEndpoints)
				r.With(RequireRole(domain.RoleOwner, domain.RoleAdmin), RequireScope(domain.ScopeWebhooksWrite)).Delete("/{id}", webhookHandler.DeleteEndpoint)
				r.With(RequireScope(domain.ScopeWebhooksRead)).Get("/{id}/deliveries", webhookHandler.ListDeliveries)
			})

			// Sandbox-only escape hatch: lets a developer drive a webhook event
			// through the test environment without moving real funds.
			r.With(RequireTestMode).Post("/test/trigger-event", webhookHandler.TriggerTestEvent)

			// Operational routes (Require not viewer for mutating calls)
			r.Group(func(r chi.Router) {
				r.Use(RequireNotViewer)
				r.With(RequireScope(domain.ScopeWalletsRead)).Route("/wallets", walletHandler.Routes())
				if beneficiaryHandler != nil {
					r.With(RequireScope(domain.ScopeBeneficiariesRead)).Route("/beneficiaries", beneficiaryHandler.Routes())
				}
				r.Route("/wallets/{id}/deposit", fiatHandler.DepositRoutes())
				r.Route("/wallets/{id}/withdraw", fiatHandler.WithdrawRoutes())
				r.Route("/webhooks/fiat", fiatHandler.WebhookRoutes())
				r.With(RequireScope(domain.ScopeFiatRead)).Route("/fiat", anchorFiatHandler.Routes())
				if paymentLinkHandler != nil {
					r.Route("/payment-links", paymentLinkHandler.Routes(
						RequireScope(domain.ScopeFiatRead),
						RequireScope(domain.ScopeFiatWrite),
					))
				}
				if refundHandler != nil {
					r.Route("/refunds", refundHandler.Routes(
						RequireScope(domain.ScopeTransfersRead),
						RequireScope(domain.ScopeTransfersWrite),
					))
				}
				r.With(RequireScope(domain.ScopeTransfersRead)).Route("/transfers", transferHandler.Routes())
				r.With(RequireScope(domain.ScopeTransfersWrite)).Route("/transfers/batch", batchHandler.Routes())
				r.With(RequireScope(domain.ScopeTransfersRead)).Route("/transactions", transferHandler.TransactionRoutes())
				r.Route("/schedules", scheduleHandler.Routes(
					RequireScope(domain.ScopeTransfersRead),
					RequireScope(domain.ScopeTransfersWrite),
				))
				r.With(RequireScope(domain.ScopeFXRead)).Route("/fx", fxHandler.Routes())
				r.With(RequireScope(domain.ScopeFeesRead)).Route("/fees", feeHandler.Routes())
				if claimableHandler != nil {
					r.Route("/claimable-balances", claimableHandler.Routes())
				}
			})

			// Administrative routes (Owner & Admin only)
			r.Group(func(r chi.Router) {
				r.Use(RequireRole(domain.RoleOwner, domain.RoleAdmin))
				r.Route("/admin/fees", feeHandler.AdminRoutes())
				r.Route("/admin/anchors", anchorHandler.AdminRoutes())
				r.Route("/admin", reconcileHandler.AdminRoutes())
				r.With(RequirePlatformOperator()).Route("/admin/treasury", treasuryHandler.AdminRoutes())
				if statusHandler != nil {
					statusHandler.RegisterAdminRoutes(r)
				}
				// Mounted at /admin/compliance, not /admin
				if complianceHandler != nil {
					r.Route("/admin/compliance", complianceHandler.AdminRoutes())
				}
			})
		})
	})

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", port),
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return &Server{router: r, http: srv}
}

func (s *Server) Start() error {
	return s.http.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.http.Shutdown(ctx)
}

func (s *Server) Router() *chi.Mux {
	return s.router
}
