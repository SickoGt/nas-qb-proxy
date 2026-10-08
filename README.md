# fnos-qb-proxy

## What is it?

fnOS 中自带了一个下载器（基于 qBittorrent 和 Aria2），但默认关闭了 WebUI，且采用动态密码。这使得我们无法在外部连接 fnOS 中的 qBittorrent（e.g. 接入 MoviePilot 或 NasTools 等）

该项目是一个简单的代理，能绕过这些限制，提供在外部访问 fnOS 的 qBittorrent 的能力同时不影响 fnOS 自身的下载器运行

## Get Started

### Manual Install

下载 binary 到 fnOS 节点上

```bash
$ wget https://github.com/xxxuuu/fnos-qb-proxy/releases/download/v0.1.0/fnos-qb-proxy_linux-amd64 -O fnos-qb-proxy
$ chmod +x fnos-qb-proxy
```

```bash
$ fnos-qb-proxy -h
NAME:
   fnos-qb-proxy - fnos-qb-proxy is a proxy for qBittorrent in fnOS

USAGE:
   fnos-qb-proxy [global options] command [command options]

COMMANDS:
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --password value, -p value  if not set, qBittorrent will login automatically [%PASSWORD%]
   --debug, -d                 verbose logging (default: false)
   --port value                proxy running port (default: 8080) [%PORT%]
   --username value            qBittorrent WebUI login username (default: "admin") [%QBIT_USERNAME%]
   --process-match value       substring used to find the qBittorrent process in /proc/*/cmdline (default: "qbittorrent-nox") [%QBIT_PROCESS_MATCH%]
   --password-source value     qBittorrent password source: cmdline (--webui-password) or environ (WEBUI_PASSWORD) (default: "cmdline") [%QBIT_PASSWORD_SOURCE%]
   --sock-param value          cmdline parameter name holding the WebUI unix socket path (default: "webui-sock-path") [%QBIT_SOCK_PARAM%]
   --sock-path value           WebUI unix socket path, skips cmdline parsing when set [%QBIT_SOCKET_PATH%]
   --help, -h                  show help
```

运行后，访问 `http://{host}:8080` 即可进入 qBittorrent WebUI。默认情况会自动登录，如果通过 `--password` 指定了密码，则只有该密码可访问；`--port` 修改运行端口

每个参数都可用同名环境变量指定，优先级为 flag > 环境变量 > 默认值。fnOS 保持默认即可；极空间等其他 NAS 通过 `QBIT_PASSWORD_SOURCE`、`QBIT_SOCK_PARAM` 等环境变量适配，说明见 [Docker-Install.md](/Docker-Install.md)

```bash
$ ./fnos-qb-proxy
proxy running on port 8080
```

### Configure Systemd Service

上面的命令会一直在前台运行，可以使用 Systemd 配置成 daemon 在后台自动运行

移动 binary 到 `/usr/bin`

```bash
$ sudo mv fnos-qb-proxy /usr/bin/
```

将以下配置写入到 `/etc/systemd/system/fnos-qb-proxy.service`，可自行修改命令参数

```
[Unit]
Description=fnOS qBittorrent Proxy Service
Before=dlcenter.service

[Service]
ExecStart=/usr/bin/fnos-qb-proxy
Restart=always

[Install]
WantedBy=multi-user.target
```

启用服务

```bash
$ sudo systemctl daemon-reload
$ sudo systemctl enable --now fnos-qb-proxy
```

查看服务状态，成功运行

```bash
$ sudo systemctl status fnos-qb-proxy
● fnos-qb-proxy.service - fnOS qBittorrent Proxy Service
     Loaded: loaded (/etc/systemd/system/fnos-qb-proxy.service; enabled; preset: enabled)
     Active: active (running) since Mon 2024-10-21 23:09:34 CST; 4s ago
   Main PID: 1801543 (fnos-qb-proxy)
      Tasks: 6 (limit: 9495)
     Memory: 6.0M
        CPU: 122ms
     CGroup: /system.slice/fnos-qb-proxy.service
             └─1801543 /usr/bin/fnos-qb-proxy

Oct 21 23:09:34 fnOS systemd[1]: Started fnos-qb-proxy.service - fnOS qBittorrent Proxy Service.
Oct 21 23:09:34 fnOS fnos-qb-proxy[1801543]: proxy running on port 8080
```

### Docker 部署

见 [Docker-Install.md](/Docker-Install.md).

## 极空间（ZSpace）

极空间的下载器同样是 qBittorrent，WebUI 只通过 unix socket 暴露（不监听 TCP 端口），因此本项目同样适用，但有两项配置与 fnOS 的默认值不同：

| 项目 | fnOS（默认值） | 极空间所需配置 |
| --- | --- | --- |
| 进程匹配串 | `trim-qbittorrent-nox` | `qbittorrent-nox`（默认值即可命中） |
| WebUI 口令来源 | 命令行 `--webui-password=`（`cmdline`） | 进程环境变量 `WEBUI_PASSWORD` → `QBIT_PASSWORD_SOURCE=environ` |
| socket 参数 | `--webui-sock-path=` | `--webui-unix-socket=` → `QBIT_SOCK_PARAM=webui-unix-socket` |
| socket 位置 | `/home`、`/usr/trim/var/downloadcenter` | `/dev/shm/qbittorrent.sock` → 容器需挂载 `/dev/shm` |

确认口令可读（返回 `Ok.` 即说明 `WEBUI_PASSWORD` 是可直接登录的明文口令）：

```bash
PID=$(pidof qbittorrent-nox | awk '{print $1}')
PW=$(tr '\0' '\n' < /proc/$PID/environ | sed -n 's/^WEBUI_PASSWORD=//p')
curl -s -i --unix-socket /dev/shm/qbittorrent.sock \
  -X POST -d "username=admin&password=$PW" \
  http://localhost/api/v2/auth/login
```

Docker 部署时的关键配置：

```yaml
environment:
  - PASSWORD=fnosnb
  - QBIT_PASSWORD_SOURCE=environ
  - QBIT_SOCK_PARAM=webui-unix-socket
volumes:
  - /dev/shm:/dev/shm:ro
```

完整配置见 [docker-compose.zspace.yml](./docker-compose.zspace.yml)，参数说明与排错见 [Docker-Install.md](/Docker-Install.md)。
