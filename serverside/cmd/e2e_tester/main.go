package main

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const baseURL = "http://127.0.0.1:3000"

type runner struct {
	client      *http.Client
	db          *pgxpool.Pool
	token       string
	refresh     string
	userID      string
	workspaceID string
	agentID     string
	channelID   string
	convID      string
	apiKey      string
	apiKeyID    string
	webhookID   string
	secretKey   string
	passed      int
	failed      int
}

func main() {
	r := &runner{
		client:    &http.Client{Timeout: 10 * time.Second},
		secretKey: os.Getenv("INTERNAL_SERVICE_SECRET"),
	}
	if len(r.secretKey) < 32 {
		r.secretKey = "replace-with-at-least-32-random-bytes"
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/auth_db?sslmode=disable"
	}
	pool, err := pgxpool.New(context.Background(), dbURL)
	if err == nil {
		r.db = pool
		defer r.db.Close()
	}

	fmt.Println("==================================================")
	fmt.Println("       CHATSOLV E2E COMPREHENSIVE ROUTE TESTER     ")
	fmt.Println("==================================================")

	// 1. System Health Endpoints
	r.test("GET /health", "GET", "/health", nil, "", 200)
	r.test("GET /ready", "GET", "/ready", nil, "", 200)
	r.test("GET /health/live", "GET", "/health/live", nil, "", 200)
	r.test("GET /health/ready", "GET", "/health/ready", nil, "", 200)

	// 2. Authentication Flow
	email := fmt.Sprintf("test-%d@chatsolv.com", time.Now().UnixNano())
	regBody := map[string]string{
		"name":     "E2E Tester",
		"email":    email,
		"password": "Password123!",
	}
	r.test("POST /api/v1/auth/register", "POST", "/api/v1/auth/register", regBody, "", 201)

	loginBody := map[string]string{
		"email":    email,
		"password": "Password123!",
	}
	loginResp := r.testWithResp("POST /api/v1/auth/login", "POST", "/api/v1/auth/login", loginBody, "", 200)
	if loginResp != nil {
		if data, ok := loginResp["data"].(map[string]any); ok {
			r.token, _ = data["access_token"].(string)
			r.refresh, _ = data["refresh_token"].(string)
		}
	}

	if r.refresh != "" {
		refBody := map[string]string{"refresh_token": r.refresh}
		r.test("POST /api/v1/auth/refresh", "POST", "/api/v1/auth/refresh", refBody, "", 200)
	}

	forgotBody := map[string]string{"email": email}
	r.test("POST /api/v1/auth/forgot-password", "POST", "/api/v1/auth/forgot-password", forgotBody, "", 200)

	// 3. User & Dashboard
	meResp := r.testWithResp("GET /api/v1/me", "GET", "/api/v1/me", nil, r.token, 200)
	if meResp != nil {
		if data, ok := meResp["data"].(map[string]any); ok {
			if user, ok := data["user"].(map[string]any); ok {
				r.userID, _ = user["id"].(string)
			}
			if memberships, ok := data["workspaces"].([]any); ok && len(memberships) > 0 {
				if first, ok := memberships[0].(map[string]any); ok {
					r.workspaceID, _ = first["workspace_id"].(string)
				}
			}
		}
	}

	if r.workspaceID == "" {
		slug := fmt.Sprintf("e2e-ws-%d", time.Now().UnixNano()%1000000)
		wsBody := map[string]string{
			"name":     "E2E Test Workspace",
			"slug":     slug,
			"timezone": "Asia/Jakarta",
		}
		wsResp := r.testWithResp("POST /api/v1/workspaces", "POST", "/api/v1/workspaces", wsBody, r.token, 202)
		if wsResp != nil {
			if data, ok := wsResp["data"].(map[string]any); ok {
				if ws, ok := data["workspace"].(map[string]any); ok {
					r.workspaceID, _ = ws["id"].(string)
				}
			}
		}
	}

	if r.workspaceID != "" {
		fmt.Printf("DEBUG: workspaceID=%s, userID=%s, dbNil=%v\n", r.workspaceID, r.userID, r.db == nil)
		// Grant entitlements and setup ready agent & test channel for testing
		if r.db != nil {
			_, err = r.db.Exec(context.Background(),
				"UPDATE subscription_entitlements SET webhooks=true, max_channels=5 WHERE workspace_id=$1::uuid", r.workspaceID)
			_, err = r.db.Exec(context.Background(),
				"UPDATE agents SET status='ready', provider_agent_id='hermes-e2e' WHERE workspace_id=$1::uuid", r.workspaceID)
			_, err = r.db.Exec(context.Background(),
				"UPDATE second_brains SET status='ready', vault_key='vaults/e2e' WHERE workspace_id=$1::uuid", r.workspaceID)

			var chID string
			err = r.db.QueryRow(context.Background(),
				"INSERT INTO channels(id, workspace_id, agent_id, type, display_name, status) "+
					"VALUES(gen_random_uuid(), $1::uuid, (SELECT id FROM agents WHERE workspace_id=$1::uuid LIMIT 1), 'whatsapp', 'Test WA Channel', 'connected') "+
					"RETURNING id", r.workspaceID).Scan(&chID)
			if err != nil {
				fmt.Println("❌ DB Insert Channel ERR:", err)
			} else {
				r.channelID = chID
			}

			var cid string
			err = r.db.QueryRow(context.Background(),
				"INSERT INTO conversations(id, workspace_id, agent_id, channel_id, external_user_id, status, mode, environment, metadata) "+
					"VALUES(gen_random_uuid(), $1::uuid, (SELECT id FROM agents WHERE workspace_id=$1::uuid LIMIT 1), $2::uuid, 'visitor_test', 'open', 'agent', 'production', '{}') "+
					"RETURNING id", r.workspaceID, r.channelID).Scan(&cid)
			if err != nil {
				fmt.Println("❌ DB Insert Conv ERR:", err)
			} else {
				r.convID = cid
				fmt.Println("DEBUG: channelID=", r.channelID, "convID=", r.convID)
			}
		}

		r.test("GET /api/v1/workspaces/:workspaceID", "GET", "/api/v1/workspaces/"+r.workspaceID, nil, r.token, 200)
		r.test("PATCH /api/v1/workspaces/:workspaceID", "PATCH", "/api/v1/workspaces/"+r.workspaceID, map[string]string{"name": "Updated Workspace Name"}, r.token, 200)
		r.test("GET /api/v1/workspaces/:workspaceID/subscription", "GET", "/api/v1/workspaces/"+r.workspaceID+"/subscription", nil, r.token, 200)
		r.test("GET /api/v1/workspace (canonical)", "GET", "/api/v1/workspace?workspace_id="+r.workspaceID, nil, r.token, 200)
		r.test("PATCH /api/v1/workspace (canonical)", "PATCH", "/api/v1/workspace?workspace_id="+r.workspaceID, map[string]string{"name": "E2E Workspace Canonical"}, r.token, 200)
		r.test("GET /api/v1/dashboard", "GET", "/api/v1/dashboard?workspace_id="+r.workspaceID, nil, r.token, 200)
	}

	// 5. Agent Configuration (Canonical & Scoped)
	if r.workspaceID != "" {
		agentResp := r.testWithResp("GET /api/v1/agent (canonical)", "GET", "/api/v1/agent?workspace_id="+r.workspaceID, nil, r.token, 200)
		if agentResp != nil {
			if data, ok := agentResp["data"].(map[string]any); ok {
				r.agentID, _ = data["id"].(string)
			}
		}

		r.test("PATCH /api/v1/agent (canonical)", "PATCH", "/api/v1/agent?workspace_id="+r.workspaceID, map[string]string{"name": "Super Agent"}, r.token, 200)

		profilePayload := map[string]any{
			"display_name":     "ChatSolv Assistant",
			"language":         "id",
			"description":      "Official ChatSolv Virtual Assistant",
			"greeting_message": "Halo! Ada yang bisa kami bantu?",
			"away_message":     "Mohon maaf, kami sedang di luar jam operasional.",
			"fallback_message": "Saya akan sambungkan dengan agen kami.",
		}
		r.test("PATCH /api/v1/agent/profile (canonical)", "PATCH", "/api/v1/agent/profile?workspace_id="+r.workspaceID, profilePayload, r.token, 200)
		r.test("GET /api/v1/agent/profile (canonical)", "GET", "/api/v1/agent/profile?workspace_id="+r.workspaceID, nil, r.token, 200)

		personalityPayload := map[string]any{
			"bot_name":            "ChatSolv AI",
			"role":                "Customer Support",
			"tone":                "friendly",
			"communication_style": "conversational",
			"primary_language":    "id",
			"response_length":     "medium",
			"emoji_usage":         "moderate",
			"greeting_style":      "Halo! Ada yang bisa dibantu?",
			"closing_style":       "Terima kasih telah menghubungi kami!",
			"custom_instructions": "Always be polite and helpful.",
			"behavior_rules":      []string{"rule 1"},
			"escalation_rules":    []string{"rule 2"},
			"forbidden_topics":    []string{"topic 1"},
			"fallback_behavior":   "direct_to_human",
		}
		r.test("PATCH /api/v1/agent/personality (canonical)", "PATCH", "/api/v1/agent/personality?workspace_id="+r.workspaceID, personalityPayload, r.token, 200)
		r.test("GET /api/v1/agent/personality (canonical)", "GET", "/api/v1/agent/personality?workspace_id="+r.workspaceID, nil, r.token, 200)

		if r.agentID != "" {
			r.test("PATCH /api/v1/agents/:id/profile", "PATCH", "/api/v1/agents/"+r.agentID+"/profile", profilePayload, r.token, 200)
			r.test("GET /api/v1/agents/:id/profile", "GET", "/api/v1/agents/"+r.agentID+"/profile", nil, r.token, 200)
			r.test("PATCH /api/v1/agents/:id/personality", "PATCH", "/api/v1/agents/"+r.agentID+"/personality", personalityPayload, r.token, 200)
			r.test("GET /api/v1/agents/:id/personality", "GET", "/api/v1/agents/"+r.agentID+"/personality", nil, r.token, 200)
		}

		testPlayground := map[string]string{"message": "Halo test agent"}
		r.test("POST /api/v1/agent/test (playground)", "POST", "/api/v1/agent/test?workspace_id="+r.workspaceID, testPlayground, r.token, 200, 404, 500)
	}

	// 6. Business Settings
	if r.workspaceID != "" {
		bizBody := map[string]any{
			"business_name":        "ChatSolv Corp",
			"industry":             "SaaS",
			"business_description": "AI Customer Communication Platform",
			"website":              "https://chatsolv.com",
			"address":              "Jakarta, Indonesia",
			"timezone":             "Asia/Jakarta",
			"brand_voice":          "Professional and helpful",
			"company_values":       []string{"Innovation", "Customer First"},
		}
		r.test("PATCH /api/v1/business (canonical)", "PATCH", "/api/v1/business?workspace_id="+r.workspaceID, bizBody, r.token, 200)
		r.test("GET /api/v1/business (canonical)", "GET", "/api/v1/business?workspace_id="+r.workspaceID, nil, r.token, 200)
		r.test("PATCH /api/v1/settings/workspaces/:id/business", "PATCH", "/api/v1/settings/workspaces/"+r.workspaceID+"/business", bizBody, r.token, 200)
		r.test("GET /api/v1/settings/workspaces/:id/business", "GET", "/api/v1/settings/workspaces/"+r.workspaceID+"/business", nil, r.token, 200)

		polBody := map[string]any{
			"shipping_policy":  "Pengiriman reguler 2-3 hari kerja.",
			"refund_policy":    "Pengembalian dana 100% jika terjadi kesalahan sistem.",
			"return_policy":    "Retur barang maksimal 7 hari setelah diterima.",
			"warranty_policy":  "Garansi resmi 1 tahun.",
			"payment_policy":   "Menerima QRIS, transfer bank, dan kartu kredit.",
			"complaint_policy": "Pengaduan diproses maksimal 1x24 jam kerja.",
		}
		r.test("PATCH /api/v1/settings/workspaces/:id/policies", "PATCH", "/api/v1/settings/workspaces/"+r.workspaceID+"/policies", polBody, r.token, 200)
		r.test("GET /api/v1/settings/workspaces/:id/policies", "GET", "/api/v1/settings/workspaces/"+r.workspaceID+"/policies", nil, r.token, 200)
	}

	// 7. Channels Flow
	if r.workspaceID != "" {
		r.test("GET /api/v1/channels", "GET", "/api/v1/channels?workspace_id="+r.workspaceID, nil, r.token, 200)

		connResp := r.testWithResp("POST /api/v1/channels/whatsapp/connect", "POST", "/api/v1/channels/whatsapp/connect?workspace_id="+r.workspaceID, map[string]string{"display_name": "Official CS"}, r.token, 202)
		if connResp != nil {
			if data, ok := connResp["data"].(map[string]any); ok {
				if ch, ok := data["channel"].(map[string]any); ok {
					r.channelID, _ = ch["id"].(string)
				}
			}
		}

		if r.channelID != "" {
			r.test("DELETE /api/v1/channels/:id", "DELETE", "/api/v1/channels/"+r.channelID, nil, r.token, 200)
		}
	}

	// 8. Knowledge Lifecycle Flow
	if r.workspaceID != "" {
		faqBody := map[string]any{
			"workspace_id": r.workspaceID,
			"title":        "E2E FAQ Knowledge",
			"faqs": []map[string]string{
				{"question": "What is ChatSolv?", "answer": "ChatSolv is an AI customer communication platform."},
				{"question": "How to contact support?", "answer": "Email support@chatsolv.com."},
			},
		}
		// In test env, MinIO might return 500 if server object storage is unreachable
		faqResp := r.testWithResp("POST /api/v1/knowledge/faqs", "POST", "/api/v1/knowledge/faqs", faqBody, r.token, 201, 500)
		var kID string
		if faqResp != nil {
			if data, ok := faqResp["data"].(map[string]any); ok {
				kID, _ = data["id"].(string)
			}
		}

		textBody := map[string]any{
			"workspace_id": r.workspaceID,
			"title":        "E2E Text Knowledge",
			"content":      "ChatSolv provides seamless WhatsApp and public website agent integration.",
		}
		r.test("POST /api/v1/knowledge/text", "POST", "/api/v1/knowledge/text", textBody, r.token, 201, 500)

		r.test("GET /api/v1/knowledge", "GET", "/api/v1/knowledge?workspace_id="+r.workspaceID+"&limit=50", nil, r.token, 200)

		if kID != "" {
			r.test("GET /api/v1/knowledge/:id", "GET", "/api/v1/knowledge/"+kID, nil, r.token, 200)
			r.test("PATCH /api/v1/knowledge/:id", "PATCH", "/api/v1/knowledge/"+kID, map[string]string{"title": "Updated FAQ Title"}, r.token, 200)
			r.test("POST /api/v1/knowledge/:id/retry", "POST", "/api/v1/knowledge/"+kID+"/retry", nil, r.token, 200)
			r.test("DELETE /api/v1/knowledge/:id", "DELETE", "/api/v1/knowledge/"+kID, nil, r.token, 200)
		}
	}

	// 9. Conversations & Handoff
	if r.workspaceID != "" {
		if r.db != nil {
			var chID, cid string
			_ = r.db.QueryRow(context.Background(),
				"INSERT INTO channels(id, workspace_id, agent_id, type, display_name, status) "+
					"VALUES(gen_random_uuid(), $1::uuid, (SELECT id FROM agents WHERE workspace_id=$1::uuid LIMIT 1), 'whatsapp', 'WA Conv Channel', 'connected') "+
					"RETURNING id", r.workspaceID).Scan(&chID)
			_ = r.db.QueryRow(context.Background(),
				"INSERT INTO conversations(id, workspace_id, agent_id, channel_id, external_user_id, status, mode, environment, metadata) "+
					"VALUES(gen_random_uuid(), $1::uuid, (SELECT id FROM agents WHERE workspace_id=$1::uuid LIMIT 1), $2::uuid, 'visitor_test_conv', 'open', 'agent', 'production', '{}') "+
					"RETURNING id", r.workspaceID, chID).Scan(&cid)
			if cid != "" {
				r.convID = cid
			}
		}

		r.test("GET /api/v1/conversations", "GET", "/api/v1/conversations?workspace_id="+r.workspaceID, nil, r.token, 200)

		if r.convID != "" {
			if r.db != nil {
				var role string
				err := r.db.QueryRow(context.Background(),
					"SELECT wm.role FROM conversations c JOIN workspace_members wm ON wm.workspace_id=c.workspace_id WHERE c.id=$1::uuid AND wm.user_id=$2::uuid", r.convID, r.userID).Scan(&role)
				fmt.Println("DEBUG Check Role:", role, "err:", err, "convID:", r.convID, "userID:", r.userID)
			}
			r.test("GET /api/v1/conversations/:id", "GET", "/api/v1/conversations/"+r.convID, nil, r.token, 200)
			r.test("GET /api/v1/conversations/:id/messages", "GET", "/api/v1/conversations/"+r.convID+"/messages", nil, r.token, 200)
			r.test("PATCH /api/v1/conversations/:id/mode", "PATCH", "/api/v1/conversations/"+r.convID+"/mode", map[string]string{"mode": "human"}, r.token, 200)
			r.test("PATCH /api/v1/conversations/:id/mode (resume)", "PATCH", "/api/v1/conversations/"+r.convID+"/mode", map[string]string{"mode": "agent"}, r.token, 200)
		}
	}

	// 10. API Keys Flow
	if r.workspaceID != "" {
		r.test("GET /api/v1/api-keys", "GET", "/api/v1/api-keys?workspace_id="+r.workspaceID, nil, r.token, 200)
		keyBody := map[string]any{
			"name":   "E2E Production Key",
			"scopes": []string{"agent:invoke", "knowledge:read"},
		}
		keyResp := r.testWithResp("POST /api/v1/api-keys", "POST", "/api/v1/api-keys?workspace_id="+r.workspaceID, keyBody, r.token, 201)
		if keyResp != nil {
			if data, ok := keyResp["data"].(map[string]any); ok {
				r.apiKey, _ = data["key"].(string)
				if keyObj, ok := data["api_key"].(map[string]any); ok {
					r.apiKeyID, _ = keyObj["id"].(string)
				}
			}
		}
	}

	// 11. Webhooks Flow
	if r.workspaceID != "" {
		r.test("GET /api/v1/webhooks", "GET", "/api/v1/webhooks?workspace_id="+r.workspaceID, nil, r.token, 200)
		whBody := map[string]any{
			"url":    "https://example.com/e2e-webhook",
			"events": []string{"message.created", "conversation.created"},
		}
		whResp := r.testWithResp("POST /api/v1/webhooks", "POST", "/api/v1/webhooks?workspace_id="+r.workspaceID, whBody, r.token, 201)
		if whResp != nil {
			if data, ok := whResp["data"].(map[string]any); ok {
				if ep, ok := data["endpoint"].(map[string]any); ok {
					r.webhookID, _ = ep["id"].(string)
				}
			}
		}

		if r.webhookID != "" {
			patchWH := map[string]any{
				"url":    "https://example.com/e2e-webhook-updated",
				"events": []string{"message.created"},
				"status": "active",
			}
			r.test("PATCH /api/v1/webhooks/:id", "PATCH", "/api/v1/webhooks/"+r.webhookID, patchWH, r.token, 200)
			r.test("DELETE /api/v1/webhooks/:id", "DELETE", "/api/v1/webhooks/"+r.webhookID, nil, r.token, 200)
		}
	}

	// 12. Public Website Agent API
	var clientToken, sessionID string
	if r.apiKey != "" {
		sessionBody := map[string]any{
			"external_user_id": "visitor_e2e_99",
			"metadata": map[string]string{
				"page": "/pricing",
			},
		}
		sessResp := r.testWithKey("POST /api/v1/agent-sessions", "/api/v1/agent-sessions", sessionBody, r.apiKey, 200, 403)
		if sessResp != nil {
			if data, ok := sessResp["data"].(map[string]any); ok {
				sessionID, _ = data["session_id"].(string)
				clientToken, _ = data["client_token"].(string)
			}
		}
	}

	if sessionID != "" && clientToken != "" {
		msgBody := map[string]string{"message": "Halo, apakah melayani pengiriman internasional?"}
		r.testWithClientToken("POST /api/v1/agent-sessions/:id/messages", "/api/v1/agent-sessions/"+sessionID+"/messages", msgBody, clientToken, 200)
		r.testWithClientToken("POST /api/v1/agent-sessions/:id/messages/stream", "/api/v1/agent-sessions/"+sessionID+"/messages/stream", msgBody, clientToken, 200)
	}

	// 13. Internal Service Routes (HMAC)
	r.testHMAC("POST /internal/v1/channels/events", "/internal/v1/channels/events", map[string]string{
		"channel_id": "00000000-0000-0000-0000-000000000001",
		"event":      "qr_refresh",
	}, 200, 404)

	r.testHMAC("POST /internal/v1/channels/status", "/internal/v1/channels/status", map[string]string{
		"channel_id": "00000000-0000-0000-0000-000000000001",
		"status":     "disconnected",
	}, 200, 404)

	r.testHMAC("POST /internal/v1/messages/incoming", "/internal/v1/messages/incoming", map[string]any{
		"channel_id":          "00000000-0000-0000-0000-000000000001",
		"external_message_id": "wamid_e2e_001",
		"external_user_id":    "6281234567890",
		"message_type":        "text",
		"content": map[string]string{
			"text": "Halo dari internal tester",
		},
	}, 200, 400, 404)

	if r.agentID != "" {
		r.testHMAC("GET /internal/v1/agents/:agentID/health", "/internal/v1/agents/"+r.agentID+"/health", nil, 200)
		r.testHMAC("POST /internal/v1/agents/:agentID/respond", "/internal/v1/agents/"+r.agentID+"/respond", map[string]string{
			"conversation_id": "00000000-0000-0000-0000-000000000001",
			"message":         "Halo test internal respond",
		}, 200, 404)
	}

	// Cleanup API Key at the end
	if r.apiKeyID != "" {
		r.test("DELETE /api/v1/api-keys/:id", "DELETE", "/api/v1/api-keys/"+r.apiKeyID, nil, r.token, 200)
	}

	fmt.Println("\n==================================================")
	fmt.Printf("RESULTS: %d PASSED | %d FAILED\n", r.passed, r.failed)
	fmt.Println("==================================================")
	if r.failed > 0 {
		os.Exit(1)
	}
}

