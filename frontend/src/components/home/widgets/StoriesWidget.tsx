import { RefreshCcwIcon, WalletCardsIcon, XIcon } from "lucide-react";
import React from "react";
import albyExtension from "src/assets/suggested-apps/alby-extension.png";
import albyGo from "src/assets/suggested-apps/alby-go.png";
import ExternalLink from "src/components/ExternalLink";
import { Badge } from "src/components/ui/badge";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";
import { Button } from "src/components/ui/button";
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogTitle,
} from "src/components/ui/dialog";
import { cn } from "src/lib/utils";

const STORIES_VIEWED_STORAGE_KEY = "alby-hub-home-stories-viewed";

type Story = {
  id: string;
  title: string;
  badge?: string;
  youtubeUrl: string;
  description: string;
  logo?: string;
  emoji?: string;
  icon?: React.ComponentType<{ className?: string }>;
  innerClassName?: string;
};

const stories: Story[] = [
  {
    id: "hub-release",
    title: "Alby Hub 1.72.1",
    badge: "NEW",
    youtubeUrl: "https://www.youtube.com/watch?v=M7lc1UVf-VE",
    description:
      "Placeholder release story. Swap this with the latest Hub update video.",
    emoji: "✨",
    innerClassName:
      "bg-amber-50 text-amber-700 dark:bg-amber-950/40 dark:text-amber-200",
  },
  {
    id: "alby-go",
    title: "Alby Go",
    youtubeUrl: "https://www.youtube.com/watch?v=ysz5S6PUM-U",
    description: "Placeholder mobile wallet story for Alby Go.",
    logo: albyGo,
    innerClassName: "bg-white dark:bg-muted",
  },
  {
    id: "alby-extension",
    title: "Alby Extension",
    youtubeUrl: "https://www.youtube.com/watch?v=jNQXAC9IVRw",
    description: "Placeholder browser extension story with a working player.",
    logo: albyExtension,
    innerClassName: "bg-white dark:bg-muted",
  },
  {
    id: "swaps",
    title: "Swaps",
    youtubeUrl: "https://www.youtube.com/watch?v=aqz-KE-bpKQ",
    description:
      "Placeholder story slot for future swap walkthroughs and tutorials.",
    icon: RefreshCcwIcon,
    innerClassName:
      "bg-emerald-50 text-emerald-600 dark:bg-emerald-950/40 dark:text-emerald-300",
  },
  {
    id: "sub-wallets",
    title: "Sub-wallets",
    youtubeUrl: "https://www.youtube.com/watch?v=ScMzIvxBSi4",
    description:
      "Placeholder story slot for sub-wallet tips and product updates.",
    icon: WalletCardsIcon,
    innerClassName:
      "bg-cyan-50 text-cyan-700 dark:bg-cyan-950/40 dark:text-cyan-200",
  },
];

function loadViewedStoryIds(): Set<string> {
  try {
    const raw = localStorage.getItem(STORIES_VIEWED_STORAGE_KEY);
    if (!raw) {
      return new Set();
    }
    const parsed = JSON.parse(raw) as unknown;
    if (!Array.isArray(parsed)) {
      return new Set();
    }
    return new Set(parsed.filter((id): id is string => typeof id === "string"));
  } catch {
    return new Set();
  }
}

function persistViewedStoryIds(ids: Set<string>) {
  try {
    localStorage.setItem(STORIES_VIEWED_STORAGE_KEY, JSON.stringify([...ids]));
  } catch {
    // ignore quota / private mode
  }
}

function getYouTubeEmbedUrl(url: string) {
  try {
    const parsedUrl = new URL(url);
    const host = parsedUrl.hostname.replace(/^www\./, "");
    let videoId = "";

    if (host === "youtu.be") {
      videoId = parsedUrl.pathname.replace("/", "");
    } else if (host.endsWith("youtube.com")) {
      if (parsedUrl.pathname === "/watch") {
        videoId = parsedUrl.searchParams.get("v") || "";
      } else if (parsedUrl.pathname.startsWith("/shorts/")) {
        videoId = parsedUrl.pathname.split("/")[2] || "";
      } else if (parsedUrl.pathname.startsWith("/embed/")) {
        videoId = parsedUrl.pathname.split("/")[2] || "";
      }
    }

    if (!videoId) {
      return url;
    }

    return `https://www.youtube-nocookie.com/embed/${videoId}?autoplay=1&rel=0`;
  } catch {
    return url;
  }
}

