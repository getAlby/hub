import { useMemo } from "react";
import Lottie from "react-lottie";
import animationDataDark from "src/assets/lotties/loading-dark.json";
import animationDataLight from "src/assets/lotties/loading-light.json";
import { useTheme } from "src/components/ui/theme-provider";

export default function LottieLoading({ size }: { size?: number }) {
  const { isDarkMode } = useTheme();

  const options = useMemo(
    () => ({
      loop: true,
      autoplay: true,
      animationData: isDarkMode ? animationDataDark : animationDataLight,
      rendererSettings: { preserveAspectRatio: "xMidYMid slice" },
    }),
    [isDarkMode]
  );

  return <Lottie options={options} height={size} width={size} />;
}
