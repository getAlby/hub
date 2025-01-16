import { Button } from "src/components/ui/button";
import { useStepper } from "src/components/ui/Stepper";

export default function StepButtons() {
  const { nextStep } = useStepper();

  return (
    <div className="w-full flex mt-4 mb-4">
      <Button size="sm" onClick={nextStep}>
        Next
      </Button>
    </div>
  );
}
