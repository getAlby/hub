"use client";

import * as React from "react";
import { Avatar, AvatarFallback, AvatarImage } from "src/components/ui/avatar";
import { cn } from "src/lib/utils";

// Types for our Stories component
export interface Story {
  id: string;
  avatar: string;
  username: string;
  seen: boolean;
  videoUrl?: string; // YouTube video URL
}

interface StoryItemProps extends React.HTMLAttributes<HTMLDivElement> {
  story: Story;
  size?: "sm" | "md" | "lg";
  onStoryClick?: (id: string) => void;
}

interface StoriesProps extends React.HTMLAttributes<HTMLDivElement> {
  stories: Story[];
  size?: "sm" | "md" | "lg";
  onStoryClick?: (id: string) => void;
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
        <Avatar className={cn("border-2 border-background size-14")}>
          <AvatarImage src={story.avatar} alt={story.username} />
          <AvatarFallback>
            {story.username.substring(0, 2).toUpperCase()}
          </AvatarFallback>
        </Avatar>
      </div>
      <span className={cn("text-center max-w-14 text-xs")}>
        {story.username}
      </span>
    </div>
  );
}

// Main Stories component with horizontal scrolling
export function Stories({
  stories,
  size = "md",
  onStoryClick,
  className,
  ...props
}: StoriesProps) {
  return (
    <div
      className={cn("flex gap-4 overflow-x-auto pb-2", className)}
      {...props}
    >
      {stories.map((story) => (
        <div key={story.id}>
          <StoryItem story={story} size={size} onStoryClick={onStoryClick} />
        </div>
      ))}
    </div>
  );
}
