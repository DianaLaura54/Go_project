package notify

import (
	"fmt"
	"os/exec"
	"runtime"
)

// Send sends a desktop notification with the given title and message.
// Falls back to printing to stdout if desktop notifications aren't available.
func Send(title, message string) {
	var err error

	switch runtime.GOOS {
	case "darwin":
		script := fmt.Sprintf(`display notification %q with title %q`, message, title)
		err = exec.Command("osascript", "-e", script).Run()
	case "linux":
		err = exec.Command("notify-send", title, message).Run()
	case "windows":
		// PowerShell toast notification
		ps := fmt.Sprintf(`
			[Windows.UI.Notifications.ToastNotificationManager, Windows.UI.Notifications, ContentType = WindowsRuntime] | Out-Null
			$template = [Windows.UI.Notifications.ToastNotificationManager]::GetTemplateContent([Windows.UI.Notifications.ToastTemplateType]::ToastText02)
			$textNodes = $template.GetElementsByTagName("text")
			$textNodes.Item(0).AppendChild($template.CreateTextNode('%s')) | Out-Null
			$textNodes.Item(1).AppendChild($template.CreateTextNode('%s')) | Out-Null
			$toast = [Windows.UI.Notifications.ToastNotification]::new($template)
			[Windows.UI.Notifications.ToastNotificationManager]::CreateToastNotifier('Pomodoro').Show($toast)
		`, title, message)
		err = exec.Command("powershell", "-Command", ps).Run()
	}

	if err != nil {
		// Graceful fallback
		fmt.Printf("\n[%s] %s\n", title, message)
	}
}
