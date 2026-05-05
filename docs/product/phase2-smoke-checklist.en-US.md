# Phase 2 Smoke Checklist

## Prepare Images

Build the sample spider images before bringing the stack up:

```bash
docker build -t crawler/go-echo:latest examples/spiders/go-echo
docker build -t crawler/python-echo:latest examples/spiders/python-echo
```

## Start the Stack

From the repository root:

```bash
make migrate
make up
```

`make migrate` starts PostgreSQL, waits for readiness, and applies the SQL migrations.
`make up` builds the service binaries, builds the web assets, re-runs migrations, and starts the Compose stack.

## Run the Manual Smoke Path

1. Open `http://localhost:3000`.
2. Open `Projects` and create a project.
3. Open `Spiders` and register a Docker spider for that project:
   - Go example image: `crawler/go-echo:latest`
   - Python example image: `crawler/python-echo:latest`
4. Open `Executions` and create a manual execution with the project ID, spider ID, image, and command.
5. Confirm the agent claims the execution and the execution moves from `pending` to `running` to `succeeded`.
6. Open the execution detail page and confirm the logs are visible.

## Expected Runtime Wiring

- PostgreSQL stores project, spider, datasource, and execution metadata.
- Redis stores node liveness and execution queue state.
- MongoDB stores execution logs.
- The agent uses:
  - `NODE_SERVICE_URL=http://node-service:8084`
  - `EXECUTION_SERVICE_URL=http://execution-service:8085`
  - `INTERNAL_API_TOKEN` for execution worker routes
- The agent container now includes the Docker CLI and mounts `/var/run/docker.sock`.

## Validation Evidence

The flow above was exercised on 2026-04-22 in this environment:

```text
Go sample:
  project_id=e7ab72ff-45a6-4e43-9520-bbaa68b92af4
  spider_id=45d697ef-d579-4b9a-8c1f-95b53b8709aa
  execution_id=2b374bd7-7b14-4c6f-9026-2837b1ea4c2e
  status=succeeded
  logs=go spider started / go spider finished

Python sample:
  project_id=bb33593a-d607-4d12-b001-8c98e79afa04
  spider_id=15fb0de2-ff4a-46dd-bc1a-0370851038c2
  execution_id=09adb472-aecc-481f-a2ff-0fa05b6612c3
  status=succeeded
  logs=python spider started / python spider finished
```

That means the current Compose stack, migrations, agent poller, Docker runtime path, and Mongo-backed log retrieval all completed successfully.
