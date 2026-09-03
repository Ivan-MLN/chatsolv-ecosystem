package agentconfig

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type GeneratedSetup struct {
	Profile struct {
		DisplayName     string `json:"display_name"`
		Description     string `json:"description"`
		GreetingMessage string `json:"greeting_message"`
		AwayMessage     string `json:"away_message"`
		FallbackMessage string `json:"fallback_message"`
		Language        string `json:"language"`
	} `json:"profile"`
	Personality struct {
		BotName            string   `json:"bot_name"`
		Role               string   `json:"role"`
		Tone               string   `json:"tone"`
		CommunicationStyle string   `json:"communication_style"`
		PrimaryLanguage    string   `json:"primary_language"`
		ResponseLength     string   `json:"response_length"`
		EmojiUsage         string   `json:"emoji_usage"`
		GreetingStyle      string   `json:"greeting_style"`
		ClosingStyle       string   `json:"closing_style"`
		CustomInstructions string   `json:"custom_instructions"`
		BehaviorRules      []string `json:"behavior_rules"`
		EscalationRules    []string `json:"escalation_rules"`
		ForbiddenTopics    []string `json:"forbidden_topics"`
		FallbackBehavior   string   `json:"fallback_behavior"`
	} `json:"personality"`
	Business struct {
		BusinessName        string   `json:"business_name"`
		Industry            string   `json:"industry"`
		BusinessDescription string   `json:"business_description"`
		Website             string   `json:"website"`
		Address             string   `json:"address"`
		Timezone            string   `json:"timezone"`
		BrandVoice          string   `json:"brand_voice"`
		CompanyValues       []string `json:"company_values"`
	} `json:"business"`
}

type AIGenerator struct {
	client *http.Client
}

func NewAIGenerator() *AIGenerator {
	return &AIGenerator{
		client: &http.Client{Timeout: 60 * time.Second},
	}
}

func (g *AIGenerator) GenerateSetup(ctx context.Context, userDescription string) (GeneratedSetup, error) {
	userDescription = strings.TrimSpace(userDescription)
	if userDescription == "" {
		return GeneratedSetup{}, errors.New("user description is required")
	}

	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		apiKey = "sk-f0af8e27b8b51cfb-y3w711-be371d84"
	}
	baseURL := os.Getenv("OPENROUTER_BASE_URL")
	if baseURL == "" {
		baseURL = "http://127.0.0.1:20128/v1"
	} else if strings.Contains(baseURL, "bansos.naeladtya.my.id") {
		baseURL = "http://127.0.0.1:20128/v1"
	}
	baseURL = strings.TrimRight(baseURL, "/")

	model := os.Getenv("HERMES_INFERENCE_MODEL")
	if model == "" {
		model = "ag/gemini-3.7-flash-high"
	}

	systemPrompt := `Anda adalah AI System Architect & Prompt Engineer ahli untuk ChatSolv (Platform Customer Service AI Otomatis).
Tugas Anda adalah membaca deskripsi bisnis dan kebutuhan customer service dari pengguna, lalu secara otomatis menghasilkan konfigurasi lengkap dalam format JSON yang valid.

Field yang HARUS diisi sesuai skema berikut:
{
  "profile": {
    "display_name": "Nama tampilan agent di chat",
    "description": "Deskripsi tugas dan peran agent",
    "greeting_message": "Pesan sapaan pembuka customer",
    "away_message": "Pesan saat di luar jam kerja",
    "fallback_message": "Pesan saat mengalihkan ke CS manusia",
    "language": "id"
  },
  "personality": {
    "bot_name": "Nama bot",
    "role": "Peran spesifik (misal: Customer Support Specialist)",
    "tone": "friendly | professional | warm | neutral | formal | casual",
    "communication_style": "conversational | casual_professional | formal | concise",
    "primary_language": "id",
    "response_length": "short | medium | detailed",
    "emoji_usage": "none | minimal | moderate",
    "greeting_style": "Sapaan singkat max 35 karakter (contoh: Halo! Ada yang bisa dibantu? 😊)",
    "closing_style": "Penutup singkat max 35 karakter (contoh: Terima kasih! Ada lagi?)",
    "custom_instructions": "Instruksi prompt mendalam untuk AI agar memahami knowledge base, cara menangani order, komplain, dll.",
    "behavior_rules": ["Aturan perilaku 1", "Aturan 2"],
    "escalation_rules": ["Aturan kapan harus dialihkan ke manusia 1", "Aturan 2"],
    "forbidden_topics": ["Topik terlarang 1", "Topik 2"],
    "fallback_behavior": "direct_to_human"
  },
  "business": {
    "business_name": "Nama Bisnis / Toko / Brand",
    "industry": "Kategori Industri (misal: E-commerce Fashion, Kuliner, SaaS)",
    "business_description": "Deskripsi lengkap tentang produk dan layanan yang ditawarkan",
    "website": "https://example.com",
    "address": "Alamat atau domisili bisnis",
    "timezone": "Asia/Jakarta",
    "brand_voice": "Karakter suara brand",
    "company_values": ["Nilai 1", "Nilai 2"]
  }
}

PENTING: Berikan HANYA format JSON valid tanpa markdown backticks atau penjelasan tambahan.`

	reqBody, _ := json.Marshal(map[string]any{
		"model": model,
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": "Berikut deskripsi bisnis dan customer service yang saya inginkan:\n" + userDescription},
		},
		"temperature": 0.4,
		"stream":      false,
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/chat/completions", bytes.NewReader(reqBody))
	if err != nil {
		return GeneratedSetup{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := g.client.Do(req)
	if err != nil {
		return GeneratedSetup{}, fmt.Errorf("call gemini: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return GeneratedSetup{}, err
	}

	if resp.StatusCode/100 != 2 {
		return GeneratedSetup{}, fmt.Errorf("gemini api error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	var completion struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err = json.Unmarshal(bodyBytes, &completion); err != nil || len(completion.Choices) == 0 {
		return GeneratedSetup{}, fmt.Errorf("invalid gemini response: %s", string(bodyBytes))
	}

	content := strings.TrimSpace(completion.Choices[0].Message.Content)
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	var result GeneratedSetup
	if err = json.Unmarshal([]byte(content), &result); err != nil {
		return GeneratedSetup{}, fmt.Errorf("parse generated json: %w (raw: %s)", err, content)
	}

	// Sanitize length restrictions
	if len(result.Personality.GreetingStyle) > 40 {
		result.Personality.GreetingStyle = result.Personality.GreetingStyle[:38] + ".."
	}
	if len(result.Personality.ClosingStyle) > 40 {
		result.Personality.ClosingStyle = result.Personality.ClosingStyle[:38] + ".."
	}

	return result, nil
}
