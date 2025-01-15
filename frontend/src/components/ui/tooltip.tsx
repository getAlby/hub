import {
  PopoverContentProps,
  PopoverProps,
  PopoverTriggerProps,
} from "@radix-ui/react-popover";
import * as TooltipPrimitive from "@radix-ui/react-tooltip";
import {
  TooltipContentProps,
  TooltipProps,
  TooltipTriggerProps,
} from "@radix-ui/react-tooltip";
import * as React from "react";
import {
  PropsWithChildren,
  createContext,
  useContext,
  useEffect,
  useState,
} from "react";
import { Popover, PopoverContent, PopoverTrigger } from "./popover";

import { cn } from "src/lib/utils";

const TooltipProvider = TooltipPrimitive.Provider;

TooltipProvider.defaultProps = {
  ...(TooltipProvider.defaultProps || {}),
  delayDuration: 200,
};

const Tooltip = TooltipPrimitive.Root;

const TooltipTrigger = TooltipPrimitive.Trigger;

const TooltipContent = React.forwardRef<
  React.ElementRef<typeof TooltipPrimitive.Content>,
  React.ComponentPropsWithoutRef<typeof TooltipPrimitive.Content>
>(({ className, sideOffset = 4, ...props }, ref) => (
  <TooltipPrimitive.Content
    ref={ref}
    sideOffset={sideOffset}
    className={cn(
      "z-50 overflow-hidden rounded-md bg-primary px-3 py-1.5 text-xs text-primary-foreground animate-in fade-in-0 zoom-in-95 data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=closed]:zoom-out-95 data-[side=bottom]:slide-in-from-top-2 data-[side=left]:slide-in-from-right-2 data-[side=right]:slide-in-from-left-2 data-[side=top]:slide-in-from-bottom-2",
      className
    )}
    {...props}
  />
));
TooltipContent.displayName = TooltipPrimitive.Content.displayName;

// Hybrid tooltip

const TouchContext = createContext<boolean | undefined>(undefined);
const useTouch = () => useContext(TouchContext);

const TouchProvider = (props: PropsWithChildren) => {
  const [isTouch, setTouch] = useState<boolean>();

  useEffect(() => {
    setTouch(window.matchMedia("(pointer: coarse)").matches);
  }, []);

  return <TouchContext.Provider value={isTouch} {...props} />;
};

const HybridTooltip = (props: TooltipProps & PopoverProps) => {
  const isTouch = useTouch();

  return isTouch ? <Popover {...props} /> : <Tooltip {...props} />;
};

const HybridTooltipTrigger = (
  props: TooltipTriggerProps & PopoverTriggerProps
) => {
  const isTouch = useTouch();

  return isTouch ? (
    <PopoverTrigger {...props} />
  ) : (
    <TooltipTrigger {...props} />
  );
};

const HybridTooltipContent = (
  props: TooltipContentProps & PopoverContentProps
) => {
  const isTouch = useTouch();

  return isTouch ? (
    <PopoverContent
      {...props}
      className={cn(
        "bg-primary text-primary-foreground text-sm",
        props.className
      )}
    />
  ) : (
    <TooltipContent {...props} />
  );
};

export {
  HybridTooltip as Tooltip,
  HybridTooltipContent as TooltipContent,
  TooltipProvider,
  HybridTooltipTrigger as TooltipTrigger,
  TouchProvider,
};
