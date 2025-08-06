# EngLog JSON Schema Documentation

**Version:** 1.0
**Date:** August 5, 2025
**Status:** Complete

This document defines the comprehensive JSON schemas used throughout the EngLog API for journal entries, AI processing results, and all request/response formats.

---

## 📋 Core Schemas

### Journal Entry Schema

The `Journal` object represents a complete journal entry in the system.

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "content": "Today was a wonderful day filled with new experiences...",
  "timestamp": "2025-08-05T10:30:00Z",
  "created_at": "2025-08-05T10:30:15Z",
  "updated_at": "2025-08-05T10:30:15Z",
  "metadata": {
    "mood": 8,
    "tags": ["work", "productivity"],
    "location": "home"
  },
  "processing_result": {
    "status": "completed",
    "sentiment_result": {
      "score": 0.75,
      "label": "positive",
      "confidence": 0.92,
      "processed_at": "2025-08-05T10:30:18Z"
    },
    "processed_at": "2025-08-05T10:30:20Z",
    "processing_time": "2.5s"
  }
}
```

#### Field Specifications

| Field               | Type   | Required | Description                 | Constraints     |
| ------------------- | ------ | -------- | --------------------------- | --------------- |
| `id`                | string | Yes      | Unique identifier (UUID v4) | Read-only       |
| `content`           | string | Yes      | Main journal text           | 1-50,000 chars  |
| `timestamp`         | string | Yes      | User's entry timestamp      | ISO 8601 format |
| `created_at`        | string | Yes      | System creation time        | ISO 8601 format |
| `updated_at`        | string | Yes      | System update time          | ISO 8601 format |
| `metadata`          | object | No       | Additional structured data  | Max 20 fields   |
| `processing_result` | object | No       | AI analysis results         | Read-only       |

---

## 📝 Request Schemas

### Create Journal Request

Used for `POST /journals` endpoint.

```json
{
  "content": "Today I learned something amazing about Go programming...",
  "metadata": {
    "mood": 7,
    "tags": ["learning", "tech"],
    "location": "office"
  }
}
```

#### Validation Rules

- **content** (required):

  - Must not be empty after trimming whitespace
  - Minimum length: 1 character
  - Maximum length: 50,000 characters
  - Cannot be only whitespace

- **metadata** (optional):
  - Maximum 20 fields allowed
  - Key constraints:
    - Cannot be empty
    - Maximum 100 characters per key
  - Value constraints:
    - String values: maximum 1,000 characters
    - Arrays: maximum 50 elements, no nested arrays/objects
    - Objects: maximum 10 fields, one level deep only
    - Supported types: string, number, boolean, null, array, object

#### Example with Complex Metadata

```json
{
  "content": "Had an amazing team meeting today where we discussed the new project roadmap.",
  "metadata": {
    "mood": 8.5,
    "energy_level": 9,
    "tags": ["work", "team", "planning", "excitement"],
    "location": "conference_room_a",
    "participants": ["alice", "bob", "charlie"],
    "duration_minutes": 45,
    "weather": "sunny",
    "follow_up_needed": true,
    "meeting_details": {
      "type": "roadmap_planning",
      "quarter": "Q3_2025",
      "priority": "high"
    }
  }
}
```

---

## 🤖 AI Processing Schemas

### Processing Result Schema

Contains complete AI analysis results for a journal entry.

```json
{
  "status": "completed",
  "sentiment_result": {
    "score": 0.75,
    "label": "positive",
    "confidence": 0.92,
    "processed_at": "2025-08-05T10:30:18Z"
  },
  "processed_at": "2025-08-05T10:30:20Z",
  "processing_time": "2.5s",
  "error": null
}
```

#### Processing Status Values

| Status        | Description                       |
| ------------- | --------------------------------- |
| `"pending"`   | Processing queued but not started |
| `"completed"` | Processing finished successfully  |
| `"failed"`    | Processing failed with error      |

#### Sentiment Result Schema

```json
{
  "score": 0.75,
  "label": "positive",
  "confidence": 0.92,
  "processed_at": "2025-08-05T10:30:18Z"
}
```

**Field Specifications:**

- **score**: Float from -1.0 (very negative) to 1.0 (very positive), 0.0 is neutral
- **label**: One of `"positive"`, `"negative"`, or `"neutral"`
- **confidence**: Float from 0.0 (no confidence) to 1.0 (maximum confidence)
- **processed_at**: ISO 8601 timestamp when analysis was performed

---

## 🎯 AI Generation Schemas

### Prompt Request Schema

Used for AI-assisted journal generation.

```json
{
  "prompt": "Write about a day when I felt grateful",
  "context": "I've been working on mindfulness practices lately",
  "metadata": {
    "mood_preference": "positive",
    "length": "medium"
  }
}
```

#### Validation Rules

- **prompt** (required):

  - Minimum length: 3 characters
  - Maximum length: 2,000 characters
  - Cannot be only whitespace

- **context** (optional):

  - Maximum length: 5,000 characters

- **metadata** (optional):
  - Maximum 10 fields allowed
  - Same value constraints as journal metadata

---

## ❌ Error Response Schemas

### Validation Error Response

Returned when request validation fails (HTTP 400).

```json
{
  "error": "Validation failed",
  "status": 400,
  "timestamp": "2025-08-05T10:30:15Z",
  "validation_errors": [
    {
      "field": "content",
      "message": "Content is required and cannot be empty",
      "code": "REQUIRED"
    },
    {
      "field": "metadata",
      "message": "Metadata key 'very_long_key_name_that_exceeds_limit' exceeds maximum length of 100 characters",
      "code": "KEY_TOO_LONG"
    }
  ]
}
```

#### Validation Error Codes

| Code                  | Description                                     |
| --------------------- | ----------------------------------------------- |
| `REQUIRED`            | Required field is missing or empty              |
| `INVALID_FORMAT`      | Field format is invalid (e.g., only whitespace) |
| `MAX_LENGTH_EXCEEDED` | Field exceeds maximum allowed length            |
| `MIN_LENGTH_NOT_MET`  | Field is below minimum required length          |
| `TOO_MANY_FIELDS`     | Object has too many fields                      |
| `INVALID_KEY`         | Object key is invalid or empty                  |
| `KEY_TOO_LONG`        | Object key exceeds maximum length               |
| `INVALID_VALUE`       | Value type is not supported                     |
| `INVALID_JSON`        | JSON syntax is malformed                        |

### General Error Response

Returned for non-validation errors.

```json
{
  "error": "Internal server error",
  "status": 500,
  "timestamp": "2025-08-05T10:30:15Z"
}
```

---

## 📊 Success Response Schemas

### Create Journal Response

Returns the created journal entry (HTTP 201).

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "content": "Today I learned something amazing about Go programming...",
  "timestamp": "2025-08-05T10:30:00Z",
  "created_at": "2025-08-05T10:30:15Z",
  "updated_at": "2025-08-05T10:30:15Z",
  "metadata": {
    "mood": 7,
    "tags": ["learning", "tech"],
    "location": "office"
  },
  "processing_result": {
    "status": "completed",
    "sentiment_result": {
      "score": 0.65,
      "label": "positive",
      "confidence": 0.88,
      "processed_at": "2025-08-05T10:30:18Z"
    },
    "processed_at": "2025-08-05T10:30:20Z",
    "processing_time": "3.2s"
  }
}
```