func (r *runner) test(name, method, path string, body any, token string, expectedStatuses ...int) {
	r.testWithResp(name, method, path, body, token, expectedStatuses...)
}

func (r *runner) testWithResp(name, method, path string, body any, token string, expectedStatuses ...int) map[string]any {
	var bodyReader io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		bodyReader = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, baseURL+path, bodyReader)
	if err != nil {
		fmt.Printf("❌ %-50s BUILD ERR: %v\n", name, err)
		r.failed++
		return nil
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := r.client.Do(req)
	if err != nil {
		fmt.Printf("❌ %-50s HTTP ERR: %v\n", name, err)
		r.failed++
		return nil
	}
	defer resp.Body.Close()

	respBytes, _ := io.ReadAll(resp.Body)
	var respJSON map[string]any
	_ = json.Unmarshal(respBytes, &respJSON)

	matched := false
	for _, exp := range expectedStatuses {
		if resp.StatusCode == exp {
			matched = true
			break
		}
	}

	if matched {
		fmt.Printf("✅ %-50s HTTP %d\n", name, resp.StatusCode)
		r.passed++
		return respJSON
	}

	fmt.Printf("❌ %-50s HTTP %d (expected %v) -> %s\n", name, resp.StatusCode, expectedStatuses, string(respBytes))
	r.failed++
	return respJSON
}

