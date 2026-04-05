# OpenClaude Bridge 🌉🤖

Um adaptador simples, poderoso e direto ao ponto escrito em **Go** que permite usar a recém-lançada CLI oficial **Claude Code** (da Anthropic) conectada a **Modelos Open Source (LLMs) rodando 100% localmente e de graça** via [Ollama](https://ollama.com/) (ou qualquer API compatível com OpenAI, como vLLM e LMStudio).

## 🚀 Como funciona?
O *Claude Code* é um projeto formidável com Agentes Autônomos feitos para ajudar desenvolvedores lendo arquivos, executando comandos via terminal (`BashTool`), manipulando arquivos, analisando logs e arquiteturas.

Originalmente ele força requisições para os endpoints pagos e proprietários da Anthropic usando formatos específicos (Ex: o "Anthropic Tool Use Protocol"). Essa Bridge:
1. Sobe um servidor proxy na sua máquina.
2. Intercepta todas as requisições, System Prompts, Arquivos Lidos e o Histórico enviadas do Claude CLI.
3. Traduz os blocos gigantes em tempo real para a especificação do **Ollama / OpenAI**, entregando ao LLM local de sua preferência.
4. Quando a sua IA decide executar uma linha de comando no seu Windows/Linux, a Bridge **traduz o "Tool Call" de volta** para que a interface React/Ink oficial do Claude obedeça sua própria IA como se ela fosse o *Claude 3.7 Sonnet*!

## 💾 Instalação

1. Clone ou baixe esse repositório e compile o projeto (requer Go instalado):
```bash
git clone https://github.com/O-Seu-Usuario/openclaude-bridge.git
cd openclaude-bridge
go mod init openclaude-bridge
go build -o openclaude-bridge.exe
```

2. Certifique-se que você tem o Ollama rodando localmente (porta 11434 é o padrão):
```bash
ollama run qwen2.5-coder:7b
# Ou qualquer LLM forte que preferir (Recomendado: Modelos com bom suporte a Tool-Calling como Qwen 2.5 Coder, Llama 3.1 8b, Mistral, etc).
```

## 🛠️ Como Usar (Conectando a Bridge no Claude)

Rode o serviço:
```bash
go run main.go
# Ele escutará na porta 4000
```

Abra **OUTRA** aba do seu terminal aonde você quer que o assistente analise seus projetos, defina as variávies mágicas que "destravam" as seguranças de endpoint e modelo da interface do Claude, e execute-o com o nome do modelo suportado pelo Ollama:

**No PowerShell (Windows):**
```powershell
$env:ANTHROPIC_BASE_URL="http://localhost:4000"
$env:ANTHROPIC_CUSTOM_MODEL_OPTION="qwen2.5-coder:7b"
$env:ANTHROPIC_API_KEY="sk-fake-key"

npx @anthropic-ai/claude-code --model qwen2.5-coder:7b
```

**No Bash / ZSH (Linux ou Mac):**
```bash
export ANTHROPIC_BASE_URL="http://localhost:4000"
export ANTHROPIC_CUSTOM_MODEL_OPTION="qwen2.5-coder:7b"
export ANTHROPIC_API_KEY="sk-fake-key"

npx @anthropic-ai/claude-code --model qwen2.5-coder:7b
```

## 🧠 Suporte Agêntico
Esta bridge suporta oficialmente O **Tool Calling**:
- `BashTool`
- `FileEditTool`
- `FileReadTool`

Quando o modelo decide executar uma ação de sistema, a resposta é entregue no padrão `event-stream` do Anthropic (`content_block_start`, `input_json_delta`, `message_delta`), o que faz a interface reativa da ferramenta desenhar e invocar funções reais no seu computador. Devido a complexidade, a bridge opera consumindo JSON integro do Ollama primeiro e então falsificação ("mock") de Server-Sent-Events (SSE) devolvendo pro CLI.

---
---
**Nota:** Modelos fracos (abaixo de 7b parâmetros ou sem ajuste fino para código/JSON) podem ficar confusos ou vomitar JSON mal formados na hora de chamar uma das ferramentas complexas que o Claude CLI injeta. Modelos modernos geralmente tiram isso de letra. Recomendado: Coder models.

## 🚀 MODO AVANÇADO: OpenRouter & Bypass de Login

Se você deseja usar modelos da Nuvem (como os gratuitos do OpenRouter) mantendo a interface oficial do Claude Code, siga estes passos para ignorar a exigência de `/login` da Anthropic:

### 1. Configure a Ponte (Terminal 1)
Defina sua chave REAL do OpenRouter no servidor da ponte:
```powershell
$env:OPENAI_API_KEY="sk-or-v1-sua-chave-aqui"
go run main.go
```

### 2. Prepare o Claude CLI (Terminal 2)
Use uma chave falsa que comece com `sk-ant-` para "enganar" a verificação de login do CLI oficial, enquanto aponta para a nossa ponte local:
```powershell
$env:ANTHROPIC_BASE_URL="http://localhost:4000"
$env:ANTHROPIC_API_KEY="sk-ant-api03-BURLADO"

# Escolha um modelo do OpenRouter (Ex: Qwen 3.6 Plus Free ou Gemma 2 9b Free)
npx @anthropic-ai/claude-code --model qwen/qwen3.6-plus:free
```

Dessa forma, o Claude CLI aceita a chave fake, envia a requisição para a Ponte, e a Ponte substitui silenciosamente pela sua chave real do OpenRouter antes de disparar para a nuvem. Inteligência de ponta com custo zero de infraestrutura local! 🚀

