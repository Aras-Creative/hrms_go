package browser

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	errors "hrms/internal/pkg/apperror"
)

// execEnv is the override. Set it to point at a browser chromedp cannot find on its own,
// for example a Flatpak install, which does not land in PATH.
const execEnv = "CHROME_PATH"

// candidates are tried in order after the environment override. The names differ by
// distribution and by whether Chrome or open-source Chromium is installed, so guessing
// only one is what makes this fail on a machine that does have a browser.
var candidates = []string{
	"google-chrome-stable",
	"google-chrome",
	"chromium",
	"chromium-browser",
	"chrome",
	"headless_shell",
	"headless-shell",
}

// Resolve locates a browser executable for chromedp to drive.
//
// chromedp's own default only looks for a couple of hardcoded names and then fails with
// "executable file not found in $PATH", which surfaces to the API caller as a bare
// "internal server error". That tells whoever is debugging nothing about which dependency
// is missing, so this reports the executable it looked for and how to point it elsewhere.
func Resolve() (string, error) {
	if custom := strings.TrimSpace(os.Getenv(execEnv)); custom != "" {
		if _, err := os.Stat(custom); err != nil {
			return "", errors.NewServiceUnavailable(
				execEnv + " points at " + custom + ", which does not exist")
		}
		return custom, nil
	}

	for _, name := range candidates {
		if path, err := exec.LookPath(name); err == nil {
			return path, nil
		}
	}

	// Flatpak installs into a per-user prefix that is never on PATH. Worth checking
	// explicitly, since it is the usual reason a browser is present but "not found".
	if home, err := os.UserHomeDir(); err == nil {
		p := filepath.Join(home, ".local", "share", "flatpak", "exports", "bin", "com.google.Chrome")
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}

	return "", errors.NewServiceUnavailable(notInstalledMessage())
}

func notInstalledMessage() string {
	var b strings.Builder
	b.WriteString("no Chrome or Chromium executable found for PDF rendering; looked for ")
	b.WriteString(strings.Join(candidates, ", "))
	if runtime.GOOS != "windows" {
		b.WriteString(" in PATH")
	}
	b.WriteString(" and the Flatpak user prefix. Install one, or set ")
	b.WriteString(execEnv)
	b.WriteString(" to its full path")
	return b.String()
}
