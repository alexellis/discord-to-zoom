# Deploy to OpenFaaS

Create the various required secrets:

```bash
export OPENFAAS_URL="https://derek.openfaas.com"
faas-cli secret create \
    discord-bot-token \
    --from-file .secrets/discord-bot-token

faas-cli secret create \
    discord-public-key \
    --from-file .secrets/discord-public-key

faas-cli secret create \
    zoom-account-id \
    --from-file .secrets/zoom-account-id

faas-cli secret create \
    zoom-client-secret \
    --from-file .secrets/zoom-client-secret

faas-cli secret create \
    zoom-client-id \
    --from-file .secrets/zoom-client-id
```

Copy `example.discord_config.yaml` to `discord_config.yaml` and edit the values.

Then copy `example.zoom_config.yaml` to `zoom_config.yaml` and edit the values.

Then deploy:

```bash
faas-cli deploy
```

# Test locally with an inlets tunnel

For local builds:

```
DOCKER_BUILDKIT=1 faas-cli build --build-arg GO111MODULE=off && \
    OPENFAAS_EXPERIMENTAL=1 faas-cli local-run discord-start-zoom --port 8082
```

Then point the tunnel to 127.0.0.1:8002.

