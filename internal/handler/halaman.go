package handler

import (
	"net/http"

	"github.com/bilaaah/SeduhPoin/internal/middleware"
	"github.com/bilaaah/SeduhPoin/internal/repository"
)

type HalamanHandler struct {
	Repo *repository.AdminRepo
	View *Renderer
}

func (h *HalamanHandler) Home(w http.ResponseWriter, r *http.Request) {
	u, _ := middleware.UserDari(r.Context())
	ringkasan, err := h.Repo.Ringkasan(r.Context())
	if err != nil {
		ringkasan.User = u
		h.View.Tampil(w, "home.html", ringkasan)
		return
	}
	ringkasan.User = u
	h.View.Tampil(w, "home.html", ringkasan)
}
