package rest

import (
	"encoding/json"
	index "github.com/Yakwilik/MRGAbackend/internal/app/templ"
	"github.com/Yakwilik/MRGAbackend/internal/model"
	"net/http"
	"strings"
)

type Handler struct {
	mux *http.ServeMux
}

func New() *Handler {
	return &Handler{mux: http.NewServeMux()}
}

func (receiver *Handler) Init() *http.ServeMux {
	receiver.mux.HandleFunc("GET /get_document", receiver.document)

	return receiver.mux
}

func (receiver *Handler) document(w http.ResponseWriter, r *http.Request) {

	sb := &strings.Builder{}
	index.Index().Render(r.Context(), sb)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(model.Document{
		Pages: []model.DocumentPage{{
			Content: sb.String(),
		}},
		FieldsBlocks: nil,
	})
}
