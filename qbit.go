package main

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type Qbit struct {
	cfg      Config
	password string
	uds      string
}

func NewQbit(cfg Config) *Qbit {
	return &Qbit{cfg: cfg}
}

func (q *Qbit) GetPassword() (string, error) {
	if err := q.refresh(); err != nil {
		return "", fmt.Errorf("refresh qbittorrent parameters: %w", err)
	}
	return q.password, nil
}

func (q *Qbit) GetUds() (string, error) {
	// uds should not be changed once set
	if q.uds != "" {
		return q.uds, nil
	}
	if err := q.refresh(); err != nil {
		return "", fmt.Errorf("refresh qbittorrent parameters: %w", err)
	}
	return q.uds, nil
}

// refresh checks if the qbittorrent-nox process is running and updates password and uds path
func (q *Qbit) refresh() error {
	fmt.Printf("refreshing qbittorrent-nox parameters\n")
	pid, cmdline, err := findQbittorrentProcess(q.cfg.ProcessMatch)
	if err != nil {
		return fmt.Errorf("find process: %w", err)
	}

	if q.password, err = q.resolvePassword(pid, cmdline); err != nil {
		return fmt.Errorf("get webui-password: %w", err)
	}
	if q.uds, err = q.resolveUds(cmdline); err != nil {
		return fmt.Errorf("get webui socket path: %w", err)
	}
	return nil
}

func (q *Qbit) resolvePassword(pid int, cmdline string) (string, error) {
	switch strings.ToLower(q.cfg.PasswordSource) {
	case "", "cmdline":
		return getCmdParams(cmdline, "webui-password")
	case "environ", "env":
		environ, err := readProcEnviron(pid)
		if err != nil {
			return "", err
		}
		password, ok := getEnvParam(environ, "WEBUI_PASSWORD")
		if !ok {
			return "", fmt.Errorf("WEBUI_PASSWORD not found in process environ")
		}
		return password, nil
	default:
		return "", fmt.Errorf("unknown password source %q, want cmdline or environ", q.cfg.PasswordSource)
	}
}

func (q *Qbit) resolveUds(cmdline string) (string, error) {
	if q.cfg.SockPath != "" {
		return q.cfg.SockPath, nil
	}
	return getCmdParams(cmdline, q.cfg.SockParam)
}

func findQbittorrentProcess(match string) (int, string, error) {
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return 0, "", err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		pid, err := strconv.Atoi(entry.Name())
		if err != nil {
			continue
		}

		content, err := os.ReadFile(fmt.Sprintf("/proc/%d/cmdline", pid))
		if err != nil {
			continue
		}
		// cmdline arguments are separated by null bytes
		cmdline := strings.ReplaceAll(string(content), "\x00", " ")
		if strings.Contains(cmdline, match) {
			return pid, cmdline, nil
		}
	}
	return 0, "", fmt.Errorf("qbittorrent process matching %q not found", match)
}

func readProcEnviron(pid int) (string, error) {
	content, err := os.ReadFile(fmt.Sprintf("/proc/%d/environ", pid))
	if err != nil {
		return "", fmt.Errorf("read process environ: %w", err)
	}
	// environ entries are separated by null bytes
	return strings.ReplaceAll(string(content), "\x00", "\n"), nil
}

func getCmdParams(cmd string, parameter string) (string, error) {
	// parse output(likes --webui-password=xxx) to get
	re := regexp.MustCompile(fmt.Sprintf(`--%s=(\S+)`, regexp.QuoteMeta(parameter)))
	matches := re.FindStringSubmatch(cmd)
	if len(matches) > 1 {
		return matches[1], nil
	}

	return "", fmt.Errorf("parameter --%s not found in process cmdline", parameter)
}

func getEnvParam(environ string, key string) (string, bool) {
	for _, line := range strings.Split(environ, "\n") {
		if value, ok := strings.CutPrefix(line, key+"="); ok {
			return value, true
		}
	}
	return "", false
}
