# NimsChatWidget Runbook

Operational procedures for the nimschatwidget service -- a thin iframe launcher that serves widget JS for embedding the webchat.

## Architecture

The service is minimal: a single HTTP server that serves a JavaScript file at `GET /widget`. The JS creates a floating button that opens an iframe to the webchat's embed mode. No NATS, no database, no message routing -- all chat logic lives in the webchat iframe.

## Deployment

nimschatwidget runs as a Docker container on the NimsForest land server (46.225.164.179), managed by land.

- **Port**: 8096
- **Config**: `/opt/nimschatwidget/config.yaml`
- **Domain**: `chatwidget.nimsforest.mynimsforest.com`
- **Container image**: from GitHub Container Registry, deployed via land

To deploy a new version:
1. Tag and push: `git tag v0.X.0 && git push origin v0.X.0`
2. GitHub Actions builds and pushes the Docker image
3. On the land server: `land plant --config /etc/land.yaml`

Never deploy manually with `docker run`. Always go through land.

## Configuration

```yaml
server:
  addr: ":8096"
webchat_url: "https://webchat.nimsforest.mynimsforest.com"
```

The `webchat_url` field is currently unused by the server (the widget JS reads `webchatURL` from `window.nimschatwidgetConfig` at runtime on the host page). It is kept in config for documentation purposes.

To update config, update the role seed in the landconfigregistry repo (`/home/claude-user/landconfigregistry`), not the file on the server directly.

## Health check

```bash
curl -s https://chatwidget.nimsforest.mynimsforest.com/health
# Expected: {"status":"ok"}
```

## Common operations

### Verify widget is serving

```bash
curl -sI https://chatwidget.nimsforest.mynimsforest.com/widget
# Should return 200 with Content-Type: application/javascript
```

### Check container status

```bash
ssh root@46.225.164.179 "docker ps | grep nimschatwidget"
```

### View logs

```bash
ssh root@46.225.164.179 "docker logs nimschatwidget --tail 50"
```

### Restart container

```bash
ssh root@46.225.164.179 "land plant --config /etc/land.yaml"
```

## Troubleshooting

### Widget button does not appear

1. Check the host page includes the script tag: `<script src="https://chatwidget.nimsforest.mynimsforest.com/widget"></script>`
2. Check browser console for JS errors
3. Check if `ncw-root` element already exists (double-init protection)

### Iframe does not load

1. Verify `window.nimschatwidgetConfig.webchatURL` is set on the host page
2. Check browser console for `[nimschatwidget] webchatURL not configured`
3. Verify the webchat URL is reachable: `curl -sI https://webchat.nimsforest.mynimsforest.com/embed`
4. Check for mixed-content blocking (host page must be HTTPS if webchat is HTTPS)

### CORS errors

The widget endpoint serves `Access-Control-Allow-Origin: *`, so CORS should not be an issue for the script itself. If the iframe has CORS issues, that is a webchat configuration problem, not a widget problem.

### Container not starting

1. Check land config: `ssh root@46.225.164.179 "cat /etc/land.yaml"`
2. Check if port 8096 is already in use: `ssh root@46.225.164.179 "ss -tlnp | grep 8096"`
3. Check container logs for startup errors
