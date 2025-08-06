# EngLog API Use Cases & Testing Guide

**Version:** 1.0
**Date:** August 5, 2025
**Status:** Complete
**Related:** PROTOTYPE-005-JSON-SCHEMA.md

This document provides practical usage examples of the EngLog API, serving as a development reference and foundation for test collection in Bruno API Client.

---

## 📋 Overview

The EngLog API provides endpoints for:

- ✅ **Journal Management**: Create, read, update, and delete diary entries
- ✅ **AI Processing**: Automatic sentiment analysis and insights
- ✅ **Health Checks**: System health monitoring
- 🔮 **AI Generation**: AI-assisted content generation (future)

---

## 🏥 Health Check Use Cases

### UC-001: Basic Health Check

**Scenario**: Check if the API is working
**Endpoint**: `GET /health`

```http
GET http://localhost:8080/health
```

**Expected Response (200 OK)**:

```json
{
  "status": "healthy",
  "timestamp": "2025-08-05T10:30:15Z",
  "version": "0.1.0"
}
```

**Bruno Collection Reference**:

- Folder: `Health Checks`
- Request: `Basic Health Check`

---

## 📖 Journal Management Use Cases

### UC-002: Create Simple Journal

**Scenario**: User creates a basic diary entry
**Endpoint**: `POST /journals`

```http
POST http://localhost:8080/journals
Content-Type: application/json

{
  "content": "Today was a productive day. I managed to complete all my tasks and felt accomplished."
}
```

**Expected Response (201 Created)**:

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

**Bruno Collection Reference**:

- Folder: `Journals`
- Request: `Create Simple Journal`

### UC-003: Create Journal with Rich Metadata

**Scenario**: User creates entry with structured information
**Endpoint**: `POST /journals`

```http
POST http://localhost:8080/journals
Content-Type: application/json

{
  "content": "Had an amazing team meeting today where we discussed the new project roadmap. The energy was high and everyone was engaged.",
  "metadata": {
    "mood": 8.5,
    "energy_level": 9,
    "tags": ["work", "team", "planning", "excitement"],
    "location": "conference_room_a",
    "participants": ["alice", "bob", "charlie", "diana"],
    "duration_minutes": 45,
    "weather": "sunny",
    "follow_up_needed": true,
    "meeting_details": {
      "type": "roadmap_planning",
      "quarter": "Q3_2025",
      "priority": "high",
      "next_steps": "document_decisions"
    },
    "mood_factors": {
      "sleep_quality": "good",
      "coffee_cups": 2,
      "exercise_morning": true
    }
  }
}
```

**Expected Response (201 Created)**:

```json
{
  "id": "660e8400-e29b-41d4-a716-446655440001",
  "content": "Had an amazing team meeting today where we discussed the new project roadmap. The energy was high and everyone was engaged.",
  "timestamp": "2025-08-05T11:00:00Z",
  "created_at": "2025-08-05T11:00:10Z",
  "updated_at": "2025-08-05T11:00:10Z",
  "metadata": {
    "mood": 8.5,
    "energy_level": 9,
    "tags": ["work", "team", "planning", "excitement"],
    "location": "conference_room_a",
    "participants": ["alice", "bob", "charlie", "diana"],
    "duration_minutes": 45,
    "weather": "sunny",
    "follow_up_needed": true,
    "meeting_details": {
      "type": "roadmap_planning",
      "quarter": "Q3_2025",
      "priority": "high",
      "next_steps": "document_decisions"
    },
    "mood_factors": {
      "sleep_quality": "good",
      "coffee_cups": 2,
      "exercise_morning": true
    }
  },
  "processing_result": {
    "status": "completed",
    "sentiment_result": {
      "score": 0.85,
      "label": "positive",
      "confidence": 0.94,
      "processed_at": "2025-08-05T11:00:13Z"
    },
    "processed_at": "2025-08-05T11:00:15Z",
    "processing_time": "3.2s"
  }
}
```

**Bruno Collection Reference**:

- Folder: `Journals`
- Request: `Create Rich Metadata Journal`

### UC-004: List All Journals

**Scenario**: User retrieves all their journal entries
**Endpoint**: `GET /journals`

