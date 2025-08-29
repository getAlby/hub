"use client";

import * as React from "react";
import { Avatar, AvatarFallback, AvatarImage } from "src/components/ui/avatar";
import { cn } from "src/lib/utils";

// Types for our Stories component
export interface Story {
  id: number;
  avatar: string | React.ComponentType<{ className?: string }>;
  title: string;
  seen: boolean;
  videoUrl?: string; // YouTube video URL
}

interface StoryItemProps extends React.HTMLAttributes<HTMLDivElement> {
  story: Story;
  onStoryClick?: (id: number) => void;
}

interface StoriesProps extends React.HTMLAttributes<HTMLDivElement> {
  stories: Story[];
  onStoryClick?: (id: number) => void;
}

// Individual story item component
export function StoryItem({
  story,
  onStoryClick,
  className,
  ...props
}: StoryItemProps) {
  const handleClick = (e: React.MouseEvent<HTMLDivElement>) => {
    if (onStoryClick) {
      onStoryClick(story.id);
    }
  };

  return (
    <div
      className={cn("flex flex-col items-center gap-1", className)}
      {...props}
    >
      <div
        className={cn(
          "rounded-full border-2 flex items-center justify-center cursor-pointer p-0.5",
          story.seen
            ? "bg-muted" // Seen state
            : "border-primary" // Unseen state with animated primary color border
        )}
        onClick={handleClick}
      >
        <Avatar className={cn("border-background size-14")}>
          {typeof story.avatar === "string" ? (
            <>
              <AvatarImage src={story.avatar} alt={story.title} />
              <AvatarFallback>
                {story.title.substring(0, 2).toUpperCase()}
              </AvatarFallback>
            </>
          ) : (
            <div className="w-full h-full flex items-center justify-center">
              <story.avatar className="size-8" />
            </div>
          )}
        </Avatar>
      </div>
      <div
        className={cn(
          "text-center max-w-18 text-xs overflow-hidden overflow-ellipsis"
        )}
      >
        {story.title}
      </div>
    </div>
  );
}

export function Stories({
  stories,
  onStoryClick,
  className,
  ...props
}: StoriesProps) {
  return (
    <div
      className={cn("flex gap-3 overflow-x-auto mb-2", className)}
      {...props}
    >
      {stories.map((story) => (
        <div key={story.id}>
          <StoryItem story={story} onStoryClick={onStoryClick} />
        </div>
      ))}
    </div>
  );
}
