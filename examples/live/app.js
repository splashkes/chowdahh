import { ChowdahhClient } from "../../src/client.js";

/* ── Config ── */
const STORY_DURATION_MS = 14000;        // time per story when audio is muted
const POST_AUDIO_PAUSE_MS = 2500;       // pause after audio ends before advancing
const STREAMS = ["top", "world", "science", "business", "culture"];

/* ── Gradient palettes keyed by rough topic category ── */
const PALETTES = {
  default:  ["#0f0c29", "#302b63", "#24243e"],
  world:    ["#0f2027", "#203a43", "#2c5364"],
  science:  ["#0d1b2a", "#1b3a4b", "#274060"],
  business: ["#0a1628", "#1a3a2a", "#2d6a4f"],
  culture:  ["#1a0533", "#3a1255", "#2d1b69"],
  politics: ["#1a0a1e", "#2e1a3e", "#1b2838"],
  tech:     ["#04111d", "#0a2540", "#1a3f5c"],
  war:      ["#1a0000", "#3a1020", "#2d1b2a"],
  climate:  ["#001a0a", "#0a3020", "#1a5040"],
};

/* ── State ── */
const client = new ChowdahhClient({ baseUrl: "" });  // proxy via same-origin server
let cards = [];
let idx = 0;
let activeSlot = "a";
let audioOn = true;
let paused = false;
let advanceTimer = null;
let nextCursor = null;

// Radio state — maps card IDs to audio URLs from the radio session
let radioTracks = new Map();   // card_id → audio_url
let radioSessionId = null;

// HTML5 Audio element for Chowdahh radio
const audio = new Audio();
audio.preload = "auto";

/* ── DOM ── */
const $ = (s) => document.querySelector(s);
const startScreen = $("#start-screen");
const storyA      = $("#story-a");
const storyB      = $("#story-b");
const progressBar = $("#progress-bar");
const counter     = $("#counter");
const muteBtn     = $("#mute-btn");
const pauseBtn    = $("#pause-btn");
const tickerTrack = $("#ticker-track");

/* ── Helpers ── */

function pickPalette(card) {
  const topics = (card.topics || []).join(" ").toLowerCase();
  if (/war|military|army|navy|block|strike|weapon|iran|conflict/.test(topics)) return PALETTES.war;
  if (/politi|elect|vote|democrat|republican|congress|parliament|governor/.test(topics)) return PALETTES.politics;
  if (/tech|ai |artificial|chip|quantum|cyber|software/.test(topics)) return PALETTES.tech;
  if (/scien|space|nasa|physics|biology|research|medical/.test(topics)) return PALETTES.science;
  if (/climat|environment|carbon|energy|solar|wind/.test(topics)) return PALETTES.climate;
  if (/business|market|stock|econ|trade|bank|gdp|startup/.test(topics)) return PALETTES.business;
  if (/cultur|art|music|film|book|museum|festival/.test(topics)) return PALETTES.culture;
  if (/world|global|europe|asia|africa|latin|middle east|un |nato/.test(topics)) return PALETTES.world;
  return PALETTES.default;
}

function gradientCSS(palette, seed) {
  const angle = 120 + (seed % 60);
  return `linear-gradient(${angle}deg, ${palette[0]} 0%, ${palette[1]} 50%, ${palette[2]} 100%)`;
}

function hashCode(s) {
  let h = 0;
  for (let i = 0; i < s.length; i++) h = ((h << 5) - h + s.charCodeAt(i)) | 0;
  return Math.abs(h);
}

function timeAgo(dateStr) {
  if (!dateStr) return "";
  const diff = Date.now() - new Date(dateStr).getTime();
  const mins = Math.floor(diff / 60000);
  if (mins < 2) return "just now";
  if (mins < 60) return `${mins}m ago`;
  const hrs = Math.floor(mins / 60);
  if (hrs < 24) return `${hrs}h ago`;
  return `${Math.floor(hrs / 24)}d ago`;
}

/* ── Fetch cards ── */

async function fetchCards() {
  // Fetch more from top (has images), less from others
  const fetches = STREAMS.map(s =>
    client.getStream(s, { limit: s === "top" ? 20 : 5 })
  );
  const results = await Promise.allSettled(fetches);

  const all = [];
  for (const r of results) {
    if (r.status === "fulfilled" && r.value?.data?.cards) {
      all.push(...r.value.data.cards);
      if (r.value.meta?.next_cursor) nextCursor = r.value.meta.next_cursor;
    }
  }

  // Dedupe by id, sort: cards with images first, then by significance
  const seen = new Set();
  cards = all
    .filter(c => { if (seen.has(c.id)) return false; seen.add(c.id); return true; })
    .sort((a, b) => {
      const aImg = a.image_url ? 1 : 0;
      const bImg = b.image_url ? 1 : 0;
      if (bImg !== aImg) return bImg - aImg;
      return (b.significance_score || 0) - (a.significance_score || 0);
    });
}