```http
GET http://localhost:8080/journals
```

**Expected Response (200 OK)**:

```json
{
  "entries": [
    {
      "id": "660e8400-e29b-41d4-a716-446655440001",
      "content": "Had an amazing team meeting today where we discussed the new project roadmap.",
      "timestamp": "2025-08-05T11:00:00Z",
      "created_at": "2025-08-05T11:00:10Z",
      "updated_at": "2025-08-05T11:00:10Z",
      "metadata": {
        "mood": 8.5,
        "energy_level": 9,
        "tags": ["work", "team", "planning"]
      },
      "processing_result": {
        "status": "completed",
        "sentiment_result": {
          "score": 0.85,
          "label": "positive",
          "confidence": 0.94
        }
      }
    },
    {
      "id": "660e8400-e29b-41d4-a716-446655440000",
      "content": "Today was a productive day! I managed to complete several important tasks and felt quite satisfied with my progress.",
      "timestamp": "2025-08-05T10:30:00Z",
      "created_at": "2025-08-05T10:30:05Z",
      "updated_at": "2025-08-05T10:30:05Z",
      "metadata": {},
      "processing_result": {
        "status": "completed",
        "sentiment_result": {
          "score": 0.78,
          "label": "positive",
          "confidence": 0.91
        }
      }
    }
  ],
  "total_count": 2,
  "returned_count": 2
}
```

**Bruno Collection Reference**:

- Folder: `Journals`
- Request: `List All Journals`

### UC-005: Get Journal by ID

**Scenario**: User views a specific entry
**Endpoint**: `GET /journals/{id}`

```http
GET http://localhost:8080/journals/550e8400-e29b-41d4-a716-446655440000
```

**Expected Response (200 OK)**:

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

**Bruno Collection Reference**:

- Folder: `Journals`
- Request: `Get Journal by ID`

### UC-006: Journal Not Found

**Scenario**: Search for journal with non-existent ID
**Endpoint**: `GET /journals/{id}`

```http
GET http://localhost:8080/journals/nonexistent-id
```

**Expected Response (404 Not Found)**:

```json
{
  "error": "Journal not found",
  "status": 404,
  "timestamp": "2025-08-05T10:30:15Z"
}
```

**Bruno Collection Reference**:

- Folder: `Journals`
- Request: `Get Nonexistent Journal`

---

## ❌ Validation Error Use Cases

### UC-007: Validation - Empty Content

**Scenario**: User tries to create journal without content
**Endpoint**: `POST /journals`

```http
POST http://localhost:8080/journals
Content-Type: application/json

{
  "content": "",
  "metadata": {
    "mood": 5
  }
}
```

**Expected Response (400 Bad Request)**:

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
    }
  ]
}
```

**Bruno Collection Reference**:

- Folder: `Validation Errors`
- Request: `Empty Content Error`

### UC-008: Validation - Whitespace Only Content

**Scenario**: User tries to create journal with only whitespace
**Endpoint**: `POST /journals`

```http
POST http://localhost:8080/journals
Content-Type: application/json

{
  "content": "   \n\t   ",
  "metadata": {
    "mood": 5
  }
}
```

**Expected Response (400 Bad Request)**:

```json
{
  "error": "Validation failed",
  "status": 400,
  "timestamp": "2025-08-05T10:30:15Z",
  "validation_errors": [
    {
      "field": "content",
      "message": "Content cannot be only whitespace",
      "code": "INVALID_FORMAT"
    }
  ]
}
```

**Bruno Collection Reference**:

- Folder: `Validation Errors`
- Request: `Whitespace Only Content Error`

### UC-009: Validation - Content Too Long

**Scenario**: User tries to create journal with content exceeding 50,000 characters
**Endpoint**: `POST /journals`

```http
POST http://localhost:8080/journals
Content-Type: application/json

