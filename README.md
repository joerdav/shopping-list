# Shopping List

## Tasks

### run

Requires: generate-templates, generate-tailwind

Env: DB=./db.db, SEED=1

```
go run .
```

### generate-templates

```
templ generate
```

### generate-tailwind

Dir: app

```
tailwindcss -o public/styles.css --minify
```

### watch

Requires: watch-templates, watch-go, watch-tailwind, watch-assets

RunDeps: async

### watch-templates

Watch for changes in `.templ` files, and inject hot reload scripts.

```
templ generate -watch -proxy=http://localhost:8080 -open-browser=false -v
```

### watch-go

Watch for changes in go files, this may be triggered by `watch-templates` generating some files.

Env: DB=./db.db, SEED=1

```
go run github.com/cosmtrek/air@v1.51.0 \
    --build.cmd "go build -tags dev -o tmp/bin/main" --build.bin "tmp/bin/main" --build.delay "100" \
    --build.include_ext "go" \
    --build.stop_on_error "false" \
    --misc.clean_on_exit true
```

### watch-tailwind

Dir: app

```
tailwindcss -o public/styles.css --minify --watch
```

### watch-assets

Watch for generated assets changing, trigger a browser reload.

```
go run github.com/cosmtrek/air@v1.51.0 \
       --build.cmd "templ generate --notify-proxy" \
       --build.bin "true" \
       --build.delay "100" \
       --build.exclude_dir "" \
       --build.include_dir "app/public" \
       --build.include_ext "js,css"
```
