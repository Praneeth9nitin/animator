# animatr

Describe something in plain English. Get a Manim animation.

**animatr** takes a natural language prompt, converts it into a structured animation spec via LLM, generates executable Manim Python code, runs it inside an isolated Docker container, and streams the resulting video back to the browser — automatically retrying and fixing errors along the way.

---

## How it works

```
Prompt → JSON spec (LLM) → Manim Python code (LLM) → Docker execution → .mp4 video
                                                              ↑
                                                    auto-fix on error (up to 3x)
```

1. **Routing** — your prompt is sent to Groq (LLaMA 3.3 70B). The model first converts it into a structured JSON animation spec, then converts that spec into a complete Manim Python script.
2. **Execution** — the script is written to `files/scene.py` and run inside a `manim-runner` Docker container with the project directory bind-mounted.
3. **Error retry** — if the container exits with stderr output, the error is fed back to the LLM to produce a fixed version of the code. This repeats up to 3 times.
4. **Streaming** — progress events are streamed to the frontend over SSE (Server-Sent Events) so you see live status updates as each stage completes.
5. **Playback** — the generated `.mp4` is served directly from the Go server and plays in the browser.

---

## Stack

| Layer | Technology |
|---|---|
| Backend | Go (standard library HTTP) |
| LLM | Groq API — `llama-3.3-70b-versatile` (for json generation) + Gemini 2.0 flash (for code generation)| 
| Animation | Manim Community Edition |
| Isolation | Docker |
| Frontend | Next.js (App Router) |

---

## Prerequisites

- Go 1.21+
- Node.js 18+
- Docker (with the `manim-runner` image built — see below)
- A [Groq API key](https://console.groq.com)
- A [Google AI Studio API key](https://aistudio.google.com/api-keys)

---

## Setup

### 1. Build the Manim Docker image

```bash
docker build -t manim-runner .
```

Your `Dockerfile` should install Manim Community Edition and set the working directory to `/app`.

### 2. Configure environment

Create a `.env` file in the `goserver/` directory:

```env
GROQ_API_KEY=your_key_here
```

### 3. Start the Go server

```bash
cd goserver
go run .
```

Server runs on `http://localhost:8080`.

### 4. Start the Next.js frontend

```bash
cd frontend
npm install
npm run dev
```

Frontend runs on `http://localhost:3000`.

---

## API

### `POST /generate`

Accepts a JSON body, streams SSE events back.

**Request**
```json
{ "prompt": "Visualise a Fourier transform decomposing a square wave" }
```

**SSE event types**

| Type | Fields | Description |
|---|---|---|
| `progress` | `message` | Status update (e.g. "Generating Manim code...") |
| `done` | `video_url` | Relative URL to the rendered `.mp4` |
| `error` | `message` | Generation failed after all retries |

### `GET /video/<path>`

Serves the generated video file. The path comes directly from the `video_url` field in the `done` event.

---

## Project structure

```
.
├── goserver/
│   ├── main.go          # Entry point — calls StartServer()
│   ├── server.go        # HTTP handlers, SSE streaming, video serving
│   ├── route.go         # LLM chain: prompt → JSON spec → Manim code
│   ├── fix.go           # Error repair: feeds stderr back to LLM
│   ├── container.go     # Docker execution, bind mount, log capture
│   └── .env             # GROQ_API_KEY and GEMINI_API_KEY
├── frontend/
│   └── app/
│       └── page.tsx     # Single-page Next.js UI
├── files/
│   ├── scene.py         # Generated Manim script (overwritten each run)
│   └── media/           # Manim output directory (auto-created)
└── Dockerfile           # manim-runner image
```

---

## Example prompts

- `Show a circle transforming into a square while its color shifts from blue to red`
- `Animate a binary search tree inserting the values 5, 3, 7, 1, 4`
- `Visualise how a Fourier transform decomposes a square wave into sine waves`
- `Draw a vector field and show a particle following the gradient`

---

## Notes

- The Manim class in every generated script is always named `SceneName`. This is hardcoded in both the generation prompt and the Docker command.
- Generated videos are not cleaned up automatically. Clear `files/media/` periodically if disk space is a concern.
- The LLM sometimes produces valid Python that uses deprecated or non-existent Manim methods. The auto-fix loop handles most of these cases within 1–2 retries.