{
  "content": "This is a very long text that repeats many times to exceed the limit..." // 50,001 characters
}
```

**Expected Response (400 Bad Request)**:

```json
{
  "error": "Validation failed",
  "status": 400,
  "timestamp": "2025-08-05T10:30:15Z",
  "validation_errors": [
    {
      "field": "content",
      "message": "Content exceeds maximum length of 50000 characters",
      "code": "MAX_LENGTH_EXCEEDED"
    }
  ]
}
```

**Bruno Collection Reference**:

- Folder: `Validation Errors`
- Request: `Content Too Long Error`

### UC-010: Validation - Metadata with Too Many Fields

**Scenario**: User tries to create journal with more than 20 metadata fields
**Endpoint**: `POST /journals`

```http
POST http://localhost:8080/journals
Content-Type: application/json

{
  "content": "Valid content",
  "metadata": {
    "field1": "value1",
    "field2": "value2",
    // ... up to field21
    "field21": "value21"
  }
}
```

**Expected Response (400 Bad Request)**:

```json
{
  "error": "Validation failed",
  "status": 400,
  "timestamp": "2025-08-05T10:30:15Z",
  "validation_errors": [
    {
      "field": "metadata",
      "message": "Metadata cannot have more than 20 fields",
      "code": "TOO_MANY_FIELDS"
    }
  ]
}
```

**Bruno Collection Reference**:

- Folder: `Validation Errors`
- Request: `Too Many Metadata Fields Error`

### UC-011: Validation - Metadata Key Too Long

**Scenario**: User tries to use metadata key with more than 100 characters
**Endpoint**: `POST /journals`

```http
POST http://localhost:8080/journals
Content-Type: application/json

{
  "content": "Valid content",
  "metadata": {
    "this_metadata_key_is_extremely_long_and_definitely_exceeds_the_maximum_allowed_limit_of_one_hundred_characters_for_keys": "value"
  }
}
```

**Expected Response (400 Bad Request)**:

```json
{
  "error": "Validation failed",
  "status": 400,
  "timestamp": "2025-08-05T10:30:15Z",
  "validation_errors": [
    {
      "field": "metadata",
      "message": "Metadata key 'this_metadata_key_is_extremely_long_and_definitely_exceeds_the_maximum_allowed_limit_of_one_hundred_characters_for_keys' exceeds maximum length of 100 characters",
      "code": "KEY_TOO_LONG"
    }
  ]
}
```

**Bruno Collection Reference**:

- Folder: `Validation Errors`
- Request: `Metadata Key Too Long Error`

### UC-012: Validation - Empty Metadata Key

**Scenario**: User tries to use empty key in metadata
**Endpoint**: `POST /journals`

```http
POST http://localhost:8080/journals
Content-Type: application/json

{
  "content": "Valid content",
  "metadata": {
    "": "value_with_empty_key",
    "valid_key": "valid_value"
  }
}
```

**Expected Response (400 Bad Request)**:

```json
{
  "error": "Validation failed",
  "status": 400,
  "timestamp": "2025-08-05T10:30:15Z",
  "validation_errors": [
    {
      "field": "metadata",
      "message": "Metadata keys cannot be empty",
      "code": "INVALID_KEY"
    }
  ]
}
```

**Bruno Collection Reference**:

- Folder: `Validation Errors`
- Request: `Empty Metadata Key Error`

### UC-013: Validation - Array with Too Many Elements

**Scenario**: User tries to use array with more than 50 elements in metadata
**Endpoint**: `POST /journals`

```http
POST http://localhost:8080/journals
Content-Type: application/json

{
  "content": "Valid content",
  "metadata": {
    "tags": ["tag1", "tag2", "tag3", "...", "tag51"] // 51 elements
  }
}
```

**Expected Response (400 Bad Request)**:

```json
{
  "error": "Validation failed",
  "status": 400,
  "timestamp": "2025-08-05T10:30:15Z",
  "validation_errors": [
    {
      "field": "metadata",
      "message": "Array 'tags' cannot have more than 50 elements",
      "code": "TOO_MANY_ELEMENTS"
    }
  ]
}
```

**Bruno Collection Reference**:

- Folder: `Validation Errors`
- Request: `Array Too Many Elements Error`

### UC-014: Validation - Multiple Errors

**Scenario**: Request with multiple validation problems
**Endpoint**: `POST /journals`

```http
POST http://localhost:8080/journals
Content-Type: application/json

