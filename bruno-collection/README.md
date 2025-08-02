# englog API - Bruno Collection

Esta é uma collection completa do Bruno para testar a API englog, baseada na documentação Swagger encontrada na pasta `api/`.

## Estrutura da Collection

A collection está organizada nas seguintes pastas:

### 🔐 Auth
- **Register User**: Registra um novo usuário e retorna tokens de autenticação (cria sessão automaticamente)
- **Login User**: Faz login e obtém tokens de acesso (cria nova sessão)
- **Get Me**: Obtém informações do perfil do usuário autenticado
- **Refresh Token**: Renova os tokens de acesso usando o refresh token (atualiza atividade da sessão)
- **Logout User**: Invalida o refresh token e faz logout (desativa sessão atual)
- **Get Active Sessions**: Lista todas as sessões ativas do usuário
- **Logout From All Devices**: Desativa todas as sessões ativas do usuário
- **Session Management Flow**: Guia completo do fluxo de gerenciamento de sessões

### 🏥 Health
- **Health Check**: Verifica se a API está funcionando
- **Readiness Check**: Verifica se a API está pronta para receber requisições

### 📝 Log Entries
- **Create Log Entry**: Cria uma nova entrada de log
- **Get Log Entries**: Lista todas as entradas de log com filtros opcionais
- **Get Log Entry**: Obtém uma entrada específica por ID
- **Update Log Entry**: Atualiza uma entrada existente
- **Delete Log Entry**: Remove uma entrada
- **Bulk Create Log Entries**: Cria múltiplas entradas de uma vez

### 📁 Projects
- **Create Project**: Cria um novo projeto
- **Get Projects**: Lista todos os projetos
- **Get Project**: Obtém um projeto específico por ID
- **Update Project**: Atualiza um projeto existente
- **Delete Project**: Remove um projeto

### 📊 Analytics
- **Get Productivity Metrics**: Obtém métricas de produtividade
- **Get Activity Summary**: Obtém resumo de atividades

### 🏷️ Tags
- **Create Tag**: Cria uma nova tag
- **Get Tags**: Lista todas as tags
- **Get Popular Tags**: Obtém as tags mais populares
- **Get Recently Used Tags**: Obtém tags recentemente utilizadas
- **Search Tags**: Busca tags por nome
- **Get User Tag Usage**: Obtém estatísticas de uso de tags
- **Get Tag**: Obtém uma tag específica por ID
- **Update Tag**: Atualiza uma tag existente
- **Delete Tag**: Remove uma tag

### � Workers
- **Get Active Workers**: Lista workers ativos e suas informações
- **Worker Health Check**: Verifica o status de saúde do sistema de workers

### 📋 Tasks (Enhanced AI Insights)
- **Request Insight Generation**: Solicita geração de insights com IA (enhanced with structured context)
- **Request Skill Development Insight**: Insights focados em desenvolvimento de habilidades
- **Request Time Management Insight**: Análise de gestão de tempo e produtividade
- **Request Team Collaboration Insight**: Insights sobre colaboração e dinâmicas de equipe
- **Request Insight with Mixed Context**: Exemplo com contexto misto e lista grande de entries
- **Request Weekly Report**: Solicita geração de relatório semanal
- **Get Task Result**: Obtém resultado de uma task específica
- **Get Task Result by Report ID**: Obtém resultado usando ID do relatório
- **Request Insight - Invalid Payload**: Teste com payload inválido (enhanced validation)
- **Request Report - Invalid Date Format**: Teste com formato de data inválido
- **Get Task Result - Nonexistent Task**: Teste com task inexistente

> 🚀 **New in v2.1**: Enhanced context support with structured JSON objects for richer AI insights!
> 📖 See `ENHANCED_CONTEXT_GUIDE.md` for detailed documentation on the new features.

### �👤 Users
- **Get Profile**: Obtém o perfil do usuário
- **Update Profile**: Atualiza informações do perfil
- **Change Password**: Altera a senha do usuário
- **Delete Account**: Remove a conta do usuário

## Como Usar

### 1. Configuração do Ambiente

O arquivo `environments/Local.bru` contém as variáveis de ambiente:

```
vars {
  base_url: http://localhost:8080
  access_token:
  refresh_token:
}
```

Ajuste a `base_url` conforme necessário para o seu ambiente.

### 2. Fluxo de Autenticação

1. **Registre um usuário** usando "Register User" ou **faça login** com "Login User"
2. Os tokens serão automaticamente salvos nas variáveis de ambiente através dos scripts pós-resposta
3. As demais requisições autenticadas usarão automaticamente o `access_token`

### 3. Variáveis Dinâmicas

Algumas requisições usam variáveis para IDs:
- `{{log_entry_id}}`: ID de uma entrada de log
- `{{project_id}}`: ID de um projeto
- `{{tag_id}}`: ID de uma tag

Você pode definir essas variáveis:
- Manualmente no ambiente Local.bru
- Capturando das respostas de criação através de scripts
- Copiando IDs das listagens

