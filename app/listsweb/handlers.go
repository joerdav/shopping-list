package listsweb

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/joerdav/shopping-list/app/middleware"
	"github.com/joerdav/shopping-list/business/auth"
	"github.com/joerdav/shopping-list/business/items"
	"github.com/joerdav/shopping-list/business/items/itemsdb"
	"github.com/joerdav/shopping-list/business/lists"
	"github.com/joerdav/shopping-list/business/lists/listsdb"
	"github.com/joerdav/shopping-list/business/recipes"
	"github.com/joerdav/shopping-list/business/recipes/recipesdb"
	"github.com/joerdav/shopping-list/business/shops"
	"github.com/joerdav/shopping-list/business/shops/shopsdb"
	"github.com/joerdav/shopping-list/foundation/routing"
)

type Config struct {
	Conn *sql.DB
}

func RegisterHandlers(mux *http.ServeMux, config Config) {
	listCore := lists.NewCore(listsdb.NewStorer(config.Conn))
	itemsCore := items.NewCore(itemsdb.NewStorer(config.Conn))
	recipeCore := recipes.NewCore(recipesdb.NewStorer(config.Conn))
	shopCore := shops.NewCore(shopsdb.NewStorer(config.Conn))

	authMiddleware := middleware.AuthMiddleware(auth.NewCore())

	routing.RegisterRoute(mux, "GET /{$}", listsListHandler(listCore), authMiddleware)
	routing.RegisterRoute(mux, "POST /lists", createListHandler(listCore), authMiddleware)
	routing.RegisterRoute(
		mux,
		"GET /lists/{listid}",
		getListHandler(listCore, itemsCore, recipeCore, shopCore),
		authMiddleware,
	)
	routing.RegisterRoute(
		mux,
		"PUT /lists/{listid}/item",
		setListItemCountHandler(listCore, itemsCore),
		authMiddleware,
	)
	routing.RegisterRoute(
		mux,
		"PUT /lists/{listid}/recipe",
		setListRecipeCountHandler(listCore, recipeCore),
		authMiddleware,
	)
}

func listsListHandler(listCore *lists.Core) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := auth.UserID(r.Context())
		lists, err := listCore.QueryAll(r.Context(), userID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		var listSummaries []ListSummary
		for _, list := range lists {
			listSummaries = append(listSummaries, ListSummary{
				ID:          list.ID.String(),
				CreatedDate: list.CreatedDate,
			})
		}
		// TODO: create render function that prevents partial renders
		if err := ListsPage(r.URL.Path, listSummaries).Render(r.Context(), w); err != nil {
			// TODO: replace with error page
			http.Error(w, "render error", http.StatusInternalServerError)
		}
	})
}

func createListHandler(listCore *lists.Core) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := auth.UserID(r.Context())
		_, err := listCore.Create(r.Context(), lists.NewList{CreatedDate: time.Now(), UserID: userID})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// TODO: avoid a redirect
		http.Redirect(w, r, "/", http.StatusSeeOther)
	})
}

