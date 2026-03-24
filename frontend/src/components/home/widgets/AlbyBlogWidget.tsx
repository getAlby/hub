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
import {
  FramerBlogPost,
  framerBlogPosts,
} from "src/data/framerBlogPosts.generated";
import { cn } from "src/lib/utils";

type BlogPost = {
  id: string;
  category: string;
  title: string;
  description: string;
  url: string;
  themeClassName: string;
  imageUrl?: string;
  publishedAt?: string;
};

const manualBlogPosts: BlogPost[] = [
  {
    id: "wavecard",
    category: "Experience",
    title: "Stop shorting bitcoin with Alby Hub and wavecard",
    description:
      "Many bitcoiners unknowingly short BTC by holding fiat. Learn how wavecard and Alby Hub let you truly live on bitcoin.",
    url: "https://blog.getalby.com/blog/stop-shorting-bitcoin-with-alby-hub-and-wavecard",
    themeClassName: "from-emerald-200 via-cyan-200 to-yellow-200",
    publishedAt: "2025-01-01T00:00:00.000Z",
  },
  {
    id: "cli-skill",
    category: "Developer",
    title: "How Pull That Up Jamie uses the Alby CLI Skill",
    description:
      "A look at how Alby tooling can power agentic workflows, fast prototyping, and local AI automation.",
    url: "https://blog.getalby.com/blog/how-pull-that-up-jamie-uses-the-alby-cli-skill",
    themeClassName: "from-orange-200 via-amber-100 to-pink-100",
    publishedAt: "2024-12-01T00:00:00.000Z",
  },
  {
    id: "modularizing",
    category: "Alby Hub",
    title: "Modularizing Alby Hub: User-Hosted Serverless Signers",
    description:
      "A deeper look at the next step in making Alby Hub more modular and easier to run across different setups.",
    url: "https://blog.getalby.com/blog/modularizing-alby-hub-user-hosted-serverless-signers",
    themeClassName: "from-slate-200 via-zinc-100 to-lime-100",
    publishedAt: "2024-11-01T00:00:00.000Z",
  },
];

const fallbackThemes = [
  "from-emerald-200 via-cyan-200 to-yellow-200",
  "from-orange-200 via-amber-100 to-pink-100",
  "from-slate-200 via-zinc-100 to-lime-100",
  "from-sky-200 via-indigo-100 to-violet-100",
];

function prettifySlugTitle(slug: string) {
  return slug
    .split("-")
    .filter(Boolean)
    .map((word) => word.charAt(0).toUpperCase() + word.slice(1))
    .join(" ");
}

function fromFramerPost(post: FramerBlogPost, index: number): BlogPost {
  const rawTitle = post.title?.trim() || "";
  const displayTitle =
    !rawTitle ||
    rawTitle === post.slug ||
    rawTitle.replace(/\s+/g, "-") === post.slug
      ? prettifySlugTitle(post.slug)
      : rawTitle;

  return {
    id: post.id,
    category: post.category || "Blog",
    title: displayTitle,
    description: post.description || "Read more on the Alby blog.",
    url: post.url,
    themeClassName: fallbackThemes[index % fallbackThemes.length],
    imageUrl: post.imageUrl,
    publishedAt: post.publishedAt,
  };
}

function pickLatestPost(posts: BlogPost[]): BlogPost {
  const dated = posts.filter((p) => p.publishedAt);
  if (dated.length > 0) {
    return [...dated].sort(
      (a, b) =>
        new Date(b.publishedAt!).getTime() - new Date(a.publishedAt!).getTime()
    )[0];
  }
  return posts[0];
}

export function AlbyBlogWidget() {
  const post = React.useMemo(() => {
    const feedPosts = framerBlogPosts.map((p, index) =>
      fromFramerPost(p, index)
    );
    const source = feedPosts.length > 0 ? feedPosts : manualBlogPosts;
    return pickLatestPost(source);
  }, []);

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
                post.themeClassName
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
