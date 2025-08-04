# Web Application Design

**Version:** 1.0
**Date:** August 4, 2025
**Author:** Senior Staff Software Architect, Garnizeh
**Status:** In Progress

---

## 📋 Overview

The EngLog Web Application is a lightweight, dependency-free Single Page Application (SPA) built with vanilla JavaScript and modern CSS frameworks. It provides an intuitive interface for journal management, AI-powered insights visualization, and collaborative group features while maintaining excellent performance and accessibility.

## 🏗️ Frontend Architecture

### Technology Stack

- **Core Framework:** Vanilla JavaScript (ES2022+)
- **CSS Framework:** Bootstrap 5 / Tailwind CSS (via CDN)
- **Build Process:** Minimal build requirements, direct browser compatibility
- **Module System:** ES6 modules with dynamic imports
- **State Management:** Custom lightweight state management
- **HTTP Client:** Native Fetch API with retry logic
- **Authentication:** JWT token management with automatic refresh

### Design Principles

1. **Zero Dependencies:** No external JavaScript frameworks or libraries
2. **Progressive Enhancement:** Works without JavaScript, enhanced with it
3. **Mobile-First Design:** Responsive design optimized for mobile devices
4. **Accessibility:** WCAG 2.1 AA compliance throughout
5. **Performance:** Fast loading with code splitting and lazy loading
6. **Offline Capability:** Service worker for offline functionality
7. **Security:** XSS protection and secure data handling

## 🎨 User Interface Design

### Design System

```css
/* CSS Custom Properties for Design System */
:root {
  /* Color Palette */
  --color-primary: #4f46e5;
  --color-primary-light: #6366f1;
  --color-primary-dark: #3730a3;
  --color-secondary: #10b981;
  --color-danger: #ef4444;
  --color-warning: #f59e0b;
  --color-success: #10b981;

  /* Neutral Colors */
  --color-gray-50: #f9fafb;
  --color-gray-100: #f3f4f6;
  --color-gray-200: #e5e7eb;
  --color-gray-300: #d1d5db;
  --color-gray-600: #4b5563;
  --color-gray-900: #111827;

  /* Typography */
  --font-family-sans: "Inter", -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
  --font-family-mono: "JetBrains Mono", "Courier New", monospace;

  /* Spacing Scale */
  --spacing-xs: 0.25rem;
  --spacing-sm: 0.5rem;
  --spacing-md: 1rem;
  --spacing-lg: 1.5rem;
  --spacing-xl: 2rem;
  --spacing-2xl: 3rem;

  /* Border Radius */
  --radius-sm: 0.25rem;
  --radius-md: 0.5rem;
  --radius-lg: 0.75rem;
  --radius-xl: 1rem;

  /* Shadows */
  --shadow-sm: 0 1px 2px 0 rgba(0, 0, 0, 0.05);
  --shadow-md: 0 4px 6px -1px rgba(0, 0, 0, 0.1);
  --shadow-lg: 0 10px 15px -3px rgba(0, 0, 0, 0.1);
}
```

### Component Library

#### Button Components

```css
.btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: var(--spacing-sm) var(--spacing-md);
  font-weight: 500;
  border-radius: var(--radius-md);
  border: 1px solid transparent;
  cursor: pointer;
  transition: all 0.2s ease-in-out;
  text-decoration: none;
}

.btn-primary {
  background-color: var(--color-primary);
  color: white;
  border-color: var(--color-primary);
}

.btn-primary:hover {
  background-color: var(--color-primary-dark);
  border-color: var(--color-primary-dark);
}

.btn-secondary {
  background-color: transparent;
  color: var(--color-primary);
  border-color: var(--color-primary);
}

.btn-ghost {
  background-color: transparent;
  color: var(--color-gray-600);
  border-color: transparent;
}
```

#### Card Components

