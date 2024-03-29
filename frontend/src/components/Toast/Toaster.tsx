import { Toaster as HotToaster } from "react-hot-toast";

export default function Toaster() {
  return (
    <HotToaster
      position="bottom-right"
      containerClassName="text-foreground"
      toastOptions={{
        duration: 4_000,
      }}
    />
  );
}
