package recipesweb

import (
	"database/sql"
	"errors"
	"net/http"
	"sort"
	"strconv"

	"github.com/google/uuid"
	"github.com/joerdav/shopping-list/business/items"
	"github.com/joerdav/shopping-list/business/items/itemsdb"
	"github.com/joerdav/shopping-list/business/recipes"
	"github.com/joerdav/shopping-list/business/recipes/recipesdb"
)

type Config struct {
	Conn *sql.DB
}

func RegisterHandlers(mux *http.ServeMux, config Config) {
	recipeCore := recipes.NewCore(recipesdb.NewStorer(config.Conn))
	itemsCore := items.NewCore(itemsdb.NewStorer(config.Conn))

	mux.Handle("GET /recipes", getRecipesHandler(recipeCore, itemsCore))
	mux.Handle("POST /recipes", createRecipeHandler(recipeCore, itemsCore))
	mux.Handle("PUT /recipes/{recipeid}", setIngredientsHandler(recipeCore, itemsCore))
}

type Item struct {
	ID   string
	Name string
}

type RecipeItem struct {
	ID       string
	Name     string
	Quantity int
}

type Recipe struct {
	ID          string
	Name        string
	Ingredients []RecipeItem
}

func getRecipesHandler(recipeCore *recipes.Core, itemsCore *items.Core) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		recipes, err := recipeCore.QueryAll(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		var recipeModels []Recipe
		for _, recipe := range recipes {
			recipeModel := Recipe{
				ID:   recipe.ID.String(),
				Name: recipe.Name,
			}
			for id, qty := range recipe.Ingredients {
				item, err := itemsCore.Query(r.Context(), id)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				recipeModel.Ingredients = append(recipeModel.Ingredients, RecipeItem{
					ID:       id.String(),
					Name:     item.Name,
					Quantity: qty,
				})
			}
			sort.Slice(recipeModel.Ingredients, func(i, j int) bool {
				return recipeModel.Ingredients[i].Name < recipeModel.Ingredients[j].Name
			})
			recipeModels = append(recipeModels, recipeModel)
		}
		sort.Slice(recipeModels, func(i, j int) bool {
			return recipeModels[i].Name < recipeModels[j].Name
		})
		var availableIngredients []Item
		ingredients, err := itemsCore.QueryAll(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		for _, item := range ingredients {
			availableIngredients = append(availableIngredients, Item{
				ID:   item.ID.String(),
				Name: item.Name,
			})
		}
		sort.Slice(availableIngredients, func(i, j int) bool {
			return availableIngredients[i].Name < availableIngredients[j].Name
		})

		if err := RecipesPage(r.URL.Path, recipeModels, availableIngredients).Render(r.Context(), w); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
}
func createRecipeHandler(recipeCore *recipes.Core, itemsCore *items.Core) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		recipeName := r.FormValue("recipeName")
		if recipeName == "" {
			http.Error(w, "recipeName is required", http.StatusBadRequest)
			return
		}
		recipe, err := recipeCore.Create(r.Context(), recipes.NewRecipe{Name: recipeName})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		var availableIngredients []Item
		ingredients, err := itemsCore.QueryAll(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		for _, item := range ingredients {
			availableIngredients = append(availableIngredients, Item{
				ID:   item.ID.String(),
				Name: item.Name,
			})
		}
		sort.Slice(availableIngredients, func(i, j int) bool {
			return availableIngredients[i].Name < availableIngredients[j].Name
		})

		if err := Ingredients(Recipe{ID: recipe.ID.String(), Name: recipeName}, availableIngredients).Render(r.Context(), w); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
}
func setIngredientsHandler(recipeCore *recipes.Core, itemsCore *items.Core) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		quantity, err := strconv.Atoi(r.FormValue("quantity"))
		if err != nil {
			http.Error(w, "invalid quantity", http.StatusBadRequest)
			return
		}
		recipeID, err := uuid.Parse(r.PathValue("recipeid"))
		if err != nil {
			http.Error(w, "malformed recipeid", http.StatusBadRequest)
			return
		}
		itemID, err := uuid.Parse(r.FormValue("itemID"))
		if err != nil {
			http.Error(w, "malformed itemID", http.StatusBadRequest)
			return
		}
		recipe, err := recipeCore.Query(r.Context(), recipeID)
		if errors.Is(err, recipes.ErrNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_, err = itemsCore.Query(r.Context(), itemID)
		if errors.Is(err, items.ErrNotFound) {
			http.Error(w, "item does not exist", http.StatusBadRequest)
			return
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		recipe.Ingredients[itemID] = quantity
		if quantity == 0 {
			delete(recipe.Ingredients, itemID)
		}
		recipeCore.Update(r.Context(), recipes.UpdateRecipe{
			ID:          recipeID,
			Ingredients: &recipe.Ingredients,
		})
		recipeModel := Recipe{
			ID:   recipe.ID.String(),
			Name: recipe.Name,
		}
		for id, qty := range recipe.Ingredients {
			item, err := itemsCore.Query(r.Context(), id)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			recipeModel.Ingredients = append(recipeModel.Ingredients, RecipeItem{
				ID:       id.String(),
				Name:     item.Name,
				Quantity: qty,
			})
		}
		sort.Slice(recipeModel.Ingredients, func(i, j int) bool {
			return recipeModel.Ingredients[i].Name < recipeModel.Ingredients[j].Name
		})
		var availableIngredients []Item
		ingredients, err := itemsCore.QueryAll(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		for _, item := range ingredients {
			availableIngredients = append(availableIngredients, Item{
				ID:   item.ID.String(),
				Name: item.Name,
			})
		}
		sort.Slice(availableIngredients, func(i, j int) bool {
			return availableIngredients[i].Name < availableIngredients[j].Name
		})
		if err := Ingredients(recipeModel, availableIngredients).Render(r.Context(), w); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
}