{
  "content": "   ",
  "metadata": {
    "": "empty_key",
    "very_long_key_that_exceeds_the_maximum_allowed_limit_of_one_hundred_characters_for_metadata_keys": "value",
    "very_long_string": "This string value is too long and definitely exceeds the maximum limit of one thousand characters..." // 1001 characters
  }
}
```

**Expected Response (400 Bad Request)**:

```json
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
      "message": "Metadata key 'very_long_key_that_exceeds_the_maximum_allowed_limit_of_one_hundred_characters_for_metadata_keys' exceeds maximum length of 100 characters",
      "code": "KEY_TOO_LONG"
    },
    {
      "field": "metadata",
      "message": "String value for key 'very_long_string' exceeds maximum length of 1000 characters",
      "code": "VALUE_TOO_LONG"
    }
  ]
}
```

**Bruno Collection Reference**:

- Folder: `Validation Errors`
- Request: `Multiple Validation Errors`

---

## 🤖 AI Processing Use Cases

### UC-015: Journal with Pending Processing

**Scenario**: Journal created but AI still processing
**Endpoint**: Can occur on any journal creation

**Expected Response (201 Created)**:

```json
{
  "id": "770e8400-e29b-41d4-a716-446655440002",
  "content": "Today I felt a little anxious about tomorrow's presentation.",
  "timestamp": "2025-08-05T12:00:00Z",
  "created_at": "2025-08-05T12:00:10Z",
  "updated_at": "2025-08-05T12:00:10Z",
  "metadata": {
    "mood": 4,
    "tags": ["anxiety", "work"]
  },
  "processing_result": {
    "status": "pending",
    "processed_at": null,
    "processing_time": null
  }
}
```

**Bruno Collection Reference**:

- Folder: `AI Processing`
- Request: `Journal with Pending Processing`

### UC-016: Journal with Processing Failed

**Scenario**: Sentiment analysis failure
**Endpoint**: Can occur on any journal creation

**Expected Response (201 Created)**:

```json
{
  "id": "880e8400-e29b-41d4-a716-446655440003",
  "content": "Reflections about today...",
  "timestamp": "2025-08-05T13:00:00Z",
  "created_at": "2025-08-05T13:00:10Z",
  "updated_at": "2025-08-05T13:00:10Z",
  "metadata": null,
  "processing_result": {
    "status": "failed",
    "error": "AI service temporarily unavailable",
    "processed_at": "2025-08-05T13:00:15Z",
    "processing_time": "5.0s"
  }
}
```

**Bruno Collection Reference**:

- Folder: `AI Processing`
- Request: `Journal with Failed Processing`

### UC-017: Negative Sentiment

**Scenario**: Journal with negative content
**Endpoint**: `POST /journals`

```http
POST http://localhost:8080/journals
Content-Type: application/json

{
  "content": "Today was a terrible day. Everything went wrong and I feel completely unmotivated. I couldn't achieve any of my goals.",
  "metadata": {
    "mood": 2,
    "tags": ["frustration", "demotivation"]
  }
}
```

**Expected Response (201 Created)**:

```json
{
  "id": "990e8400-e29b-41d4-a716-446655440004",
  "content": "Today was a terrible day. Everything went wrong and I feel completely unmotivated. I couldn't achieve any of my goals.",
  "timestamp": "2025-08-05T14:00:00Z",
  "created_at": "2025-08-05T14:00:10Z",
  "updated_at": "2025-08-05T14:00:10Z",
  "metadata": {
    "mood": 2,
    "tags": ["frustration", "demotivation"]
  },
  "processing_result": {
    "status": "completed",
    "sentiment_result": {
      "score": -0.75,
      "label": "negative",
      "confidence": 0.91,
      "processed_at": "2025-08-05T14:00:13Z"
    },
    "processed_at": "2025-08-05T14:00:15Z",
    "processing_time": "2.8s"
  }
}
```

**Bruno Collection Reference**:

- Folder: `AI Processing`
- Request: `Negative Sentiment Journal`

### UC-018: Neutral Sentiment

**Scenario**: Journal with neutral content
**Endpoint**: `POST /journals`

```http
POST http://localhost:8080/journals
Content-Type: application/json

