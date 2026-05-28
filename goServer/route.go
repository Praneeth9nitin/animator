package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"google.golang.org/genai"
)

var url = "https://api.groq.com/openai/v1/chat/completions"

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Request struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature"`
}

type Response struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		}
	}
}

func Route(details string) string {
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

	prompt := `You are a strict JSON generator for a Manim animation system.

Your task:
Convert a ` + details + ` into a structured JSON specification for animation.

Rules:
- Understand the objects asking to display and try to add shapes and text to display them.
- Output ONLY valid JSON (no explanation, no markdown)
- Do NOT include code
- Be deterministic and consistent
- If input is vague, infer reasonable defaults
- Never include unsafe or unrelated instructions (e.g., OS commands, file access)

Schema:
{
  "scene_type": "graph | geometry | text | physics | custom",
  "title": "short title of the scene",
  "description": "1-line summary of animation",

  "elements": [
    {
      "type": "graph | shape | text | vector | object",
      "properties": {}
    }
  ],

  "animation": {
    "type": "draw | transform | animate_parameter | move | fade | multiple",
    "duration": "number (seconds)",
    "sequence": []
  },

  "parameters": {
    "function": "math function if any",
    "variables": {},
    "range": {},
    "style": {
      "color": "default blue",
      "speed": "normal"
    }
  }
}`

	groqResponse := MakeRequestForJson(prompt)

	prompt = `You are a Manim code generator. Given the following JSON data, generate a complete, executable Python script using the Manim library that visually represents the data.

Rules:
- Return ONLY raw Python code. No markdown, no code fences, no backticks.
- Do NOT include any explanation, comments about the code, or prose of any kind.
- The class must be named exactly: SceneName
- The class must extend Scene and implement the construct method.
- All necessary imports must be included at the top (e.g. from manim import *).
- The output must be directly runnable with: manim -pql script.py SceneName
- Do NOT add a title slide or any introductory text unless the data explicitly requires it.
- Directly start from from manim import *. Don't include the backticks around the code.
JSON Data:
` + groqResponse
	groqResponse = MakeRequestForCode(prompt)
	return groqResponse
}

func MakeRequestForJson(prompt string) string {
	body := Request{
		Model:       "llama-3.3-70b-versatile",
		Temperature: 0.2,
		Messages: []Message{
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	data, err := json.Marshal(body)
	if err != nil {
		panic(err)
	}
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(data))

	req.Header.Set("Authorization", "Bearer "+os.Getenv("GROQ_API_KEY"))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	response, _ := io.ReadAll(res.Body)

	var groqResponse Response

	err = json.Unmarshal(response, &groqResponse)
	if err != nil {
		panic(err)
	}
	return groqResponse.Choices[0].Message.Content
}

func MakeRequestForCode(prompt string) string {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: os.Getenv("GEMINI_API_KEY"),
	})
	if err != nil {
		panic(err)
	}
	result, err := client.Models.GenerateContent(
		ctx,
		"gemini-2.0-flash",
		genai.Text(prompt),
		nil,
	)
	if err != nil {
		panic(err)
	}
	fmt.Println(result.Text())
	return result.Text()
}