async function fetchMore() {
  try {
    const r = await client.getStream("top", { limit: 10, cursor: nextCursor || undefined });
    if (r.data?.cards) {
      const existing = new Set(cards.map(c => c.id));
      const fresh = r.data.cards.filter(c => !existing.has(c.id));
      cards.push(...fresh);
      if (r.meta?.next_cursor) nextCursor = r.meta.next_cursor;
      buildTicker();
    }
  } catch { /* silent */ }
}

/* ── Radio session — get audio URLs for cards ── */

async function startRadio() {
  try {
    const session = await client.startRadioSession({
      mode: "headlines",
      duration_minutes: 15,
    });
    radioSessionId = session.data.radio_session_id;

    // API may return tracks[] with audio_url, or queue[] of card IDs
    const tracks = session.data.tracks || [];
    for (const t of tracks) {
      if (t.id && t.audio_url) {
        radioTracks.set(t.id, t.audio_url);
      }
    }
    // Also map queue IDs → audio URLs via the helper
    const queue = session.data.queue || [];
    for (const id of queue) {
      if (!radioTracks.has(id)) {
        radioTracks.set(id, client.audioUrl(id));
      }
    }
  } catch (err) {
    console.warn("Radio session failed, falling back to silent mode:", err.message);
  }
}

/* ── Render story ── */

function renderStory(layer, card) {
  const bg      = layer.querySelector(".story-bg");
  const meta    = layer.querySelector(".story-meta");
  const hl      = layer.querySelector(".headline");
  const sum     = layer.querySelector(".summary");
  const topics  = layer.querySelector(".topics");

  // Use image_url when available, gradient fallback
  if (card.image_url) {
    bg.style.backgroundImage = `url(${card.image_url})`;
    bg.style.backgroundSize = "cover";
    bg.style.backgroundPosition = "center";
    bg.classList.add("has-image");
  } else {
    const palette = pickPalette(card);
    bg.style.backgroundImage = "";
    bg.style.background = gradientCSS(palette, hashCode(card.id));
    bg.classList.remove("has-image");
  }

  // Reset Ken Burns
  bg.style.animation = "none";
  bg.offsetHeight; // force reflow
  bg.style.animation = "";

  meta.textContent = `${card.source_count || 0} sources  ·  ${timeAgo(card.latest_source_at)}`;
  hl.textContent = card.headline;
  sum.textContent = card.summary || "";
  topics.innerHTML = (card.topics || [])
    .slice(0, 5)
    .map(t => `<span class="topic-chip">${t}</span>`)
    .join("");
}

/* ── Navigate to a specific story ── */

function goTo(index) {
  audio.pause();
  clearTimeout(advanceTimer);
  idx = ((index % cards.length) + cards.length) % cards.length;
  showStory(idx);
}

/* ── Show story (crossfade) ── */

function showStory(index) {
  const card = cards[index];
  if (!card) return;

  const incoming = activeSlot === "a" ? storyB : storyA;
  const outgoing = activeSlot === "a" ? storyA : storyB;

  renderStory(incoming, card);
  incoming.classList.add("active");
  outgoing.classList.remove("active");
  activeSlot = activeSlot === "a" ? "b" : "a";

  counter.textContent = `${index + 1} / ${cards.length}`;
  highlightTickerItem(index);

  // Record "seen" signal (fire-and-forget)
  client.recordSignals([{ signal_type: "seen", card_id: card.id }]).catch(() => {});

  // Play audio or use timer
  if (audioOn) {
    playAudio(card);
  } else {
    scheduleAdvance(STORY_DURATION_MS);
  }

  // Prefetch more if running low
  if (index > cards.length - 5) fetchMore();
}

/* ── Audio playback via Chowdahh Radio ── */

