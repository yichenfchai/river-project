package service

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/your-org/grand-canal-guardian/internal/model"
)

const canalSpritePrompt = `你是"运河小精灵"，一只生活在京杭大运河畔的千年水精灵。
你活泼可爱，对大运河的历史、生态、文化了如指掌。
你说话带点古风，喜欢用成语和诗句，但也能和小朋友亲切交流。
你的使命是引导人们了解和保护大运河。
当你不知道答案时，会诚实地说"此事小精灵尚不知晓，待我再去运河里探探"。`

// LLMService LLM 服务接口
type LLMService interface {
	Chat(ctx context.Context, input ChatInput, onToken func(token string) error) error
	GenerateStory(ctx context.Context, input StoryInput) (*StoryOutput, error)
	Health(ctx context.Context) bool
}

type ChatInput struct {
	Message   string
	SessionID string
}

type StoryInput struct {
	Topic   string
	Keyword string
}

type StoryOutput struct {
	Title       string `json:"title"`
	Content     string `json:"content"`
	ImagePrompt string `json:"image_prompt,omitempty"`
}

type llmService struct {
	baseURL     string
	apiKey      string
	model       string
	maxTokens   int
	temperature float64
	client      *http.Client
	logger      *zap.Logger
}

type LLMConfig struct {
	BaseURL     string
	APIKey      string
	Model       string
	Timeout     time.Duration
	MaxTokens   int
	Temperature float64
}

func NewLLMService(cfg LLMConfig, logger *zap.Logger) LLMService {
	return &llmService{
		baseURL:     strings.TrimRight(cfg.BaseURL, "/"),
		apiKey:      cfg.APIKey,
		model:       cfg.Model,
		maxTokens:   cfg.MaxTokens,
		temperature: cfg.Temperature,
		client:      &http.Client{Timeout: cfg.Timeout},
		logger:      logger,
	}
}

func (s *llmService) isConfigured() bool {
	return s.apiKey != ""
}

// ---- Chat (SSE streaming) ----

func (s *llmService) Chat(ctx context.Context, input ChatInput, onToken func(token string) error) error {
	if !s.isConfigured() {
		return s.fallbackChat(input.Message, onToken)
	}

	messages := []map[string]string{
		{"role": "system", "content": canalSpritePrompt},
		{"role": "user", "content": input.Message},
	}

	body := map[string]interface{}{
		"model":       s.model,
		"messages":    messages,
		"stream":      true,
		"max_tokens":  s.maxTokens,
		"temperature": s.temperature,
	}

	reqBody, _ := json.Marshal(body)
	req, err := http.NewRequestWithContext(ctx, "POST", s.baseURL+"/chat/completions", bytes.NewReader(reqBody))
	if err != nil {
		s.logger.Error("构造LLM请求失败", zap.Error(err))
		return s.fallbackChat(input.Message, onToken)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.apiKey)

	resp, err := s.client.Do(req)
	if err != nil {
		s.logger.Warn("LLM API 不可用，启用降级", zap.Error(err))
		return s.fallbackChat(input.Message, onToken)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		s.logger.Warn("LLM API 返回非200", zap.Int("status", resp.StatusCode))
		return s.fallbackChat(input.Message, onToken)
	}

	return s.readSSEStream(resp.Body, onToken)
}

func (s *llmService) readSSEStream(reader io.Reader, onToken func(token string) error) error {
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			return nil
		}

		var chunk struct {
			Choices []struct {
				Delta struct {
					Content string `json:"content"`
				} `json:"delta"`
			} `json:"choices"`
		}
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			continue
		}
		if len(chunk.Choices) > 0 && chunk.Choices[0].Delta.Content != "" {
			if err := onToken(chunk.Choices[0].Delta.Content); err != nil {
				return err
			}
		}
	}
	return scanner.Err()
}

func (s *llmService) fallbackChat(_message string, onToken func(token string) error) error {
	reply := "运河之水，千年流长。小精灵正在运河里玩耍，稍后便回～\n\n如需帮助，请稍后再试，或前往时空地图探索大运河的千年变迁。"
	for _, r := range reply {
		if err := onToken(string(r)); err != nil {
			return err
		}
	}
	return nil
}

// ---- Story Generate ----

