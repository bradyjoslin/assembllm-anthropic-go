ANTHROPIC_API_KEY := $(shell echo $$ANTHROPIC_API_KEY)

build:
	@mkdir -p ./dist
	tinygo build -target wasi -o ./dist/assembllm-anthropic-go.wasm main.go

test:
	extism call ./dist/assembllm-anthropic-go.wasm models --wasi --log-level info
	@extism call ./dist/assembllm-anthropic-go.wasm completion --input "tell me a joke" --wasi --allow-host "api.anthropic.com" --log-level info --set-config='{"api_key": "$(ANTHROPIC_API_KEY)"}'