# Contributing

## `.env` file

```ini
PRETTY=1
DEBUG=1
REGION=us-east-1
AWS_SERVICE=dynamodb
PROXY_ENDPOINT=http://localhost:8888
CONTROL_PLANE_ADDR=http://localhost:8888
```

To start the server with the env file run `task` (install Taskfile)

in one terminal run
```
bun --watch example_control_plane/index.ts
```

in another run
```
bun example_control_plane/demo_client.ts
```