import Lottie from "lottie-react";
import animationData from "src/assets/lotties/success-check.json";

export default function LottieSuccess({ size = 288 }: { size?: number }) {
  return (
    <div className="[&_path[fill='rgb(75,177,0)']]:fill-positive-foreground [&_path[stroke='rgb(75,177,0)']]:stroke-positive-foreground [&_path[stroke='rgb(255,255,255)']]:stroke-card">
      <Lottie
        animationData={animationData}
        loop={false}
        autoplay
        rendererSettings={{ preserveAspectRatio: "xMidYMid meet" }}
        style={{ width: size, height: size }}
      />
    </div>
  );
}
