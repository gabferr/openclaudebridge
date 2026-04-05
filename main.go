package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

// ========= Anthropic (Claude) Structs =========
const SYSTEM_TURBO = `
[MODE: TURBO]
- Action-oriented
- Execute directly
- 1-3 lines max for conversation
`
const SYSTEM_DEBUG = `
[MODE: DEBUG]
- Find problems and fix
- Show issue + fix
- Minimal explanation
`
const SYSTEM_BALANCED = `
[MODE: BALANCED]
- Be concise
- Max 5 bullets
- Focus on useful info
`

func detectMode(messages []AnthropicMessage) string {
	if len(messages) == 0 {
		return SYSTEM_BALANCED
	}

	envMode := os.Getenv("CLAUDE_MODE")
	if envMode == "turbo" {
		return SYSTEM_TURBO
	} else if envMode == "debug" {
		return SYSTEM_DEBUG
	}

	last := ""
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role != "user" {
			continue
		}

		if str, ok := messages[i].Content.(string); ok {
			last = strings.ToLower(str)
			break
		} else if contentArr, ok := messages[i].Content.([]interface{}); ok {
			for _, itemRaw := range contentArr {
				item, okMap := itemRaw.(map[string]interface{})
				if !okMap {
					continue
				}
				if blockType, _ := item["type"].(string); blockType == "text" {
					if text, ok := item["text"].(string); ok {
						last += strings.ToLower(text) + " "
					}
				}
			}
			break
		}
	}

	last = strings.TrimSpace(last)
	if strings.Contains(last, "erro") || strings.Contains(last, "bug") {
		return SYSTEM_DEBUG
	}

	if len(last) > 0 && len(last) < 60 {
		return SYSTEM_TURBO
	}

	return SYSTEM_BALANCED
}

// ========= Structs =========
type AnthropicTool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	InputSchema map[string]interface{} `json:"input_schema"`
}

type AnthropicMessage struct {
	Role    string      `json:"role"`
	Content interface{} `json:"content"` // Pode ser string ou []map[string]interface{}
}

type AnthropicRequest struct {
	Model     string             `json:"model"`
	Messages  []AnthropicMessage `json:"messages"`
	System    interface{}        `json:"system,omitempty"`
	Stream    bool               `json:"stream"`
	MaxTokens int                `json:"max_tokens,omitempty"`
	Tools     []AnthropicTool    `json:"tools,omitempty"`
}

// ========= OpenAI / Ollama Structs =========
type OpenAITool struct {
	Type     string `json:"type"`
	Function struct {
		Name        string                 `json:"name"`
		Description string                 `json:"description,omitempty"`
		Parameters  map[string]interface{} `json:"parameters"`
	} `json:"function"`
}

type OpenAIToolCall struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}

type OpenAIMessage struct {
	Role       string           `json:"role"`
	Content    string           `json:"content"`
	ToolCallID string           `json:"tool_call_id,omitempty"`
	ToolCalls  []OpenAIToolCall `json:"tool_calls,omitempty"`
}

type OllamaRequest struct {
	Model    string                 `json:"model"`
	Messages []OpenAIMessage        `json:"messages"`
	Tools    []OpenAITool           `json:"tools,omitempty"`
	Stream   bool                   `json:"stream"`
	Options  map[string]interface{} `json:"options,omitempty"`
}

type OpenAIResponse struct {
	Model   string `json:"model"`
	Choices []struct {
		Message OpenAIMessage `json:"message"`
	} `json:"choices"`
}

