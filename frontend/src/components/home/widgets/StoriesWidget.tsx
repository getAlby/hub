import { XIcon } from "lucide-react";
import React from "react";
import ExternalLink from "src/components/ExternalLink";
import { Button } from "src/components/ui/button";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";
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
  avatar: string;
  videoUrl?: string;
};

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
  return (
    <div
      className={cn(
        "relative box-border flex size-[60px] shrink-0 items-center justify-center rounded-full border-2",
        viewed ? "border-accent" : "border-primary"
      )}
    >
      <div className="relative flex size-[56px] items-center justify-center overflow-hidden rounded-full bg-white dark:bg-muted">
        <img
          src={story.avatar}
          alt={`${story.title} story`}
          className="size-[56px] rounded-full object-cover"
        />
      </div>
    </div>
  );
}

export function StoriesWidget() {
  const [stories, setStories] = React.useState<Story[]>([]);
  const [isLoading, setIsLoading] = React.useState(true);
  const [activeStory, setActiveStory] = React.useState<Story | null>(null);
  const [viewedIds, setViewedIds] =
    React.useState<Set<string>>(loadViewedStoryIds);

  React.useEffect(() => {
    const loadStories = async () => {
      try {
        const response = await fetch("/api/alby/stories");
        if (!response.ok) {
          throw new Error(`Failed to fetch stories: ${response.status}`);
        }
        const payload = (await response.json()) as Array<{
          id: number;
          title: string;
          avatar: string;
          videoUrl?: string;
        }>;
        const mappedStories = payload.map((story) => ({
          id: String(story.id),
          title: story.title,
          avatar: story.avatar,
          videoUrl: story.videoUrl,
        }));
        setStories(mappedStories);
      } catch {
        setStories([]);
      } finally {
        setIsLoading(false);
      }
    };

    void loadStories();
  }, []);

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

  if (!isLoading && stories.length === 0) {
    return null;
  }

  return (
    <>
      <Card className="overflow-hidden rounded-[14px] shadow-none">
        <CardHeader className="px-6 pb-0">
          <CardTitle className="text-base font-semibold">Stories</CardTitle>
        </CardHeader>
        <CardContent className="px-0 py-0">
          <div className="flex gap-3 overflow-x-auto px-6 pb-1">
            {isLoading && (
              <span className="text-sm text-muted-foreground">
                Loading stories...
              </span>
            )}
            {!isLoading &&
              stories.map((story) => {
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
                        "w-full text-xs leading-normal",
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
                    Watch the latest update
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

              {activeStory.videoUrl && (
                <div className="aspect-video w-full bg-black">
                  <iframe
                    className="size-full"
                    src={getYouTubeEmbedUrl(activeStory.videoUrl)}
                    title={activeStory.title}
                    allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share"
                    allowFullScreen
                  />
                </div>
              )}

              {activeStory.videoUrl && (
                <div className="flex items-center justify-end gap-3 px-5 py-4">
                  <ExternalLink
                    to={activeStory.videoUrl}
                    className="inline-flex h-9 items-center rounded-md border border-white/15 px-4 text-sm font-medium text-white transition-colors hover:bg-white/10"
                  >
                    Watch on YouTube
                  </ExternalLink>
                </div>
              )}
            </div>
          )}
        </DialogContent>
      </Dialog>
    </>
  );
}
