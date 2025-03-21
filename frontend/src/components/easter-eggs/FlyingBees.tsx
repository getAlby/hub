import * as React from "react";
import bee from "src/assets/images/easter-eggs/flying-bee.png";

type FlyingBee = { id: number; top: number };

export type FlyingBeesRef = {
  addFlyingBee: () => void;
};

export const FlyingBees = React.forwardRef<FlyingBeesRef>((_, ref) => {
  const [flyingBees, setFlyingBees] = React.useState<FlyingBee[]>([]);
  const idCounter = React.useRef(0);

  const addFlyingBee = React.useCallback(() => {
    const randomTop = Math.floor(Math.random() * window.innerHeight * 0.8);
    const newId = idCounter.current++;
    setFlyingBees((prev) => [...prev, { id: newId, top: randomTop }]);
  }, []);

  const handleAnimationEnd = (id: number) => {
    setFlyingBees((prev) => prev.filter((item) => item.id !== id));
  };

  React.useImperativeHandle(ref, () => ({
    addFlyingBee,
  }));

  return (
    <>
      {flyingBees.map((item) => (
        <img
          key={item.id}
          src={bee}
          alt="🐝"
          className="fixed left-0 w-32 pointer-events-none z-[9999] animate-easter-egg-fly-right"
          style={{ top: item.top }}
          onAnimationEnd={() => handleAnimationEnd(item.id)}
        />
      ))}
    </>
  );
});
