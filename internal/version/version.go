package version

// Version is the build version of sloper, injected at build time via
//
//	go build -ldflags "-X github.com/ButterHost69/sloper/internal/version.Version=..."
//
// It defaults to "dev" for local builds that don't pass ldflags.
var Version = "dev"