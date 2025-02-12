import React from "react";
import PasswordViewAdornment from "src/components/password/PasswordAdornment";
import { Input } from "src/components/ui/input";

type PasswordInputProps = {
  onChange: (value: string) => void;
  id: string;
};

export default function PasswordInput({ onChange, id }: PasswordInputProps) {
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
      onChange={(e) => setPassword(e.target.value)}
      placeholder="Password"
      endAdornment={
        <PasswordViewAdornment
          isRevealed={passwordVisible}
          onChange={setPasswordVisible}
        />
      }
    />
  );
}
