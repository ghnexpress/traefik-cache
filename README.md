### Local Mode

Traefik also offers a developer mode that can be used for temporary testing of plugins not hosted on GitHub.
To use a plugin in local mode, the Traefik static configuration must define the module name (as is usual for Go packages) and a path to a [Go workspace](https://golang.org/doc/gopath_code.html#Workspaces), which can be the local GOPATH or any directory.

The plugins must be placed in `./plugins-local` directory,
which should be in the working directory of the process running the Traefik binary.
The source code of the plugin should be organized as follows:

```
 ├── docker-compose.yml
 └── plugins-local
    └── src
        └── github.com
            └── ghnexpress
                └── traefik-cache
                    ├── main.go
                    ├── vendor
                    ├── go.mod
                    └── ...

```

```yaml
# docker-compose.yml
version: "3.6"

services:
  memcached:
    image: launcher.gcr.io/google/memcached1
    container_name: some-memcached
    ports:
      - "11211:11211"
    networks:
      - traefik-network
  traefik:
    image: traefik:v2.9.6 #v3.0.0-beta2
    container_name: traefik
    depends_on:
    - memcached
    command:
      # - --log.level=DEBUG
      - --api
      - --api.dashboard
      - --api.insecure=true
      - --providers.docker=true
      - --entrypoints.web.address=:80
      - --experimental.localPlugins.plugindemo.moduleName=github.com/ghnexpress/traefik-cache
    ports:
      - "80:80"
      - "8080:8080"
    networks:
      - traefik-network
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
      - ./plugins-local/src/github.com/ghnexpress/traefik-cache:/plugins-local/src/github.com/ghnexpress/traefik-cache
    labels:
      - traefik.http.middlewares.my-plugindemo.plugin.plugindemo.hashkey.header.enable=true
      - traefik.http.middlewares.my-plugindemo.plugin.plugindemo.hashkey.header.fields=Token,User-Agent
      - traefik.http.middlewares.my-plugindemo.plugin.plugindemo.hashkey.header.ignoreFields=X-Request-Id,Postman-Token,Content-Length
      - traefik.http.middlewares.my-plugindemo.plugin.plugindemo.hashkey.body.enable=false
      - traefik.http.middlewares.my-plugindemo.plugin.plugindemo.hashkey.method.enable=true
      - traefik.http.middlewares.my-plugindemo.plugin.plugindemo.memcached.address=some-memcached:11211
      - traefik.http.middlewares.my-plugindemo.plugin.plugindemo.alert.telegram.chatId=-795576798
      - traefik.http.middlewares.my-plugindemo.plugin.plugindemo.alert.telegram.token=xxx
      - traefik.http.middlewares.my-plugindemo.plugin.plugindemo.env=dev
      - traefik.http.middlewares.my-plugindemo.plugin.plugindemo.forceCache.enable=true
      - traefik.http.middlewares.my-plugindemo.plugin.plugindemo.forceCache.expiredTime=10

  whoami:
    image: traefik/whoami
    container_name: simple-service
    depends_on:
    - traefik
    networks:
      - traefik-network
    labels:
      - traefik.enable=true
      - traefik.http.routers.whoami.rule=Host(`localhost`)
      - traefik.http.routers.whoami.entrypoints=web
      - traefik.http.routers.whoami.middlewares=my-plugindemo
networks:
  traefik-network:
    driver: bridge
```

### K8s

```yaml
# cache-middleware.yaml
apiVersion: traefik.containo.us/v1alpha1
kind: Middleware
metadata:
  annotations: {}
  name: ghn-cache
  namespace: default
spec:
  plugin:
    plugin-cache:
      memcached:
        address: xxx:11211
      hashkey:
        body:
          enable: true
        header:
          enable: true
          fields: Token,User-Agent
          ignoreFields: X-Request-Id,Postman-Token,Content-Length
        method:
          enable: true
      alert:
        telegram:
          chatId: -795576798
          token: xxx
      env: dev
      forceCache:
        enable: true
        expiredTime: 100 #second
```

