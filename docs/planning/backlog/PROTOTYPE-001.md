**Task_ID:** PROTOTYPE-001
**Feature_Name:** Basic REST API Foundation
**Task_Title:** Create Simple Go REST API with In-Memory Storage

**Task_Description:**
Develop a basic Go REST API server that provides essential journal endpoints without authentication or persistence. The API should use in-memory storage (maps/slices) to store journal data temporarily, enabling rapid development and testing without database dependencies. This forms the core foundation for validating the basic journal management concepts before adding complexity.

**Acceptance_Criteria:**

- [ ] Go HTTP server starts and responds to requests on port 8080
- [ ] In-memory data structures store journal entries (slice of structs)
- [ ] Basic JSON request/response handling is functional
- [ ] Server gracefully handles startup and shutdown
- [ ] Structured logging outputs request information for debugging
- [ ] Basic error handling returns appropriate HTTP status codes
- [ ] No authentication or authorization required
- [ ] Data persists only during server runtime (lost on restart)

**Technical_Specifications:**

- **Component(s):** Go HTTP Server, In-Memory Storage
- **API Endpoint(s):** Base server setup, no specific endpoints yet
- **Data Model(s):** Basic Journal struct with ID, Content, Timestamp fields
- **Key Logic:** HTTP server setup, in-memory data structures (map[string]Journal), JSON marshaling/unmarshaling
- **Non-Functional Requirements:** Server startup time <5 seconds, basic error handling, structured logging

**Dependencies:**

- None (This is the foundational task)

**Estimated_Effort:** Small

--- END OF TASK ---
