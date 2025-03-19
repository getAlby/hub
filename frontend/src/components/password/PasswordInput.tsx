import React from "react";
import RevealPasswordToggle from "src/components/password/RevealPasswordToggle";
import { Input } from "src/components/ui/input";

type PasswordInputProps = {
  onChange: (value: string) => void;
  id: string;
  placeholder?: string;
};

export default function PasswordInput({
  onChange,
  id,
  placeholder = "Password",
}: PasswordInputProps) {
  const [password, setPassword] = React.useState("");
  const [passwordVisible, setPasswordVisible] = React.useState(false);

  React.useEffect(() => {
    onChange(password);
  }, [password, onChange]);

  return (
    <Input
      id={id}
      type={passwordVisible ? "text" : "password"}
      name="password"
      value={password}
      required={true}
      onChange={(e) => setPassword(e.target.value)}
      placeholder={placeholder}
      endAdornment={
        <RevealPasswordToggle
          isRevealed={passwordVisible}
          onChange={setPasswordVisible}
        />
      }
    />
  );
}
