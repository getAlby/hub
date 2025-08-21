import * as React from "react";
import { Input } from "src/components/ui/input";
import { cn } from "src/lib/utils";

export interface InputWithAdornmentProps extends React.ComponentProps<"input"> {
  endAdornment: React.ReactNode;
}

const InputWithAdornment = React.forwardRef<
  HTMLInputElement,
  InputWithAdornmentProps
>(({ className, type, endAdornment, ...props }, ref) => {
  return (
    <div className="relative flex items-center w-full">
      <Input
        type={type}
        ref={ref}
        className={cn(
          "[appearance:textfield] [&::-webkit-outer-spin-button]:appearance-none [&::-webkit-inner-spin-button]:appearance-none",
          className
        )}
        {...props}
      />
      {endAdornment && (
        <span className="absolute right-1 flex items-center">
          {endAdornment}
        </span>
      )}
    </div>
  );
});

InputWithAdornment.displayName = "InputWithAdornment";

export { InputWithAdornment };
