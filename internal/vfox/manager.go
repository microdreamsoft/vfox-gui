package vfox

import (
	"bytes"
	"os/exec"
	"strings"
	"syscall"
)

type Manager struct {
	command string
}

func NewManager() *Manager {
	return &Manager{command: "vfox"}
}

func (m *Manager) Run(args ...string) (string, string, error) {
	cmd := exec.Command(m.command, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return strings.TrimSpace(stdout.String()), strings.TrimSpace(stderr.String()), err
}

func (m *Manager) Available() (string, error) {
	stdout, _, err := m.Run("available")
	return stdout, err
}

func (m *Manager) Add(name string) (string, error) {
	stdout, _, err := m.Run("add", name)
	return stdout, err
}

func (m *Manager) Remove(name string) (string, error) {
	stdout, _, err := m.Run("remove", name)
	return stdout, err
}

func (m *Manager) Update(name string) (string, error) {
	stdout, _, err := m.Run("update", name)
	return stdout, err
}

func (m *Manager) UpdateAll() (string, error) {
	stdout, _, err := m.Run("update", "--all")
	return stdout, err
}

func (m *Manager) Search(name string) (string, error) {
	stdout, _, err := m.Run("search", name)
	return stdout, err
}

func (m *Manager) Install(sdk string) (string, error) {
	stdout, _, err := m.Run("install", "--yes", sdk)
	return stdout, err
}

func (m *Manager) Uninstall(sdk string) (string, error) {
	stdout, _, err := m.Run("uninstall", sdk)
	return stdout, err
}

func (m *Manager) Use(sdk string, scope string) (string, error) {
	args := []string{"use"}
	switch scope {
	case "global":
		args = append(args, "--global")
	case "project":
		args = append(args, "--project")
	case "session":
		args = append(args, "--session")
	}
	args = append(args, sdk)
	stdout, _, err := m.Run(args...)
	return stdout, err
}

func (m *Manager) Unuse(sdk string, scope string) (string, error) {
	args := []string{"unuse"}
	switch scope {
	case "global":
		args = append(args, "--global")
	case "project":
		args = append(args, "--project")
	case "session":
		args = append(args, "--session")
	}
	args = append(args, sdk)
	stdout, _, err := m.Run(args...)
	return stdout, err
}

func (m *Manager) List() (string, error) {
	stdout, _, err := m.Run("list")
	return stdout, err
}

func (m *Manager) ListSDK(name string) (string, error) {
	stdout, _, err := m.Run("list", name)
	return stdout, err
}

func (m *Manager) Current() (string, error) {
	stdout, _, err := m.Run("current")
	return stdout, err
}

func (m *Manager) CurrentSDK(name string) (string, error) {
	stdout, _, err := m.Run("current", name)
	return stdout, err
}

func (m *Manager) Info(sdk string) (string, error) {
	stdout, _, err := m.Run("info", sdk)
	return stdout, err
}

func (m *Manager) ConfigGet(key string) (string, error) {
	stdout, _, err := m.Run("config", key)
	return stdout, err
}

func (m *Manager) ConfigSet(key, value string) (string, error) {
	stdout, _, err := m.Run("config", key, value)
	return stdout, err
}

func (m *Manager) Upgrade() (string, error) {
	stdout, _, err := m.Run("upgrade")
	return stdout, err
}

func (m *Manager) Exec(sdk, command string) (string, error) {
	stdout, _, err := m.Run("exec", sdk, "--", command)
	return stdout, err
}

func (m *Manager) CdPlugin(sdk string) (string, error) {
	stdout, _, err := m.Run("cd", "--plugin", sdk)
	return stdout, err
}

func (m *Manager) ParseVersions(output string) []string {
	var versions []string
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "SDK") || strings.HasPrefix(line, "---") {
			continue
		}
		versions = append(versions, line)
	}
	return versions
}

func (m *Manager) ParseInstalled(output string) []string {
	var items []string
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "SDK") || strings.HasPrefix(line, "---") {
			continue
		}
		items = append(items, line)
	}
	return items
}

func (m *Manager) ParseAvailable(output string) []string {
	var items []string
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "SDK") || strings.HasPrefix(line, "---") {
			continue
		}
		items = append(items, line)
	}
	return items
}
