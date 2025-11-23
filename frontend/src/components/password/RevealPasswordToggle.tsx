import { EyeIcon, EyeOffIcon } from "lucide-react";
import { useEffect, useState } from "react";
import { cn } from "src/lib/utils";

type Props = {
  onChange: (isRevealed: boolean) => void;
  isRevealed?: boolean;
  iconClass?: string;
};

export default function RevealPasswordToggle({
  onChange,
  isRevealed,
  iconClass,
}: Props) {
  const [_isRevealed, setRevealed] = useState(false);

  // toggle the button if password view is handled by component itself
  useEffect(() => {
    if (typeof isRevealed !== "undefined") {
      // eslint-disable-next-line react-hooks/set-state-in-effect
      setRevealed(isRevealed);
    }
  }, [isRevealed]);

  return (
    <button
      type="button"
      tabIndex={-1}
      className="flex justify-center items-center w-10 h-8"
      onClick={() => {
        setRevealed(!_isRevealed);
        onChange(!_isRevealed);
      }}
    >
      {_isRevealed ? (
        <EyeOffIcon className={cn("h-4 w-4", iconClass)} />
      ) : (
        <EyeIcon className={cn("h-4 w-4", iconClass)} />
      )}
    </button>
  );
}
