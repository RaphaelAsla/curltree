# Project Specification for Terminal-Based Linktree Application

## Overview
Develop a terminal user interface (TUI) application using Go, Wish (SSH manager), and Bubbletea, designed as a Linktree alternative for terminal users. The application runs on a server accessed via SSH and authenticates users by their SSH public keys.

The project consists of two main components:
1. **TUI Client Application**: Runs upon SSH login, enabling users to create, view, and manage their profiles and associated links.
2. **Backend Service**: A REST-like HTTP service implemented in Go, serving user profiles in a plain-text, tree-like format via HTTP endpoints, accessed through `curl`.

---

## TUI Client Application Specification

### Environment and Libraries
- Language: Go
- Libraries/Frameworks:
  - **Wish**: For managing SSH server sessions and authentication.
  - **Bubbletea**: For building the interactive TUI interface.

### Authentication
- Authenticate users by verifying their SSH public key on login.
- The SSH public key serves as the unique user identifier.
- Public keys are stored securely in a database (e.g., SQL or NoSQL).
- If the public key is not found in the database, prompt the user to create a profile.

### User Profile Data Model
- `Full Name` (string): User's real name.
- `Username` (string): Unique identifier, used to construct public URLs as `curltree.dev/<Username>`.
- `About` (string): Short description or biography.
- `Links` (list of tuples): Each tuple contains:
  - `Link Name` (string): Descriptive label for the link.
  - `Link URL` (string): The actual URL.

### Profile Creation and Validation
- Profile creation flow triggered if SSH public key is not recognized.
- Input fields for full name, username, about, and links.
- Username uniqueness enforced:
  - On username duplication, notify user and request a different username.
- Allow adding multiple links.
- Link management controls:
  - `Ctrl+N`: Add a new link input field.
  - `Ctrl+D`: Delete the currently focused link.

### Profile Viewing and Editing
- If the SSH public key exists in the database:
  - Display a read-only view of the user profile.
  - User controls:
    - `Ctrl+E`: Enter edit mode, allowing updates to all fields.
    - `Ctrl+C`: Cancel and exit the application.
    - `Ctrl+Delete` (or `Ctrl+D` on profile view): Delete profile and remove SSH public key from database permanently.

### UI/UX Requirements
- Responsive and intuitive keyboard controls using Bubbletea conventions.
- Clear notifications for validation errors (e.g., duplicate usernames).
- Confirmation prompts for destructive actions (profile deletion).
- Persist data changes to the backend database immediately on save.

### Security Considerations
- Ensure public keys are stored and compared securely.
- Sanitize user inputs to prevent injection attacks.
- Handle SSH session lifecycle gracefully.

---

## Backend Service Specification

### Environment and Libraries
- Language: Go
- Frameworks: Use standard Go HTTP libraries or a lightweight web framework.
- Database: Same as TUI (SQL/NoSQL) with secure connection pooling.

### Endpoints

#### Profile Retrieval
- HTTP GET `/username`
- Returns a plain-text, tree-like structured response containing:
  - Full Name
  - Username
  - About
  - List of links formatted as:
    ```
    Link Name
      └─ URL
    ```
- Profiles are publicly accessible via `curl curltree.dev/<Username>` only.

#### Profile Management API (Optional for Admin / TUI)
- Support create, update, delete operations over secured internal API (not public-facing).

### Rate Limiting
- Implement rate limiting on all public endpoints to prevent abuse.
- Suggested policies:
  - Per-IP rate limit (e.g., 60 requests/minute).
  - Burst handling with token bucket or leaky bucket algorithm.
- Return appropriate HTTP status codes on rate limit exceedance (e.g., 429 Too Many Requests).

### Security Considerations
- Validate all input parameters.
- Sanitize and escape outputs to prevent header or injection attacks.
- Log requests and errors for monitoring.

---

## Database Design Notes

- Use a schema that supports:
  - Storing SSH public keys linked to profiles.
  - Unique index on usernames.
  - Efficient query by public key or username.
- Example fields:
  - `id` (UUID or auto-increment)
  - `ssh_public_key` (text, unique)
  - `full_name` (text)
  - `username` (text, unique)
  - `about` (text)
  - `links` (JSON array or normalized table with name/url pairs)

---

## Best Practices and Additional Considerations

- **Code Quality**
  - Use idiomatic Go error handling.
  - Modularize code: separate SSH handling, UI logic, backend API calls, and database access layers.
  - Write unit and integration tests for critical components.
  - Use context for request cancellations and timeouts.
  
- **User Experience**
  - Provide clear instructions on each screen in the TUI.
  - Ensure smooth keyboard navigation and minimal input errors.
  
- **Deployment**
  - Use environment variables or config files for database and server configuration.
  - Graceful shutdown handling for both TUI and backend service.
  
- **Logging and Monitoring**
  - Structured logging for backend and TUI events.
  - Metrics collection for rate limiting and usage statistics.
  
- **Security**
  - Protect backend API endpoints not meant for public access.
  - Use HTTPS with TLS for the backend service.
  - Regularly audit stored public keys and profile data.

---

This specification should provide a comprehensive foundation for implementing a production-ready terminal Linktree application with a Go-based backend and TUI frontend integrated through SSH sessions.
