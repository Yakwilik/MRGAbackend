package rest

import (
	"encoding/json"
	"github.com/Yakwilik/MRGAbackend/internal/model"
	"net/http"
)

func (receiver *Handler) document(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(model.Document{
		FieldsBlocks: nil,
	})

}
