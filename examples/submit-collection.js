import { ChowdahhClient } from "../src/index.js";

const client = new ChowdahhClient({
  baseUrl: process.env.CHOWDAHH_BASE_URL || "https://chowdahh.com",
  apiKey: process.env.CHOWDAHH_API_KEY
});

const result = await client.submitCollection([
  {
    title: "Roman Empire overview",
    source_url: "https://example.com/roman-empire-intro"
  },
  {
    title: "Roman Empire podcast",
    source_url: "https://example.com/roman-empire-podcast"
  }
]);

console.log("Total:", result.data.total);
console.log("Accepted:", result.data.accepted);
console.log("Guidance:", result.guidance?.status_explanation);
for (const r of result.data.results || []) {
  console.log(`  ${r.title}: ${r.status} ${r.submission_id || ""}`);
}
