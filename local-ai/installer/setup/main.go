//go:build windows

// Gaon-Setup.exe – מחלץ את קבצי ההתקנה ומפעיל את אשף ההתקנה הגרפי בעברית.
package main

import (
	"embed"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	"gaon/installer/winapi"
)

//go:embed all:payload
var payload embed.FS

func main() {
	dir := filepath.Join(os.TempDir(), "GaonSetup")
	_ = os.RemoveAll(dir)
	err := fs.WalkDir(payload, "payload", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel("payload", filepath.FromSlash(p))
		target := filepath.Join(dir, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		data, err := payload.ReadFile(p)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, 0o644)
	})
	if err != nil {
		winapi.Error("התקנת גאון", "שגיאה בחילוץ קבצי ההתקנה:\n"+err.Error())
		return
	}
	cmd := exec.Command("powershell.exe", "-NoProfile", "-ExecutionPolicy", "Bypass", "-STA",
		"-WindowStyle", "Hidden", "-File", filepath.Join(dir, "install.ps1"))
	cmd.Dir = dir
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: 0x08000000}
	if err := cmd.Run(); err != nil {
		winapi.Error("התקנת גאון", "אשף ההתקנה נסגר עם שגיאה.\nפרטים בקובץ:\n"+filepath.Join(dir, "install.log"))
	}
}
