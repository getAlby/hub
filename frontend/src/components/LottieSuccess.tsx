import { useMemo } from "react";
import Lottie from "react-lottie";
import animationData from "src/assets/lotties/success-check.json";

export default function LottieSuccess({ size = 288 }: { size?: number }) {
  const options = useMemo(
    () => ({
      loop: false,
      autoplay: true,
      animationData,
      rendererSettings: { preserveAspectRatio: "xMidYMid meet" },
    }),
    []
  );

  return (
    <div className="[&_path[fill='rgb(75,177,0)']]:fill-positive-foreground [&_path[stroke='rgb(75,177,0)']]:stroke-positive-foreground [&_path[stroke='rgb(255,255,255)']]:stroke-card">
      <Lottie options={options} height={size} width={size} />
    </div>
  );
}
