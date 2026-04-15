import { ChowdahhClient } from "../src/index.js";

// Replay requires a person token
const client = new ChowdahhClient({
  baseUrl: process.env.CHOWDAHH_BASE_URL || "https://chowdahh.com",
  apiKey: process.env.CHOWDAHH_API_KEY // must be a ch_person_* token
});

try {
  const result = await client.getReplay({
    signal_type: "share",
    period: "this_month"
  });
  console.log("Events:", result.data.count);
  console.log("Guidance:", result.guidance?.status_explanation);
  for (const event of result.data.events || []) {
    console.log(`  ${event.signal_type}: ${event.headline || event.card_id} (${event.occurred_at})`);
  }
} catch (err) {
  if (err.code === "unauthorized") {
    console.log("Replay requires authentication.");
    console.log("Guidance:", err.guidance?.status_explanation);
    console.log("Suggestion:", err.guidance?.next_best_actions?.[0]?.user_facing_prompt);
  } else {
    throw err;
  }
}
