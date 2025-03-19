import React from "react";
import RevealPasswordToggle from "src/components/password/RevealPasswordToggle";
import { Input } from "src/components/ui/input";

type PasswordInputProps = Omit<
  React.InputHTMLAttributes<HTMLInputElement>,
  "type" | "onChange"
> & {
  onChange?: (value: string) => void;
};

export default function PasswordInput({
  onChange,
  placeholder,
  name = "password",
  value = "",
  ...restProps
}: PasswordInputProps) {
  const [password, setPassword] = React.useState(String(value));
  const [passwordVisible, setPasswordVisible] = React.useState(false);

  React.useEffect(() => {
    if (onChange) {
      onChange(password);
    }
  }, [password, onChange]);

  return (
    <Input
      type={passwordVisible ? "text" : "password"}
      name={name}
      value={password}
      required
      onChange={(e) => setPassword(e.target.value)}
      placeholder={placeholder}
      {...restProps}
      endAdornment={
        <RevealPasswordToggle
          isRevealed={passwordVisible}
          onChange={setPasswordVisible}
        />
      }
    />
  );
}