func main() {
	port := "4000"
	if p := os.Getenv("PORT"); p != "" {
		port = p
	}

	http.HandleFunc("/v1/messages/", handleMessages)
	http.HandleFunc("/v1/messages", handleMessages)

	fmt.Printf("=================================================================\n")
	fmt.Printf("🚀 OpenClaude Bridge (Ponte Nativa para OpenRouter) O\n")
	fmt.Printf("🔥 Rodando na Porta: %s | Modo: Burlar Login Ativo\n", port)
	fmt.Printf("=================================================================\n")
	fmt.Printf("Comandos para o PowerShell (Copie e Cole):\n\n")
	fmt.Printf("$env:ANTHROPIC_BASE_URL=\"http://localhost:%s\"\n", port)
	fmt.Printf("$env:ANTHROPIC_API_KEY=\"sk-ant-api03-BURLADO\"\n")
	fmt.Printf("npx @anthropic-ai/claude-code --model qwen/qwen3.6-plus:free\n")
	fmt.Printf("=================================================================\n")

	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func handleMessages(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	anthropicKey := r.Header.Get("x-api-key")
	if anthropicKey == "" {
		anthropicKey = strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	}

	// SUPORTE IGUAL AO VÍDEO: Busca a chave real nas variáveis preferidas do usuário
	realKey := os.Getenv("OPENAI_API_KEY")
	if realKey == "" {
		realKey = os.Getenv("OPENROUTER_API_KEY")
	}

	if realKey != "" && strings.HasPrefix(anthropicKey, "sk-ant-") {
		anthropicKey = realKey
	}

	var req AnthropicRequest
	if err := json.Unmarshal(body, &req); err != nil {
		log.Printf("Erro parse Anthropic: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ollamaURL := os.Getenv("OLLAMA_URL")
	if ollamaURL == "" {
		ollamaURL = "https://openrouter.ai/api" // Mudança feita: Foco em performance via Nuvem!
	}

	// 0. Inteligência Genuína: Burlar o "Model Not Found" exigido pelo LM Studio
	actualModel := req.Model
	if !strings.Contains(ollamaURL, "openrouter.ai") {
		if mResp, mErr := http.Get(ollamaURL + "/v1/models"); mErr == nil {
			defer mResp.Body.Close()
			var mData struct {
				Data []struct {
					ID string `json:"id"`
				} `json:"data"`
			}
			json.NewDecoder(mResp.Body).Decode(&mData)
			if len(mData.Data) > 0 {
				actualModel = mData.Data[0].ID
			}
		}
	}

	ollamaReq := OllamaRequest{
		Model:  actualModel, // Injeção do Nome Correto Invisível pro Claude
		Stream: false,
		Options: map[string]interface{}{
			"num_ctx": 8192,
		},
	}

	// 1. Traduzir System Prompt
	if req.System != nil {
		systemText := ""
		systemArray, ok := req.System.([]interface{})
		if ok {
			for _, block := range systemArray {
				if blockMap, isMap := block.(map[string]interface{}); isMap && blockMap["text"] != nil {
					systemText += blockMap["text"].(string) + "\n\n"
				}
			}
		} else if systemStr, ok := req.System.(string); ok {
			systemText = systemStr
		}
		if systemText != "" {
			// Injeção de personalidade forte
			mode := detectMode(req.Messages)

			systemText += "\n\n[CRITICAL SYSTEM INSTRUCTION]: You are Claude Code running on Windows PowerShell. DO NOT use Linux root paths like /Users/. Use standard Windows paths or relative paths. You have access to bash/powershell tools. When asked to execute tasks, use the tools directly instead of just giving text instructions. Avoid web searching before acting. If the user asks a direct conversational question, you may reply with text."
			systemText += "\n\n" + mode

			ollamaReq.Messages = append(ollamaReq.Messages, OpenAIMessage{
				Role:    "system",
				Content: systemText,
			})
		}
	}

	// 2. Traduzir Mensagens
	for _, msg := range req.Messages {
		if contentStr, ok := msg.Content.(string); ok {
			ollamaReq.Messages = append(ollamaReq.Messages, OpenAIMessage{
				Role:    msg.Role,
				Content: contentStr,
			})
		} else if contentArr, ok := msg.Content.([]interface{}); ok {
			// Pode ser ToolUse, ToolResult, ou Text
			for _, itemRaw := range contentArr {
				item, _ := itemRaw.(map[string]interface{})
				blockType, _ := item["type"].(string)

				if blockType == "text" {
					text, _ := item["text"].(string)
					ollamaReq.Messages = append(ollamaReq.Messages, OpenAIMessage{
						Role:    msg.Role,
						Content: text,
					})
				} else if blockType == "tool_use" {
					id, _ := item["id"].(string)
					name, _ := item["name"].(string)
					inputMap, _ := item["input"].(map[string]interface{})
					inputBytes, _ := json.Marshal(inputMap)

					tollCall := OpenAIToolCall{
						ID:   id,
						Type: "function",
					}
					tollCall.Function.Name = name
					tollCall.Function.Arguments = string(inputBytes)

					ollamaReq.Messages = append(ollamaReq.Messages, OpenAIMessage{
						Role:      "assistant",
						ToolCalls: []OpenAIToolCall{tollCall},
					})
				} else if blockType == "tool_result" {
					toolCallID, _ := item["tool_use_id"].(string)
					contentRaw := item["content"]
					contentStr := ""

					// Tool result pode ser string pura ou array de texto
					if str, ok := contentRaw.(string); ok {
						contentStr = str
					} else if arr, ok := contentRaw.([]interface{}); ok {
						for _, rRaw := range arr {
							rMap, _ := rRaw.(map[string]interface{})
							if text, ok := rMap["text"].(string); ok {
								contentStr += text
							}
						}
					}

					ollamaReq.Messages = append(ollamaReq.Messages, OpenAIMessage{
						Role:       "tool",
						Content:    contentStr,
						ToolCallID: toolCallID,
					})
				}
			}
		}
	}

	// 3. Filtro Dourado de Ferramentas (Velocidade Extrema)
	var toolNamesLog []string

	// Lista das ÚNICAS ferramentas que deixaremos o Qwen ver.
	// Isso poda as 28 opções absurdas pra apenas as essenciais de código, matando 80% do texto do prompt!
	// Adicionado 'Skill' para suportar o ecossistema de Automations/Skills da Anthropic!
	essentialTools := map[string]bool{
		"Bash":         true,
		"FileEdit":     true,
		"FileRead":     true,
		"FileWrite":    true,
		"Glob":         true,
		"Grep":         true,
		"NotebookEdit": true,
		"Replace":      true,
		"StrReplace":   true,
	}

	for _, tool := range req.Tools {
		// Pular (ignorar) ferramentas que não estão na "lista VIP"
		if !essentialTools[tool.Name] {
			continue
		}

		toolNamesLog = append(toolNamesLog, tool.Name)
		oTool := OpenAITool{Type: "function"}
		oTool.Function.Name = tool.Name
		oTool.Function.Description = tool.Description
		oTool.Function.Parameters = tool.InputSchema
		ollamaReq.Tools = append(ollamaReq.Tools, oTool)
	}

	if len(toolNamesLog) > 0 {
		log.Printf("🔎 [CLI TOOLS MANTIDAS NO FILTRO]: %s", strings.Join(toolNamesLog, ", "))
	}

	// 4. Mandar pro Ollama via API Compatível com OpenAI (/v1/chat/completions ou /api/chat)
	// Vamos usar /api/chat do Ollama que já suporta OpenAI format e tools
	ollamaBody, _ := json.Marshal(ollamaReq)

	toolsLog := "(Sem Tools)"
	if len(ollamaReq.Tools) > 0 {
		toolsLog = fmt.Sprintf("(%d Tools enviadas)", len(ollamaReq.Tools))
	}
	log.Printf("[REQ] Processando Mensagem -> Modelo: %s %s", req.Model, toolsLog)

	reqDownstream, _ := http.NewRequest("POST", ollamaURL+"/v1/chat/completions", bytes.NewReader(ollamaBody))
	reqDownstream.Header.Set("Content-Type", "application/json")
	if anthropicKey != "" && anthropicKey != "sk-fake-key" && anthropicKey != "sk-nota10" {
		reqDownstream.Header.Set("Authorization", "Bearer "+anthropicKey)
		reqDownstream.Header.Set("HTTP-Referer", "http://localhost") // Recomendado pelo OpenRouter
		reqDownstream.Header.Set("X-Title", "OpenClaude Bridge")
	}

	resp, err := http.DefaultClient.Do(reqDownstream)
	if err != nil {
		log.Printf("Ollama/LMStudio Erro: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	var llmResp OpenAIResponse
	bodyResp, _ := io.ReadAll(resp.Body)

	// >>> DEBUG PROFUNDO DE RESPOSTA!
	log.Printf("🔍 [DEBUG LM STUDIO RAW]: %s", string(bodyResp))
	// <<<

	// Tratar erros explícitos (LM Studio formato String OU OpenRouter formato Objeto)
	var errResp map[string]interface{}
	json.Unmarshal(bodyResp, &errResp)

	if errObj, ok := errResp["error"].(map[string]interface{}); ok {
		if msg, okMsg := errObj["message"].(string); okMsg {
			log.Printf("🚨 OPENROUTER RECUSOU NA NUVEM: %s", msg)
			http.Error(w, "Erro do OpenRouter: "+msg, http.StatusInternalServerError)
			return
		}
	} else if errStr, ok := errResp["error"].(string); ok {
		log.Printf("🚨 LM STUDIO LOCAL RECUSOU O PROMPT: %s", errStr)
		http.Error(w, errStr, http.StatusInternalServerError)
		return
	}

	if err := json.Unmarshal(bodyResp, &llmResp); err != nil {
		log.Printf("Ollama/LMStudio Resposta Unmarshal Erro: %v | Raw: %s", err, string(bodyResp))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var parsedMessage OpenAIMessage
	if len(llmResp.Choices) > 0 {
		parsedMessage = llmResp.Choices[0].Message
	}

	// =========================================================================
	// 🛠️ HACK PARA IA LOCAL: Fallback de Tool Hallucination
	// =========================================================================
	if len(parsedMessage.ToolCalls) == 0 && parsedMessage.Content != "" {
		contentStr := strings.TrimSpace(parsedMessage.Content)

		// 1. Tentar limpar blocos markdown (```json ... ```) se existirem
		if strings.HasPrefix(contentStr, "```") {
			lines := strings.Split(contentStr, "\n")
			if len(lines) > 2 {
				// Remove a primeira linha (```json) e a última (```)
				contentStr = strings.Join(lines[1:len(lines)-1], "\n")
				contentStr = strings.TrimSpace(contentStr)
			}
		}

		// 2. Se agora começa com "{", vamos tentar parsear
		if strings.HasPrefix(contentStr, "{") {
			var fallbackTool struct {
				Name      string                 `json:"name"`
				Arguments map[string]interface{} `json:"arguments"`
			}
			if err := json.Unmarshal([]byte(contentStr), &fallbackTool); err == nil && fallbackTool.Name != "" {

				// 3. Garantir os nomes corretos das ferramentas Anthropic
				toolName := fallbackTool.Name
				if strings.Contains(strings.ToLower(toolName), "bash") || toolName == "Execute" || toolName == "Run" {
					toolName = "Bash"
				} else if toolName == "List" {
					toolName = "Glob"
				}

				if toolName != fallbackTool.Name {
					log.Printf("🛠️ [FALLBACK TOOL]: Extraído %s -> Mapeado para %s", fallbackTool.Name, toolName)
				} else {
					log.Printf("🛠️ [FALLBACK TOOL]: Acionando %s", toolName)
				}

				argsBytes, _ := json.Marshal(fallbackTool.Arguments)
				parsedMessage.ToolCalls = append(parsedMessage.ToolCalls, OpenAIToolCall{
					ID:   "call_fallback_123",
					Type: "function",
					Function: struct {
						Name      string `json:"name"`
						Arguments string `json:"arguments"`
					}{
						Name:      toolName,
						Arguments: string(argsBytes),
					},
				})
				parsedMessage.Content = "" // Limpa o texto
			}
		}
	}

	log.Printf("[RES] Sucesso. Respondendo ao Claude CLI...")

	// 5. Devolver no formato do Anthropic
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	mockID := "msg_mock_ollama_123"

	// Start message
	fmt.Fprintf(w, "event: message_start\ndata: {\"type\":\"message_start\",\"message\":{\"id\":\"%s\",\"type\":\"message\",\"role\":\"assistant\",\"content\":[],\"model\":\"%s\",\"stop_reason\":null,\"usage\":{\"input_tokens\":0,\"output_tokens\":0}}}\n\n", mockID, req.Model)

	// Block Start (Text)
	if parsedMessage.Content == "" && len(parsedMessage.ToolCalls) == 0 {
		parsedMessage.Content = "✔️ Ação concluída (O Qwen/LMStudio não quis falar mais nada)."
	}

	if parsedMessage.Content != "" {
		fmt.Fprintf(w, "event: content_block_start\ndata: {\"type\":\"content_block_start\",\"index\":0,\"content_block\":{\"type\":\"text\",\"text\":\"\"}}\n\n")

		safeText, _ := json.Marshal(parsedMessage.Content)
		fmt.Fprintf(w, "event: content_block_delta\ndata: {\"type\":\"content_block_delta\",\"index\":0,\"delta\":{\"type\":\"text_delta\",\"text\":%s}}\n\n", string(safeText))

		fmt.Fprintf(w, "event: content_block_stop\ndata: {\"type\":\"content_block_stop\",\"index\":0}\n\n")
	}

	stopReason := "end_turn"

	// Tools Call Response
	if len(parsedMessage.ToolCalls) > 0 {
		stopReason = "tool_use"
		for i, tc := range parsedMessage.ToolCalls {
			index := i
			if parsedMessage.Content != "" {
				index = i + 1
			}

			// A CLI do Claude entra em crash se o ID for vazio!
			toolID := tc.ID
			if toolID == "" {
				toolID = fmt.Sprintf("call_mock_%d", index)
			}

			// START
			fmt.Fprintf(w, "event: content_block_start\ndata: {\"type\":\"content_block_start\",\"index\":%d,\"content_block\":{\"type\":\"tool_use\",\"id\":\"%s\",\"name\":\"%s\",\"input\":{}}}\n\n", index, toolID, tc.Function.Name)

			// DELTA JSON ARGUMENTS (Raw string directly as partial_json)
			argsStr := strings.TrimSpace(tc.Function.Arguments)
			if argsStr == "" {
				argsStr = "{}"
			}
			
			fmt.Fprintf(w, "event: content_block_delta\ndata: {\"type\":\"content_block_delta\",\"index\":%d,\"delta\":{\"type\":\"input_json_delta\",\"partial_json\":%s}}\n\n", index, argsStr)

			// STOP
			fmt.Fprintf(w, "event: content_block_stop\ndata: {\"type\":\"content_block_stop\",\"index\":%d}\n\n", index)
		}
	}

	// Message Stop
	fmt.Fprintf(w, "event: message_delta\ndata: {\"type\":\"message_delta\",\"delta\":{\"stop_reason\":\"%s\"},\"usage\":{\"input_tokens\":0,\"output_tokens\":0}}\n\n", stopReason)
	fmt.Fprintf(w, "event: message_stop\ndata: {\"type\":\"message_stop\"}\n\n")

	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}
}
