import Lottie from "lottie-react";
import animationDataDark from "src/assets/lotties/loading-dark.json";
import animationDataLight from "src/assets/lotties/loading-light.json";
import { useTheme } from "src/components/ui/theme-provider";

export default function LottieLoading({ size }: { size?: number }) {
  const { isDarkMode } = useTheme();

  return (
    <Lottie
      animationData={isDarkMode ? animationDataDark : animationDataLight}
      loop
      autoplay
      rendererSettings={{ preserveAspectRatio: "xMidYMid slice" }}
      style={{ width: size ?? "100%", height: size ?? "100%" }}
    />
  );
}
