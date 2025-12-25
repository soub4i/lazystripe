package version

// go build -ldflags "-X 'github.com/soub4i/internal/version.Version=1.2.3'"
var Version = "1.0.0"

func String() string {
	return Version
}
