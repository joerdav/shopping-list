package itemsweb

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/joerdav/shopping-list/business/items"
	"github.com/joerdav/shopping-list/business/items/itemsdb"
	"github.com/joerdav/shopping-list/business/shops"
	"github.com/joerdav/shopping-list/business/shops/shopsdb"
)

type Config struct {
	Conn *sql.DB
}

func RegisterHandlers(mux *http.ServeMux, config Config) {
	shopsCore := shops.NewCore(shopsdb.NewStorer(config.Conn))
	itemsCore := items.NewCore(itemsdb.NewStorer(config.Conn))

	mux.Handle("POST /items", createItemHandler(itemsCore, shopsCore))
}

func createItemHandler(itemsCore *items.Core, shopsCore *shops.Core) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		itemName := r.FormValue("itemName")
		if itemName == "" {
			http.Error(w, "itemName is required", http.StatusBadRequest)
			return
		}
		shopID, err := uuid.Parse(r.FormValue("shopID"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		_, err = shopsCore.Query(r.Context(), shopID)
		if errors.Is(err, shops.ErrNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_, err = itemsCore.Create(r.Context(), items.NewItem{Name: itemName, ShopID: shopID})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err := item(itemName).Render(r.Context(), w); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
}
