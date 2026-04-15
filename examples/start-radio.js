import { ChowdahhClient } from "../src/index.js";

const client = new ChowdahhClient({
  baseUrl: process.env.CHOWDAHH_BASE_URL || "https://chowdahh.com",
  apiKey: process.env.CHOWDAHH_API_KEY
});

// 1. Start a radio session — the server builds a queue of tracks with audio URLs
const session = await client.startRadioSession({
  mode: "briefing",
  duration_minutes: 5
});

const radioId = session.data.radio_session_id;
console.log("State:", session.data.state);
console.log("Tracks:", session.data.queue_length);

// 2. Each track has an audio_url that streams MP3
for (const track of session.data.tracks || []) {
  const fullUrl = client.audioUrl(track.id);
  console.log(`  ${track.headline} — ${fullUrl}`);
  console.log(`    topics: ${track.topics?.join(", ") || "—"}  sources: ${track.source_count}`);
}

// 3. Check session state
const status = await client.getRadioSession(radioId);
console.log("\nPosition:", status.data.position, "of", status.data.queue_length);

// 4. Skip to next track
const skipped = await client.updateRadioSession(radioId, { action: "skip" });
console.log("After skip — position:", skipped.data.position);
if (skipped.data.tracks?.length > 0) {
  console.log("Now playing:", skipped.data.tracks[0].headline);
}
