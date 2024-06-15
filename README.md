# assembllm Anthropic Go Plug-in

Plugin for [assembllm](https://github.com/bradyjoslin/assembllm)

Requires [tinygo](https://tinygo.org/).

Uses [Extism PDK](https://github.com/extism/go-pdk). Requires WASI.

## Building

```bash
make build
```

Built wasm file will be in `dist/assembllm-anthropic-go.wasm`.

## Testing

```bash
make test
```

## Add to assembllm

Sample configuration update to `~/.assembllm/config.yaml`:

```yaml
  - name: anthropic
    source: https://github.com/bradyjoslin/assembllm-anthropic-go/releases/latest/download/assembllm-anthropic-go.wasm
    hash: 
    apiKey: ANTHROPIC_API_KEY
    url: api.anthropic.com
    model: claude-3-opus-20240229
    wasi: true
```
