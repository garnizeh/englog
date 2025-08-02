# Relatório de Levantamento: Configuração de Modelo LLM via Variáveis de Ambiente

> "The best way to predict the future is to implement it." - Alan Kay ⚙️

## Resumo Executivo

Este relatório identifica todas as alterações necessárias para transformar o modelo LLM hardcoded atual (`qwen2.5-coder:7b`) em um sistema configurável via variáveis de ambiente, permitindo flexibilidade na escolha de modelos para diferentes ambientes e casos de uso.

## Situação Atual

### Modelo Hardcoded Identificado

Atualmente o modelo está hardcoded em **2 locais principais**:

1. **`internal/ai/ollama.go:171`**
   ```go
   modelName := "qwen2.5-coder:7b"
   ```

2. **`internal/worker/client.go:257`**
   ```go
   "ai_model": "qwen2.5-coder:7b",
   ```

### Arquivos de Teste Afetados

3. **`internal/ai/ollama_test.go:60`**
   ```go
   assert.Equal(t, "qwen2.5-coder:7b", service.modelName)
   ```

## Análise de Impacto

### 🔍 Arquivos que Precisam de Alteração

#### 1. **Configuração Principal** (`internal/config/config.go`)

**Localização**: Linhas 115-123 (struct WorkerConfig)

**Alterações Necessárias**:
- Adicionar campos para configuração de modelo LLM na struct `WorkerConfig`
- Implementar carregamento das variáveis de ambiente
- Adicionar valores padrão apropriados

**Campos a Adicionar**:
```go
type WorkerConfig struct {
    // ... campos existentes ...

    // Configuração de Modelo LLM
    DefaultModel         string        // Modelo padrão
    FallbackModel        string        // Modelo de fallback
    ModelTimeout         time.Duration // Timeout para operações do modelo
    ModelMaxRetries      int           // Máximo de tentativas
    ModelTemperature     float64       // Temperatura do modelo
    ModelMaxTokens       int           // Máximo de tokens
}
```

**Implementação de Load()** (Linha ~205):
```go
Worker: WorkerConfig{
    // ... campos existentes ...
    DefaultModel:         getEnv("LLM_DEFAULT_MODEL", "qwen2.5-coder:7b"),
    FallbackModel:        getEnv("LLM_FALLBACK_MODEL", "llama3.2:3b"),
    ModelTimeout:         getDurationEnv("LLM_MODEL_TIMEOUT", 120*time.Second),
    ModelMaxRetries:      getIntEnv("LLM_MODEL_MAX_RETRIES", 3),
    ModelTemperature:     getFloatEnv("LLM_MODEL_TEMPERATURE", 0.7),
    ModelMaxTokens:       getIntEnv("LLM_MODEL_MAX_TOKENS", 2048),
},
```

**Funções Helper Necessárias**:
```go
func getDurationEnv(key string, defaultValue time.Duration) time.Duration
func getFloatEnv(key string, defaultValue float64) float64
```

#### 2. **Serviço Ollama** (`internal/ai/ollama.go`)

**Alterações Principais**:

**A. Modificar Constructor** (Linha 159):
```go
// Antes
func NewOllamaService(ctx context.Context, baseURL string, logger *logging.Logger) (*OllamaService, error)

// Depois
func NewOllamaService(ctx context.Context, baseURL string, modelName string, logger *logging.Logger) (*OllamaService, error)
```

**B. Remover Hardcode** (Linha 171):
```go
// Remover esta linha
modelName := "qwen2.5-coder:7b"

// Usar o parâmetro recebido
if modelName == "" {
    return nil, fmt.Errorf("model name cannot be empty")
}
```

**C. Adicionar Validação de Modelo**:
- Implementar função para validar se o modelo existe no Ollama
- Adicionar fallback automático se modelo principal falhar

