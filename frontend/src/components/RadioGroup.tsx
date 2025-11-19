import * as RadioGroupPrimitive from "@radix-ui/react-radio-group";
import { Label } from "src/components/ui/label";
import { RadioGroupItem as ShadcnRadioGroupItem } from "src/components/ui/radio-group";
import { cn } from "src/lib/utils";

export function RadioGroupItem(
  props: React.ComponentProps<typeof RadioGroupPrimitive.Item> & {
    selected?: boolean;
  }
) {
  const { selected, ...otherProps } = props;

  return (
    <Label
      htmlFor={otherProps.id}
      className={cn(
        "flex items-center gap-3 p-4 rounded-lg hover:bg-primary/5 dark:hover:bg-primary/10 cursor-pointer",
        selected &&
          "border border-primary bg-primary/5 dark:bg-primary/10 text-sm font-medium text-primary"
      )}
    >
      <ShadcnRadioGroupItem {...otherProps} />
      <span>{props.children}</span>
    </Label>
  );
}