{
  "content": "Today was a regular day. I woke up at 7am, had coffee, went to work, had lunch at the usual time and came home at 6pm.",
  "metadata": {
    "mood": 5,
    "tags": ["routine", "daily"]
  }
}
```

**Expected Response (201 Created)**:

```json
{
  "id": "aa0e8400-e29b-41d4-a716-446655440005",
  "content": "Today was a regular day. I woke up at 7am, had coffee, went to work, had lunch at the usual time and came home at 6pm.",
  "timestamp": "2025-08-05T15:00:00Z",
  "created_at": "2025-08-05T15:00:10Z",
  "updated_at": "2025-08-05T15:00:10Z",
  "metadata": {
    "mood": 5,
    "tags": ["routine", "daily"]
  },
  "processing_result": {
    "status": "completed",
    "sentiment_result": {
      "score": 0.05,
      "label": "neutral",
      "confidence": 0.83,
      "processed_at": "2025-08-05T15:00:13Z"
    },
    "processed_at": "2025-08-05T15:00:15Z",
    "processing_time": "3.1s"
  }
}
```

**Bruno Collection Reference**:

- Folder: `AI Processing`
- Request: `Neutral Sentiment Journal`

---

## 🎯 Edge Cases Use Cases

### UC-019: Malformed JSON

**Scenario**: Request with invalid JSON
**Endpoint**: `POST /journals`

```http
POST http://localhost:8080/journals
Content-Type: application/json

