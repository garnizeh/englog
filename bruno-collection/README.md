# EngLog API - Bruno Collection

Esta é a collection completa do Bruno para testar todos os endpoints da API EngLog.

## 📋 Estrutura da Collection

```
bruno-collection/
├── bruno.json                     # Configuração da collection
├── environments/
│   └── Local Development.bru      # Variáveis de ambiente
├── Health Checks/
│   └── Basic Health Check.bru     # Verificação de saúde da API
├── Journals/
│   ├── Create Simple Journal.bru  # Criar journal básico
│   ├── Create Journal with Rich Metadata.bru  # Journal com metadados
│   ├── Get All Journals.bru       # Listar todos os journals
│   ├── Get Journal by ID.bru      # Buscar journal específico
│   ├── Update Journal.bru         # Atualizar journal
│   └── Delete Journal.bru         # Deletar journal
├── AI Processing/
│   ├── Direct Sentiment Analysis.bru  # Análise de sentimento direta
│   ├── AI Health Check.bru        # Verificação de saúde da IA
│   └── AI Generate Journal.bru    # Geração de journal por IA
├── Error Scenarios/
│   ├── Invalid JSON Request.bru   # Teste de JSON inválido
│   ├── Empty Content Journal.bru  # Teste de conteúdo vazio
│   └── Journal Not Found.bru      # Teste de recurso não encontrado
└── Performance Tests/
    ├── Large Content Journal.bru  # Teste com conteúdo grande
    └── Rapid Sequential Requests.bru  # Teste de requisições rápidas
```

## 🚀 Como Usar

### 1. Instalar o Bruno

Baixe e instale o Bruno API Client:

- Site oficial: https://usebruno.com/
- GitHub: https://github.com/usebruno/bruno

### 2. Importar a Collection

1. Abra o Bruno
2. Clique em "Open Collection"
3. Navegue até a pasta `bruno-collection`
4. Selecione a pasta completa

### 3. Configurar o Ambiente

1. Certifique-se que o Docker está rodando:

   ```bash
   docker compose -f docker-compose.dev.yml up -d
   ```

2. Verifique se a API está rodando em `http://localhost:8080`

3. No Bruno, selecione o ambiente "Local Development"

### 4. Executar os Testes

#### Fluxo Recomendado de Teste:

1. **Health Check**

   - Execute "Basic Health Check" para verificar se a API está funcionando

2. **Criação de Journals**

   - Execute "Create Simple Journal" para criar uma entrada básica
   - Execute "Create Journal with Rich Metadata" para testar metadados

3. **Consulta de Journals**

   - Execute "Get All Journals" para ver todas as entradas
   - Copie um ID de journal dos resultados
   - Execute "Get Journal by ID" usando o ID copiado

4. **Atualização e Remoção**

   - Execute "Update Journal" com um ID válido
   - Execute "Delete Journal" com um ID válido

5. **AI Processing**

   - Execute "AI Health Check" para verificar o serviço de IA
   - Execute "Direct Sentiment Analysis" para testar análise
   - Execute "AI Generate Journal" para geração de conteúdo

6. **Testes de Erro**

   - Execute os cenários em "Error Scenarios" para validar tratamento de erros

7. **Testes de Performance**
   - Execute "Large Content Journal" para testar performance
   - Execute "Rapid Sequential Requests" várias vezes rapidamente

## 📊 Monitoramento de Performance

Durante os testes, observe:

- **Tempo de Resposta**:

  - Health checks: < 100ms
  - Operações básicas de journal: < 500ms
  - Processamento de IA: 1-3 segundos

- **Status Codes Esperados**:
  - 200 OK: Operações de leitura e atualização
  - 201 Created: Criação de journals
  - 204 No Content: Deleção bem-sucedida
  - 400 Bad Request: Erros de validação
  - 404 Not Found: Recursos não encontrados

## 🔧 Personalização

### Variáveis de Ambiente

Edite `environments/Local Development.bru` para alterar:

```bru
vars {
  baseUrl: http://localhost:8080
  # Adicione outras variáveis conforme necessário
}
```

### Adicionando Novos Requests

1. Clique com o botão direito na pasta apropriada
2. Selecione "New Request"
3. Configure método, URL e body
4. Adicione documentação na seção `docs`

## 🐛 Troubleshooting

### API não responde

```bash
# Verificar se os containers estão rodando
docker compose -f docker-compose.dev.yml ps

# Verificar logs
docker compose -f docker-compose.dev.yml logs api
```

### Erro de conexão com IA

```bash
# Verificar logs do Ollama
docker compose -f docker-compose.dev.yml logs ollama

# Verificar se o modelo está carregado
curl http://localhost:11434/api/tags
```

### Collection não carrega no Bruno

- Certifique-se de selecionar a pasta `bruno-collection` completa
- Verifique se o arquivo `bruno.json` está presente
- Tente fechar e reabrir o Bruno

## 📝 Contribuindo

Para adicionar novos testes:

1. Identifique a pasta apropriada ou crie uma nova
2. Crie um arquivo `.bru` com a estrutura padrão
3. Inclua documentação clara na seção `docs`
4. Teste o endpoint antes de commit
5. Atualize este README se necessário

## 📚 Referências

- **Documentação Completa**: `/docs/hands-on/PROTOTYPE-005-USE-CASES.md`
- **Schema JSON**: `/docs/hands-on/PROTOTYPE-005-JSON-SCHEMA.md`
- **Setup Docker**: `/docs/hands-on/DOCKER.md`
- **Task Original**: `/docs/planning/backlog/PROTOTYPE-008.md`
