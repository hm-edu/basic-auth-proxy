# basic-auth-proxy

Simple reverse proxy that protects all incoming requests with HTTP Basic Auth.

## Configuration

Create a YAML file (for example `config.yaml`) based on [`config.example.yaml`](./config.example.yaml):

```yaml
listen: ":8080"
forward_to: "http://localhost:3000"
credentials:
  - username: "admin"
    password: "secret"
```

- `listen`: address the proxy server binds to (default: `:8080` if omitted)
- `forward_to`: target base URL to proxy to
- `credentials`: allowed Basic Auth username/password pairs

## Run

```bash
go run . -config config.yaml
```