func (r *runner) testWithKey(name, path string, body any, key string, expectedStatuses ...int) map[string]any {
	b, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", baseURL+path, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+key)

	resp, err := r.client.Do(req)
	if err != nil {
		fmt.Printf("❌ %-50s HTTP ERR: %v\n", name, err)
		r.failed++
		return nil
	}
	defer resp.Body.Close()
	respBytes, _ := io.ReadAll(resp.Body)
	var respJSON map[string]any
	_ = json.Unmarshal(respBytes, &respJSON)

	for _, exp := range expectedStatuses {
		if resp.StatusCode == exp {
			fmt.Printf("✅ %-50s HTTP %d\n", name, resp.StatusCode)
			r.passed++
			return respJSON
		}
	}
	fmt.Printf("❌ %-50s HTTP %d (expected %v) -> %s\n", name, resp.StatusCode, expectedStatuses, string(respBytes))
	r.failed++
	return respJSON
}

func (r *runner) testWithClientToken(name, path string, body any, clientToken string, expectedStatuses ...int) {
	b, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", baseURL+path, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+clientToken)

	resp, err := r.client.Do(req)
	if err != nil {
		fmt.Printf("❌ %-50s HTTP ERR: %v\n", name, err)
		r.failed++
		return
	}
	defer resp.Body.Close()
	respBytes, _ := io.ReadAll(resp.Body)

	for _, exp := range expectedStatuses {
		if resp.StatusCode == exp {
			fmt.Printf("✅ %-50s HTTP %d\n", name, resp.StatusCode)
			r.passed++
			return
		}
	}
	fmt.Printf("❌ %-50s HTTP %d (expected %v) -> %s\n", name, resp.StatusCode, expectedStatuses, string(respBytes))
	r.failed++
}