### 4. Scripts Automáticos

As requisições de autenticação incluem scripts que automaticamente:
- Salvam tokens nas variáveis de ambiente após login/registro
- Limpam tokens após logout

### 5. Testes

Cada requisição inclui testes básicos que verificam:
- Status codes esperados
- Presença de campos obrigatórios nas respostas
- Estrutura básica dos dados retornados

## Gerenciamento de Sessões

A API englog implementa um sistema robusto de gerenciamento de sessões que oferece:

### Recursos de Sessão

- **Criação Automática**: Sessões são criadas automaticamente durante registro e login
- **Rastreamento de Atividade**: Cada requisição autenticada atualiza a atividade da sessão
- **Multi-device Support**: Suporte para múltiplas sessões simultâneas
- **Informações de Dispositivo**: Rastreamento de User-Agent e IP para segurança
- **Logout Granular**: Logout individual ou de todos os dispositivos

### Fluxo de Sessão

1. **Autenticação** → Cria nova sessão automaticamente
2. **Uso da API** → Atualiza `last_activity` a cada requisição
3. **Monitoramento** → Visualiza sessões ativas via `GET /v1/auth/sessions`
4. **Logout** → Desativa sessão atual ou todas as sessões

### Segurança

- Tokens hasheados antes do armazenamento
- Validação automática de sessões ativas
- Limpeza automática de sessões inativas
- Auditoria completa de atividade por dispositivo

### Endpoints de Sessão

| Endpoint | Propósito |
|----------|-----------|
| `GET /v1/auth/sessions` | Lista sessões ativas |
| `POST /v1/auth/logout-all` | Logout de todos os dispositivos |

## Exemplos de Dados

### Registro de Usuário
```json
{
  "email": "user@example.com",
  "password": "securePassword123",
  "first_name": "John",
  "last_name": "Doe",
  "timezone": "UTC"
}
```

### Entrada de Log
```json
{
  "title": "Implementação de autenticação JWT",
  "description": "Desenvolvimento completo do sistema de autenticação usando JWT com refresh tokens e middleware de proteção",
  "type": "development",
  "value_rating": "high",
  "impact_level": "team",
  "project_id": null,
  "start_time": "2025-07-31T09:00:00Z",
  "end_time": "2025-07-31T11:30:00Z",
  "tags": ["jwt", "authentication", "security", "backend"]
}
```

### Projeto
```json
{
  "name": "englog API",
  "description": "Engineering log management API",
  "repository_url": "https://github.com/garnizeh/englog",
  "status": "active",
  "tags": ["go", "api", "postgresql"]
}
```

## Dicas

### Campos de Entrada de Log

- **type**: Tipo de atividade (development, meeting, code_review, debugging, documentation, testing, deployment, research, planning, learning, maintenance, support, other)
- **value_rating**: Valor percebido da atividade (low, medium, high, critical)
- **impact_level**: Nível de impacto (personal, team, department, company)
- **start_time/end_time**: Horários em formato ISO 8601 (ex: 2025-07-31T09:00:00Z)
- **duration_minutes**: Calculado automaticamente pelo sistema
- **tags**: Array de strings para categorização flexível

## 🚀 Novos Endpoints: Workers e Tasks

### Workers API

Os endpoints de workers permitem monitorar e gerenciar o sistema distribuído:

```bash
# Listar workers ativos
GET /v1/workers

# Verificar saúde do sistema de workers
GET /v1/workers/health
```

### Tasks API

Os endpoints de tasks permitem solicitar processamento assíncrono com IA:

```bash
# Solicitar geração de insights
POST /v1/tasks/insights
{
  "user_id": "uuid-do-usuario",
  "entry_ids": ["uuid1", "uuid2"],
  "insight_type": "productivity",
  "context": "Análise semanal"
}

# Solicitar relatório semanal
POST /v1/tasks/reports
{
  "user_id": "uuid-do-usuario",
  "week_start": "2025-07-21",
  "week_end": "2025-07-27"
}

# Obter resultado da task
GET /v1/tasks/{task_id}/result
```

### Fluxo de Trabalho com Tasks

1. **Autenticar**: Fazer login para obter tokens
2. **Criar entrada de log**: Ter algumas entradas para análise
3. **Solicitar insight**: Usar POST /v1/tasks/insights
4. **Aguardar processamento**: A task é processada pelo worker
5. **Obter resultado**: Usar GET /v1/tasks/{task_id}/result

### Exemplos de Insight Types

- `productivity` - Análise de produtividade geral
- `patterns` - Identificação de padrões de trabalho
- `recommendations` - Sugestões de melhoria
- `summary` - Resumo de atividades

## Dicas

### Campos de Entrada de Log

- **type**: Tipo de atividade (development, meeting, code_review, debugging, documentation, testing, deployment, research, planning, learning, maintenance, support, other)
- **value_rating**: Valor percebido da atividade (low, medium, high, critical)
- **impact_level**: Nível de impacto (personal, team, department, company)
- **start_time/end_time**: Horários em formato ISO 8601 (ex: 2025-07-31T09:00:00Z)
- **duration_minutes**: Calculado automaticamente pelo sistema
- **tags**: Array de strings para categorização flexível

