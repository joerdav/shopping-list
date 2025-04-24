load("@gazelle//:def.bzl", "gazelle")
load("@rules_go//go:def.bzl", "go_binary", "go_library")

gazelle(name = "gazelle")

go_library(
    name = "shopping-list_lib",
    srcs = ["main.go"],
    embedsrcs = [
        "migrations/001_init.down.sql",
        "migrations/001_init.up.sql",
        "migrations/002_bought_items.down.sql",
        "migrations/002_bought_items.up.sql",
    ],
    importpath = "github.com/joerdav/shopping-list",
    visibility = ["//visibility:private"],
    deps = [
        "//app",
        "@com_github_golang_migrate_migrate_v4//:migrate",
        "@com_github_golang_migrate_migrate_v4//database/sqlite3",
        "@com_github_golang_migrate_migrate_v4//source/iofs",
        "@com_github_joho_godotenv//:godotenv",
        "@com_github_mattn_go_sqlite3//:go-sqlite3",
    ],
)

go_binary(
    name = "shopping-list",
    embed = [":shopping-list_lib"],
    visibility = ["//visibility:public"],
)
