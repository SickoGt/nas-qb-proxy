package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/urfave/cli/v2"
)

type Config struct {
	Debug            bool
	Port             int
	ExpectedPassword string
	Username         string
	ProcessMatch     string
	PasswordSource   string
	SockParam        string
	SockPath         string
}

func configFromContext(ctx *cli.Context) Config {
	cfg := Config{
		Debug:            ctx.Bool("debug"),
		Port:             ctx.Int("port"),
		ExpectedPassword: ctx.String("password"),
		Username:         ctx.String("username"),
		ProcessMatch:     ctx.String("process-match"),
		PasswordSource:   ctx.String("password-source"),
		SockParam:        ctx.String("sock-param"),
		SockPath:         ctx.String("sock-path"),
	}
	// env vars set to an empty string bypass flag defaults
	if cfg.Username == "" {
		cfg.Username = "admin"
	}
	if cfg.ProcessMatch == "" {
		cfg.ProcessMatch = "qbittorrent-nox"
	}
	if cfg.PasswordSource == "" {
		cfg.PasswordSource = "cmdline"
	}
	if cfg.SockParam == "" {
		cfg.SockParam = "webui-sock-path"
	}
	return cfg
}

func proxyCmd(ctx *cli.Context) error {
	cfg := configFromContext(ctx)
	proxy := NewFnosProxy(cfg)
	fmt.Printf("proxy running on port %d\n", cfg.Port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", cfg.Port), proxy)
	if err != nil {
		return fmt.Errorf("listen and serve: %w", err)
	}
	return nil
}

func main() {
	app := &cli.App{
		Name:   "fnos-qb-proxy",
		Usage:  "fnos-qb-proxy is a proxy for qBittorrent in fnOS",
		Action: proxyCmd,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "password",
				Aliases: []string{"p"},
				EnvVars: []string{"PASSWORD"},
				Usage:   "if not set, qBittorrent will login automatically",
				Value:   "",
			},
			&cli.BoolFlag{
				Name:    "debug",
				Aliases: []string{"d"},
				Usage:   "verbose logging",
				Value:   false,
			},
			&cli.IntFlag{
				Name:    "port",
				EnvVars: []string{"PORT"},
				Usage:   "proxy running port",
				Value:   8080,
			},
			&cli.StringFlag{
				Name:    "username",
				EnvVars: []string{"QBIT_USERNAME"},
				Usage:   "qBittorrent WebUI login username",
				Value:   "admin",
			},
			&cli.StringFlag{
				Name:    "process-match",
				EnvVars: []string{"QBIT_PROCESS_MATCH"},
				Usage:   "substring used to find the qBittorrent process in /proc/*/cmdline",
				Value:   "qbittorrent-nox",
			},
			&cli.StringFlag{
				Name:    "password-source",
				EnvVars: []string{"QBIT_PASSWORD_SOURCE"},
				Usage:   "qBittorrent password source: cmdline (--webui-password) or environ (WEBUI_PASSWORD)",
				Value:   "cmdline",
			},
			&cli.StringFlag{
				Name:    "sock-param",
				EnvVars: []string{"QBIT_SOCK_PARAM"},
				Usage:   "cmdline parameter name holding the WebUI unix socket path",
				Value:   "webui-sock-path",
			},
			&cli.StringFlag{
				Name:    "sock-path",
				EnvVars: []string{"QBIT_SOCKET_PATH"},
				Usage:   "WebUI unix socket path, skips cmdline parsing when set",
				Value:   "",
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
