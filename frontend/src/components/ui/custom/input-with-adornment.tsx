import * as React from "react";
import { Input } from "src/components/ui/input";

export interface InputWithAdornmentProps extends React.ComponentProps<"input"> {
  endAdornment: React.ReactNode;
}

const InputWithAdornment = React.forwardRef<
  HTMLInputElement,
  InputWithAdornmentProps
>(({ className, type, endAdornment, ...props }, ref) => {
  return (
    <div className="relative flex items-center w-full">
      <Input type={type} ref={ref} {...props} />
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