```css
.card {
  background: white;
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-md);
  overflow: hidden;
  transition: transform 0.2s ease-in-out;
}

.card:hover {
  transform: translateY(-2px);
  box-shadow: var(--shadow-lg);
}

.card-header {
  padding: var(--spacing-lg);
  border-bottom: 1px solid var(--color-gray-200);
}

.card-body {
  padding: var(--spacing-lg);
}

.card-footer {
  padding: var(--spacing-lg);
  background-color: var(--color-gray-50);
  border-top: 1px solid var(--color-gray-200);
}
```

## 📱 Application Structure

### Single Page Application Layout

```html
<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>EngLog - Personal Journal with AI Insights</title>

    <!-- CSS Framework (CDN) -->
    <link
      href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/css/bootstrap.min.css"
      rel="stylesheet"
    />
    <!-- Custom Styles -->
    <link rel="stylesheet" href="/assets/css/app.css" />

    <!-- Service Worker Registration -->
    <script>
      if ("serviceWorker" in navigator) {
        navigator.serviceWorker.register("/sw.js");
      }
    </script>
  </head>
  <body>
    <!-- Application Shell -->
    <div id="app">
      <!-- Navigation -->
      <nav id="navbar"></nav>

      <!-- Main Content Area -->
      <main id="main-content"></main>

      <!-- Modal Container -->
      <div id="modal-container"></div>

      <!-- Toast Notifications -->
      <div id="toast-container"></div>

      <!-- Loading Overlay -->
      <div id="loading-overlay" class="d-none"></div>
    </div>

    <!-- JavaScript Modules -->
    <script type="module" src="/js/app.js"></script>
  </body>
</html>
```

### File Structure

```
web/
├── index.html
├── assets/
│   ├── css/
│   │   ├── app.css
│   │   ├── components.css
│   │   └── utilities.css
│   ├── icons/
│   │   └── *.svg
│   └── images/
├── js/
│   ├── app.js                 # Main application entry point
│   ├── router.js              # Client-side routing
│   ├── state.js               # State management
│   ├── api.js                 # API client
│   ├── auth.js                # Authentication handling
│   ├── components/
│   │   ├── navbar.js
│   │   ├── journal-editor.js
│   │   ├── insights-panel.js
│   │   ├── group-manager.js
│   │   └── tag-system.js
│   ├── pages/
│   │   ├── dashboard.js
│   │   ├── journal.js
│   │   ├── insights.js
│   │   ├── groups.js
│   │   ├── settings.js
│   │   └── auth.js
│   └── utils/
│       ├── helpers.js
│       ├── validators.js
│       └── formatters.js
├── sw.js                      # Service Worker
└── manifest.json              # PWA Manifest
```

## 🔄 State Management

### Lightweight State Management System

```javascript
// state.js - Custom state management
class AppState {
  constructor() {
    this.state = {
      user: null,
      journals: [],
      groups: [],
      tags: [],
      insights: {},
      ui: {
        currentPage: "dashboard",
        loading: false,
        errors: [],
        notifications: [],
      },
    };
    this.subscribers = new Map();
  }

  subscribe(key, callback) {
    if (!this.subscribers.has(key)) {
      this.subscribers.set(key, []);
    }
    this.subscribers.get(key).push(callback);

    // Return unsubscribe function
    return () => {
      const callbacks = this.subscribers.get(key);
      const index = callbacks.indexOf(callback);
      if (index > -1) {
        callbacks.splice(index, 1);
      }
    };
  }

  setState(updates) {
    const prevState = { ...this.state };
    this.state = { ...this.state, ...updates };

    // Notify subscribers of changes
    Object.keys(updates).forEach((key) => {
      if (this.subscribers.has(key)) {
        this.subscribers.get(key).forEach((callback) => {
          callback(this.state[key], prevState[key]);
        });
      }
    });
  }

  getState(key) {
    return key ? this.state[key] : this.state;
  }
}

// Global state instance
export const appState = new AppState();
```

### State Updates Pattern