func (s *llmService) GenerateStory(ctx context.Context, input StoryInput) (*StoryOutput, error) {
	if !s.isConfigured() {
		return s.fallbackStory(input), nil
	}

	topicDescriptions := map[string]string{
		"history":    "京杭大运河的历史变迁与重大事件",
		"ecology":    "大运河沿岸的生态环境与生物多样性",
		"culture":    "大运河沿线的文化遗产与民俗风情",
		"legend":     "与大运河相关的神话传说与民间故事",
		"technology": "古代运河工程技术与现代保护科技",
	}

	topicDesc := topicDescriptions[input.Topic]
	keywordHint := ""
	if input.Keyword != "" {
		keywordHint = fmt.Sprintf("，请围绕关键词「%s」展开", input.Keyword)
	}

	userPrompt := fmt.Sprintf(
		"请生成一篇关于大运河的科普故事。主题方向：%s%s。要求：标题吸引人、内容生动有趣、适合大众阅读、字数500-800字。请以JSON格式返回，字段为title和content。",
		topicDesc, keywordHint,
	)

	messages := []map[string]string{
		{"role": "system", "content": "你是大运河文化科普专家，擅长用生动有趣的方式讲述运河故事。请严格返回JSON格式。"},
		{"role": "user", "content": userPrompt},
	}

	body := map[string]interface{}{
		"model":       s.model,
		"messages":    messages,
		"stream":      false,
		"max_tokens":  s.maxTokens,
		"temperature": 0.8,
	}

	reqBody, _ := json.Marshal(body)
	req, err := http.NewRequestWithContext(ctx, "POST", s.baseURL+"/chat/completions", bytes.NewReader(reqBody))
	if err != nil {
		return s.fallbackStory(input), nil
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.apiKey)

	resp, err := s.client.Do(req)
	if err != nil {
		s.logger.Warn("LLM 故事生成失败，使用降级", zap.Error(err))
		return s.fallbackStory(input), nil
	}
	defer resp.Body.Close()

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return s.fallbackStory(input), nil
	}

	if len(result.Choices) == 0 {
		return s.fallbackStory(input), nil
	}

	content := result.Choices[0].Message.Content
	content = extractJSON(content)

	var gen model.StoryGenerateResponse
	if err := json.Unmarshal([]byte(content), &gen); err != nil {
		return &StoryOutput{
			Title:   "大运河故事",
			Content: content,
		}, nil
	}

	return &StoryOutput{
		Title:   gen.Title,
		Content: gen.Content,
	}, nil
}

func extractJSON(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "```json") {
		s = strings.TrimPrefix(s, "```json")
		s = strings.TrimSuffix(s, "```")
	} else if strings.HasPrefix(s, "```") {
		s = strings.TrimPrefix(s, "```")
		s = strings.TrimSuffix(s, "```")
	}
	return strings.TrimSpace(s)
}

func (s *llmService) fallbackStory(input StoryInput) *StoryOutput {
	stories := map[string]StoryOutput{
		"history": {
			Title:   "千年运河，流动的史诗",
			Content: "京杭大运河始建于春秋时期，是世界上里程最长、工程最大的古代运河...它南起余杭（今杭州），北到涿郡（今北京），途经今浙江、江苏、山东、河北四省及天津、北京两市，贯通海河、黄河、淮河、长江、钱塘江五大水系，全长约1797公里。大运河对中国南北地区之间的经济、文化发展与交流有着巨大贡献。",
		},
		"ecology": {
			Title:   "运河边的生态家园",
			Content: "大运河不仅是一条黄金水道，更是一条生态走廊。运河两岸绿树成荫，芦苇摇曳，白鹭翩跹。近年来，随着'绿水青山就是金山银山'理念的深入，运河沿线实施了大规模生态修复工程。水质明显改善，鱼类种类从治理前的不足20种恢复到现在的60余种。",
		},
		"culture": {
			Title:   "运河水韵，千年文脉",
			Content: "大运河是流动的文化长廊。沿岸孕育了独具特色的运河文化——扬州园林的精致、苏州评弹的婉转、天津相声的幽默、山东快书的豪迈，无不浸润着运河水韵。运河沿线的非物质文化遗产更是灿若星河。",
		},
		"legend": {
			Title:   "运河龙王的传说",
			Content: "相传大运河开通之初，东海龙王派九子镇守各段河道。其中最小的龙子贪玩好动，常化作白衣少年在运河边与孩童嬉戏。每遇旱涝，他便呼风唤雨，护佑两岸百姓。至今在扬州一带，还保留着'祭小龙王'的民俗活动。",
		},
		"technology": {
			Title:   "古代智慧与现代科技的交响",
			Content: "大运河是古代水利工程的巅峰之作。古人设计了'闸坝体系'来克服地形高差——通过分段筑闸蓄水，使船只逐级翻越。山东南旺分水枢纽被誉为'运河之心'，其设计之精妙至今令人叹服。今天，北斗卫星导航、数字孪生等现代技术正为古老运河注入新的活力。",
		},
	}

	if s, ok := stories[input.Topic]; ok {
		return &s
	}
	story := stories["history"]
	return &story
}

// ---- Health ----

func (s *llmService) Health(ctx context.Context) bool {
	if !s.isConfigured() {
		return false
	}

	req, err := http.NewRequestWithContext(ctx, "GET", s.baseURL+"/models", nil)
	if err != nil {
		return false
	}
	req.Header.Set("Authorization", "Bearer "+s.apiKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return false
	}
	resp.Body.Close()
	return resp.StatusCode == 200
}
