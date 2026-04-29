package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/tamcore/motus/internal/api"
	"github.com/tamcore/motus/internal/audit"
	"github.com/tamcore/motus/internal/demo"
	"github.com/tamcore/motus/internal/validation"
	"golang.org/x/crypto/bcrypt"
)

type updateProfileRequest struct {
	Name            string `json:"name"`
	Email           string `json:"email"`
	CurrentPassword string `json:"currentPassword"`
	NewPassword     string `json:"newPassword"`
}

// UpdateProfile allows the authenticated user to update their own name, email, and password.
// PUT /api/profile
func (h *UserHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	user := api.UserFromContext(r.Context())
	if user == nil {
		api.RespondError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	if demo.IsEnabled() && demo.IsDemoAccount(user.Email) {
		api.RespondError(w, http.StatusForbidden, "demo accounts cannot be modified")
		return
	}

	var req updateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.RespondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	existing, err := h.users.GetByID(r.Context(), user.ID)
	if err != nil {
		api.RespondError(w, http.StatusInternalServerError, "failed to fetch user")
		return
	}

	changes := map[string]interface{}{}

	if req.Email != "" && req.Email != existing.Email {
		if err := validation.ValidateEmail(req.Email); err != nil {
			api.RespondError(w, http.StatusBadRequest, err.Error())
			return
		}
		changes["oldEmail"] = existing.Email
		changes["newEmail"] = req.Email
		existing.Email = req.Email
	}
	if req.Name != "" && req.Name != existing.Name {
		if err := validation.ValidateName(req.Name); err != nil {
			api.RespondError(w, http.StatusBadRequest, err.Error())
			return
		}
		changes["oldName"] = existing.Name
		changes["newName"] = req.Name
		existing.Name = req.Name
	}

	if len(changes) > 0 {
		if err := h.users.Update(r.Context(), existing); err != nil {
			api.RespondError(w, http.StatusInternalServerError, "failed to update profile")
			return
		}
	}

	if req.NewPassword != "" {
		if req.CurrentPassword == "" {
			api.RespondError(w, http.StatusBadRequest, "current password is required to set a new password")
			return
		}
		if err := bcrypt.CompareHashAndPassword([]byte(existing.PasswordHash), []byte(req.CurrentPassword)); err != nil {
			api.RespondError(w, http.StatusBadRequest, "current password is incorrect")
			return
		}
		if err := validation.ValidatePassword(req.NewPassword); err != nil {
			api.RespondError(w, http.StatusBadRequest, err.Error())
			return
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
		if err != nil {
			api.RespondError(w, http.StatusInternalServerError, "failed to hash password")
			return
		}
		if err := h.users.UpdatePassword(r.Context(), existing.ID, string(hash)); err != nil {
			api.RespondError(w, http.StatusInternalServerError, "failed to update password")
			return
		}
		changes["passwordChanged"] = true
	}

	if h.audit != nil && len(changes) > 0 {
		h.audit.LogFromRequest(r, &existing.ID, audit.ActionUserUpdate, audit.ResourceUser, &existing.ID, changes)
	}

	existing.PopulateTraccarFields()
	api.RespondJSON(w, http.StatusOK, existing)
}
