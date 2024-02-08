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

Requires: watch-templates, watch-go, watch-tailwind

RunDeps: async

### watch-templates

```
templ generate -watch -proxy=http://localhost:8080
```

### watch-go

```
air
```

### watch-tailwind

Dir: app

```
tailwindcss -o public/styles.css --minify --watch
```
