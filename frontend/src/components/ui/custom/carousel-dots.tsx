"use client";

import * as React from "react";

import { cn } from "src/lib/utils";

// Dot Button Hook
export type UseDotButtonType = {
  selectedIndex: number;
  scrollSnaps: number[];
  onDotButtonClick: (index: number) => void;
};

// Dot Button Component
const CarouselDotButton = React.forwardRef<
  HTMLButtonElement,
  React.ComponentPropsWithRef<"button">
>(({ className, ...props }, ref) => {
  return (
    <button
      ref={ref}
      type="button"
      className={cn(
        "h-2 w-2 rounded-full border-0 bg-muted transition-colors hover:bg-muted-foreground/50 focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:pointer-events-none disabled:opacity-50",
        "data-[selected=true]:bg-primary",
        className
      )}
      {...props}
    />
  );
});
CarouselDotButton.displayName = "CarouselDotButton";

// Dots Container Component
const CarouselDots = React.forwardRef<
  HTMLDivElement,
  React.HTMLAttributes<HTMLDivElement>
>(({ className, ...props }, ref) => (
  <div
    ref={ref}
    className={cn("flex items-center justify-center gap-2 p-2", className)}
    {...props}
  />
));
CarouselDots.displayName = "CarouselDots";

export { CarouselDotButton, CarouselDots };
