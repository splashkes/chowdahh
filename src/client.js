export class ChowdahhClient {
  // Auth: pass `apiKey` to send it as `Authorization: Bearer <key>` on every
  // request. Chowdahh also accepts the same token as `?key=…` on GET URLs
  // (header wins on conflict — see ADR-0140). Use `pasteUrl()` to build a
  // shareable URL that carries the key in the query string.
  constructor({ baseUrl, apiKey } = {}) {
    this.baseUrl = (baseUrl !== undefined ? baseUrl : "https://chowdahh.com").replace(/\/+$/, "");
    this.apiKey = apiKey || null;
  }

  // --- Discovery ---

  // listStreams returns the public stream catalog (slug/label/description).
  async listStreams() {
    return this.#request("/api/v1/streams");
  }

  // Deprecated alias: `/api/v1/categories` is not a live endpoint. Use
  // listStreams() instead. Kept for one release as a no-op alias that calls
  // the real endpoint.
  async getCategories() {
    return this.listStreams();
  }

  async getStream(slug = "top", params = {}) {
    const query = new URLSearchParams({ limit: "20", ...params });
    return this.#request(`/api/v1/streams/${encodeURIComponent(slug)}?${query}`);
  }

  async getTopic(topicId, params = {}) {
    const query = new URLSearchParams({ limit: "20", ...params });
    return this.#request(`/api/v1/topics/${encodeURIComponent(topicId)}?${query}`);
  }

  async search(params) {
    const query = new URLSearchParams(params);
    return this.#request(`/api/v1/search?${query}`);
  }

  async getCurator(curatorId) {
    return this.#request(`/api/v1/curators/${encodeURIComponent(curatorId)}`);
  }

  // --- Feed Sessions ---

  async startFeedSession(payload) {
    return this.#request("/api/v1/feed-sessions", {
      method: "POST",
      body: JSON.stringify(payload)
    });
  }

  async getFeedSession(sessionId) {
    return this.#request(`/api/v1/feed-sessions/${encodeURIComponent(sessionId)}`);
  }

  async sendMore(sessionId, params = {}) {
    const query = new URLSearchParams(params).toString();
    return this.#request(`/api/v1/feed-sessions/${encodeURIComponent(sessionId)}/more${query ? `?${query}` : ""}`, {
      method: "POST"
    });
  }

  async updateControls(sessionId, payload) {
    return this.#request(`/api/v1/feed-sessions/${encodeURIComponent(sessionId)}/controls`, {
      method: "PATCH",
      body: JSON.stringify(payload)
    });
  }

  // --- Signals & Replay ---

  async recordSignals(signals) {
    return this.#request("/api/v1/signals", {
      method: "POST",
      body: JSON.stringify(signals)
    });
  }

  async getReplay(params) {
    const query = new URLSearchParams(params).toString();
    return this.#request(`/api/v1/replay${query ? `?${query}` : ""}`);
  }

  async getActivityStats(params) {
    const query = new URLSearchParams(params).toString();
    return this.#request(`/api/v1/stats/activity${query ? `?${query}` : ""}`);
  }

  // --- Feedback ---

  async submitFeedback(payload) {
    return this.#request("/api/v1/feedback", {
      method: "POST",
      body: JSON.stringify(payload)
    });
  }

  // --- Preferences ---

  async getPreferences(personId) {
    return this.#request(`/api/v1/preferences/${encodeURIComponent(personId)}`);
  }

  async setPreferences(personId, payload) {
    return this.#request(`/api/v1/preferences/${encodeURIComponent(personId)}`, {
      method: "PUT",
      body: JSON.stringify(payload)
    });
  }

  // --- Submissions ---

  async submitItem(payload) {
    return this.#request("/api/v1/submissions/items", {
      method: "POST",
      body: JSON.stringify(payload)
    });
  }

  async submitCollection(items) {
    return this.#request("/api/v1/submissions/collections", {
      method: "POST",
      body: JSON.stringify(items)
    });
  }

  async getSubmission(submissionId) {
    return this.#request(`/api/v1/submissions/${encodeURIComponent(submissionId)}`);
  }

  // --- Radio ---

  async startRadioSession(payload) {
    return this.#request("/api/v1/radio-sessions", {
      method: "POST",
      body: JSON.stringify(payload)
    });
  }

  async getRadioSession(radioSessionId) {
    return this.#request(`/api/v1/radio-sessions/${encodeURIComponent(radioSessionId)}`);
  }

  async updateRadioSession(radioSessionId, payload) {
    return this.#request(`/api/v1/radio-sessions/${encodeURIComponent(radioSessionId)}`, {
      method: "PATCH",
      body: JSON.stringify(payload)
    });
  }

  audioUrl(trackId) {
    return `${this.baseUrl}/audio/${encodeURIComponent(trackId)}`;
  }

  // pasteUrl builds a fully-qualified URL with the API key encoded as `?key=…`
  // — the form an end-user can paste into an LLM (Hermes, OpenClaw, Claude,
  // ChatGPT, Cursor MCP) without writing header-injection glue. GET only;
  // do not use for writes.
  pasteUrl(path, params = {}) {
    const query = new URLSearchParams(params);
    if (this.apiKey) query.set("key", this.apiKey);
    const sep = path.startsWith("/") ? "" : "/";
    const qs = query.toString();
    return `${this.baseUrl}${sep}${path}${qs ? `?${qs}` : ""}`;
  }

  // --- Internal ---

  async #request(path, init = {}) {
    // All paths are relative to baseUrl — never send tokens to arbitrary URLs
    const url = `${this.baseUrl}${path}`;
    const headers = { "content-type": "application/json", ...(init.headers || {}) };
    if (this.apiKey) {
      headers.authorization = `Bearer ${this.apiKey}`;
    }
    const response = await fetch(url, { ...init, headers });

    let body;
    try {
      body = await response.json();
    } catch {
      throw new Error(`HTTP ${response.status}: non-JSON response`);
    }

    if (!response.ok) {
      const err = new Error(body.error?.message || `HTTP ${response.status}`);
      err.status = response.status;
      err.code = body.error?.code;
      err.details = body.error?.details;
      err.guidance = body.guidance;
      err.meta = body.meta;
      throw err;
    }

    return body;
  }
}
