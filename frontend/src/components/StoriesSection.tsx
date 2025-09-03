import { ChevronDown } from "lucide-react";
import React from "react";
import albyGo from "src/assets/suggested-apps/alby-go.png";
import { AlbyHubIcon } from "src/components/icons/AlbyHubIcon";
import { AlbyHead } from "src/components/images/AlbyHead";
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
  avatar: React.ComponentType<{ className?: string }>;
  title: string;
  seen: boolean;
  videoUrl?: string;
}

export default function StoriesSection() {
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

  const initializeStories = (): Story[] => {
    const readIds = getReadStoryIds();
    return [
      {
        id: 1,
        title: "Update",
        avatar: AlbyHubIcon,
        seen: readIds.includes(1),
        videoUrl: "https://www.youtube.com/embed/Nw8vU46KoTY",
      },
      {
        id: 2,
        title: "getalby.com",
        avatar: AlbyHead,
        seen: readIds.includes(2),
        videoUrl: "https://www.youtube.com/embed/Nw8vU46KoTY",
      },
      {
        id: 3,
        title: "Auto-Swaps",
        avatar: AlbyHubIcon,
        seen: readIds.includes(3),
        videoUrl: "https://www.youtube.com/embed/Nw8vU46KoTY",
      },
      {
        id: 4,
        title: "Alby Go",
        avatar: () => <img src={albyGo} alt="Alby Go" className="size-8" />,
        seen: readIds.includes(4),
        videoUrl: "https://www.youtube.com/embed/Nw8vU46KoTY",
      },
    ];
  };

  const [stories, setStories] = React.useState<Story[]>(initializeStories());

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
              <story.avatar className="size-8" />
            </div>
          </Avatar>
        </div>
        <div className="text-center max-w-18 text-xs overflow-hidden overflow-ellipsis">
          {story.title}
        </div>
      </div>
    );
  };

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
            {[...stories]
              .sort((a, b) => (a.seen === b.seen ? 0 : a.seen ? 1 : -1))
              .map(renderStoryItem)}
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