```javascript
// Example: Journal state management
import { appState } from "./state.js";
import { apiClient } from "./api.js";

export class JournalManager {
  async loadJournals() {
    appState.setState({ ui: { ...appState.getState("ui"), loading: true } });

    try {
      const journals = await apiClient.get("/journals");
      appState.setState({
        journals,
        ui: { ...appState.getState("ui"), loading: false },
      });
    } catch (error) {
      appState.setState({
        ui: {
          ...appState.getState("ui"),
          loading: false,
          errors: [...appState.getState("ui").errors, error.message],
        },
      });
    }
  }

  async createJournal(journalData) {
    try {
      const newJournal = await apiClient.post("/journals", journalData);
      const currentJournals = appState.getState("journals");
      appState.setState({
        journals: [newJournal, ...currentJournals],
      });
      return newJournal;
    } catch (error) {
      throw error;
    }
  }
}
```

## 🧩 Component Architecture

### Component Base Class

```javascript
// components/base-component.js
export class BaseComponent {
  constructor(element) {
    this.element = element;
    this.subscriptions = [];
  }

  render(data) {
    // Override in subclasses
    throw new Error("render method must be implemented");
  }

  subscribe(stateKey, callback) {
    const unsubscribe = appState.subscribe(stateKey, callback);
    this.subscriptions.push(unsubscribe);
  }

  destroy() {
    // Cleanup subscriptions
    this.subscriptions.forEach((unsubscribe) => unsubscribe());
    this.subscriptions = [];

    // Remove event listeners
    this.element.innerHTML = "";
  }

  emit(eventName, data) {
    this.element.dispatchEvent(
      new CustomEvent(eventName, {
        detail: data,
        bubbles: true,
      })
    );
  }
}
```

### Journal Editor Component

```javascript
// components/journal-editor.js
import { BaseComponent } from "./base-component.js";
import { JournalManager } from "../services/journal-manager.js";

export class JournalEditor extends BaseComponent {
  constructor(element) {
    super(element);
    this.journalManager = new JournalManager();
    this.initializeEditor();
  }

  initializeEditor() {
    this.render();
    this.attachEventListeners();
  }

  render() {
    this.element.innerHTML = `
            <div class="journal-editor">
                <div class="editor-header">
                    <h2>New Journal Entry</h2>
                    <div class="editor-actions">
                        <button class="btn btn-secondary" id="save-draft">Save Draft</button>
                        <button class="btn btn-primary" id="publish">Publish</button>
                    </div>
                </div>

                <div class="editor-body">
                    <div class="form-group mb-3">
                        <label for="entry-title" class="form-label">Title (Optional)</label>
                        <input type="text" class="form-control" id="entry-title"
                               placeholder="What's on your mind?">
                    </div>

                    <div class="form-group mb-3">
                        <label for="entry-content" class="form-label">Content</label>
                        <textarea class="form-control" id="entry-content" rows="10"
                                  placeholder="Write your thoughts here..."></textarea>
                    </div>

                    <div class="editor-metadata">
                        <div class="row">
                            <div class="col-md-6">
                                <label for="mood-slider" class="form-label">Mood</label>
                                <input type="range" class="form-range" id="mood-slider"
                                       min="1" max="10" value="5">
                                <div class="mood-labels">
                                    <span>😞</span>
                                    <span>😐</span>
                                    <span>😊</span>
                                </div>
                            </div>

                            <div class="col-md-6">
                                <label for="tags-input" class="form-label">Tags</label>
                                <input type="text" class="form-control" id="tags-input"
                                       placeholder="Add tags...">
                                <div id="tag-suggestions" class="tag-suggestions"></div>
                            </div>
                        </div>
                    </div>
                </div>

                <div class="editor-footer">
                    <div class="entry-stats">
                        <span id="word-count">0 words</span>
                        <span id="char-count">0 characters</span>
                    </div>
                </div>
            </div>
        `;
  }

  attachEventListeners() {
    const contentTextarea = this.element.querySelector("#entry-content");
    const publishBtn = this.element.querySelector("#publish");
    const saveDraftBtn = this.element.querySelector("#save-draft");
    const tagsInput = this.element.querySelector("#tags-input");

    // Real-time word count
    contentTextarea.addEventListener("input", this.updateWordCount.bind(this));

    // Publish journal entry
    publishBtn.addEventListener("click", this.publishEntry.bind(this));

    // Save as draft
    saveDraftBtn.addEventListener("click", this.saveDraft.bind(this));

    // Tag suggestions
    tagsInput.addEventListener("input", this.handleTagInput.bind(this));
  }

  updateWordCount() {
    const content = this.element.querySelector("#entry-content").value;
    const words = content
      .trim()
      .split(/\s+/)
      .filter((word) => word.length > 0);
    const chars = content.length;

    this.element.querySelector(
      "#word-count"
    ).textContent = `${words.length} words`;
    this.element.querySelector(
      "#char-count"
    ).textContent = `${chars} characters`;
  }

  async publishEntry() {
    const entryData = this.gatherEntryData();

    try {
      await this.journalManager.createJournal(entryData);
      this.emit("journal-published", entryData);
      this.clearEditor();
      this.showSuccessMessage("Journal entry published successfully!");
    } catch (error) {
      this.showErrorMessage(
        "Failed to publish journal entry: " + error.message
      );
    }
  }

  gatherEntryData() {
    const title = this.element.querySelector("#entry-title").value;
    const content = this.element.querySelector("#entry-content").value;
    const mood = parseInt(this.element.querySelector("#mood-slider").value);
    const tags = this.element
      .querySelector("#tags-input")
      .value.split(",")
      .map((tag) => tag.trim())
      .filter((tag) => tag.length > 0);

    return {
      raw_data: {
        title: title || null,
        content,
        mood,
        tags,
        timestamp: new Date().toISOString(),
        word_count: content
          .trim()
          .split(/\s+/)
          .filter((w) => w.length > 0).length,
      },
      entry_type: "text",
    };
  }
}
```

