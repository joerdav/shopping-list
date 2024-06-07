package shopsweb

import (
	"database/sql"
	"log/slog"
	"net/http"
	"sort"

	"github.com/joerdav/shopping-list/app/itemsweb"
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

	routing.RegisterRoute(mux, "GET /shops", getShopsHandler(shopsCore, itemsCore), authMiddleware)
	routing.RegisterRoute(mux, "POST /shops", createShopHandler(shopsCore), authMiddleware)
}

type Shop struct {
	ID    string
	Name  string
	Items []itemsweb.Item
}

func getShopsHandler(shopsCore *shops.Core, itemsCore *items.Core) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var shopModels []Shop
		userID := auth.UserID(r.Context())
		shops, err := shopsCore.QueryAll(r.Context(), userID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		for _, shop := range shops {
			shopModels = append(shopModels, Shop{
				ID:   shop.ID.String(),
				Name: shop.Name,
			})
			shopModel := &shopModels[len(shopModels)-1]
			items, err := itemsCore.QueryByShopID(r.Context(), shop.ID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			for _, item := range items {
				shopModel.Items = append(
					shopModel.Items,
					itemsweb.Item{ID: item.ID.String(), Name: item.Name},
				)
			}
			sort.Slice(shopModel.Items, func(i, j int) bool {
				return shopModel.Items[i].Name < shopModel.Items[j].Name
			})
		}
		if err := ShopsPage(r.URL.Path, shopModels).Render(r.Context(), w); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
}

func createShopHandler(shopsCore *shops.Core) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := auth.UserID(r.Context())
		if err := r.ParseForm(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		shopName := r.FormValue("shopName")
		if shopName == "" {
			http.Error(w, "shopName is required", http.StatusBadRequest)
			return
		}
		slog.Info("Creating shop", "shopName", shopName)
		_, err := shopsCore.Create(r.Context(), shops.NewShop{Name: shopName, UserID: userID})
		if err != nil {
			slog.Error("Failed to create shop", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/shops", http.StatusSeeOther)
	})
}
