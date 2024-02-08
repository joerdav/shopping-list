package shopsweb

import (
	"database/sql"
	"log/slog"
	"net/http"
	"sort"

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

	mux.Handle("GET /shops", getShopsHandler(shopsCore, itemsCore))
	mux.Handle("POST /shops", createShopHandler(shopsCore))
}

type Shop struct {
	ID    string
	Name  string
	Items []string
}

func getShopsHandler(shopsCore *shops.Core, itemsCore *items.Core) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var shopModels []Shop
		shops, err := shopsCore.QueryAll(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		for _, shop := range shops {
			shopModels = append(shopModels, Shop{
				ID:   shop.ID.String(),
				Name: shop.Name,
			})
			items, err := itemsCore.QueryByShopID(r.Context(), shop.ID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			for _, item := range items {
				shopModels[len(shopModels)-1].Items = append(shopModels[len(shopModels)-1].Items, item.Name)
			}
			sort.Strings(shopModels[len(shopModels)-1].Items)
		}
		if err := ShopsPage(r.URL.Path, shopModels).Render(r.Context(), w); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
}
func createShopHandler(shopsCore *shops.Core) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
		_, err := shopsCore.Create(r.Context(), shops.NewShop{Name: shopName})
		if err != nil {
			slog.Error("Failed to create shop", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/shops", http.StatusSeeOther)
	})
}
