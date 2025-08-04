**Task_ID:** PROTOTYPE-003
**Feature_Name:** Ollama AI Integration
**Task_Title:** Implement Direct Ollama Integration for Basic Sentiment Analysis

**Task_Description:**
Create a direct integration with Ollama for basic sentiment analysis of journal entries. This should be a simple HTTP client that sends journal content to a local Ollama instance and receives sentiment analysis results. The integration should be synchronous and straightforward, without complex abstraction layers, to quickly validate the AI processing concept.

**Acceptance_Criteria:**

- [ ] HTTP client can communicate with local Ollama instance
- [ ] Basic prompt is crafted for sentiment analysis of journal content
- [ ] Ollama response is parsed and structured into sentiment data
- [ ] Integration handles Ollama API errors gracefully
- [ ] Processing results include sentiment score and confidence level
- [ ] Simple retry mechanism for failed API calls (max 3 attempts)
- [ ] Processing time is logged for performance monitoring
- [ ] Configuration allows setting Ollama endpoint URL
- [ ] Basic validation ensures journal content is suitable for processing

**Technical_Specifications:**

- **Component(s):** Ollama HTTP Client, AI Processing Service
- **API Endpoint(s):** Internal service calls to Ollama (typically localhost:11434)
- **Data Model(s):** SentimentResult struct with Score, Label, Confidence fields
- **Key Logic:** HTTP client for Ollama, prompt engineering for sentiment analysis, response parsing
- **Non-Functional Requirements:** Processing time <10 seconds per journal, 90%+ success rate with local Ollama

**Dependencies:**

- Requires local Ollama installation and running instance
- PROTOTYPE-002: Journal endpoints must exist to provide content for processing

**Estimated_Effort:** Medium

--- END OF TASK ---
