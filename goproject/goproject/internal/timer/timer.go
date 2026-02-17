package timer

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"
)

// Config holds Pomodoro timer settings.
type Config struct {
	WorkMinutes       int
	ShortBreakMinutes int
	LongBreakMinutes  int
	RoundsBeforeLong  int
	TotalSessions     int // 0 = run forever
	PlaySound         bool
}

// Timer runs the Pomodoro workflow.
type Timer struct {
	cfg Config
}

// New creates a new Timer.
func New(cfg Config) *Timer {
	return &Timer{cfg: cfg}
}

type phase struct {
	name     string
	emoji    string
	minutes  int
	isWork   bool
}

// Run starts the Pomodoro loop.
func (t *Timer) Run() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	session := 0
	round := 0

	fmt.Println()
	printBanner()

	for {
		if t.cfg.TotalSessions > 0 && session >= t.cfg.TotalSessions {
			fmt.Printf("\n✅ All %d session(s) complete! Great work!\n\n", t.cfg.TotalSessions)
			return
		}

		round++
		if t.cfg.TotalSessions > 0 {
			fmt.Printf("\n📋 Session %d/%d — Round %d\n", session+1, t.cfg.TotalSessions, round)
		} else {
			fmt.Printf("\n📋 Round %d\n", round)
		}

		// Work phase
		p := phase{"Focus Time", "🍅", t.cfg.WorkMinutes, true}
		if done := t.runPhase(p, sig); !done {
			return
		}

		notify("Pomodoro Complete! 🍅", fmt.Sprintf("Round %d done — time for a break!", round), t.cfg.PlaySound)

		// Break phase
		var brk phase
		if round%t.cfg.RoundsBeforeLong == 0 {
			brk = phase{"Long Break", "🌿", t.cfg.LongBreakMinutes, false}
			session++
		} else {
			brk = phase{"Short Break", "☕", t.cfg.ShortBreakMinutes, false}
		}

		if done := t.runPhase(brk, sig); !done {
			return
		}

		notify("Break Over!", "Time to focus again 🍅", t.cfg.PlaySound)
	}
}

func (t *Timer) runPhase(p phase, sig chan os.Signal) bool {
	total := time.Duration(p.minutes) * time.Minute
	end := time.Now().Add(total)

	fmt.Printf("\n%s %s — %d min\n", p.emoji, p.name, p.minutes)

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-sig:
			fmt.Println("\n\n👋 Pomodoro interrupted. Goodbye!")
			return false
		case now := <-ticker.C:
			remaining := end.Sub(now)
			if remaining <= 0 {
				clearLine()
				fmt.Printf("\r%s %s — Done!                          \n", p.emoji, p.name)
				return true
			}
			printProgress(p, total, remaining)
		}
	}
}

func printProgress(p phase, total, remaining time.Duration) {
	elapsed := total - remaining
	pct := float64(elapsed) / float64(total)

	barWidth := 30
	filled := int(pct * float64(barWidth))
	bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)

	mins := int(remaining.Minutes())
	secs := int(remaining.Seconds()) % 60

	clearLine()
	fmt.Printf("\r  [%s] %02d:%02d remaining  (%.0f%%)", bar, mins, secs, pct*100)
}

func clearLine() {
	fmt.Print("\r\033[K")
}

func printBanner() {
	fmt.Println("┌─────────────────────────────────┐")
	fmt.Println("│     🍅  Pomodoro Timer  🍅       │")
	fmt.Println("│  Stay focused. Take real breaks. │")
	fmt.Println("└─────────────────────────────────┘")
	fmt.Println("  Press Ctrl+C to stop at any time.")
}

// notify sends a desktop notification via the OS notification system.
func notify(title, message string, playSound bool) {
	if playSound {
		fmt.Print("\a") // terminal bell
	}

	switch runtime.GOOS {
	case "darwin":
		script := fmt.Sprintf(`display notification "%s" with title "%s" sound name "Glass"`, message, title)
		exec.Command("osascript", "-e", script).Run() //nolint
	case "linux":
		exec.Command("notify-send", "-t", "5000", title, message).Run() //nolint
	case "windows":
		// PowerShell toast notification
		ps := fmt.Sprintf(`
[Windows.UI.Notifications.ToastNotificationManager, Windows.UI.Notifications, ContentType = WindowsRuntime] | Out-Null
$template = [Windows.UI.Notifications.ToastNotificationManager]::GetTemplateContent([Windows.UI.Notifications.ToastTemplateType]::ToastText02)
$template.GetElementsByTagName("text")[0].AppendChild($template.CreateTextNode('%s')) | Out-Null
$template.GetElementsByTagName("text")[1].AppendChild($template.CreateTextNode('%s')) | Out-Null
$notifier = [Windows.UI.Notifications.ToastNotificationManager]::CreateToastNotifier('Pomodoro Timer')
$notifier.Show([Windows.UI.Notifications.ToastNotification]::new($template))`, title, message)
		exec.Command("powershell", "-Command", ps).Run() //nolint
	}

	// Always print to terminal regardless
	fmt.Printf("\n  🔔 %s — %s\n", title, message)
}
