**Task_ID:** PROTOTYPE-002
**Feature_Name:** Journal CRUD Endpoints
**Task_Title:** Implement Basic Journal Creation and Retrieval Endpoints

**Task_Description:**
Create the essential REST API endpoints for journal management including creating new journal entries and retrieving existing ones. The endpoints should accept simple JSON data structures and store them in memory without any authentication requirements. This enables testing of the core journaling functionality and provides the foundation for AI processing integration.

**Acceptance_Criteria:**

- [ ] POST /journals endpoint creates new journal entries
- [ ] GET /journals endpoint returns list of all journals
- [ ] GET /journals/{id} endpoint returns specific journal by ID
- [ ] JSON request/response format is consistent and well-defined
- [ ] Unique IDs are generated for each journal entry (UUID or simple counter)
- [ ] Basic input validation prevents empty or malformed requests
- [ ] HTTP status codes are appropriate (200, 201, 400, 404)
- [ ] Request and response bodies are logged for debugging
- [ ] Error responses include helpful error messages
- [ ] Journal data structure includes content, timestamp, and metadata fields

**Technical_Specifications:**

- **Component(s):** Go HTTP Handlers, JSON Processing
- **API Endpoint(s):** `POST /journals`, `GET /journals`, `GET /journals/{id}`
- **Data Model(s):** Journal struct with ID, Content, CreatedAt, Metadata fields
- **Key Logic:** HTTP routing, JSON validation, in-memory CRUD operations, ID generation
- **Non-Functional Requirements:** Response time <100ms, handles 100+ journals in memory, clear error messages

**Dependencies:**

- PROTOTYPE-001: Basic API foundation must be established

**Estimated_Effort:** Medium

--- END OF TASK ---
