package injection

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

type clipboardBackend struct {
	paste    bool
	shortcut string
}

func NewClipboardBackend(paste bool, shortcut string) Backend {
	return &clipboardBackend{paste: paste, shortcut: normalizeClipboardShortcut(shortcut)}
}

func (c *clipboardBackend) Name() string {
	return "clipboard"
}

func (c *clipboardBackend) Available() error {
	if _, err := exec.LookPath("wl-copy"); err != nil {
		return fmt.Errorf("wl-copy not found: %w (install wl-clipboard)", err)
	}

	if os.Getenv("WAYLAND_DISPLAY") == "" {
		return fmt.Errorf("WAYLAND_DISPLAY not set - clipboard operations require Wayland session")
	}

	if os.Getenv("XDG_RUNTIME_DIR") == "" {
		return fmt.Errorf("XDG_RUNTIME_DIR not set - clipboard operations require proper session environment")
	}

	if c.paste {
		if err := NewYdotoolBackend().Available(); err != nil {
			return fmt.Errorf("clipboard paste requires ydotool: %w", err)
		}
	}

	return nil
}

func (c *clipboardBackend) Inject(ctx context.Context, text string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	if err := c.Available(); err != nil {
		return err
	}

	cmd := exec.CommandContext(ctx, "wl-copy")
	cmd.Stdin = strings.NewReader(text)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("wl-copy failed: %w", err)
	}

	if c.paste {
		cmd = exec.CommandContext(ctx, "ydotool", append([]string{"key"}, clipboardShortcutKeySequence(c.shortcut)...)...)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("clipboard paste via ydotool failed: %w", err)
		}
	}

	return nil
}

func normalizeClipboardShortcut(shortcut string) string {
	switch strings.ToLower(strings.TrimSpace(shortcut)) {
	case "", "ctrl+v":
		return "ctrl+v"
	case "ctrl+shift+v":
		return "ctrl+shift+v"
	default:
		return "ctrl+v"
	}
}

func clipboardShortcutKeySequence(shortcut string) []string {
	switch normalizeClipboardShortcut(shortcut) {
	case "ctrl+shift+v":
		return []string{"29:1", "42:1", "47:1", "47:0", "42:0", "29:0"}
	default:
		return []string{"29:1", "47:1", "47:0", "29:0"}
	}
}
