# 🚀 Como Executar o OpenClaude Bridge

Este guia rápido vai te ensinar a configurar e executar a sua Bridge para conectar o **Claude Code CLI** com o seu LLM local no **Ollama**.

## 📋 Pré-requisitos
1. **Ollama** instalado e rodando na sua máquina.
2. Pelo menos um modelo baixado no Ollama (Ex: `qwen2.5-coder:7b`, `llama3.1:8b`).
3. O pacote do **Claude Code** acessível (seja via `npx @anthropic-ai/claude-code` ou compilado localmente).
4. **Go (Golang)** instalado para rodar a Bridge.

---

## ▶️ Passo 1: Iniciando a Bridge (Proxy)
Abra um terminal (PowerShell ou Bash) na pasta onde este arquivo está (`openclaude-bridge`) e execute o servidor em Go:

```bash
go run main.go
```
*Você verá uma mensagem no console dizendo que o servidor está rodando em http://localhost:4000.*
**(Deixe este terminal aberto e rodando em segundo plano)**

---

## 🤖 Passo 2: Iniciando o Ollama
No seu terminal (pode ser o mesmo que você usa no dia-a-dia), garanta que o modelo que você quer usar está pronto:
```bash
ollama run qwen2.5-coder:7b
```
Você pode digitar `/bye` para sair se ele já estiver salvo no seu sistema, o importante é o serviço do Ollama estar ativo na porta padrão `11434`.

---

## 💻 Passo 3: Conectando o Claude Code
Abra um **NOVO TERMINAL** (na pasta do seu projeto principal, onde você quer que a IA te ajude a codar) e defina as variáveis de ambiente que "enganam" a SDK do Claude para olhar para a nossa Bridge.

**No PowerShell (Windows) - Opção 1: Qwen 2.5 Coder (Melhor para Código):**
```powershell
$env:ANTHROPIC_BASE_URL="http://localhost:4000"
$env:ANTHROPIC_CUSTOM_MODEL_OPTION="qwen2.5-coder:7b"
$env:ANTHROPIC_API_KEY="sk-nota10"

npx @anthropic-ai/claude-code --model qwen2.5-coder:7b
```

**No PowerShell (Windows) - Opção 2: Llama 3.1 (Melhor para Conversa geral):**
```powershell
$env:ANTHROPIC_BASE_URL="http://localhost:4000"
$env:ANTHROPIC_CUSTOM_MODEL_OPTION="llama3.1:8b"
$env:ANTHROPIC_API_KEY="sk-nota10"

npx @anthropic-ai/claude-code --model llama3.1:8b
```

**No PowerShell (Windows) - Opção 3: Usando o LM Studio (Phi-4, modelos Uncensored, etc):**
1. Inicie o Local Server no LM Studio (ele roda na porta `1234`).
2. Feche a Bridge Go (Ctrl+C) e inicie ela novamente passando a porta do LM Studio:
```powershell
$env:OLLAMA_URL="http://127.0.0.1:1234"
go run main.go
```
3. Inicie o Claude CLI normalmente, mas passe o nome do modelo carregado no LM Studio nas variáveis (Exemplo para o Qwen3.5):
```powershell
$env:ANTHROPIC_BASE_URL="http://localhost:4000"
$env:ANTHROPIC_CUSTOM_MODEL_OPTION="qwen3.5-9b"
$env:ANTHROPIC_API_KEY="sk-nota10"

npx @anthropic-ai/claude-code --model qwen3.5-9b
```

**No Bash / ZSH / Linux / MacOS:**
```bash
export ANTHROPIC_BASE_URL="http://localhost:4000"
export ANTHROPIC_CUSTOM_MODEL_OPTION="qwen2.5-coder:7b"
export ANTHROPIC_API_KEY="sk-nota10"

npx @anthropic-ai/claude-code --model qwen2.5-coder:7b
```

> ⚠️ IMPORTANTE: Sempre que for abrir o Claude Code, lembre-se de rodar primeiro os comandos de export/variável de ambiente no terminal. Sem isso, ele tentará conectar aos servidores originais da Anthropic e dará erro de autenticação!

---

## 🎉 Tudo Pronto!
Você será recebido pela tela do Claude Code. 
Se você olhar no terminal onde a Bridge está rodando (Passo 1), verá os logs de requisições sendo interceptadas e enviadas ao Ollama toda vez que enviar uma mensagem!
