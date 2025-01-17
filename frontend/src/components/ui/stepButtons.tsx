import { Button } from "src/components/ui/button";
import { useStepper } from "src/components/ui/Stepper";

export default function StepButtons() {
  const { nextStep, prevStep, isLastStep, isDisabledStep } = useStepper();
  return (
    <>
      <div className="w-full flex justify-start gap-2 mt-8 mb-2">
        <>
          {isLastStep ? (
            <Button
              disabled={isDisabledStep}
              onClick={prevStep}
              variant="secondary"
            >
              Back
            </Button>
          ) : (
            <Button onClick={nextStep}>Next</Button>
          )}
        </>
      </div>
    </>
  );
}