### Workflow Recomendado

1. **Ordem das Requisições**: Comece sempre com autenticação
2. **IDs de Teste**: Use os IDs de exemplo fornecidos ou capture-os das respostas
3. **Filtros**: Muitas listagens suportam filtros via query parameters
4. **Paginação**: Use `limit` e `offset` para paginar resultados
5. **Tokens**: Os tokens são automaticamente gerenciados pelos scripts
6. **Tasks Assíncronas**: Use os endpoints de tasks para processamento com IA
7. **Horários**: Use sempre formato ISO 8601 com timezone (Z para UTC)
8. **Contexto Estruturado**: Use objetos JSON para contexto rico nos insights

## 🧠 Enhanced AI Insights (v2.1)

### New Features

A collection agora demonstra as funcionalidades avançadas de geração de insights com contexto estruturado:

#### Structured Context Support
- **JSON Objects**: Contexto estruturado com dados aninhados
- **Backward Compatibility**: Mantém suporte a contexto string simples
- **Type-Specific Guidelines**: Instruções específicas por tipo de insight

#### Insight Types with Specialized Context

1. **Productivity Analysis**
   ```json
   {
     "time_blocks": ["morning", "afternoon"],
     "focus_areas": ["development", "meetings"],
     "performance_metrics": {
       "avg_daily_hours": 8.5,
       "productivity_score": 0.85
     }
   }
   ```

2. **Skill Development**
   ```json
   {
     "focus_areas": ["golang", "system_design"],
     "learning_goals": ["Improve concurrency patterns"],
     "current_skill_level": {
       "golang": "intermediate"
     }
   }
   ```

3. **Time Management**
   ```json
   {
     "date_range": {
       "start": "2025-07-01",
       "end": "2025-07-31"
     },
     "work_schedule": {
       "preferred_deep_work_hours": ["09:00-12:00"]
     }
   }
   ```

4. **Team Collaboration**
   ```json
   {
     "team_size": 6,
     "role": "senior_developer",
     "collaboration_types": ["code_reviews", "mentoring"]
   }
   ```

#### Enhanced Prompt Generation

The system now creates rich, structured prompts that include:
- User identification and context
- Insight type specific guidelines
- Entry count and smart ID truncation
- Structured context data
- Consistent output format instructions

#### Testing the Enhanced Features

1. **Run the enhanced requests** to see structured context in action
2. **Compare results** with simple string context
3. **Check task results** for improved AI responses
4. **Experiment with different** context structures

For detailed documentation, see `ENHANCED_CONTEXT_GUIDE.md`.

## Testando o Gerenciamento de Sessões

### Fluxo Completo de Teste

Para testar o sistema de sessões completo, execute as requisições nesta ordem:

1. **Register User** ou **Login User**
   - Cria automaticamente uma nova sessão
   - Salva os tokens nas variáveis de ambiente

2. **Get Active Sessions**
   - Deve mostrar 1 sessão ativa
   - Note o `created_at` e `last_activity`

3. **Get Me** (ou qualquer endpoint protegido)
   - Atualiza automaticamente a atividade da sessão
   - Session tracking transparente

4. **Get Active Sessions** novamente
   - Observe que `last_activity` foi atualizado
   - Demonstra o rastreamento automático

5. **Login User** novamente (mesmo usuário)
   - Cria uma segunda sessão para simular múltiplos dispositivos
   - Cada login = nova sessão

6. **Get Active Sessions**
   - Deve mostrar 2 sessões ativas agora
   - Diferentes IDs e timestamps

7. **Logout User**
   - Desativa apenas a sessão atual
   - Outras sessões permanecem ativas

8. **Logout From All Devices**
   - Desativa TODAS as sessões
   - Tokens ficam inválidos imediatamente

9. **Get Active Sessions** (final)
   - Deve mostrar 0 sessões ativas
   - Confirma que todas foram desativadas

### Cenários de Segurança

- **Dispositivo comprometido**: Use "Logout From All Devices"
- **Múltiplos dispositivos**: Monitore via "Get Active Sessions"
- **Atividade suspeita**: Verifique `last_activity` e `user_agent`
- **Limpeza periódica**: Logout automático de sessões antigas

## Troubleshooting

- **401 Unauthorized**: Verifique se o token está válido ou renove-o
- **Session not found**: Sessão pode ter sido desativada ou expirada - faça login novamente
- **404 Not Found**: Verifique se os IDs usados existem
- **Structured Context Issues**: Verifique se o JSON está bem formado
- **400 Bad Request**: Verifique a estrutura do JSON enviado
- **Multiple sessions**: Normal ter várias sessões ativas de diferentes dispositivos/logins
- **Conexão recusada**: Verifique se a API está rodando na URL configurada

---

Esta collection cobre todos os endpoints documentados na API englog e está pronta para uso em desenvolvimento e testes.
