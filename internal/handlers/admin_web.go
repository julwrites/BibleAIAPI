package handlers

import (
	_ "embed"
	"net/http"
)

//go:embed templates/admin.html
var adminHTML []byte

func (h *AdminHandler) ServeAdminUI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "text/html")
	w.Write(adminHTML)
}