// Move logic out
func getListHandler(
	listCore *lists.Core,
	itemsCore *items.Core,
	recipeCore *recipes.Core,
	shopCore *shops.Core,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := auth.UserID(r.Context())
		listID, err := uuid.Parse(r.PathValue("listid"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		list, err := listCore.Query(r.Context(), listID)
		if errors.Is(err, lists.ErrNotFound) {
			http.Error(w, "list not found", http.StatusNotFound)
			return
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if list.UserID != userID {
			http.Error(w, "list not found", http.StatusNotFound)
			return
		}
		listModel := List{
			ID:          list.ID.String(),
			CreatedDate: list.CreatedDate,
			Recipes:     []ListRecipe{},
			Items:       []ListItem{},
			Shops:       []ListShop{},
		}
		items := map[string]ListItem{}
		selectedItems := map[string]ListItem{}
		recipes := map[string]ListRecipe{}
		for id, count := range list.Items {
			ci, err := itemsCore.Query(r.Context(), id)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			sid := id.String()
			item := ListItem{
				ID:       sid,
				Name:     ci.Name,
				Quantity: count,
				ShopID:   ci.ShopID.String(),
			}
			items[sid] = item
			selectedItems[sid] = item
		}
		for rid, rcount := range list.Recipes {
			slog.Info("Loading recipe", "recipeID", rid)
			recipe, err := recipeCore.Query(r.Context(), rid)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			recipes[rid.String()] = ListRecipe{
				ID:       rid.String(),
				Name:     recipe.Name,
				Quantity: rcount,
			}
			for id, count := range recipe.Ingredients {
				count *= rcount
				sid := id.String()
				item, ok := items[sid]
				if !ok {
					ci, err := itemsCore.Query(r.Context(), id)
					if err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
						return
					}
					items[sid] = ListItem{
						ID:       sid,
						Name:     ci.Name,
						Quantity: count,
						ShopID:   ci.ShopID.String(),
					}
					continue
				}
				item.Quantity += count
				items[sid] = item
			}
		}
		shops := map[string]ListShop{}
		for _, item := range items {
			shop, ok := shops[item.ShopID]
			if !ok {
				slog.Info("Loading shop", "shopID", item.ShopID)
				id, _ := uuid.Parse(item.ShopID)
				s, err := shopCore.Query(r.Context(), id)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				shop = ListShop{
					Name: s.Name,
				}
			}
			shop.Items = append(shop.Items, item)
			shops[item.ShopID] = shop
		}
		for _, recipe := range recipes {
			listModel.Recipes = append(listModel.Recipes, recipe)
		}
		sort.Slice(listModel.Recipes, func(i, j int) bool {
			return listModel.Recipes[i].Name < listModel.Recipes[j].Name
		})
		for _, item := range selectedItems {
			listModel.Items = append(listModel.Items, item)
		}
		sort.Slice(listModel.Items, func(i, j int) bool {
			return listModel.Items[i].Name < listModel.Items[j].Name
		})
		for _, shop := range shops {
			sort.Slice(shop.Items, func(i, j int) bool {
				return shop.Items[i].Name < shop.Items[j].Name
			})
			listModel.Shops = append(listModel.Shops, shop)
		}
		sort.Slice(listModel.Shops, func(i, j int) bool {
			return listModel.Shops[i].Name < listModel.Shops[j].Name
		})
		availableItems := []Item{}
		citems, err := itemsCore.QueryAll(r.Context(), userID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		for _, item := range citems {
			availableItems = append(availableItems, Item{
				ID:   item.ID.String(),
				Name: item.Name,
			})
		}
		availableRecipes := []Recipe{}
		crecipes, err := recipeCore.QueryAll(r.Context(), userID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		for _, recipe := range crecipes {
			availableRecipes = append(availableRecipes, Recipe{
				ID:   recipe.ID.String(),
				Name: recipe.Name,
			})
		}
		if err := ListPage(r.URL.Path, listModel, availableRecipes, availableItems).Render(r.Context(), w); err != nil {
			// TODO: replace with error page
			http.Error(w, "render error", http.StatusInternalServerError)
		}
	})
}

// Move logic out
func setListItemCountHandler(listCore *lists.Core, itemCore *items.Core) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := auth.UserID(r.Context())
		listID, err := uuid.Parse(r.PathValue("listid"))
		if err != nil {
			// TODO: replace with error page
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		quantity, err := strconv.Atoi(r.FormValue("quantity"))
		if err != nil {
			slog.Error("failed to parse quantity", "error", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		itemID, err := uuid.Parse(r.FormValue("itemID"))
		if err != nil {
			slog.Error("failed to parse itemID", "error", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		list, err := listCore.Query(r.Context(), listID)
		if errors.Is(err, lists.ErrNotFound) {
			http.Error(w, "list not found", http.StatusNotFound)
			return
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if list.UserID != userID {
			http.Error(w, "list not found", http.StatusNotFound)
			return
		}
		_, err = itemCore.Query(r.Context(), itemID)
		if errors.Is(err, items.ErrNotFound) {
			http.Error(w, "item not found", http.StatusNotFound)
			return
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		slog.Info("Adding item", "listID", listID, "itemID", itemID, "quantity", quantity)
		list.Items[itemID] = quantity
		if quantity == 0 {
			delete(list.Items, itemID)
		}
		_, err = listCore.Update(r.Context(), lists.UpdateList{
			ID:    listID,
			Items: &list.Items,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, fmt.Sprintf("/lists/%s", listID), http.StatusSeeOther)
	})
}

// Move logic out
func setListRecipeCountHandler(listCore *lists.Core, recipeCore *recipes.Core) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := auth.UserID(r.Context())
		listID, err := uuid.Parse(r.PathValue("listid"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		recipeID, err := uuid.Parse(r.FormValue("recipeID"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		quantity, err := strconv.Atoi(r.FormValue("quantity"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		list, err := listCore.Query(r.Context(), listID)
		if errors.Is(err, lists.ErrNotFound) {
			http.Error(w, "list not found", http.StatusNotFound)
			return
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if list.UserID != userID {
			http.Error(w, "list not found", http.StatusNotFound)
			return
		}
		_, err = recipeCore.Query(r.Context(), recipeID)
		if errors.Is(err, recipes.ErrNotFound) {
			http.Error(w, "recipe not found", http.StatusNotFound)
			return
		}
		slog.Info("Adding recipe", "listID", listID, "recipeID", recipeID, "quantity", quantity)
		list.Recipes[recipeID] = quantity
		if quantity == 0 {
			delete(list.Recipes, recipeID)
		}
		_, err = listCore.Update(r.Context(), lists.UpdateList{
			ID:      listID,
			Recipes: &list.Recipes,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, fmt.Sprintf("/lists/%s", listID), http.StatusSeeOther)
	})
}

type ListSummary struct {
	ID          string
	CreatedDate time.Time
}

type List struct {
	ID          string
	CreatedDate time.Time
	Shops       []ListShop
	Items       []ListItem
	Recipes     []ListRecipe
}

type ListShop struct {
	Name  string
	Items []ListItem
}

type ListItem struct {
	ID       string
	ShopID   string
	Name     string
	Quantity int
}

type ListRecipe struct {
	ID       string
	Name     string
	Quantity int
}

type Item struct {
	ID   string
	Name string
}

type Recipe struct {
	ID   string
	Name string
}
