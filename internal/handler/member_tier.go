package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/bilaaah/SeduhPoin/internal/middleware"
	"github.com/bilaaah/SeduhPoin/internal/model"
	"github.com/bilaaah/SeduhPoin/internal/repository"
)

type MemberTierHandler struct {
	Repo *repository.MemberTierRepo
	View *Renderer
}

// Halaman: GET /member/tier[?member=ID]
// ADMIN & KASIR boleh melihat semua member; MEMBER hanya miliknya sendiri.
func (h *MemberTierHandler) Halaman(w http.ResponseWriter, r *http.Request) {
	u, _ := middleware.UserDari(r.Context())
	ctx := r.Context()

	onlyID := 0
	if u.Role == "MEMBER" {
		id, err := h.Repo.MemberIDByEmail(ctx, u.Email)
		if err != nil {
			http.Error(w, "Kartu member untuk akun ini belum terdaftar", http.StatusNotFound)
			return
		}
		onlyID = id
	}

	daftar, err := h.Repo.DaftarMember(ctx, onlyID)
	if err != nil {
		gagal(w, err)
		return
	}

	var pilih model.MemberTier
	var batch []model.BatchPoin
	if len(daftar) > 0 {
		pilih = daftar[0]
		if want, _ := strconv.Atoi(r.URL.Query().Get("member")); want > 0 && onlyID == 0 {
			for _, m := range daftar {
				if m.ID == want {
					pilih = m
					break
				}
			}
		}
		if batch, err = h.Repo.BatchPoin(ctx, pilih.ID); err != nil {
			gagal(w, err)
			return
		}
	}

	var totalPoin, totalRupiah, gold, silver, perluSinkron int
	for _, m := range daftar {
		totalPoin += m.PoinAktif
		totalRupiah += m.NilaiRupiah
		if m.TingkatTersimpan == "GOLD" {
			gold++
		} else {
			silver++
		}
		if m.PerluSinkron {
			perluSinkron++
		}
	}

	h.View.Tampil(w, "poin_tier.html", struct {
		User         model.User
		Daftar       []model.MemberTier
		Pilih        model.MemberTier
		Batch        []model.BatchPoin
		TotalPoin    int
		TotalRupiah  int
		Gold, Silver int
		PerluSinkron int
		Info         string
	}{u, daftar, pilih, batch, totalPoin, totalRupiah, gold, silver, perluSinkron,
		r.URL.Query().Get("info")})
}

// Evaluasi: POST /member/tier/evaluasi  (khusus ADMIN)
// Form field opsional "member" = ID; kosong berarti evaluasi semua member.
func (h *MemberTierHandler) Evaluasi(w http.ResponseWriter, r *http.Request) {
	u, _ := middleware.UserDari(r.Context())
	if u.Role != "ADMIN" {
		http.Error(w, "403 Forbidden: hanya Admin yang dapat mengevaluasi tingkat member", http.StatusForbidden)
		return
	}

	var n int64
	var err error
	redirect := "/member/tier"
	if id, _ := strconv.Atoi(r.FormValue("member")); id > 0 {
		n, err = h.Repo.EvaluasiMember(r.Context(), id)
		redirect += "?member=" + strconv.Itoa(id) + "&"
	} else {
		n, err = h.Repo.EvaluasiSemua(r.Context())
		redirect += "?"
	}
	if err != nil {
		gagal(w, err)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("%sinfo=%d", redirect, n), http.StatusSeeOther)
}