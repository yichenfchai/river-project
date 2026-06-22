package model

type ChatRequest struct {
	Message  string `json:"message" binding:"required,min=1,max=2000"`
	SessionID string `json:"session_id"`
}

type ChatEvent struct {
	Type  string `json:"type"`  // "token" | "done" | "error"
	Token string `json:"token,omitempty"`
	Index int    `json:"index,omitempty"`
	Done  bool   `json:"done,omitempty"`
	Error string `json:"error,omitempty"`
}

type StoryGenerateRequest struct {
	Topic   string `json:"topic" binding:"required,oneof=history ecology culture legend technology"`
	Keyword string `json:"keyword"`
}

type StoryGenerateResponse struct {
	Title   string `json:"title"`
	Content string `json:"content"`
	ImagePrompt string `json:"image_prompt,omitempty"`
}