### Insights Panel Component

```javascript
// components/insights-panel.js
import { BaseComponent } from "./base-component.js";

export class InsightsPanel extends BaseComponent {
  constructor(element) {
    super(element);
    this.subscribe("insights", this.renderInsights.bind(this));
  }

  renderInsights(insights) {
    if (!insights || Object.keys(insights).length === 0) {
      this.renderEmptyState();
      return;
    }

    this.element.innerHTML = `
            <div class="insights-panel">
                <div class="insights-header">
                    <h3>AI Insights</h3>
                    <span class="insights-timestamp">
                        Last updated: ${new Date(
                          insights.last_updated
                        ).toLocaleString()}
                    </span>
                </div>

                <div class="insights-grid">
                    ${this.renderSentimentInsight(insights.sentiment)}
                    ${this.renderThemeInsight(insights.themes)}
                    ${this.renderPatternInsight(insights.patterns)}
                    ${this.renderRecommendations(insights.recommendations)}
                </div>
            </div>
        `;
  }

  renderSentimentInsight(sentiment) {
    if (!sentiment) return "";

    const sentimentColor = this.getSentimentColor(sentiment.score);
    const sentimentEmoji = this.getSentimentEmoji(sentiment.score);

    return `
            <div class="insight-card sentiment-card">
                <div class="insight-header">
                    <span class="insight-icon">${sentimentEmoji}</span>
                    <h4>Overall Sentiment</h4>
                </div>
                <div class="insight-content">
                    <div class="sentiment-score" style="color: ${sentimentColor}">
                        ${(sentiment.score * 100).toFixed(0)}%
                    </div>
                    <div class="sentiment-label">${sentiment.label}</div>
                    <div class="sentiment-trend">
                        ${
                          sentiment.trend > 0
                            ? "↗"
                            : sentiment.trend < 0
                            ? "↘"
                            : "→"
                        }
                        ${Math.abs(sentiment.trend * 100).toFixed(
                          1
                        )}% from last week
                    </div>
                </div>
            </div>
        `;
  }

  renderThemeInsight(themes) {
    if (!themes || themes.length === 0) return "";

    return `
            <div class="insight-card themes-card">
                <div class="insight-header">
                    <span class="insight-icon">🏷️</span>
                    <h4>Common Themes</h4>
                </div>
                <div class="insight-content">
                    <div class="theme-list">
                        ${themes
                          .slice(0, 5)
                          .map(
                            (theme) => `
                            <div class="theme-item">
                                <span class="theme-name">${theme.name}</span>
                                <div class="theme-progress">
                                    <div class="theme-bar" style="width: ${
                                      theme.frequency * 100
                                    }%"></div>
                                </div>
                                <span class="theme-count">${theme.count}</span>
                            </div>
                        `
                          )
                          .join("")}
                    </div>
                </div>
            </div>
        `;
  }

  getSentimentColor(score) {
    if (score >= 0.7) return "#10B981"; // Green
    if (score >= 0.4) return "#F59E0B"; // Yellow
    return "#EF4444"; // Red
  }

  getSentimentEmoji(score) {
    if (score >= 0.8) return "😊";
    if (score >= 0.6) return "🙂";
    if (score >= 0.4) return "😐";
    if (score >= 0.2) return "🙁";
    return "😞";
  }
}
```