### Get All Journals Response

Returns array of journal entries (HTTP 200).

```json
[
  {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "content": "First journal entry...",
    "timestamp": "2025-08-05T10:30:00Z",
    "created_at": "2025-08-05T10:30:15Z",
    "updated_at": "2025-08-05T10:30:15Z",
    "metadata": { "mood": 7 },
    "processing_result": {
      "status": "completed",
      "sentiment_result": {
        "score": 0.65,
        "label": "positive",
        "confidence": 0.88,
        "processed_at": "2025-08-05T10:30:18Z"
      }
    }
  },
  {
    "id": "660e8400-e29b-41d4-a716-446655440001",
    "content": "Second journal entry...",
    "timestamp": "2025-08-05T11:00:00Z",
    "created_at": "2025-08-05T11:00:10Z",
    "updated_at": "2025-08-05T11:00:10Z",
    "metadata": null,
    "processing_result": {
      "status": "failed",
      "error": "AI service temporarily unavailable"
    }
  }
]
```

---

## 🔧 Schema Evolution Guidelines

### Phase 0 (Current) Scope

- Basic journal CRUD operations
- Simple sentiment analysis
- Flexible metadata support
- Comprehensive validation

### Future Phase Considerations

The current schema is designed to evolve gracefully:

1. **Extensible Metadata**: The flexible `metadata` object allows future features without breaking changes
2. **Processing Results**: Additional AI analysis types can be added to `processing_result`
3. **Versioning Ready**: Schema structure supports API versioning when needed

