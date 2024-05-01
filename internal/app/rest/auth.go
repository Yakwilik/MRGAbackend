package rest

import (
	"github.com/Yakwilik/MRGAbackend/internal/logger"
	"github.com/Yakwilik/MRGAbackend/internal/pkg/helper"
	"net/http"
	"time"
)

func (receiver *Handler) logout(w http.ResponseWriter, r *http.Request) {
	sessionID, err := helper.GetSessionIDFromContext(r.Context())
	if err == nil {
		logger.Info(r.Context(), "Logout_GetSessionIDFromContext", "err", err)
		defer func() {
			err := receiver.useCase.DeleteSession(r.Context(), sessionID)
			if err != nil {
				logger.Info(r.Context(), "Logout_DeleteSession", "err", err)
			}
		}()
	}
	c := helper.GetSessionCookie(receiver.cfg.Domain(r.Context()), sessionID, time.Date(1980, time.January, 1, 0, 0, 0, 0, time.UTC))

	w.Header().Set("Set-Cookie", c.String())
	w.WriteHeader(http.StatusOK)
}
