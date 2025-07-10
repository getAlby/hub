import * as ProgressPrimitive from "@radix-ui/react-progress";
import * as React from "react";

import { cn } from "src/lib/utils";

function Progress({
  className,
  value,
  ...props
}: React.ComponentProps<typeof ProgressPrimitive.Root>) {
  return (
    <ProgressPrimitive.Root
      data-slot="progress"
      className={cn(
        "bg-primary/20 relative h-2 w-full overflow-hidden rounded-full",
        className
      )}
      {...props}
    >
      <ProgressPrimitive.Indicator
        data-slot="progress-indicator"
        className="bg-primary h-full w-full flex-1 transition-all"
        style={{ transform: `translateX(-${100 - (value || 0)}%)` }}
      />
    </ProgressPrimitive.Root>
  );
}

function CircleProgress({
  className,
  value,
  ...props
}: React.ComponentProps<typeof ProgressPrimitive.Root>) {
  <ProgressPrimitive.Root
    className={cn(
      `relative h-20 w-20 overflow-hidden rounded-full flex justify-center items-center`,
      className
    )}
    {...props}
    style={{
      background: `radial-gradient(closest-side, hsl(var(--background)) 79%, transparent 80% 100%), conic-gradient(hsl(var(--primary)) ${value || 0}%, hsl(var(--secondary)) 0)`,
    }}
  >
    {props.children || <div className="">{`${value || 0}%`}</div>}
  </ProgressPrimitive.Root>;
}

export { CircleProgress, Progress };
