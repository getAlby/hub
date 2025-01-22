import { EyeIcon, EyeOffIcon } from "lucide-react";
import { useEffect, useState } from "react";

type Props = {
  onChange: (viewingPassword: boolean) => void;
  isRevealed?: boolean;
};

export default function PasswordViewAdornment({ onChange, isRevealed }: Props) {
  const [_isRevealed, setRevealed] = useState(false);

  // toggle the button if password view is handled by component itself
  useEffect(() => {
    if (typeof isRevealed !== "undefined") {
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
        <EyeOffIcon className="h-4 w-4" />
      ) : (
        <EyeIcon className="h-4 w-4" />
      )}
    </button>
  );
}
