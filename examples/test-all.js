#!/usr/bin/env node
/**
 * Comprehensive test of all Chowdahh agent API endpoints against a live server.
 *
 * Usage:
 *   CHOWDAHH_BASE_URL=https://chowdahh.com node examples/test-all.js
 *
 * Optional:
 *   CHOWDAHH_API_KEY=ch_person_xxx  (for authenticated endpoints)
 */
import { ChowdahhClient } from "../src/index.js";

const BASE = process.env.CHOWDAHH_BASE_URL || "https://chowdahh.com";
const client = new ChowdahhClient({ baseUrl: BASE, apiKey: process.env.CHOWDAHH_API_KEY });

let passed = 0;
let failed = 0;

async function test(name, fn) {
  try {
    await fn();
    console.log(`  PASS  ${name}`);
    passed++;
  } catch (err) {
    console.log(`  FAIL  ${name}: ${err.message}`);
    failed++;
  }
}

function assert(cond, msg) {
  if (!cond) throw new Error(msg || "assertion failed");
}

console.log(`\nTesting against ${BASE}\n`);

// --- Discovery ---

await test("GET /streams/top", async () => {
  const r = await client.getStream("top", { limit: 3 });
  assert(r.data, "missing data");
  assert(r.guidance, "missing guidance");
  assert(r.meta?.request_id, "missing request_id");
  assert(r.guidance.status_explanation, "missing status_explanation");
  assert(r.guidance.account_state, "missing account_state");
  assert(typeof r.data.count === "number", "count should be number");
});

await test("GET /streams/science", async () => {
  const r = await client.getStream("science", { limit: 2 });
  assert(r.data.stream === "science", "wrong stream");
  assert(Array.isArray(r.data.cards), "cards should be array");
});

await test("GET /streams/world", async () => {
  const r = await client.getStream("world", { limit: 2 });
  assert(r.data.stream === "world", "wrong stream");
});

await test("GET /streams (bad slug returns 404)", async () => {
  try {
    await client.getStream("nonexistent");
    throw new Error("should have thrown");
  } catch (err) {
    assert(err.code === "not_found", `expected not_found, got ${err.code}`);
  }
});

await test("GET /search", async () => {
  const r = await client.search({ q: "NASA", limit: 3 });
  assert(r.data.query === "NASA", "wrong query");
  assert(typeof r.data.count === "number", "count should be number");
});

await test("GET /topics/{id}", async () => {
  const r = await client.getTopic("Artemis II", { limit: 5 });
  assert(r.data.topic === "Artemis II", "wrong topic");
  assert(r.guidance, "missing guidance");
});

// --- Feed Sessions ---

let sessionId;
await test("POST /feed-sessions", async () => {
  const r = await client.startFeedSession({
    intent: "browse",
    budget_minutes: 3,
    include_controls: true
  });
  assert(r.data.session_id, "missing session_id");
  assert(typeof r.data.count === "number", "count should be number");
  assert(r.guidance.next_best_actions?.length > 0, "should have next actions");
  sessionId = r.data.session_id;
});

await test("GET /feed-sessions/{id}", async () => {
  const r = await client.getFeedSession(sessionId);
  assert(r.data.session_id === sessionId, "wrong session");
  assert(r.data.state === "active", "should be active");
});

await test("POST /feed-sessions/{id}/more", async () => {
  const r = await client.sendMore(sessionId, { limit: 3 });
  assert(typeof r.data.count === "number", "count should be number");
  assert(typeof r.data.position === "number", "position should be number");
  assert(r.guidance, "missing guidance");
});

await test("PATCH /feed-sessions/{id}/controls", async () => {
  const r = await client.updateControls(sessionId, {
    interests: ["science"],
    rank_mode: "latest"
  });
  assert(r.data.state === "active", "should be active");
  assert(r.data.controls.interests?.includes("science"), "science not in controls");
  assert(r.guidance.next_best_actions?.some(a => a.action_id === "send_more"), "should suggest send_more");
});

// --- Signals ---

await test("POST /signals", async () => {
  const r = await client.recordSignals([
    { signal_type: "seen", card_id: "test-card-1" },
    { signal_type: "open", card_id: "test-card-1" }
  ]);
  assert(r.data.recorded === 2, `expected 2 recorded, got ${r.data.recorded}`);
});

// --- Feedback ---

await test("POST /feedback (content_request)", async () => {
  const r = await client.submitFeedback({
    feedback_type: "content_request",
    title: "SDK test: more science coverage",
    detail: "Automated test from chowdahh_recipes test suite."
  });
  assert(r.data.status === "received", "should be received");
  assert(r.guidance.status_explanation.includes("content request"), "should mention content request");
});

await test("POST /feedback (validation error)", async () => {
  try {
    await client.submitFeedback({ feedback_type: "content_request" });
    throw new Error("should have thrown");
  } catch (err) {
    assert(err.code === "validation_error", `expected validation_error, got ${err.code}`);
  }
});

// --- Radio ---

await test("POST /radio-sessions", async () => {
  const r = await client.startRadioSession({
    mode: "briefing",
    duration_minutes: 3
  });
  assert(r.data.state === "ready", "should be ready");
  assert(typeof r.data.queue_length === "number", "queue_length should be number");
});

// --- Submissions ---

await test("POST /submissions/items", async () => {
  const r = await client.submitItem({
    title: "SDK test submission",
    source_url: "https://example.com/sdk-test-" + Date.now()
  });
  assert(r.data.status, "should have status");
  assert(r.guidance, "missing guidance");
});

// --- Replay (auth-required, may fail for anonymous) ---

await test("GET /replay (anonymous = 401)", async () => {
  if (process.env.CHOWDAHH_API_KEY) {
    const r = await client.getReplay({ period: "today" });
    assert(r.data, "missing data");
  } else {
    try {
      await client.getReplay({ period: "today" });
      throw new Error("should have thrown");
    } catch (err) {
      assert(err.code === "unauthorized", `expected unauthorized, got ${err.code}`);
      assert(err.guidance?.next_best_actions?.length > 0, "should suggest creating token");
    }
  }
});

// --- Rate limit visibility ---

await test("Rate limit in guidance", async () => {
  const r = await client.getStream("top", { limit: 1 });
  const rl = r.guidance?.account_state?.rate_limit;
  assert(rl, "missing rate_limit in guidance");
  assert(typeof rl.limit === "number", "limit should be number");
  assert(typeof rl.remaining === "number", "remaining should be number");
  assert(rl.reset_at, "missing reset_at");
});

// --- Envelope consistency ---

await test("Error envelope has meta.request_id", async () => {
  try {
    await client.getStream("nonexistent");
  } catch (err) {
    assert(err.meta?.request_id, "error should have meta.request_id");
  }
});

// --- Summary ---

console.log(`\n${passed + failed} tests: ${passed} passed, ${failed} failed\n`);
process.exit(failed > 0 ? 1 : 0);