## 🔐 Authentication Flow

### JWT Token Management

```javascript
// auth.js
class AuthManager {
  constructor() {
    this.token = localStorage.getItem("auth_token");
    this.refreshToken = localStorage.getItem("refresh_token");
    this.user = null;
    this.isRefreshing = false;
  }

  async login(provider) {
    try {
      if (provider === "otp") {
        return this.loginWithOTP();
      } else {
        return this.loginWithOAuth(provider);
      }
    } catch (error) {
      console.error("Login failed:", error);
      throw error;
    }
  }

  async loginWithOAuth(provider) {
    const authUrl = `${API_BASE_URL}/auth/login?provider=${provider}`;

    // Open OAuth popup
    const popup = window.open(authUrl, "oauth", "width=500,height=600");

    return new Promise((resolve, reject) => {
      const checkClosed = setInterval(() => {
        if (popup.closed) {
          clearInterval(checkClosed);
          reject(new Error("Authentication cancelled"));
        }
      }, 1000);

      window.addEventListener("message", (event) => {
        if (event.origin !== window.location.origin) return;

        clearInterval(checkClosed);
        popup.close();

        if (event.data.success) {
          this.setTokens(event.data.token, event.data.refreshToken);
          this.user = event.data.user;
          resolve(event.data.user);
        } else {
          reject(new Error(event.data.error || "Authentication failed"));
        }
      });
    });
  }

  async loginWithOTP() {
    const email = prompt("Enter your email address:");
    if (!email) return;

    try {
      // Request OTP
      await apiClient.post("/auth/otp/request", { email });

      const otpCode = prompt("Enter the 6-digit code sent to your email:");
      if (!otpCode) return;

      // Verify OTP
      const response = await apiClient.post("/auth/otp/verify", {
        email,
        otp_code: otpCode,
      });

      this.setTokens(response.token, response.refresh_token);
      this.user = response.user;
      return response.user;
    } catch (error) {
      throw error;
    }
  }

  setTokens(token, refreshToken) {
    this.token = token;
    this.refreshToken = refreshToken;
    localStorage.setItem("auth_token", token);
    localStorage.setItem("refresh_token", refreshToken);

    appState.setState({ user: this.user });
  }

  async refreshAuthToken() {
    if (this.isRefreshing) return;

    this.isRefreshing = true;

    try {
      const response = await fetch(`${API_BASE_URL}/auth/refresh`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${this.refreshToken}`,
        },
      });

      if (response.ok) {
        const data = await response.json();
        this.setTokens(data.token, data.refresh_token);
      } else {
        this.logout();
      }
    } catch (error) {
      this.logout();
    } finally {
      this.isRefreshing = false;
    }
  }

  logout() {
    this.token = null;
    this.refreshToken = null;
    this.user = null;
    localStorage.removeItem("auth_token");
    localStorage.removeItem("refresh_token");

    appState.setState({ user: null });
    router.navigate("/auth");
  }

  isAuthenticated() {
    return !!this.token;
  }

  getAuthHeader() {
    return this.token ? `Bearer ${this.token}` : null;
  }
}

