package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"goproject/internal/notify"
)

type Session struct {
	Work       time.Duration
	ShortBreak time.Duration
	LongBreak  time.Duration
	Rounds     int // long break every N rounds
}

func defaultSession() Session {
	return Session{
		Work:       25 * time.Minute,
		ShortBreak: 5 * time.Minute,
		LongBreak:  15 * time.Minute,
		Rounds:     4,
	}
}

func runTimer(label string, d time.Duration) {
	fmt.Printf("\n⏱  %s — %v\n", label, d)

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	done := time.After(d)
	remaining := d

	// Handle Ctrl+C gracefully
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	for {
		select {
		case <-sigCh:
			fmt.Println("\nTimer cancelled.")
			os.Exit(0)
		case <-done:
			fmt.Printf("\r%s ✅ Done!                    \n", label)
			return
		case <-ticker.C:
			remaining -= time.Second
			m := int(remaining.Minutes())
			s := int(remaining.Seconds()) % 60
			fmt.Printf("\r  %s — %02d:%02d remaining   ", label, m, s)
		}
	}
}

func main() {
	work := flag.Int("work", 25, "Work session duration in minutes")
	short := flag.Int("short", 5, "Short break duration in minutes")
	long := flag.Int("long", 15, "Long break duration in minutes")
	rounds := flag.Int("rounds", 4, "Rounds before long break")
	sessions := flag.Int("sessions", 4, "Total pomodoro sessions to run")
	flag.Parse()

	s := Session{
		Work:       time.Duration(*work) * time.Minute,
		ShortBreak: time.Duration(*short) * time.Minute,
		LongBreak:  time.Duration(*long) * time.Minute,
		Rounds:     *rounds,
	}

	fmt.Println("🍅 Pomodoro Timer")
	fmt.Printf("   Work: %v | Short break: %v | Long break: %v\n", s.Work, s.ShortBreak, s.LongBreak)
	fmt.Printf("   Sessions: %d | Long break every %d rounds\n", *sessions, s.Rounds)
	fmt.Println("   Press Ctrl+C to quit.")

	for i := 1; i <= *sessions; i++ {
		fmt.Printf("\n─────────────────────────────\n")
		fmt.Printf("🍅 Pomodoro %d of %d\n", i, *sessions)

		notify.Send("🍅 Pomodoro", fmt.Sprintf("Session %d starting — focus time!", i))
		runTimer(fmt.Sprintf("Work Session %d", i), s.Work)
		notify.Send("✅ Done!", fmt.Sprintf("Session %d complete! Time for a break.", i))

		if i == *sessions {
			break
		}

		if i%s.Rounds == 0 {
			fmt.Printf("\n🌟 Long break after %d sessions!\n", s.Rounds)
			notify.Send("🌟 Long Break", fmt.Sprintf("You've done %d sessions! Take a long break.", i))
			runTimer("Long Break", s.LongBreak)
		} else {
			notify.Send("☕ Short Break", "Take a short break!")
			runTimer("Short Break", s.ShortBreak)
		}
	}

	fmt.Printf("\n─────────────────────────────\n")
	fmt.Printf("🎉 All %d sessions complete! Great work!\n", *sessions)
	notify.Send("🎉 All Done!", fmt.Sprintf("Completed all %d pomodoro sessions!", *sessions))
}
