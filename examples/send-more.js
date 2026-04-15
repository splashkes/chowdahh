import { ChowdahhClient } from "../src/index.js";

const client = new ChowdahhClient({
  baseUrl: process.env.CHOWDAHH_BASE_URL || "https://chowdahh.com",
  apiKey: process.env.CHOWDAHH_API_KEY
});

// First create a session
const session = await client.startFeedSession({
  intent: "browse",
  budget_minutes: 3
});

const sessionId = session.data.session_id;
console.log("Session:", sessionId, "— initial cards:", session.data.count);

// Then send more
const more = await client.sendMore(sessionId, { limit: 3 });
console.log("More cards:", more.data.count, "— position:", more.data.position);
console.log("Guidance:", more.guidance?.status_explanation);
console.log("Suggestions:", more.guidance?.next_best_actions?.map(a => `${a.action_id} (${a.priority})`));
