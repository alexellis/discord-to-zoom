
Local iteration:

```
DOCKER_BUILDKIT=1 faas-cli build --build-arg GO111MODULE=off && \
    OPENFAAS_EXPERIMENTAL=1 faas-cli local-run discord-start-zoom --port 8082
```
