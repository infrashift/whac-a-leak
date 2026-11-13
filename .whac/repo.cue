package whac

// whac scans its own repository in CI. These directories hold fixtures:
// every value in them is fake by design.
allowlist: paths: {
	"testdata/**":              "test corpus: every value is fake"
	"internal/rules/builtin/**": "builtin rule samples: every value is fake"
	"docs/src/demo/**":          "landing-page demo: every value is fake"
}
