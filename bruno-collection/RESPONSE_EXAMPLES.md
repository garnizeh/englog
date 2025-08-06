# Exemplos de Respostas da API EngLog

Este arquivo contém exemplos das respostas esperadas da API para facilitar o entendimento dos resultados.

## Health Check

**Request**: `GET /health`

**Response (200 OK)**:

```json
{
  "status": "healthy",
  "timestamp": "2025-08-05T10:30:15Z",
  "version": "0.1.0"
}
```

## Journal Creation

**Request**: `POST /journals`

```json
{
  "content": "Today was a productive day. I managed to complete all my tasks and felt accomplished."
}
```

**Response (201 Created)**:

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "content": "Today was a productive day. I managed to complete all my tasks and felt accomplished.",
  "timestamp": "2025-08-05T10:30:00Z",
  "created_at": "2025-08-05T10:30:15Z",
  "updated_at": "2025-08-05T10:30:15Z",
  "metadata": null,
  "processing_result": {
    "status": "completed",
    "sentiment_result": {
      "score": 0.75,
      "label": "positive",
      "confidence": 0.88,
      "processed_at": "2025-08-05T10:30:18Z"
    },
    "processed_at": "2025-08-05T10:30:20Z",
    "processing_time": "2.5s"
  }
}
```

## AI Sentiment Analysis

**Request**: `POST /ai/analyze-sentiment`

```json
{
  "text": "I'm feeling quite anxious about the presentation tomorrow."
}
```

**Response (200 OK)**:

```json
{
  "sentiment_result": {
    "score": -0.45,
    "label": "negative",
    "confidence": 0.76,
    "processed_at": "2025-08-05T10:35:22Z"
  },
  "processing_time": "1.8s"
}
```

## Error Examples

### Invalid JSON (400 Bad Request)

```json
{
  "error": "Invalid JSON format",
  "details": "Unexpected end of JSON input",
  "timestamp": "2025-08-05T10:30:15Z"
}
```

### Journal Not Found (404 Not Found)

```json
{
  "error": "Journal not found",
  "id": "00000000-0000-0000-0000-000000000000",
  "timestamp": "2025-08-05T10:30:15Z"
}
```

### Validation Error (400 Bad Request)

```json
{
  "error": "Validation failed",
  "details": "Content cannot be empty",
  "timestamp": "2025-08-05T10:30:15Z"
}
```
