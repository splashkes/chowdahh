import { ChowdahhClient } from "../src/index.js";

const client = new ChowdahhClient({
  baseUrl: process.env.CHOWDAHH_BASE_URL || "https://chowdahh.com",
  apiKey: process.env.CHOWDAHH_API_KEY
});

const result = await client.startFeedSession({
  intent: "browse",
  budget_minutes: 5,
  include_controls: true
});

console.log("Session:", result.data.session_id);
console.log("Cards:", result.data.count);
console.log("Guidance:", result.guidance?.status_explanation);
console.log("Next actions:", result.guidance?.next_best_actions?.map(a => a.action_id));

for (const card of result.data.items || result.data.cards || []) {
  console.log(`  ${card.headline}`);
  if (card.image_url) console.log(`    image: ${card.image_url}`);
}

if (result.data.controls) {
  console.log("Available controls:", JSON.stringify(result.data.controls, null, 2));
}