### Backward Compatibility

- All new fields will be optional
- Existing field types and constraints will not change
- Deprecation warnings will be provided before removing fields

---

## 📚 Usage Examples

### Complete Journal Creation Flow

```bash
# Create a journal entry
curl -X POST http://localhost:8080/journals \
  -H "Content-Type: application/json" \
  -d '{
    "content": "Today was incredibly productive. I completed three major tasks and felt really accomplished.",
    "metadata": {
      "mood": 9,
      "energy_level": 8,
      "tags": ["productivity", "work", "accomplishment"],
      "location": "home_office",
      "work_hours": 8,
      "completed_tasks": ["feature_implementation", "code_review", "documentation"],
      "mood_factors": {
        "weather": "sunny",
        "sleep_quality": "good",
        "exercise": true
      }
    }
  }'

# Expected Response (201 Created)
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "content": "Today was incredibly productive. I completed three major tasks and felt really accomplished.",
  "timestamp": "2025-08-05T10:30:00Z",
  "created_at": "2025-08-05T10:30:15Z",
  "updated_at": "2025-08-05T10:30:15Z",
  "metadata": {
    "mood": 9,
    "energy_level": 8,
    "tags": ["productivity", "work", "accomplishment"],
    "location": "home_office",
    "work_hours": 8,
    "completed_tasks": ["feature_implementation", "code_review", "documentation"],
    "mood_factors": {
      "weather": "sunny",
      "sleep_quality": "good",
      "exercise": true
    }
  },
  "processing_result": {
    "status": "completed",
    "sentiment_result": {
      "score": 0.85,
      "label": "positive",
      "confidence": 0.94,
      "processed_at": "2025-08-05T10:30:18Z"
    },
    "processed_at": "2025-08-05T10:30:20Z",
    "processing_time": "2.8s"
  }
}
```

### Validation Error Example

```bash
# Invalid request with empty content
curl -X POST http://localhost:8080/journals \
  -H "Content-Type: application/json" \
  -d '{
    "content": "   ",
    "metadata": {
      "mood": "invalid_type",
      "": "empty_key_not_allowed",
      "very_long_key_that_definitely_exceeds_the_maximum_allowed_length_of_one_hundred_characters": "value"
    }
  }'

# Expected Response (400 Bad Request)
{
  "error": "Validation failed",
  "status": 400,
  "timestamp": "2025-08-05T10:30:15Z",
  "validation_errors": [
    {
      "field": "content",
      "message": "Content cannot be only whitespace",
      "code": "INVALID_FORMAT"
    },
    {
      "field": "metadata",
      "message": "Metadata keys cannot be empty",
      "code": "INVALID_KEY"
    },
    {
      "field": "metadata",
      "message": "Metadata key 'very_long_key_that_definitely_exceeds_the_maximum_allowed_length_of_one_hundred_characters' exceeds maximum length of 100 characters",
      "code": "KEY_TOO_LONG"
    }
  ]
}
```

---

## 🎯 Implementation Notes

### Performance Considerations

- **Validation Time**: Schema validation completes in <10ms for typical requests
- **JSON Parsing**: Efficient streaming parser handles large content gracefully
- **Memory Usage**: Validation reuses structures to minimize allocations

### Security Features

- **Input Sanitization**: All string fields are validated for length and content
- **XSS Prevention**: Content is properly escaped in responses
- **Data Validation**: Strict type checking prevents injection attacks

### Monitoring and Debugging

- **Structured Logging**: All validation errors are logged with context
- **Error Codes**: Consistent error codes enable client-side error handling
- **Field-Level Errors**: Precise error messages help with debugging

---

This schema documentation serves as the definitive reference for all JSON structures used in the EngLog API during Phase 0 development and provides the foundation for future enhancements.
