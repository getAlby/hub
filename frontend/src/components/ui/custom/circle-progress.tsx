import * as ProgressPrimitive from "@radix-ui/react-progress";
import * as React from "react";
import { cn } from "src/lib/utils";

function CircleProgress({
  className,
  value,
  ...props
}: React.ComponentProps<typeof ProgressPrimitive.Root>) {
  return (
    <ProgressPrimitive.Root
      className={cn(
        `relative h-20 w-20 overflow-hidden rounded-full flex justify-center items-center`,
        className
      )}
      {...props}
      style={{
        background: `radial-gradient(closest-side, var(--color-background) 79%, transparent 80% 100%), conic-gradient(var(--color-primary) ${value || 0}%, var(--color-secondary) 0)`,
      }}
    >
      {props.children || <div className="">{`${value || 0}%`}</div>}
    </ProgressPrimitive.Root>
  );
}

export default CircleProgress;
