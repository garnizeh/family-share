# FamilyShare Project - Copilot Instructions

You are acting as an expert Go and UX Developer. This project is a lightweight, self-hosted photo sharing app designed for low-resource VPS environments. The main goal is to allow users to upload and share photo albums with family members via expiring links, while optimizing for minimal storage usage and simplicity.

## Golden Rules
0. Before making recommendations or code changes, meticulously verify the current project state (files, branches, build/tests, and recent edits) to avoid incorrect assumptions.
1. Always write in English — all generated content (code, comments, filenames, documentation) and chat responses must be in English, even if you receive instructions in another language.
2. When in doubt, ask simple yes/no questions before making any change; wait for the user's answer and incorporate it into your action.
3. MVP-first: prioritize only what is required for a working MVP; defer non-essential features until explicitly requested.
4. Prioritize storage efficiency above all else. Every byte saved matters.
5. Keep the tech stack minimal and easy to maintain.
6. Focus on server-side rendering with minimal JavaScript.
7. Ensure a smooth and intuitive user experience for non-technical family members.

## Core Tech Stack
- **Backend:** Go (Golang) using `net/http` or `chi` router.
- **Database:** SQLite (embedded). Use `sqlc`. No heavy ORMs like GORM. No CGO.
- **Frontend:** Go `html/template` (Server-Side Rendering).
- **Interactivity:** HTMX for AJAX/Dynamic updates and Alpine.js for client-side state (modals, lightboxes).
- **Styling:** TailwindCSS (Utility-first).

## Critical Constraints & Patterns

### 1. Storage Optimization (Top Priority)
- **Zero-Waste Uploads:** Never save original high-res photos. 
- All uploads must be processed through an image pipeline: Resize to max 1920px (width/height) and convert to **WebP** (80% quality).
- Use `github.com/disintegration/imaging` for resizing and `github.com/chai2010/webp` for encoding.
- Implement a background Goroutine "Janitor" to clean up expired links and files from disk.

### 2. Frontend Interactivity (The GOTH Stack)
- **HTMX:** Use for form submissions, infinite scrolling, and partial page reloads.
- **Alpine.js:** Use only for UI-only states (e.g., toggling a mobile menu, opening a lightbox gallery, handling client-side image previews).
- **No SPAs:** Do not suggest React, Vue, or heavy JavaScript bundles. All routing is handled by Go.

### 3. Sharing Logic
- **No Visitor Logins:** Access is granted solely via URL tokens (`/s/{token}`).
- **View Counting:** Implement logic to increment views only for unique sessions (using cookies or basic fingerprinting) to prevent "Refresh" abuse.
- **Expiry:** Always check `max_views` and `expires_at` before serving any shared resource.

### 4. Code Style (Go)
- Keep it simple and idiomatic. 
- Avoid unnecessary abstractions. 
- Use `internal/` folder pattern for core logic.
- Prefer `context.Context` for cancellation and timeouts.
- Use `fs.FS` to embed templates and static assets into the Go binary.

## UI/UX Guidelines
- **Mobile First:** Family members will likely access via phones.
- **HTMX Progress Bars:** Always provide visual feedback during photo uploads.
- **Empty States:** Gracefully handle "No photos found" or "Link expired" screens.

## Security
- Tokens must be generated using `crypto/rand`.
- Admin area must be protected by a simple but secure session-based authentication.
- Mitigate brute-force attempts on tokens using a simple rate-limiting middleware.