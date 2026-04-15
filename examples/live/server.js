#!/usr/bin/env node
/**
 * Dev server for Chowdahh Live — a full-screen news broadcast demo.
 *
 * Serves static files from the repo root (so src/client.js is accessible)
 * and proxies /api/* and /audio/* to chowdahh.com.
 *
 * Usage:
 *   node examples/live/server.js
 *   open http://localhost:4000
 */
import { createServer } from "node:http";
import { readFile } from "node:fs/promises";
import { extname, join, resolve } from "node:path";

const PORT = process.env.PORT || 4000;
const API_ORIGIN = "https://chowdahh.com";

// Repo root is two levels up from examples/live/
const REPO_ROOT = resolve(import.meta.dirname, "../..");

const MIME = {
  ".html": "text/html",
  ".css":  "text/css",
  ".js":   "application/javascript",
  ".json": "application/json",
  ".png":  "image/png",
  ".svg":  "image/svg+xml",
  ".mp3":  "audio/mpeg",
};

const server = createServer(async (req, res) => {
  const url = new URL(req.url, `http://localhost:${PORT}`);

  // Redirect root to the live app
  if (url.pathname === "/") {
    res.writeHead(302, { location: "/examples/live/index.html" });
    res.end();
    return;
  }

  // Proxy audio streams to chowdahh.com (streaming)
  if (url.pathname.startsWith("/audio/")) {
    try {
      const target = `${API_ORIGIN}${url.pathname}`;
      const upstream = await fetch(target);
      const ct = upstream.headers.get("content-type") || "audio/mpeg";
      const cl = upstream.headers.get("content-length");
      const head = {
        "content-type": ct,
        "access-control-allow-origin": "*",
        "cache-control": "public, max-age=86400",
      };
      if (cl) head["content-length"] = cl;
      res.writeHead(upstream.status, head);
      const reader = upstream.body.getReader();
      async function pump() {
        while (true) {
          const { done, value } = await reader.read();
          if (done) { res.end(); return; }
          res.write(value);
        }
      }
      pump().catch(() => res.end());
    } catch (err) {
      res.writeHead(502, { "content-type": "text/plain" });
      res.end(err.message);
    }
    return;
  }

  // Proxy API calls to chowdahh.com
  if (url.pathname.startsWith("/api/")) {
    try {
      const target = `${API_ORIGIN}${url.pathname}${url.search}`;
      const headers = { "content-type": "application/json" };
      if (req.headers.authorization) headers.authorization = req.headers.authorization;

      let body;
      if (req.method === "POST" || req.method === "PATCH" || req.method === "PUT") {
        body = await new Promise((resolve) => {
          const chunks = [];
          req.on("data", (c) => chunks.push(c));
          req.on("end", () => resolve(Buffer.concat(chunks).toString()));
        });
      }

      const upstream = await fetch(target, {
        method: req.method,
        headers,
        body: body || undefined,
      });

      const data = await upstream.text();
      res.writeHead(upstream.status, {
        "content-type": "application/json",
        "access-control-allow-origin": "*",
      });
      res.end(data);
    } catch (err) {
      res.writeHead(502, { "content-type": "application/json" });
      res.end(JSON.stringify({ error: { message: err.message } }));
    }
    return;
  }

  // Static files from repo root
  const fullPath = join(REPO_ROOT, url.pathname);
  try {
    const data = await readFile(fullPath);
    const mime = MIME[extname(fullPath)] || "application/octet-stream";
    res.writeHead(200, { "content-type": mime });
    res.end(data);
  } catch {
    res.writeHead(404, { "content-type": "text/plain" });
    res.end("Not found");
  }
});

server.listen(PORT, () => {
  console.log(`\n  Chowdahh Live → http://localhost:${PORT}\n`);
});
