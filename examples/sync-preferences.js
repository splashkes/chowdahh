import { ChowdahhClient } from "../src/index.js";

// Preferences require a person token
const client = new ChowdahhClient({
  baseUrl: process.env.CHOWDAHH_BASE_URL || "https://chowdahh.com",
  apiKey: process.env.CHOWDAHH_API_KEY // must be a ch_person_* token
});

const personId = process.env.CHOWDAHH_PERSON_ID || "person_123";

try {
  const result = await client.setPreferences(personId, {
    topics_followed: ["science", "health", "canada"],
    topics_avoided: ["celebrity-gossip"],
    tone_preferences: ["uplifting", "grounded"],
    delivery_preferences: {
      default_budget_minutes: 8,
      default_delivery_mode: "brief"
    }
  });

  console.log("Status:", result.data.status);
  console.log("Guidance:", result.guidance?.status_explanation);
  console.log("Suggestions:", result.guidance?.next_best_actions?.map(a => a.user_facing_prompt));
} catch (err) {
  if (err.code === "unauthorized" || err.code === "forbidden") {
    console.log("Preferences require a person token matching the person_id.");
    console.log("Guidance:", err.guidance?.status_explanation);
  } else {
    throw err;
  }
}
