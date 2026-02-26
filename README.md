# fal-cli

A CLI for the [fal.ai](https://fal.ai) API — run generative AI models, manage queue requests, and browse the model catalog.

## Install

```bash
git clone https://github.com/the20100/fal-cli
cd fal-cli
go build -o fal .
mv fal /usr/local/bin/fal
```

## Authentication

Get your API key at https://fal.ai/dashboard/keys, then:

```bash
fal auth set-key your_api_key_here
# or
export FAL_KEY=your_api_key_here
```

Key resolution order:
1. `FAL_KEY` env var
2. Config file (`fal auth set-key`)

Config is stored at:
- macOS: `~/Library/Application Support/fal/config.json`
- Linux: `~/.config/fal/config.json`
- Windows: `%AppData%\fal\config.json`

## Commands

### Shortcut: generate images

```bash
# nano-banana-pro — text-to-image ($0.15/image)
fal generate "a cat wearing a hat"
fal generate "golden gate bridge at sunset" --aspect 16:9
fal generate "portrait of a woman" --resolution 2K --num 2
fal generate "futuristic city" --format webp --seed 42
fal generate "latest AI news illustration" --web-search
fal generate "epic artwork" --resolution 4K --queue --logs
```

### Shortcut: edit images

```bash
# nano-banana-pro/edit — image-to-image ($0.15/image)
fal edit "make it night time" --image https://example.com/photo.jpg
fal edit "add snow" --image https://example.com/city.jpg --aspect 16:9
fal edit "remove background" --image https://example.com/portrait.jpg
fal edit "composite together" --image https://img1.jpg --image https://img2.jpg
```

### Run any model

```bash
# Synchronous
fal run fal-ai/nano-banana-pro --input '{"prompt":"a cat"}'
fal run fal-ai/flux/dev --input '{"prompt":"a cat","image_size":"landscape_4_3"}'

# Queue + poll until done
fal run fal-ai/flux/dev --input '{"prompt":"a cat"}' --queue
fal run fal-ai/flux/schnell --input '{"prompt":"a cat"}' --queue --logs
```

### Queue management

```bash
# Submit, then manage separately
fal queue status fal-ai/flux/dev <request-id>
fal queue status fal-ai/flux/dev <request-id> --logs
fal queue result fal-ai/flux/dev <request-id>
fal queue cancel fal-ai/flux/dev <request-id>
fal queue poll   fal-ai/flux/dev <request-id> --logs
```

### Model catalog

```bash
fal models list
fal models list --category text-to-image
fal models list --search "flux" --limit 10
fal models pricing fal-ai/nano-banana-pro
fal models pricing fal-ai/flux/dev fal-ai/flux/schnell
```

### Auth

```bash
fal auth set-key <api-key>
fal auth status
fal auth logout
```

### Info

```bash
fal info   # config path, key source, env vars
```

## Global flags

| Flag | Description |
|------|-------------|
| `--json` | Force JSON output |
| `--pretty` | Force pretty-printed JSON output |

Output is **auto-detected**: JSON when stdout is piped, human-readable in a terminal.

## Scripting

```bash
# Extract image URL with jq
fal generate "a cat" --json | jq -r '.images[0].url'

# Batch generate
for prompt in "a cat" "a dog" "a bird"; do
  fal generate "$prompt" --json | jq -r '.images[0].url'
done

# Submit to queue, capture request ID, poll later
fal run fal-ai/flux/dev --input '{"prompt":"a cat"}' --queue 2>&1 | grep "Queued:" | awk '{print $2}'
fal queue poll fal-ai/flux/dev <request-id>
```

## generate / edit flags

| Flag | Default | Description |
|------|---------|-------------|
| `--aspect` | `1:1` / `auto` | `21:9 16:9 3:2 4:3 5:4 1:1 4:5 3:4 2:3 9:16 auto` |
| `--resolution` | `1K` | `1K`, `2K`, `4K` (4K = 2× price) |
| `--num` | `1` | Images to generate (1–4) |
| `--format` | `png` | `jpeg`, `png`, `webp` |
| `--safety` | `4` | `1` (strictest) → `6` (least strict) |
| `--seed` | `0` | Random seed (0 = random) |
| `--web-search` | off | Web search grounding (+$0.015/image) |
| `--google-search` | off | Google search grounding |
| `--queue` | off | Use queue instead of sync |
| `--logs` | off | Show model logs while polling |