function StoryAvatar({ story, viewed }: { story: Story; viewed: boolean }) {
  const Icon = story.icon;

  return (
    <div
      className={cn(
        "relative box-border flex size-[60px] shrink-0 items-center justify-center rounded-full border-2",
        viewed ? "border-accent" : "border-primary"
      )}
    >
      <div
        className={cn(
          "relative flex size-[56px] items-center justify-center overflow-hidden rounded-full",
          story.innerClassName
        )}
      >
        {story.logo ? (
          <img
            src={story.logo}
            alt={`${story.title} story`}
            className="size-[56px] rounded-full object-cover"
          />
        ) : story.emoji ? (
          <span className="text-[28px] leading-none">{story.emoji}</span>
        ) : Icon ? (
          <Icon className="size-7" />
        ) : null}
      </div>
      {story.badge && !viewed && (
        <Badge className="absolute -left-1 top-0 rounded-full px-1.5 py-0 text-[9px] leading-4">
          {story.badge}
        </Badge>
      )}
    </div>
  );
}

export function StoriesWidget() {
  const [activeStory, setActiveStory] = React.useState<Story | null>(null);
  const [viewedIds, setViewedIds] =
    React.useState<Set<string>>(loadViewedStoryIds);

  const markStoryViewed = React.useCallback((storyId: string) => {
    setViewedIds((prev) => {
      if (prev.has(storyId)) {
        return prev;
      }
      const next = new Set(prev);
      next.add(storyId);
      persistViewedStoryIds(next);
      return next;
    });
  }, []);

  return (
    <>
      <Card className="overflow-hidden rounded-[14px] shadow-none">
        <CardHeader className="px-6 pb-0">
          <CardTitle className="text-base font-semibold">Stories</CardTitle>
        </CardHeader>
        <CardContent className="px-0 py-0">
          <div className="flex gap-3 overflow-x-auto px-6 pb-1">
            {stories.map((story) => {
              const viewed = viewedIds.has(story.id);
              return (
                <button
                  key={story.id}
                  type="button"
                  onClick={() => {
                    markStoryViewed(story.id);
                    setActiveStory(story);
                  }}
                  className="flex w-[73px] shrink-0 flex-col items-center gap-2 text-center"
                >
                  <StoryAvatar story={story} viewed={viewed} />
                  <span
                    className={cn(
                      "w-full text-[10px] leading-none",
                      viewed
                        ? "font-medium text-muted-foreground"
                        : "font-semibold text-foreground"
                    )}
                  >
                    {story.title}
                  </span>
                </button>
              );
            })}
          </div>
        </CardContent>
      </Card>

      <Dialog
        open={!!activeStory}
        onOpenChange={(open) => !open && setActiveStory(null)}
      >
        <DialogContent
          showCloseButton={false}
          className="max-w-4xl overflow-hidden border-0 bg-zinc-950 p-0 text-white"
        >
          {activeStory && (
            <div className="flex flex-col">
              <div className="flex items-start justify-between gap-4 border-b border-white/10 px-5 py-4">
                <div className="min-w-0">
                  <DialogTitle className="truncate text-lg font-semibold text-white">
                    {activeStory.title}
                  </DialogTitle>
                  <DialogDescription className="mt-1 text-sm text-zinc-300">
                    {activeStory.description}
                  </DialogDescription>
                </div>
                <DialogClose asChild>
                  <Button
                    type="button"
                    variant="ghost"
                    size="icon"
                    className="shrink-0 text-white hover:bg-white/10 hover:text-white"
                  >
                    <XIcon className="size-5" />
                    <span className="sr-only">Close story</span>
                  </Button>
                </DialogClose>
              </div>

              <div className="aspect-video w-full bg-black">
                <iframe
                  className="size-full"
                  src={getYouTubeEmbedUrl(activeStory.youtubeUrl)}
                  title={activeStory.title}
                  allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share"
                  allowFullScreen
                />
              </div>

              <div className="flex items-center justify-end gap-3 px-5 py-4">
                <ExternalLink
                  to={activeStory.youtubeUrl}
                  className="inline-flex h-9 items-center rounded-md border border-white/15 px-4 text-sm font-medium text-white transition-colors hover:bg-white/10"
                >
                  Watch on YouTube
                </ExternalLink>
              </div>
            </div>
          )}
        </DialogContent>
      </Dialog>
    </>
  );
}
