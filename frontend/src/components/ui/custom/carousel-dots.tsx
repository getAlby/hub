"use client";

import type { EmblaCarouselType } from "embla-carousel";
import * as React from "react";

import { cn } from "src/lib/utils";

// Dot Button Hook
type UseDotButtonType = {
  selectedIndex: number;
  scrollSnaps: number[];
  onDotButtonClick: (index: number) => void;
};

export const useDotButton = (
  emblaApi: EmblaCarouselType | undefined
): UseDotButtonType => {
  const [selectedIndex, setSelectedIndex] = React.useState(0);
  const [scrollSnaps, setScrollSnaps] = React.useState<number[]>([]);

  const onDotButtonClick = React.useCallback(
    (index: number) => {
      if (!emblaApi) {
        return;
      }
      emblaApi.scrollTo(index);
    },
    [emblaApi]
  );

  const onInit = React.useCallback((emblaApi: EmblaCarouselType) => {
    setScrollSnaps(emblaApi.scrollSnapList());
  }, []);

  const onSelect = React.useCallback((emblaApi: EmblaCarouselType) => {
    setSelectedIndex(emblaApi.selectedScrollSnap());
  }, []);

  React.useEffect(() => {
    if (!emblaApi) {
      return;
    }

    onInit(emblaApi);
    onSelect(emblaApi);
    emblaApi.on("reInit", onInit).on("reInit", onSelect).on("select", onSelect);
  }, [emblaApi, onInit, onSelect]);

  return {
    selectedIndex,
    scrollSnaps,
    onDotButtonClick,
  };
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
