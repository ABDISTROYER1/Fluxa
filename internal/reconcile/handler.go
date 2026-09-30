package reconcile

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/fluxa/fluxa/internal/api"
	"github.com/fluxa/fluxa/internal/domain"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type Handler struct {
	svc   *Service
	audit interface {
		Log(r *http.Request, action, resourceType, resourceID string, metadata map[string]interface{})
	}
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) WithAuditLogger(audit interface {
	Log(r *http.Request, action, resourceType, resourceID string, metadata map[string]interface{})
}) *Handler {
	h.audit = audit
	return h
}

func (h *Handler) AdminRoutes() func(r chi.Router) {
	return func(r chi.Router) {
		r.Get("/reconciliation/summary", h.summary)
		r.Get("/reconciliation/drift", h.drift)
		r.Post("/reconciliation/run", h.run)
		r.Post("/transfers/{transferID}/force-settle", h.forceSettle)
		r.Post("/reconcile/wallet/{walletID}/run", h.runReconcile)
	}
}

func (h *Handler) summary(w http.ResponseWriter, r *http.Request) {
	daysStr := r.URL.Query().Get("days")
	days := 7
	if daysStr != "" {
		if d, err := strconv.Atoi(daysStr); err == nil && d > 0 && d <= 90 {
			days = d
		}
	}

	summary, err := h.svc.GetSummary(r.Context(), days)
	if err != nil {
		api.InternalError(w, err)
		return
	}

	api.JSON(w, http.StatusOK, summary)
}

func (h *Handler) drift(w http.ResponseWriter, r *http.Request) {
	snapshots, err := h.svc.GetDrift(r.Context())
	if err != nil {
		api.InternalError(w, err)
		return
	}
	// Always emit a JSON array, never null, so clients can iterate the result
	// without a nil check when there is no drift.
	if snapshots == nil {
		snapshots = []*DriftSnapshot{}
	}
	api.JSON(w, http.StatusOK, map[string]interface{}{
		"drift":   snapshots,
		"count":   len(snapshots),
		"checked": time.Now().UTC(),
	})
}

func (h *Handler) run(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.RunAll(r.Context()); err != nil {
		api.InternalError(w, err)
		return
	}
	api.JSON(w, http.StatusAccepted, map[string]interface{}{"status": "triggered"})
}

func (h *Handler) forceSettle(w http.ResponseWriter, r *http.Request) {
	transferID := chi.URLParam(r, "transferID")
	if transferID == "" {
		api.Error(w, http.StatusBadRequest, "TRANSFER_ID_REQUIRED", "transferID is required")
		return
	}
	if _, err := uuid.Parse(transferID); err != nil {
		api.Error(w, http.StatusBadRequest, "TRANSFER_ID_INVALID", "transferID must be a valid UUID")
		return
	}

	actor := api.ActorFromContext(r.Context())
	if err := h.svc.EnqueueForceSettle(r.Context(), transferID, actor); err != nil {
		if errors.Is(err, domain.ErrConcurrentUpdate) {
			api.Error(w, http.StatusConflict, "CONFLICT", "transfer is not in a retryable state")
			return
		}
		api.WriteError(w, r, err)
		return
	}
	if h.audit != nil {
		h.audit.Log(r, "transfer.retry", "transfer", transferID, map[string]interface{}{"operation": "retry_failed_settlement"})
	}

	api.JSON(w, http.StatusAccepted, map[string]interface{}{"status": "retry_queued", "transfer_id": transferID})
}

func (h *Handler) runReconcile(w http.ResponseWriter, r *http.Request) {
	walletID := chi.URLParam(r, "walletID")
	if walletID == "" {
		api.Error(w, http.StatusBadRequest, "WALLET_ID_REQUIRED", "walletID is required")
		return
	}

	actor := api.ActorFromContext(r.Context())
	if err := h.svc.EnqueueWalletReconcile(r.Context(), walletID, actor); err != nil {
		api.InternalError(w, err)
		return
	}

	api.JSON(w, http.StatusOK, map[string]interface{}{"status": "enqueued", "wallet_id": walletID})
}
