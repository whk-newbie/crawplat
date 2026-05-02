# MVP Smoke Checklist

## Run The Stack

Start the full MVP stack from the repository root:

```bash
make up
```

`make up` first compiles Linux binaries into `.docker-bin/` and then builds the Compose images from those artifacts.
It also runs `npm --prefix apps/web run build` so the web container serves freshly built static assets.
The Compose stack uses normal service networking, with the gateway exposed on `http://localhost:8080` and the web shell exposed on `http://localhost:3000`.

Stop and clean up the stack when finished:

```bash
make down
```

## Run The Smoke Check

With the stack running, execute:

```bash
bash deploy/scripts/smoke-mvp.sh
```

The script waits for the gateway to begin serving traffic on `http://localhost:8080` and then verifies:

- `GET /api/v1/projects` succeeds through the gateway and returns the empty project list from the fresh MVP stack
- `POST /api/v1/auth/login` succeeds through the gateway using the seeded admin credentials `admin` / `admin123`
- `GET /api/v1/datasources` succeeds through the gateway and returns the empty datasource list from the fresh MVP stack
- `GET /` on the web container succeeds and serves the `Crawler Platform` HTML shell

## Notes

- The Compose stack enables `IAM_ENABLE_SEED_ADMIN=true` and provides a development `JWT_SECRET` so the seeded login is available without extra setup.
- The agent uses `NODE_SERVICE_URL=http://node-service:8084` and `NODE_NAME=mvp-node` so it can resolve the node service over normal Compose DNS.
- Gateway upstreams also resolve over Compose DNS by default; override a service with `GATEWAY_UPSTREAM_<SERVICE>` if needed.
- Monitor alert polling supports separate cadence controls:
  - `MONITOR_ALERT_POLL_INTERVAL` (default `15s`, generic alert checks)
  - `MONITOR_NODE_OFFLINE_ALERT_POLL_INTERVAL` (default `5s`, faster node-offline checks)
