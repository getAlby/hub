import * as React from "react";
import { type PropsWithChildren, createContext, useContext } from "react";
import {
  Popover as PopoverPrimitive,
  Tooltip as TooltipPrimitive,
} from "radix-ui";

import { cn } from "src/lib/utils";

import { Popover, PopoverContent, PopoverTrigger } from "./popover";

function TooltipProvider({
  delayDuration = 0,
  ...props
}: React.ComponentProps<typeof TooltipPrimitive.Provider>) {
  return (
    <TooltipPrimitive.Provider
      data-slot="tooltip-provider"
      delayDuration={delayDuration}
      {...props}
    />
  );
}

function RadixTooltipRootWithProvider(
  props: React.ComponentProps<typeof TooltipPrimitive.Root>
) {
  return (
    <TooltipProvider delayDuration={200}>
      <TooltipPrimitive.Root data-slot="tooltip" {...props} />
    </TooltipProvider>
  );
}

function BaseTooltipContent({
  className,
  sideOffset = 0,
  children,
  ...props
}: React.ComponentProps<typeof TooltipPrimitive.Content>) {
  return (
    <TooltipPrimitive.Portal>
      <TooltipPrimitive.Content
        data-slot="tooltip-content"
        sideOffset={sideOffset}
        className={cn(
          "z-50 w-fit origin-(--radix-tooltip-content-transform-origin) animate-in rounded-md bg-foreground px-3 py-1.5 text-xs text-balance text-background fade-in-0 zoom-in-95 data-[side=bottom]:slide-in-from-top-2 data-[side=left]:slide-in-from-right-2 data-[side=right]:slide-in-from-left-2 data-[side=top]:slide-in-from-bottom-2 data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=closed]:zoom-out-95",
          className
        )}
        {...props}
      >
        {children}
        <TooltipPrimitive.Arrow className="z-50 size-2.5 translate-y-[calc(-50%_-_2px)] rotate-45 rounded-[2px] bg-foreground fill-foreground" />
      </TooltipPrimitive.Content>
    </TooltipPrimitive.Portal>
  );
}

// Hybrid tooltip for touch devices

const TouchContext = createContext<boolean | undefined>(undefined);

const useTouch = () => useContext(TouchContext);

function TouchProvider(props: PropsWithChildren) {
  const isTouch = React.useMemo(
    () => window.matchMedia("(pointer: coarse)").matches,
    []
  );

  return <TouchContext.Provider value={isTouch} {...props} />;
}

type HybridTooltipProps = React.ComponentProps<typeof TooltipPrimitive.Root> &
  React.ComponentProps<typeof PopoverPrimitive.Root>;

function HybridTooltip(props: HybridTooltipProps) {
  const isTouch = useTouch();

  return isTouch ? (
    <Popover {...props} />
  ) : (
    <RadixTooltipRootWithProvider {...props} />
  );
}

type HybridTooltipTriggerProps = React.ComponentProps<
  typeof TooltipPrimitive.Trigger
> &
  React.ComponentProps<typeof PopoverPrimitive.Trigger>;

function HybridTooltipTrigger(props: HybridTooltipTriggerProps) {
  const isTouch = useTouch();

  return isTouch ? (
    <PopoverTrigger {...props} />
  ) : (
    <TooltipPrimitive.Trigger data-slot="tooltip-trigger" {...props} />
  );
}

type HybridTooltipContentProps = React.ComponentProps<
  typeof TooltipPrimitive.Content
> &
  React.ComponentProps<typeof PopoverPrimitive.Content>;

function HybridTooltipContent({
  className,
  ...props
}: HybridTooltipContentProps) {
  const isTouch = useTouch();

  return isTouch ? (
    <PopoverContent
      {...props}
      className={cn("bg-foreground text-background text-sm", className)}
    />
  ) : (
    <BaseTooltipContent {...props} className={className} />
  );
}

export {
  HybridTooltip as Tooltip,
  HybridTooltipContent as TooltipContent,
  TooltipProvider,
  HybridTooltipTrigger as TooltipTrigger,
  TouchProvider,
};