**D. Estrutura Expandida**:
```go
type OllamaService struct {
    logger          *logging.Logger
    baseURL         string
    modelName       string
    fallbackModel   string
    timeout         time.Duration
    maxRetries      int
    temperature     float64
    maxTokens       int
    llm             llms.Model
}
```

#### 3. **Worker Client** (`internal/worker/client.go`)

**Alteração Necessária** (Linha 257):
```go
// Antes
"ai_model": "qwen2.5-coder:7b",

// Depois
"ai_model": c.config.Worker.DefaultModel,
```

**Metadata Adicional**:
```go
Metadata: map[string]string{
    "ai_model":        c.config.Worker.DefaultModel,
    "fallback_model":  c.config.Worker.FallbackModel,
    "model_timeout":   c.config.Worker.ModelTimeout.String(),
    "max_tasks":       fmt.Sprintf("%d", c.config.Worker.MaxConcurrentTasks),
    "environment":     c.config.Environment,
},
```

#### 4. **Main Worker** (`cmd/worker/main.go`)

**Alteração Necessária** (Linha 51):
```go
// Antes
aiService, err := ai.NewOllamaService(ctx, cfg.Worker.OllamaURL, logger)

// Depois
aiService, err := ai.NewOllamaService(ctx, cfg.Worker.OllamaURL, cfg.Worker.DefaultModel, logger)
```

#### 5. **Arquivos de Teste**

**A. `internal/ai/ollama_test.go`**:
- **Linha 48**: Atualizar chamada `NewOllamaService` para incluir parâmetro de modelo
- **Linha 60**: Atualizar assertion para usar modelo configurável
- **Linha 152**: Atualizar chamada no teste de integração

**B. Adicionar novos casos de teste**:
- Teste com modelo customizado
- Teste com modelo inválido
- Teste de fallback automático

#### 6. **Arquivos de Ambiente**

**A. `deployments/environments/.env.example`**:

**Adições Necessárias** (após linha 42):
```bash
# =============================================================================
# LLM Model Configuration
# =============================================================================
LLM_DEFAULT_MODEL=qwen2.5-coder:7b
LLM_FALLBACK_MODEL=llama3.2:3b
LLM_MODEL_TIMEOUT=120s
LLM_MODEL_MAX_RETRIES=3
LLM_MODEL_TEMPERATURE=0.7
LLM_MODEL_MAX_TOKENS=2048

# Task-specific model configurations (JSON format)
LLM_TASK_TYPE_CONFIGS={"TASK_TYPE_INSIGHT_GENERATION":{"model":"qwen2.5-coder:7b","timeout":"45s"},"TASK_TYPE_WEEKLY_REPORT":{"model":"llama3.2:7b","timeout":"120s"}}

# Model-specific parameters (JSON format)
LLM_MODEL_PARAMS={"qwen2.5-coder:7b":{"temperature":0.7,"max_tokens":2048},"llama3.2:7b":{"temperature":0.6,"max_tokens":4096}}
```

**B. `deployments/environments/production/.env.worker.example`**:

**Substituir Seção Ollama** (Linhas 24-29):
```bash
# LLM Model Configuration
LLM_DEFAULT_MODEL=qwen2.5-coder:7b
LLM_FALLBACK_MODEL=llama3.2:3b
LLM_MODEL_TIMEOUT=120s
LLM_MODEL_MAX_RETRIES=3
LLM_MODEL_TEMPERATURE=0.7
LLM_MODEL_MAX_TOKENS=2048

# Ollama Server Configuration
OLLAMA_URL=http://ollama:11434
OLLAMA_TIMEOUT=120s
OLLAMA_MAX_RETRIES=3
```

### 🔧 Alterações Opcionais (Futuras)

#### 7. **Configuração Hierárquica Avançada**

Para implementação futura conforme TODO Task 0125:

**A. Protocol Buffer** (`proto/worker.proto`):
- Adicionar message `LLMConfig`
- Expandir `RegisterWorkerRequest` com `supported_models`

