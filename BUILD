load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library")

go_binary(
    name = "main",
    srcs = [
        "db.go",
        "main.go",
        "server.go",
    ],
    deps = [
	"@org_openmined_tcn_psi//tcn_psi/go/server",
        "@org_openmined_tcn_psi//tcn_psi/go/tcn",
        "@com_github_gin_gonic_gin//:go_default_library",
        "@com_github_jmoiron_sqlx//:go_default_library",
        "@com_github_lib_pq//:go_default_library",
        "@com_github_urfave_cli//:go_default_library",
    ],
)