function playAudio(card) {
  audio.pause();
  clearTimeout(advanceTimer);

  // Try radio track audio first, fall back to audioUrl helper with card id
  const trackUrl = radioTracks.get(card.id) || client.audioUrl(card.id);

  audio.src = trackUrl;
  audio.currentTime = 0;

  const onEnded = () => {
    cleanup();
    if (!paused) scheduleAdvance(POST_AUDIO_PAUSE_MS);
  };

  const onError = () => {
    cleanup();
    // Audio unavailable for this card — advance on timer
    scheduleAdvance(STORY_DURATION_MS);
  };

  // Once we know duration, set progress bar
  const onLoadedMetadata = () => {
    if (audio.duration && isFinite(audio.duration)) {
      startProgress(audio.duration * 1000 + POST_AUDIO_PAUSE_MS);
    }
  };

  function cleanup() {
    audio.removeEventListener("ended", onEnded);
    audio.removeEventListener("error", onError);
    audio.removeEventListener("loadedmetadata", onLoadedMetadata);
  }

  audio.addEventListener("ended", onEnded);
  audio.addEventListener("error", onError);
  audio.addEventListener("loadedmetadata", onLoadedMetadata);

  audio.play().catch(() => {
    // Autoplay blocked or network error — fall back to timer
    cleanup();
    scheduleAdvance(STORY_DURATION_MS);
  });

  // Fallback progress bar if metadata doesn't load quickly
  startProgress(STORY_DURATION_MS);
}

/* ── Advance / progress ── */

function scheduleAdvance(ms) {
  clearTimeout(advanceTimer);
  startProgress(ms);
  advanceTimer = setTimeout(advance, ms);
}

function advance() {
  if (paused) return;
  idx = (idx + 1) % cards.length;
  showStory(idx);
}

function startProgress(durationMs) {
  progressBar.style.transition = "none";
  progressBar.style.width = "0%";
  progressBar.offsetHeight;
  progressBar.style.transition = `width ${durationMs}ms linear`;
  progressBar.style.width = "100%";
}

/* ── Ticker ── */

function buildTicker() {
  // Build clickable ticker items with data-index
  const items = cards.map((c, i) =>
    `<span class="ticker-item" data-card-index="${i}">${c.headline}</span><span class="ticker-sep">&bull;</span>`
  ).join("");

  // Duplicate for seamless loop — second copy offsets indices by cards.length
  const items2 = cards.map((c, i) =>
    `<span class="ticker-item" data-card-index="${i}">${c.headline}</span><span class="ticker-sep">&bull;</span>`
  ).join("");

  tickerTrack.innerHTML = items + items2;

  const duration = Math.max(40, cards.length * 5);
  tickerTrack.style.animationDuration = `${duration}s`;
}

function highlightTickerItem(index) {
  tickerTrack.querySelectorAll(".ticker-item.active").forEach(el => el.classList.remove("active"));
  tickerTrack.querySelectorAll(`.ticker-item[data-card-index="${index}"]`).forEach(el => el.classList.add("active"));
}

// Click ticker headline → jump to that story
tickerTrack.addEventListener("click", (e) => {
  const item = e.target.closest(".ticker-item");
  if (!item) return;
  const cardIndex = parseInt(item.dataset.cardIndex, 10);
  if (!isNaN(cardIndex)) goTo(cardIndex);
});

/* ── Keyboard navigation ── */

document.addEventListener("keydown", (e) => {
  if (startScreen && !startScreen.classList.contains("hidden")) return;

  switch (e.key) {
    case "ArrowRight":
    case "ArrowDown":
      e.preventDefault();
      goTo(idx + 1);
      break;
    case "ArrowLeft":
    case "ArrowUp":
      e.preventDefault();
      goTo(idx - 1);
      break;
    case " ":
      e.preventDefault();
      pauseBtn.click();
      break;
    case "m":
      muteBtn.click();
      break;
  }
});

/* ── Controls ── */

muteBtn.addEventListener("click", () => {
  audioOn = !audioOn;
  muteBtn.textContent = audioOn ? "\u{1f50a}" : "\u{1f507}";
  if (!audioOn) {
    audio.pause();
    scheduleAdvance(STORY_DURATION_MS);
  }
});

pauseBtn.addEventListener("click", () => {
  paused = !paused;
  pauseBtn.textContent = paused ? "\u25b6" : "\u23f8";
  if (paused) {
    clearTimeout(advanceTimer);
    audio.pause();
  } else {
    if (audioOn && audio.src && !audio.ended) {
      audio.play().catch(() => {});
    }
    scheduleAdvance(STORY_DURATION_MS / 2);
  }
});

/* ── Boot ── */

startScreen.addEventListener("click", async () => {
  startScreen.classList.add("hidden");

  try {
    // Fetch cards and start radio session in parallel
    await Promise.all([fetchCards(), startRadio()]);
  } catch (err) {
    startScreen.classList.remove("hidden");
    startScreen.querySelector("p").textContent = `Failed to load: ${err.message}. Click to retry.`;
    return;
  }

  if (cards.length === 0) {
    startScreen.classList.remove("hidden");
    startScreen.querySelector("p").textContent = "No stories available. Click to retry.";
    return;
  }

  buildTicker();
  showStory(0);
});