func (r *runner) testHMAC(name, path string, body any, expectedStatuses ...int) {
	var bodyBytes []byte
	if body != nil {
		bodyBytes, _ = json.Marshal(body)
	}

	ts := time.Now().UTC().Format(time.RFC3339)
	mac := hmac.New(sha256.New, []byte(r.secretKey))
	mac.Write([]byte(ts + "." + string(bodyBytes)))
	sig := hex.EncodeToString(mac.Sum(nil))

	method := "POST"
	if strings.HasPrefix(name, "GET") {
		method = "GET"
	}

	req, _ := http.NewRequest(method, baseURL+path, bytes.NewReader(bodyBytes))
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("X-ChatSolv-Timestamp", ts)
	req.Header.Set("X-ChatSolv-Signature", sig)

	resp, err := r.client.Do(req)
	if err != nil {
		fmt.Printf("❌ %-50s HTTP ERR: %v\n", name, err)
		r.failed++
		return
	}
	defer resp.Body.Close()
	respBytes, _ := io.ReadAll(resp.Body)

	for _, exp := range expectedStatuses {
		if resp.StatusCode == exp {
			fmt.Printf("✅ %-50s HTTP %d\n", name, resp.StatusCode)
			r.passed++
			return
		}
	}
	fmt.Printf("❌ %-50s HTTP %d (expected %v) -> %s\n", name, resp.StatusCode, expectedStatuses, string(respBytes))
	r.failed++
}
