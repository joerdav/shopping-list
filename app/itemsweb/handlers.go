package itemsweb

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/joerdav/shopping-list/app/middleware"
	"github.com/joerdav/shopping-list/business/auth"
	"github.com/joerdav/shopping-list/business/items"
	"github.com/joerdav/shopping-list/business/items/itemsdb"
	"github.com/joerdav/shopping-list/business/shops"
	"github.com/joerdav/shopping-list/business/shops/shopsdb"
	"github.com/joerdav/shopping-list/foundation/routing"
)

type Config struct {
	Conn *sql.DB
}

func RegisterHandlers(mux *http.ServeMux, config Config) {
	shopsCore := shops.NewCore(shopsdb.NewStorer(config.Conn))
	itemsCore := items.NewCore(itemsdb.NewStorer(config.Conn))

	authMiddleware := middleware.AuthMiddleware(auth.NewCore())

	routing.RegisterRoute(
		mux,
		"POST /items",
		createItemHandler(itemsCore, shopsCore),
		authMiddleware,
	)
}

func createItemHandler(itemsCore *items.Core, shopsCore *shops.Core) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := auth.UserID(r.Context())
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
		shop, err := shopsCore.Query(r.Context(), shopID)
		if errors.Is(err, shops.ErrNotFound) {
			http.Error(w, "shop not found", http.StatusBadRequest)
			return
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if shop.UserID != userID {
			http.Error(w, "shop not found", http.StatusBadRequest)
			return
		}
		_, err = itemsCore.Create(r.Context(), items.NewItem{Name: itemName, ShopID: shopID, UserID: userID})
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
