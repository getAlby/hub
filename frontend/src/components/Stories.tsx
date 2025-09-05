import { ChevronDown } from "lucide-react";
import React from "react";
import { Avatar } from "src/components/ui/avatar";
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from "src/components/ui/collapsible";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "src/components/ui/dialog";
import { localStorageKeys } from "src/constants";
import { cn } from "src/lib/utils";

export interface Story {
  id: number;
  avatar: string;
  title: string;
  seen: boolean;
  videoUrl?: string;
}

interface ApiStory {
  id: number;
  avatar: string;
  title: string;
  videoUrl?: string;
}

export default function Stories() {
  const [isStoriesOpen, setIsStoriesOpen] = React.useState(true);
  const [selectedStory, setSelectedStory] = React.useState<Story | null>(null);
  const [dialogOpen, setDialogOpen] = React.useState(false);

  const getReadStoryIds = (): number[] => {
    try {
      const stored = localStorage.getItem(localStorageKeys.readStories);
      return stored ? JSON.parse(stored) : [];
    } catch {
      return [];
    }
  };

  const saveReadStoryIds = (ids: number[]) => {
    try {
      localStorage.setItem(localStorageKeys.readStories, JSON.stringify(ids));
    } catch (error) {
      console.warn("Failed to save read stories to localStorage:", error);
    }
  };

  const loadStoriesFromAPI = React.useCallback(async (): Promise<Story[]> => {
    try {
      const response = await fetch("/api/alby/stories");
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }
      const apiStories: ApiStory[] = await response.json();
      const readIds = getReadStoryIds();

      return apiStories.map((story) => ({
        id: story.id,
        title: story.title,
        avatar: story.avatar,
        seen: readIds.includes(story.id),
        videoUrl: story.videoUrl,
      }));
    } catch (error) {
      console.error("Failed to load stories from API:", error);
      // Fallback to empty array or default stories
      return [];
    }
  }, []);

  const initializeStories = (): Story[] => {
    // Return empty array initially, will be populated by useEffect
    return [];
  };

  const [stories, setStories] = React.useState<Story[]>(initializeStories());
  const [isLoading, setIsLoading] = React.useState(true);

  React.useEffect(() => {
    const fetchStories = async () => {
      setIsLoading(true);
      const loadedStories = await loadStoriesFromAPI();
      setStories(loadedStories);
      setIsLoading(false);
    };

    fetchStories();
  }, [loadStoriesFromAPI]);

  const handleStoryClick = (id: number) => {
    const story = stories.find((s) => s.id === id);
    if (story) {
      setSelectedStory(story);
      setDialogOpen(true);

      if (!story.seen) {
        const currentReadIds = getReadStoryIds();
        const updatedReadIds = [...currentReadIds, id];
        saveReadStoryIds(updatedReadIds);

        setStories((prevStories) =>
          prevStories.map((s) => (s.id === id ? { ...s, seen: true } : s))
        );
      }
    }
  };

  const renderStoryItem = (story: Story) => {
    return (
      <div key={story.id} className="flex flex-col items-center gap-1">
        <div
          className={cn(
            "rounded-full border-2 flex items-center justify-center cursor-pointer p-0.5",
            story.seen ? "bg-muted" : "border-primary"
          )}
          onClick={() => handleStoryClick(story.id)}
        >
          <Avatar className={cn("border-background size-14")}>
            <div className="w-full h-full flex items-center justify-center">
              <img
                src={story.avatar}
                alt={story.title}
                className="size-8 rounded-full object-cover"
              />
            </div>
          </Avatar>
        </div>
        <div className="text-center max-w-18 text-xs overflow-hidden overflow-ellipsis">
          {story.title}
        </div>
      </div>
    );
  };

  if (!isLoading && stories.length === 0) {
    return null;
  }

  return (
    <>
      <Collapsible open={isStoriesOpen} onOpenChange={setIsStoriesOpen}>
        <CollapsibleTrigger>
          <div className="flex gap-2 items-center">
            <ChevronDown className="size-4" />
            Stories
          </div>
        </CollapsibleTrigger>
        <CollapsibleContent>
          <div className="flex gap-3 overflow-x-auto mb-2">
            {isLoading ? (
              <div className="text-sm text-muted-foreground">
                Loading stories...
              </div>
            ) : (
              [...stories]
                .sort((a, b) => (a.seen === b.seen ? 0 : a.seen ? 1 : -1))
                .map(renderStoryItem)
            )}
          </div>
        </CollapsibleContent>
      </Collapsible>

      <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
        <DialogContent className="sm:max-w-[800px]">
          <DialogHeader>
            <DialogTitle>{selectedStory?.title}</DialogTitle>
          </DialogHeader>
          <div className="aspect-video w-full">
            {selectedStory?.videoUrl && (
              <iframe
                src={selectedStory.videoUrl}
                title={`${selectedStory.title}`}
                className="w-full h-full rounded-md"
                allowFullScreen
              ></iframe>
            )}
          </div>
        </DialogContent>
      </Dialog>
    </>
  );
}
