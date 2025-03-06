import { Button } from "src/components/ui/button";
import { useStepper } from "src/components/ui/Stepper";

type StepButtonProps = {
  onNextClick?: () => void;
};

export default function StepButtons({ onNextClick }: StepButtonProps) {
  const { nextStep } = useStepper();
  return (
    <>
      <div className="w-full flex justify-start gap-2 mb-2">
        <>
          <Button
            onClick={() => {
              if (onNextClick) {
                onNextClick();
              }
              nextStep();
            }}
          >
            Next
          </Button>
        </>
      </div>
    </>
  );
}
