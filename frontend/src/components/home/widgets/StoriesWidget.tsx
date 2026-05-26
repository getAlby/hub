import { XIcon } from "lucide-react";
import React from "react";
import { Link } from "react-router";
import useSWR from "swr";
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
import { swrFetcher } from "src/utils/swr";

const STORIES_VIEWED_STORAGE_KEY = "alby-hub-home-stories-viewed";

type StoryCta = {
  label: string;
  url: string;
  openInNewTab: boolean;
};

type Story = {
  id: string;
  title: string;
  avatar: string;
  videoId?: string;
  cta?: StoryCta;
};

type StoryApiResponse = {
  id: number;
  title: string;
  avatar: string;
  videoId?: string;
  cta?: StoryCta;
};

function youTubeEmbedUrl(videoId: string) {
  return `https://www.youtube-nocookie.com/embed/${videoId}?autoplay=1&rel=0`;
}

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

function StoryAvatar({ story, viewed }: { story: Story; viewed: boolean }) {
  return (
    <div
      className={cn(
        "relative box-border flex size-16 shrink-0 items-center justify-center rounded-full border-2 p-0.5",
        viewed ? "border-accent" : "border-primary"
      )}
    >
      <div className="relative flex size-full items-center justify-center overflow-hidden rounded-full bg-white dark:bg-muted">
        <img
          src={story.avatar}
          alt={`${story.title} story`}
          className="size-full rounded-full object-cover"
        />
      </div>
    </div>
  );
}

export function StoriesWidget() {
  const { data, error, isLoading } = useSWR<StoryApiResponse[]>(
    "/api/alby/stories",
    swrFetcher
  );
  const [activeStory, setActiveStory] = React.useState<Story | null>(null);
  const [viewedIds, setViewedIds] =
    React.useState<Set<string>>(loadViewedStoryIds);

  const stories = React.useMemo<Story[]>(
    () =>
      error || !data
        ? []
        : data.map((story) => ({
            id: String(story.id),
            title: story.title,
            avatar: story.avatar,
            videoId: story.videoId,
            cta: story.cta,
          })),
    [data, error]
  );

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
                    className="flex w-16 shrink-0 flex-col items-center gap-2 text-center"
                  >
                    <StoryAvatar story={story} viewed={viewed} />
                    <span
                      className={cn(
                        "w-full text-xs leading-tight",
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
          className="w-[95vw] max-w-[min(95vw,calc((90vh-80px)*16/9))] sm:max-w-[min(95vw,calc((90vh-80px)*16/9))] max-h-[90vh] overflow-hidden border-0 bg-zinc-950 p-0 text-white sm:rounded-2xl"
        >
          {activeStory && (
            <div className="flex flex-col">
              <DialogTitle className="sr-only">{activeStory.title}</DialogTitle>
              <DialogDescription className="sr-only">
                Watch the latest update
              </DialogDescription>

              {activeStory.videoId && (
                <div className="relative aspect-video w-full overflow-hidden rounded-t-2xl bg-black [transform:translateZ(0)]">
                  <iframe
                    className="absolute inset-0 size-full"
                    src={youTubeEmbedUrl(activeStory.videoId)}
                    title={activeStory.title}
                    allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share"
                    allowFullScreen
                  />
                  <DialogClose asChild>
                    <Button
                      type="button"
                      variant="ghost"
                      size="icon"
                      className="absolute right-3 top-3 z-10 rounded-full bg-black/60 text-white backdrop-blur hover:bg-black/80 hover:text-white"
                    >
                      <XIcon className="size-5" />
                      <span className="sr-only">Close story</span>
                    </Button>
                  </DialogClose>
                </div>
              )}

              {activeStory.videoId && (
                <div className="flex items-center justify-between gap-3 px-6 py-4">
                  <div className="min-w-0">
                    <div className="truncate text-base font-semibold text-white">
                      {activeStory.title}
                    </div>
                  </div>
                  {activeStory.cta && (
                    <div className="flex items-center gap-2">
                      {activeStory.cta.openInNewTab ? (
                        <ExternalLink
                          to={activeStory.cta.url}
                          className="inline-flex h-9 items-center rounded-md bg-primary px-4 text-sm font-medium text-primary-foreground transition-colors hover:bg-primary/90"
                        >
                          {activeStory.cta.label}
                        </ExternalLink>
                      ) : (
                        <Link
                          to={activeStory.cta.url}
                          className="inline-flex h-9 items-center rounded-md bg-primary px-4 text-sm font-medium text-primary-foreground transition-colors hover:bg-primary/90"
                        >
                          {activeStory.cta.label}
                        </Link>
                      )}
                    </div>
                  )}
                </div>
              )}
            </div>
          )}
        </DialogContent>
      </Dialog>
    </>
  );
}
