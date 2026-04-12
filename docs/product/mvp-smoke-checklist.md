# MVP Smoke Checklist

## Run The Stack

Start the full MVP stack from the repository root:

```bash
make up
```

`make up` first compiles Linux binaries into `.docker-bin/` and then builds the Compose images from those artifacts.
The Compose file uses host networking so each service binds directly to its existing Go port on the local machine, including the gateway on `:8080`.

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

## Notes

- The Compose stack enables `IAM_ENABLE_SEED_ADMIN=true` and provides a development `JWT_SECRET` so the seeded login is available without extra setup.
- The agent uses `NODE_SERVICE_URL=http://127.0.0.1:8084` and `NODE_NAME=mvp-node` so it can start alongside the node service inside Compose without relying on Docker bridge DNS.