{
  "content": "Valid content",
  "metadata": {
    "mood": 5,
    "tags": ["test"
  }
  // Malformed JSON - missing closing bracket and brace
```

**Expected Response (400 Bad Request)**:

```json
{
  "error": "Invalid JSON format",
  "status": 400,
  "timestamp": "2025-08-05T10:30:15Z"
}
```

**Bruno Collection Reference**:

- Folder: `Edge Cases`
- Request: `Malformed JSON Error`

### UC-020: Missing Content-Type

**Scenario**: Request without Content-Type header
**Endpoint**: `POST /journals`

```http
POST http://localhost:8080/journals

{
  "content": "Content without content-type"
}
```

**Expected Response (400 Bad Request)**:

```json
{
  "error": "Content-Type must be application/json",
  "status": 400,
  "timestamp": "2025-08-05T10:30:15Z"
}
```

**Bruno Collection Reference**:

- Folder: `Edge Cases`
- Request: `Missing Content-Type Error`

### UC-021: Unsupported HTTP Method

**Scenario**: Use unimplemented HTTP method
**Endpoint**: `PATCH /journals`

```http
PATCH http://localhost:8080/journals
Content-Type: application/json

{
  "content": "PATCH attempt"
}
```

**Expected Response (405 Method Not Allowed)**:

```json
{
  "error": "Method not allowed",
  "status": 405,
  "timestamp": "2025-08-05T10:30:15Z"
}
```

**Bruno Collection Reference**:

- Folder: `Edge Cases`
- Request: `Method Not Allowed Error`

### UC-022: Non-existent Endpoint

**Scenario**: Access route that doesn't exist
**Endpoint**: `GET /nonexistent`

```http
GET http://localhost:8080/nonexistent
```

**Expected Response (404 Not Found)**:

```json
{
  "error": "Endpoint not found",
  "status": 404,
  "timestamp": "2025-08-05T10:30:15Z"
}
```

**Bruno Collection Reference**:

- Folder: `Edge Cases`
- Request: `Endpoint Not Found Error`

---

## 📊 Performance Use Cases

### UC-023: Journal with Maximum Content

**Scenario**: Create journal with 50,000 characters (maximum limit)
**Endpoint**: `POST /journals`

```http
POST http://localhost:8080/journals
Content-Type: application/json

{
  "content": "This is a very long text with exactly 50000 characters..." // Exactly 50,000 characters
}
```

**Expected Response (201 Created)**:

```json
{
  "id": "bb0e8400-e29b-41d4-a716-446655440006",
  "content": "This is a very long text with exactly 50000 characters...",
  "timestamp": "2025-08-05T16:00:00Z",
  "created_at": "2025-08-05T16:00:10Z",
  "updated_at": "2025-08-05T16:00:10Z",
  "metadata": null,
  "processing_result": {
    "status": "completed",
    "sentiment_result": {
      "score": 0.15,
      "label": "neutral",
      "confidence": 0.76,
      "processed_at": "2025-08-05T16:00:25Z"
    },
    "processed_at": "2025-08-05T16:00:30Z",
    "processing_time": "15.2s"
  }
}
```

**Bruno Collection Reference**:

- Folder: `Performance`
- Request: `Maximum Content Length Journal`
- **Expectation**: Response time < 30 seconds

### UC-024: Metadata with 20 Fields

**Scenario**: Use maximum allowed fields in metadata
**Endpoint**: `POST /journals`

```http
POST http://localhost:8080/journals
Content-Type: application/json

{
  "content": "Journal with complex metadata at maximum limit.",
  "metadata": {
    "field1": "value1",
    "field2": "value2",
    "field3": "value3",
    "field4": "value4",
    "field5": "value5",
    "field6": "value6",
    "field7": "value7",
    "field8": "value8",
    "field9": "value9",
    "field10": "value10",
    "field11": "value11",
    "field12": "value12",
    "field13": "value13",
    "field14": "value14",
    "field15": "value15",
    "field16": "value16",
    "field17": "value17",
    "field18": "value18",
    "field19": "value19",
    "field20": "value20"
  }
}
```

**Bruno Collection Reference**:

- Folder: `Performance`
- Request: `Maximum Metadata Fields Journal`
- **Expectation**: Response time < 5 seconds

---

## 🧪 Bruno Collection Structure

```
EngLog API Collection/
├── 📁 Health Checks/
│   └── Basic Health Check
├── 📁 Journals/
│   ├── Create Simple Journal
│   ├── Create Rich Metadata Journal
│   ├── Get All Journals
│   ├── Get Journal by ID
│   └── Get Nonexistent Journal
├── 📁 Validation Errors/
│   ├── Empty Content Error
│   ├── Whitespace Only Content Error
│   ├── Content Too Long Error
│   ├── Too Many Metadata Fields Error
│   ├── Metadata Key Too Long Error
│   ├── Empty Metadata Key Error
│   ├── Array Too Many Elements Error
│   └── Multiple Validation Errors
├── 📁 AI Processing/
│   ├── Journal with Pending Processing
│   ├── Journal with Failed Processing
│   ├── Negative Sentiment Journal
│   └── Neutral Sentiment Journal
├── 📁 Edge Cases/
│   ├── Malformed JSON Error
│   ├── Missing Content-Type Error
│   ├── Method Not Allowed Error
│   └── Endpoint Not Found Error
└── 📁 Performance/
    ├── Maximum Content Length Journal
    └── Maximum Metadata Fields Journal
```

---

## 🎯 Testing Guidelines

### Bruno Environment Configuration

1. **Base URL**: `http://localhost:8080`
2. **Global Headers**:
   ```
   Content-Type: application/json
   Accept: application/json
   ```

### Suggested Environment Variables

```javascript
{
  "baseUrl": "http://localhost:8080",
  "journalId": "550e8400-e29b-41d4-a716-446655440000",
  "timestamp": "{{$moment.format('YYYY-MM-DDTHH:mm:ss.SSSZ')}}"
}
```

### Automated Tests (Bruno Scripts)

Each request should include basic tests:

```javascript
// For success requests
test("should return success status", function () {
  expect(res.getStatus()).to.equal(200); // or 201
});

test("should return valid JSON", function () {
  expect(res.getHeaders()["content-type"]).to.contain("application/json");
});

// For created journals
test("should have valid UUID", function () {
  const body = res.getBody();
  expect(body.id).to.match(
    /^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i
  );
});

// For errors
test("should return error object", function () {
  const body = res.getBody();
  expect(body.error).to.be.a("string");
  expect(body.status).to.be.a("number");
  expect(body.timestamp).to.be.a("string");
});
```

### Performance Benchmarks

- **Health Check**: < 50ms
- **Simple Journal Creation**: < 5s
- **Complex Journal Creation**: < 10s
- **Get All Journals**: < 2s
- **Maximum Content Journal**: < 30s

---

This document serves as a complete reference for testing all aspects of the EngLog Phase 0 API, ensuring coverage of functionalities, validations, AI processing and edge cases.
