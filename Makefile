ANTHROPIC_API_KEY := $(shell echo $$ANTHROPIC_API_KEY)

build:
	@mkdir -p ./dist
	tinygo build -target wasi -o ./dist/assembllm-anthropic-go.wasm main.go

test:
	extism call ./dist/assembllm-anthropic-go.wasm models --wasi --log-level info
	@extism call ./dist/assembllm-anthropic-go.wasm completion --input "tell me a joke" --wasi --allow-host "api.anthropic.com" --log-level info --set-config='{"api_key": "$(ANTHROPIC_API_KEY)"}'
	@extism call ./dist/assembllm-anthropic-go.wasm completionWithTools --input '{"tools": [{"name": "get_weather","description": "Get the current weather in a given location","input_schema": {"type": "object","properties": {"location": {"type": "string","description": "The city and state, e.g. San Francisco, CA"},"unit": {"type": "string","enum": ["celsius"],"description": "The unit of temperature, always celsius"}},"required": ["location", "unit"]}}],"messages": [{"role": "user","content": "What is the weather like in San Francisco?"}]}' --wasi --allow-host "api.anthropic.com" --log-level info --set-config='{"api_key": "$(ANTHROPIC_API_KEY)"}'
	@extism call ./dist/assembllm-anthropic-go.wasm completionWithTools --input '{"tools": [{"name": "print_sentiment_scores","description": "Prints the sentiment scores of a given text.","input_schema": {"type": "object","properties": {"positive_score": {"type": "number","description": "The positive sentiment score, ranging from 0.0 to 1.0."},"negative_score": {"type": "number","description": "The negative sentiment score, ranging from 0.0 to 1.0."},"neutral_score": {"type": "number","description": "The neutral sentiment score, ranging from 0.0 to 1.0."}},"required": ["positive_score", "negative_score", "neutral_score"]}}],"messages": [{"role": "user","content": "Im a HUGE hater of pickles.  I actually despise pickles.  They are garbage."}]}' --wasi --allow-host "api.anthropic.com" --log-level info --set-config='{"api_key": "$(ANTHROPIC_API_KEY)"}'