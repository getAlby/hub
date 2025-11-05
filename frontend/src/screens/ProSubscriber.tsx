import { Sparkles } from "lucide-react";
import React, { useState } from "react";
import { AlbyIcon } from "src/components/icons/Alby";
import { useAlbyMe } from "src/hooks/useAlbyMe";
import { cn } from "src/lib/utils";

export function ProSubscriber() {
  const { data: albyMe } = useAlbyMe();
  const [rotation, setRotation] = useState({
    x: 0,
    y: 0,
    glareX: "50%",
    glareY: "50%",
  });
  const [isFlipped, setIsFlipped] = useState(false);

  // Handle mouse move for 3D effect
  const handleMouseMove = (e: React.MouseEvent<HTMLDivElement>) => {
    const card = e.currentTarget.getBoundingClientRect();
    const x = e.clientX - card.left;
    const y = e.clientY - card.top;

    // Calculate position as a percentage of the card dimensions
    const xPercent = x / card.width;
    const yPercent = y / card.height;

    // Calculate rotation values - max rotation of 10 degrees in each direction
    // Map x and y from 0-1 to -10 to 10 degrees
    const rotateY = (0.5 - xPercent) * 20; // Left is positive, right is negative
    const rotateX = (yPercent - 0.5) * 20; // Top is negative, bottom is positive

    // Add subtle movement to simulate light reflection
    const glareX = (xPercent * 100).toFixed(0) + "%";
    const glareY = (yPercent * 100).toFixed(0) + "%";

    // Update card rotation and glare position
    setRotation({
      x: rotateX,
      y: rotateY,
      glareX,
      glareY,
    });
  };

  // Reset rotation when mouse leaves
  const handleMouseLeave = () => {
    setRotation({ x: 0, y: 0, glareX: "50%", glareY: "50%" });
  };

  // Flip card with click effect
  const flipCard = () => {
    setIsFlipped(!isFlipped);
  };

  // No confetti effect

  return (
    <div className="container max-w-4xl mx-auto px-4 py-8">
      <div className="flex flex-col items-center justify-center mb-12">
        <div className="flex items-center gap-2 mb-2 animate-fade-in">
          <Sparkles className="h-6 w-6 text-amber-400" />
          <h1 className="text-3xl font-bold">Welcome Home</h1>
        </div>
        <p className="text-muted-foreground text-center max-w-md animate-fade-in-delayed">
          We're genuinely thrilled you're here with us. Your support makes
          everything we do possible, and we've crafted these benefits just for
          you.
        </p>
      </div>

      {/* Premium Credit Card - Centered and Prominent */}
      <div className="flex justify-center mb-16">
        <div
          className="perspective-1000 cursor-pointer w-full max-w-md animate-float p-8 card-click"
          onClick={(e) => {
            e.currentTarget.classList.remove("card-click");
            // Force a reflow
            void e.currentTarget.offsetWidth;
            e.currentTarget.classList.add("card-click");
            flipCard();
          }}
        >
          <div
            className="relative w-full transition-all duration-500 preserve-3d"
            style={{
              aspectRatio:
                "1.586/1" /* Standard credit card ratio (85.6mm × 53.98mm) */,
              transform: `rotateY(${isFlipped ? 180 : 0}deg)`,
              transformStyle: "preserve-3d",
              transformOrigin: "center center",
            }}
            onMouseMove={handleMouseMove}
            onMouseLeave={handleMouseLeave}
          >
            {/* Front of card */}
            <div
              className={cn(
                "absolute inset-0 rounded-xl p-6 backface-hidden",
                "border shadow-md shadow-amber-500/20",
                "flex flex-col justify-between text-black"
              )}
              style={{
                background: "linear-gradient(to right, #f59e0b, #fbbf24)",
                backgroundImage: `
                  linear-gradient(to right, #f59e0b, #fbbf24)
                `,
                border: "1px solid rgba(245, 158, 11, 0.3)",
                boxShadow:
                  "0 25px 50px -12px rgba(0, 0, 0, 0.5), inset 0 0 20px rgba(255,255,255,0.2)",
                transform: `rotateX(${rotation.x}deg) rotateY(${rotation.y}deg)`,
                backfaceVisibility: "hidden",
                transition: "transform 0.1s ease-out, box-shadow 0.2s ease-out",
                transformOrigin: "center center",
              }}
            >
              {/* Light reflection effect */}
              <div
                className="absolute inset-0 rounded-xl overflow-hidden pointer-events-none"
                style={{
                  background:
                    "linear-gradient(105deg, transparent 20%, rgba(255, 255, 255, 0.6) 40%, rgba(255, 255, 255, 0.4) 60%, transparent 80%)",
                  backgroundPosition: `${rotation.glareX} ${rotation.glareY}`,
                  backgroundSize: "200% 200%",
                  mixBlendMode: "soft-light",
                  opacity: 0.9,
                }}
              />

              {/* Alby logo with chip */}
              <div className="flex justify-between items-start">
                <div className="flex items-center">
                  <div className="h-12 w-12 rounded-full bg-gradient-to-r from-amber-400 to-amber-600 flex items-center justify-center shadow-lg">
                    <AlbyIcon className="h-8 w-8" />
                  </div>
                  <div className="ml-3">
                    <div className="text-xs font-mono text-muted-foreground">
                      ALBY
                    </div>
                    <div className="text-sm font-bold tracking-widest flex items-center">
                      PRO <Sparkles className="h-3 w-3 ml-1" />
                    </div>
                  </div>
                </div>
              </div>

              {/* Member ID */}
              <div className="mt-8">
                <div className="text-xs font-mono text-muted-foreground mb-1">
                  MEMBER ID
                </div>
                <div className="font-mono text-xl tracking-widest">
                  {albyMe?.nostr_pubkey?.slice(0, 4) || "••••"} •••• ••••{" "}
                  {albyMe?.identifier?.slice(-4) || "1234"}
                </div>
              </div>

              {/* Member name and since */}
              <div className="flex justify-between mt-6">
                <div>
                  <div className="text-xs font-mono text-muted-foreground">
                    MEMBER NAME
                  </div>
                  <div className="font-medium tracking-wide">
                    {albyMe?.name || "ALBY USER"}
                  </div>
                </div>
                <div>
                  <div className="text-xs font-mono text-muted-foreground">
                    MEMBER SINCE
                  </div>
                  <div className="font-medium tracking-wide">
                    {new Date().getFullYear()}
                  </div>
                </div>
              </div>
            </div>

            {/* Back of card */}
            <div
              className={cn(
                "absolute inset-0 rounded-xl p-6 backface-hidden",
                "border shadow-md shadow-amber-500/20",
                "flex flex-col justify-between text-black"
              )}
              style={{
                background: "linear-gradient(to right, #f59e0b, #d97706)",
                backgroundImage: `
                  linear-gradient(to right, #f59e0b, #d97706)
                `,
                border: "1px solid rgba(245, 158, 11, 0.3)",
                boxShadow:
                  "0 25px 50px -12px rgba(0, 0, 0, 0.5), inset 0 0 20px rgba(255,255,255,0.2)",
                transform: `rotateY(180deg) rotateX(${rotation.x}deg)`,
                backfaceVisibility: "hidden",
                transition: "transform 0.1s ease-out, box-shadow 0.2s ease-out",
                transformOrigin: "center center",
              }}
            >
              {/* Light reflection effect */}
              <div
                className="absolute inset-0 rounded-xl overflow-hidden pointer-events-none"
                style={{
                  background:
                    "linear-gradient(105deg, transparent 20%, rgba(255, 255, 255, 0.5) 40%, rgba(255, 255, 255, 0.3) 60%, transparent 80%)",
                  backgroundPosition: `${rotation.glareX} ${rotation.glareY}`,
                  backgroundSize: "200% 200%",
                  mixBlendMode: "soft-light",
                  opacity: 0.8,
                }}
              />

              <div className="flex flex-col gap-4">
                <div>
                  <div className="text-xs font-mono text-muted-foreground mb-1">
                    LIGHTNING ADDRESS
                  </div>
                  <div className="font-mono text-sm truncate">
                    {albyMe?.lightning_address || "user@getalby.com"}
                  </div>
                </div>

                <div>
                  <div className="text-xs font-mono text-muted-foreground mb-1">
                    EMAIL
                  </div>
                  <div className="font-mono text-sm truncate">
                    {albyMe?.email || "user@example.com"}
                  </div>
                </div>

                <div className="mt-auto pt-4">
                  <div className="flex justify-between items-center">
                    <div className="text-xs font-mono text-muted-foreground">
                      Give your card a tap to see more
                    </div>
                    <div className="flex items-center">
                      <Sparkles className="h-4 w-4 text-amber-400 mr-1" />
                      <span className="text-xs font-medium tracking-wider text-amber-400">
                        Pro <Sparkles className="inline h-3 w-3" />
                      </span>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* CSS for 3D card effect and animations */}
      <style>{`
        .perspective-1000 {
          perspective: 2000px;
        }
        .preserve-3d {
          transform-style: preserve-3d;
          position: relative;
          width: 100%;
          height: 100%;
        }
        .backface-hidden {
          backface-visibility: hidden;
          position: absolute;
          top: 0;
          left: 0;
          width: 100%;
          height: 100%;
        }
        .rotate-y-180 {
          transform: rotateY(180deg);
        }
        
        
        /* Card click effect */
        .card-click {
          animation: card-click 0.3s ease-out;
        }
        
        @keyframes card-click {
          0% { transform: scale(1); }
          50% { transform: scale(0.95); }
          100% { transform: scale(1); }
        }
        
        @keyframes float {
          0% { transform: translateY(0px); }
          50% { transform: translateY(-8px); }
          100% { transform: translateY(0px); }
        }
        
        .animate-float {
          animation: float 6s ease-in-out infinite;
        }
        
        .animate-fade-in {
          opacity: 0;
          animation: fade-in 0.5s ease-out forwards;
        }
        
        .animate-fade-in-delayed {
          opacity: 0;
          animation: fade-in 0.5s ease-out 0.3s forwards;
        }
        
        @keyframes fade-in {
          0% { opacity: 0; transform: translateY(20px); }
          100% { opacity: 1; transform: translateY(0); }
        }
      `}</style>
    </div>
  );
}

export default ProSubscriber;
