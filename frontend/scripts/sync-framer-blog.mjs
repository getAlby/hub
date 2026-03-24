import { existsSync, readFileSync } from "node:fs";
import { writeFile } from "node:fs/promises";
import path from "node:path";
import process from "node:process";
import { fileURLToPath } from "node:url";
import { connect } from "framer-api";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const outputPath = path.resolve(
  __dirname,
  "../src/data/framerBlogPosts.generated.ts"
);

/** Load frontend/.env.local into process.env (for local testing; file is gitignored). */
function loadEnvLocal() {
  const envPath = path.resolve(__dirname, "../.env.local");
  if (!existsSync(envPath)) {
    return;
  }
  const content = readFileSync(envPath, "utf8");
  for (const line of content.split("\n")) {
    const trimmed = line.trim();
    if (!trimmed || trimmed.startsWith("#")) {
      continue;
    }
    const eq = trimmed.indexOf("=");
    if (eq === -1) {
      continue;
    }
    const key = trimmed.slice(0, eq).trim();
    let value = trimmed.slice(eq + 1).trim();
    if (
      (value.startsWith('"') && value.endsWith('"')) ||
      (value.startsWith("'") && value.endsWith("'"))
    ) {
      value = value.slice(1, -1);
    }
    if (key && process.env[key] === undefined) {
      process.env[key] = value;
    }
  }
}

loadEnvLocal();

const FRAMER_PROJECT_URL =
  process.env.FRAMER_PROJECT_URL || "https://framer.com/projects/Alby--YmsZ3ggcU5mFk4BS0krt-4lvaW";
const FRAMER_API_KEY = process.env.FRAMER_API_KEY;
const FRAMER_BLOG_COLLECTION = process.env.FRAMER_BLOG_COLLECTION;
const POSTS_LIMIT = Number(process.env.FRAMER_BLOG_LIMIT || 8);

if (!FRAMER_API_KEY) {
  console.error("Missing FRAMER_API_KEY environment variable.");
  process.exit(1);
}

function normalizeText(value) {
  if (typeof value === "string") {
    return value.trim();
  }
  if (typeof value === "number" || typeof value === "boolean") {
    return String(value);
  }
  if (Array.isArray(value)) {
    return value.map(normalizeText).filter(Boolean).join(" ").trim();
  }
  if (value && typeof value === "object") {
    const candidates = [
      value.title,
      value.name,
      value.label,
      value.text,
      value.value,
      value.description,
      value.markdown,
      value.html,
      value.plainText,
      value.content,
    ];
    for (const candidate of candidates) {
      const normalized = normalizeText(candidate);
      if (normalized) {
        return normalized;
      }
    }
  }
  return "";
}

function extractImageUrl(value) {
  if (!value) {
    return undefined;
  }
  if (typeof value === "string") {
    const t = value.trim();
    if (t.startsWith("https://") || t.startsWith("http://")) {
      return t;
    }
    if (t.startsWith("//")) {
      return `https:${t}`;
    }
    return undefined;
  }
  if (Array.isArray(value)) {
    for (const entry of value) {
      const image = extractImageUrl(entry);
      if (image) {
        return image;
      }
    }
    return undefined;
  }
  if (typeof value === "object") {
    const candidates = [
      value.url,
      value.src,
      value.image,
      value.file?.url,
      value.asset?.url,
      value.original?.url,
      value.optimized?.url,
      value.thumbnail?.url,
      value.image?.url,
      value.media?.url,
      value.default?.url,
      value.large?.url,
      value.medium?.url,
      value.small?.url,
    ];

    for (const candidate of candidates) {
      const image = extractImageUrl(candidate);
      if (image) {
        return image;
      }
    }
  }
  return undefined;
}

const MAX_IMAGE_DEPTH = 14;

function extractImageUrlDeep(value, depth = 0) {
  if (depth > MAX_IMAGE_DEPTH || value === null || value === undefined) {
    return undefined;
  }
  const direct = extractImageUrl(value);
  if (direct) {
    return direct;
  }
  if (Array.isArray(value)) {
    for (const entry of value) {
      const found = extractImageUrlDeep(entry, depth + 1);
      if (found) {
        return found;
      }
    }
    return undefined;
  }
  if (typeof value === "object") {
    for (const key of Object.keys(value)) {
      const found = extractImageUrlDeep(value[key], depth + 1);
      if (found) {
        return found;
      }
    }
  }
  return undefined;
}