export const authManager = new AuthManager();
```

## 📱 Progressive Web App Features

### Service Worker

```javascript
// sw.js
const CACHE_NAME = "englog-v1";
const urlsToCache = [
  "/",
  "/assets/css/app.css",
  "/js/app.js",
  "/js/router.js",
  "/js/state.js",
  "/js/api.js",
  "/js/auth.js",
];

self.addEventListener("install", (event) => {
  event.waitUntil(
    caches.open(CACHE_NAME).then((cache) => cache.addAll(urlsToCache))
  );
});

self.addEventListener("fetch", (event) => {
  event.respondWith(
    caches.match(event.request).then((response) => {
      // Return cached version or fetch from network
      return response || fetch(event.request);
    })
  );
});

// Background sync for offline journal entries
self.addEventListener("sync", (event) => {
  if (event.tag === "journal-sync") {
    event.waitUntil(syncJournalEntries());
  }
});

async function syncJournalEntries() {
  const offlineEntries = await getOfflineEntries();

  for (const entry of offlineEntries) {
    try {
      await fetch("/api/v1/journals", {
        method: "POST",
        body: JSON.stringify(entry),
        headers: {
          "Content-Type": "application/json",
          Authorization: getStoredAuthToken(),
        },
      });

      await removeOfflineEntry(entry.id);
    } catch (error) {
      console.log("Failed to sync entry:", entry.id);
    }
  }
}
```

### PWA Manifest

```json
{
  "name": "EngLog - Personal Journal with AI",
  "short_name": "EngLog",
  "description": "Personal journal management with AI-powered insights",
  "start_url": "/",
  "display": "standalone",
  "background_color": "#ffffff",
  "theme_color": "#4F46E5",
  "orientation": "portrait-primary",
  "icons": [
    {
      "src": "/assets/icons/icon-72x72.png",
      "sizes": "72x72",
      "type": "image/png"
    },
    {
      "src": "/assets/icons/icon-144x144.png",
      "sizes": "144x144",
      "type": "image/png"
    },
    {
      "src": "/assets/icons/icon-512x512.png",
      "sizes": "512x512",
      "type": "image/png"
    }
  ],
  "categories": ["productivity", "lifestyle", "utilities"],
  "screenshots": [
    {
      "src": "/assets/screenshots/desktop.png",
      "sizes": "1280x720",
      "type": "image/png",
      "form_factor": "wide"
    },
    {
      "src": "/assets/screenshots/mobile.png",
      "sizes": "375x812",
      "type": "image/png",
      "form_factor": "narrow"
    }
  ]
}
```

## 🎯 Performance Optimization

### Code Splitting and Lazy Loading

```javascript
// Dynamic imports for code splitting
class Router {
  async loadPage(pageName) {
    try {
      const pageModule = await import(`./pages/${pageName}.js`);
      return pageModule.default;
    } catch (error) {
      console.error(`Failed to load page: ${pageName}`, error);
      return null;
    }
  }

  async navigate(path) {
    const pageName = this.getPageFromPath(path);

    // Show loading state
    appState.setState({ ui: { ...appState.getState("ui"), loading: true } });

    // Load page component
    const PageComponent = await this.loadPage(pageName);

    if (PageComponent) {
      this.renderPage(PageComponent);
    }

    // Hide loading state
    appState.setState({ ui: { ...appState.getState("ui"), loading: false } });
  }
}
```

### Caching Strategy

```javascript
// api.js - API client with caching
class APIClient {
  constructor() {
    this.cache = new Map();
    this.cacheTimeout = 5 * 60 * 1000; // 5 minutes
  }

