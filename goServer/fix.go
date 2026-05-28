package main

func FixCode(code, stderr string) string {
	prompt := `You are a Manim code repair specialist.

The following Python script failed when run with: manim -pql script.py SceneName

Rules:
- Return ONLY raw Python code. No markdown, no backticks, no explanations.
- The class must still be named exactly: SceneName
- Start directly from "from manim import *"
- Fix ONLY what caused the error. Do not rewrite unrelated parts.
- If the error is a missing import, add it. If it's a method that does not exist in Manim, replace it with the correct one.

Broken code:
` + code + `

Error output:
` + stderr + `

Return ONLY the fixed Python code.
`

	return MakeRequestForCode(prompt)
}