function decodeHtmlAttrEntities(s) {
  return s
    .replace(/&amp;/g, "&")
    .replace(/&quot;/g, '"')
    .replace(/&#39;/g, "'")
    .replace(/&lt;/g, "<")
    .replace(/&gt;/g, ">");
}

/**
 * Live blog pages expose og:image (Framer-hosted). Used when CMS image field is empty.
 */
async function fetchOgImageUrl(pageUrl) {
  try {
    const res = await fetch(pageUrl, {
      headers: {
        Accept: "text/html,application/xhtml+xml",
        "User-Agent": "AlbyHubBlogSync/1.0 (+https://getalby.com)",
      },
      redirect: "follow",
    });
    if (!res.ok) {
      return undefined;
    }
    const html = await res.text();
    let m =
      html.match(/property=["']og:image["'][^>]*\scontent=["']([^"']+)["']/i) ||
      html.match(/content=["']([^"']+)["'][^>]*property=["']og:image["']/i);
    if (!m) {
      m = html.match(
        /name=["']twitter:image(?::src)?["'][^>]*\scontent=["']([^"']+)["']/i
      );
    }
    if (m?.[1]) {
      return decodeHtmlAttrEntities(m[1].trim());
    }
  } catch (err) {
    console.warn("og:image fetch failed:", pageUrl, err?.message || err);
  }
  return undefined;
}

function resolveImageUrlFromFramerItem(item, fields) {
  const data = item?.fieldData || {};
  const imageNameMatchers = [
    /cover/i,
    /hero/i,
    /featured/i,
    /card\s*image/i,
    /thumbnail/i,
    /^image$/i,
    /photo/i,
    /banner/i,
    /picture/i,
    /og\s*image/i,
    /social\s*image/i,
  ];

  for (const matcher of imageNameMatchers) {
    const field = fields.find((f) => matcher.test(f?.name || ""));
    if (!field?.id) {
      continue;
    }
    const url = extractImageUrlDeep(data[field.id]);
    if (url) {
      return url;
    }
  }

  for (const field of fields) {
    const n = field?.name || "";
    if (
      !/image|cover|hero|thumb|photo|banner|picture|media|graphic|visual/i.test(
        n
      )
    ) {
      continue;
    }
    const url = extractImageUrlDeep(data[field.id]);
    if (url) {
      return url;
    }
  }

  return extractImageUrlDeep(data);
}

function findFieldId(fields, matchers) {
  for (const matcher of matchers) {
    const field = fields.find((entry) => matcher.test(entry?.name || ""));
    if (field?.id) {
      return field.id;
    }
  }
  return undefined;
}

/** True when value is clearly the URL slug, not a human-written title. */
function isSlugLike(value, slug) {
  if (!value || !slug) {
    return true;
  }
  const v = value.trim().toLowerCase();
  const s = slug.trim().toLowerCase();
  if (v === s) {
    return true;
  }
  if (v === s.replace(/-/g, " ")) {
    return true;
  }
  if (v.replace(/\s+/g, "-") === s) {
    return true;
  }
  return false;
}

/**
 * Framer field names vary; `name` is too broad and can steal the match from real Title fields.
 * Try explicit title-like fields first, then any field whose value is not slug-like.
 */
function isNonTitleFieldName(name) {
  return /category|tag|topic|type|slug|author|date|image|photo|cover|hero|excerpt|summary|description|teaser|subtitle|link|url|seo|meta|status|draft/i.test(
    name || ""
  );
}

function resolveTitle(item, fields, slug) {
  const data = item?.fieldData || {};
  const titleMatchers = [
    /^title$/i,
    /^post title$/i,
    /^article title$/i,
    /^heading$/i,
    /^headline$/i,
    /seo title/i,
    /display title/i,
    /^h1$/i,
  ];

  for (const matcher of titleMatchers) {
    const field = fields.find(
      (f) => matcher.test(f?.name || "") && !isNonTitleFieldName(f?.name || "")
    );
    if (!field?.id) {
      continue;
    }
    const raw = normalizeText(data[field.id]);
    if (raw && !isSlugLike(raw, slug)) {
      return raw;
    }
  }

  for (const field of fields) {
    const n = field?.name || "";
    if (isNonTitleFieldName(n)) {
      continue;
    }
    if (!/title|heading|headline|subject/i.test(n)) {
      continue;
    }
    const raw = normalizeText(data[field.id]);
    if (raw && !isSlugLike(raw, slug)) {
      return raw;
    }
  }

  return slug;
}

function scoreCollection(collection, fields) {
  const name = (collection?.name || "").toLowerCase();
  let score = 0;

  if (name.includes("blog") || name.includes("post") || name.includes("article")) {
    score += 4;
  }

  const fieldNames = fields.map((field) => (field?.name || "").toLowerCase());
  if (fieldNames.some((nameValue) => nameValue.includes("title"))) score += 2;
  if (
    fieldNames.some(
      (nameValue) =>
        nameValue.includes("description") ||
        nameValue.includes("excerpt") ||
        nameValue.includes("summary")
    )
  )
    score += 2;
  if (fieldNames.some((nameValue) => nameValue.includes("image"))) score += 1;
  if (fieldNames.some((nameValue) => nameValue.includes("date"))) score += 1;

  return score;
}

function asIsoDate(value) {
  if (!value) {
    return undefined;
  }
  const normalized = normalizeText(value);
  if (!normalized) {
    return undefined;
  }
  const parsed = new Date(normalized);
  if (Number.isNaN(parsed.getTime())) {
    return undefined;
  }
  return parsed.toISOString();
}

function buildOutputContent(posts) {
  const generatedAt = new Date().toISOString();
  const entries = JSON.stringify(posts, null, 2);

  return `// This file is auto-generated by \`yarn sync:framer-blog\`.
// Do not edit manually.

export type FramerBlogPost = {
  id: string;
  slug: string;
  title: string;
  description: string;
  url: string;
  category?: string;
  publishedAt?: string;
  imageUrl?: string;
};

export const generatedAt = "${generatedAt}";

export const framerBlogPosts: FramerBlogPost[] = ${entries};
`;
}

async function run() {
  const framer = await connect(FRAMER_PROJECT_URL, FRAMER_API_KEY);
  try {
    const collections = await framer.getCollections();
    if (!collections?.length) {
      throw new Error("No collections found in the Framer project.");
    }

    const candidates = [];
    for (const collection of collections) {
      const fields = await collection.getFields();
      const score = scoreCollection(collection, fields || []);
      candidates.push({ collection, fields: fields || [], score });
    }

    let selected = null;
    if (FRAMER_BLOG_COLLECTION) {
      selected =
        candidates.find(
          ({ collection }) =>
            collection?.id === FRAMER_BLOG_COLLECTION ||
            collection?.name === FRAMER_BLOG_COLLECTION
        ) || null;
    }

    if (!selected) {
      selected = candidates.sort((a, b) => b.score - a.score)[0] || null;
    }

    if (!selected) {
      throw new Error("Unable to determine the blog collection.");
    }

    const { collection, fields } = selected;
    const items = await collection.getItems();

    const descriptionFieldId = findFieldId(fields, [
      /excerpt/i,
      /summary/i,
      /description/i,
      /subtitle/i,
      /teaser/i,
    ]);
    const categoryFieldId = findFieldId(fields, [/category/i, /tag/i, /topic/i]);
    const dateFieldId = findFieldId(fields, [/publish/i, /date/i, /created/i]);

    const posts = (items || [])
      .filter((item) => !item?.draft)
      .map((item) => {
        const title = resolveTitle(item, fields, item?.slug || "") || item.slug;
        const description =
          normalizeText(item?.fieldData?.[descriptionFieldId]) || "";
        const category = normalizeText(item?.fieldData?.[categoryFieldId]) || undefined;
        const publishedAt = asIsoDate(item?.fieldData?.[dateFieldId]);
        const imageUrl = resolveImageUrlFromFramerItem(item, fields);
        const slug = item?.slug || "";

        return {
          id: item?.id || slug,
          slug,
          title,
          description,
          category,
          publishedAt,
          imageUrl,
          url: `https://getalby.com/blog/${slug}`,
        };
      })
      .filter((item) => item.slug && item.title);

    const ogDelayMs = Number(process.env.FRAMER_OG_FETCH_DELAY_MS || 120);
    for (const post of posts) {
      if (!post.imageUrl) {
        post.imageUrl = await fetchOgImageUrl(post.url);
      }
      if (ogDelayMs > 0) {
        await new Promise((r) => setTimeout(r, ogDelayMs));
      }
    }

    posts.sort((a, b) => {
      const aTs = a.publishedAt ? new Date(a.publishedAt).getTime() : 0;
      const bTs = b.publishedAt ? new Date(b.publishedAt).getTime() : 0;
      return bTs - aTs;
    });

    const limitedPosts = posts.slice(0, POSTS_LIMIT);
    const content = buildOutputContent(limitedPosts);
    await writeFile(outputPath, content, "utf8");

    console.log(
      `Synced ${limitedPosts.length} posts from "${collection?.name}" (${collection?.id}).`
    );
    console.log(`Generated: ${outputPath}`);
  } finally {
    await framer.disconnect();
  }
}

run().catch((error) => {
  console.error("Failed to sync Framer blog posts.");
  console.error(error);
  process.exit(1);
});