**B. Banco de Dados**:
- Tabela `user_llm_preferences`
- Configurações por tipo de task
- Preferências individuais de usuário

**C. API Endpoints**:
- Gerenciamento de preferências de modelo
- Listagem de modelos disponíveis
- Configuração por task type

## Resumo de Mudanças por Arquivo

| Arquivo | Tipo de Alteração | Prioridade | Esforço |
|---------|-------------------|------------|---------|
| `internal/config/config.go` | Estrutural | Alta | Médio |
| `internal/ai/ollama.go` | Funcional | Alta | Alto |
| `internal/worker/client.go` | Simples | Alta | Baixo |
| `cmd/worker/main.go` | Simples | Alta | Baixo |
| `internal/ai/ollama_test.go` | Teste | Média | Médio |
| `.env.example` | Configuração | Alta | Baixo |
| `.env.worker.example` | Configuração | Alta | Baixo |

## Ordem de Implementação Recomendada

### Fase 1: Configuração Básica
1. **Atualizar `internal/config/config.go`** - Adicionar campos e env loading
2. **Atualizar arquivos `.env`** - Definir variáveis de ambiente
3. **Modificar `cmd/worker/main.go`** - Passar parâmetro de modelo

### Fase 2: Serviço Core
4. **Refatorar `internal/ai/ollama.go`** - Aceitar modelo como parâmetro
5. **Atualizar `internal/worker/client.go`** - Usar modelo configurado

### Fase 3: Testes e Validação
6. **Corrigir `internal/ai/ollama_test.go`** - Adaptar testes existentes
7. **Adicionar novos testes** - Validação de configuração
8. **Teste de integração** - Verificar funcionamento end-to-end

## Validação de Funcionalidades

### Cenários de Teste Necessários

1. **Modelo Padrão**: Usar configuração default
2. **Modelo Customizado**: Configurar via env var
3. **Modelo Inexistente**: Validar tratamento de erro
4. **Fallback**: Testar recuperação automática
5. **Diferentes Ambientes**: Dev, staging, production

### Verificações de Compatibilidade

- ✅ Retrocompatibilidade com configuração atual
- ✅ Graceful degradation se env var não definida
- ✅ Validação de modelo no startup
- ✅ Logs informativos sobre modelo em uso

## Benefícios da Implementação

### Operacionais
- **Flexibilidade de Deploy**: Diferentes modelos por ambiente
- **Testes A/B**: Comparação de modelos facilmente
- **Otimização de Recursos**: Modelos menores para dev/test

### Desenvolvimento
- **Configuração Simples**: Via variáveis de ambiente
- **Debugging Melhorado**: Logs claros sobre modelo usado
- **Manutenibilidade**: Código menos acoplado

### Produção
- **Fallback Automático**: Alta disponibilidade
- **Monitoramento**: Métricas por modelo
- **Escalabilidade**: Suporte a múltiplos workers com modelos diferentes

## Riscos e Mitigações

### Riscos Identificados
1. **Breaking Changes**: Alteração na assinatura de funções
2. **Configuração Incorreta**: Modelo inexistente
3. **Performance**: Overhead de validação

### Mitigações
1. **Versionamento**: Manter compatibilidade com valores default
2. **Validação**: Startup checks para modelos
3. **Caching**: Validação uma vez por startup

## Conclusão

A implementação de configuração de modelo via variáveis de ambiente é **altamente viável** e requer alterações em **7 arquivos principais** com esforço **médio** de desenvolvimento. A mudança proporcionará **flexibilidade significativa** sem comprometer a estabilidade atual do sistema.

**Tempo Estimado**: 2-3 dias de desenvolvimento + 1 dia de testes
**Complexidade**: Média
**Impacto**: Alto benefício, baixo risco

---

*Relatório gerado em: 02 de Agosto de 2025*
*Versão do Sistema: feat-session-control branch*
*Status: Pronto para implementação*
