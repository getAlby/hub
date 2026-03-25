import { SquareArrowOutUpRightIcon } from "lucide-react";
import React from "react";
import {
  Card,
  CardContent,
  CardFooter,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";
import { ExternalLinkButton } from "src/components/ui/custom/external-link-button";
import { cn } from "src/lib/utils";

type BlogPost = {
  id: string;
  title: string;
  description: string;
  url: string;
  imageUrl?: string;
  publishedAt?: string;
};

const ALBY_BLOG_ENDPOINT =
  import.meta.env.VITE_ALBY_BLOG_ENDPOINT ||
  "https://getalby.com/api/hub/blog/latest";

const fallbackThemes = [
  "from-emerald-200 via-cyan-200 to-yellow-200",
  "from-orange-200 via-amber-100 to-pink-100",
  "from-slate-200 via-zinc-100 to-lime-100",
  "from-sky-200 via-indigo-100 to-violet-100",
];

function toStringValue(value: unknown): string | undefined {
  return typeof value === "string" && value.trim() ? value.trim() : undefined;
}

function normalizePost(input: unknown): BlogPost | null {
  if (!input || typeof input !== "object") {
    return null;
  }
  const item = input as Record<string, unknown>;
  const id = toStringValue(item.id) || toStringValue(item.slug) || "latest";
  const title = toStringValue(item.title);
  const description =
    toStringValue(item.lead) ||
    toStringValue(item.description) ||
    toStringValue(item.excerpt);
  const url = toStringValue(item.url) || toStringValue(item.link);
  const imageUrl =
    toStringValue(item.imageUrl) ||
    toStringValue(item.image_url) ||
    toStringValue(item.coverImage) ||
    toStringValue(item.cover_image);
  const publishedAt =
    toStringValue(item.publishedAt) || toStringValue(item.published_at);

  if (!title || !url || !description) {
    return null;
  }

  return {
    id,
    title,
    description,
    url,
    imageUrl,
    publishedAt,
  };
}

function pickLatestPost(posts: BlogPost[]): BlogPost {
  const dated = posts.filter((p) => p.publishedAt);
  if (dated.length > 0) {
    return [...dated].sort(
      (a, b) =>
        new Date(b.publishedAt || "").getTime() -
        new Date(a.publishedAt || "").getTime()
    )[0];
  }
  return posts[0];
}

async function fetchBlogPosts(): Promise<BlogPost[]> {
  const response = await fetch(ALBY_BLOG_ENDPOINT);
  if (!response.ok) {
    throw new Error(`Failed to fetch blog posts: ${response.status}`);
  }

  const payload = (await response.json()) as unknown;
  const candidates = Array.isArray(payload)
    ? payload
    : Array.isArray((payload as { posts?: unknown[] })?.posts)
      ? (payload as { posts: unknown[] }).posts
      : [payload];

  return candidates
    .map(normalizePost)
    .filter((post): post is BlogPost => !!post);
}

export function AlbyBlogWidget() {
  const [post, setPost] = React.useState<BlogPost | null>(null);
  const [themeClassName, setThemeClassName] = React.useState(fallbackThemes[0]);

  React.useEffect(() => {
    const loadPost = async () => {
      try {
        const posts = await fetchBlogPosts();
        if (!posts.length) {
          setPost(null);
          return;
        }
        const latest = pickLatestPost(posts);
        setPost(latest);
        const themeIndex = Math.abs(
          [...latest.id].reduce((sum, ch) => sum + ch.charCodeAt(0), 0)
        );
        setThemeClassName(fallbackThemes[themeIndex % fallbackThemes.length]);
      } catch {
        setPost(null);
      }
    };

    void loadPost();
  }, []);

  if (!post) {
    return null;
  }

  return (
    <Card className="overflow-hidden rounded-[14px] shadow-none">
      <CardHeader className="px-6 pb-0">
        <CardTitle className="text-base font-semibold">Alby Blog</CardTitle>
      </CardHeader>
      <CardContent className="px-6 pt-0">
        <div className="relative h-[247px] overflow-hidden rounded-xl border">
          {post.imageUrl ? (
            <img
              src={post.imageUrl}
              alt={post.title}
              className="absolute inset-0 size-full object-cover"
            />
          ) : (
            <div
              className={cn(
                "absolute inset-0 bg-gradient-to-br",
                themeClassName
              )}
            >
              <div className="absolute -left-10 top-6 size-36 rounded-full bg-white/35 blur-3xl" />
              <div className="absolute -right-8 bottom-2 size-40 rounded-full bg-white/20 blur-3xl" />
              <div className="absolute inset-0 bg-white/10" />
            </div>
          )}
        </div>
      </CardContent>
      <CardFooter className="flex flex-col items-start gap-4 px-6 pb-6 pt-0">
        <div className="space-y-1">
          <p className="text-xl font-semibold leading-7 text-foreground">
            {post.title}
          </p>
          <p className="text-base leading-6 text-muted-foreground">
            {post.description}
          </p>
        </div>
        <div className="flex w-full justify-end">
          <ExternalLinkButton to={post.url} variant="outline">
            Read on Alby Blog
            <SquareArrowOutUpRightIcon />
          </ExternalLinkButton>
        </div>
      </CardFooter>
    </Card>
  );
}