  async get(endpoint, options = {}) {
    const cacheKey = `${endpoint}_${JSON.stringify(options)}`;

    // Check cache first
    if (!options.bypassCache && this.cache.has(cacheKey)) {
      const cached = this.cache.get(cacheKey);
      if (Date.now() - cached.timestamp < this.cacheTimeout) {
        return cached.data;
      }
    }

    // Fetch from API
    const response = await this.request("GET", endpoint, null, options);

    // Cache the response
    this.cache.set(cacheKey, {
      data: response,
      timestamp: Date.now(),
    });

    return response;
  }

  invalidateCache(pattern) {
    for (const key of this.cache.keys()) {
      if (key.includes(pattern)) {
        this.cache.delete(key);
      }
    }
  }
}
```

## ♿ Accessibility Features

### ARIA Labels and Semantic HTML

```html
<!-- Accessible journal editor -->
<div class="journal-editor" role="main" aria-label="Journal Entry Editor">
  <h2 id="editor-title">Create New Journal Entry</h2>

  <form aria-labelledby="editor-title">
    <div class="form-group">
      <label for="entry-content" class="form-label">
        Journal Content
        <span class="required" aria-label="required">*</span>
      </label>
      <textarea
        id="entry-content"
        class="form-control"
        aria-describedby="content-help"
        required
        aria-invalid="false"
      >
      </textarea>
      <div id="content-help" class="form-text">
        Write your thoughts and experiences here
      </div>
    </div>

    <div class="form-group">
      <fieldset>
        <legend>Mood Rating</legend>
        <input
          type="range"
          id="mood-slider"
          min="1"
          max="10"
          value="5"
          aria-describedby="mood-value"
          aria-label="Rate your current mood from 1 to 10"
        />
        <output id="mood-value" for="mood-slider">5</output>
      </fieldset>
    </div>

    <button
      type="submit"
      class="btn btn-primary"
      aria-describedby="publish-help"
    >
      Publish Entry
    </button>
    <div id="publish-help" class="sr-only">
      This will save your journal entry and make it available for AI analysis
    </div>
  </form>
</div>
```

### Keyboard Navigation

```javascript
// Keyboard navigation handling
class KeyboardNavigation {
  constructor() {
    this.focusableElements =
      'button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])';
    this.addEventListeners();
  }

  addEventListeners() {
    document.addEventListener("keydown", this.handleKeyDown.bind(this));
  }

  handleKeyDown(event) {
    // Escape key to close modals
    if (event.key === "Escape") {
      this.closeActiveModal();
    }

    // Arrow keys for navigation
    if (event.key === "ArrowUp" || event.key === "ArrowDown") {
      this.handleArrowNavigation(event);
    }

    // Enter key for activation
    if (
      event.key === "Enter" &&
      event.target.getAttribute("role") === "button"
    ) {
      event.preventDefault();
      event.target.click();
    }
  }

  trapFocus(container) {
    const focusableEls = container.querySelectorAll(this.focusableElements);
    const firstFocusableEl = focusableEls[0];
    const lastFocusableEl = focusableEls[focusableEls.length - 1];

    container.addEventListener("keydown", (e) => {
      if (e.key === "Tab") {
        if (e.shiftKey) {
          if (document.activeElement === firstFocusableEl) {
            lastFocusableEl.focus();
            e.preventDefault();
          }
        } else {
          if (document.activeElement === lastFocusableEl) {
            firstFocusableEl.focus();
            e.preventDefault();
          }
        }
      }
    });
  }
}
```

---

## 🔗 Related Documents

- **[API Service Design](./API_SERVICE.md)** - Backend API integration
- **[Authentication](../design/AUTHENTICATION.md)** - Authentication flows and security
- **[Security](../operations/SECURITY.md)** - Frontend security considerations
- **[Testing](../operations/TESTING.md)** - Frontend testing strategies

---

**Document Status:** 🚧 In Progress
**Next Review:** 2025-09-04
**Last Updated:** 2025-08-04

---

_This document details the web application design and frontend architecture. It serves as the technical specification for implementing the user interface and client-side functionality._
