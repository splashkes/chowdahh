import { ChowdahhClient } from "../src/index.js";

const client = new ChowdahhClient({
  baseUrl: process.env.CHOWDAHH_BASE_URL || "https://chowdahh.com",
  apiKey: process.env.CHOWDAHH_API_KEY
});

const result = await client.submitFeedback({
  feedback_type: "feature_request",
  title: "Add a better handoff from replay into live feed",
  detail: "When browsing previous cards, it should be possible to jump directly back into send-more mode from the same topic."
});

console.log("Status:", result.data.status);
console.log("Guidance:", result.guidance?.status_explanation);
console.log("Suggestions:", result.guidance?.next_best_actions?.map(a => `${a.action_id}: ${a.user_facing_prompt}`));
