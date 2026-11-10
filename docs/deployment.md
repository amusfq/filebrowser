## Finder fork: Docker and Dokploy

The workflow in `.github/workflows/ci.yaml` tests and builds this fork. Pushes to
`master`, or a manual run on `master`, publish these ARM64 images for OCI:

- `ghcr.io/amusfq/filebrowser:latest`
- `ghcr.io/amusfq/filebrowser:sha-<12-character-commit>` for rollback

After publication succeeds, the workflow posts to the deployment webhook stored
in the GitHub Actions secret `DOKPLOY_WEBHOOK_URL`. Use the compose deployment
webhook from Dokploy, rather than a notification webhook. Runtime settings,
registry credentials, domains, and persistent mounts stay in Dokploy. A private
GHCR image requires a registry credential with `read:packages` in Dokploy.
Set the GitHub Actions variable `DOKPLOY_ORIGIN_IP` to the Dokploy server IP.
The webhook connects directly to that origin while retaining the URL hostname
and TLS certificate verification, avoiding Cloudflare browser challenges in CI.

For a separate Docker installation, run `docker compose up -d` with `compose.yaml`.
It persists served files at `/srv`, configuration at `/config`, and the database
at `/database`. Preserve the existing mounts when upgrading an installation.

The Dockerfile expects a prebuilt static Linux binary in the build context, as
the release workflow does. To build the image locally on ARM64:

```sh
cd frontend
pnpm install --frozen-lockfile
pnpm run build
cd ..
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o filebrowser .
docker build -t ghcr.io/amusfq/filebrowser:local .
```

## Self-Registration (Signup)

File Browser allows you to enable user self-registration (signup). This can be enabled via **Settings → Global Settings**, or with `filebrowser config set --signup`. Self-registered users inherit the configured **user defaults**, including the scope.

> [!WARNING]
>
> By default, the user scope is the server's root, so a self-registered user could read,
> modify, and delete every file File Browser serves. To prevent this, either:
>
> a. Enable `createUserDir` so each user gets their own directory; or
> b. If users are meant to share files, set the default scope to something other than the root.

## Fail2ban

File Browser does not natively support protection against brute force attacks. Therefore, we suggest using something like [fail2ban](https://github.com/fail2ban/fail2ban), which takes care of that by tracking the logs of your File Browser instance. For more information on how fail2ban works, please refer to their [wiki](https://github.com/fail2ban/fail2ban/wiki).

### Filter Configuration

An example filter configuration targeted at matching File Browser's logs.

```ini
[INCLUDES]
before = common.conf

[Definition]
datepattern = `^%%Y\/%%m\/%%d %%H:%%M:%%S`
failregex   = `\/api\/login: 403 <HOST> *`
```

### Jail Configuration

An example jail configuration. You should fill it with the path of the logs of File Browser, as well as the port where it is running at.

```ini
[filebrowser]

enabled = true
port = [your_port]
filter = filebrowser
logpath = [your_log_path]
maxretry = 10
bantime = 10m
findtime = 10m
banaction = iptables-allports
banaction_allports = iptables-allports
```
