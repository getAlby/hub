import {
  Check,
  CreditCard,
  ExternalLink as ExternalLinkIcon,
  Globe,
  Layers,
  Lock,
  ShieldCheck,
  Smartphone,
  Sparkles,
  Zap,
} from "lucide-react";
import { useState } from "react";
import ExternalLink from "src/components/ExternalLink";
import { Badge } from "src/components/ui/badge";
import { Button } from "src/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";
import UserAvatar from "src/components/UserAvatar";
import { useAlbyMe } from "src/hooks/useAlbyMe";
import { cn } from "src/lib/utils";

export function CardPage() {
  const { data: albyMe } = useAlbyMe();
  const [rotateX, setRotateX] = useState(0);
  const [rotateY, setRotateY] = useState(0);
  const [isHovering, setIsHovering] = useState(false);

  const handleMouseMove = (e: React.MouseEvent<HTMLDivElement>) => {
    const card = e.currentTarget;
    const rect = card.getBoundingClientRect();
    const x = e.clientX - rect.left;
    const y = e.clientY - rect.top;
    const centerX = rect.width / 2;
    const centerY = rect.height / 2;
    const rotateXValue = ((y - centerY) / centerY) * -15;
    const rotateYValue = ((x - centerX) / centerX) * 15;

    setRotateX(rotateXValue);
    setRotateY(rotateYValue);
    setIsHovering(true);
  };

  const handleMouseLeave = () => {
    setRotateX(0);
    setRotateY(0);
    setIsHovering(false);
  };

  const features = [
    {
      icon: ShieldCheck,
      title: "100% Anonymous — No KYC",
      description: "No identity verification. No documents. No personal data.",
    },
    {
      icon: Zap,
      title: "Bitcoin and Crypto Payments",
      description: "Top up instantly using Bitcoin — fast, simple, and secure.",
    },
    {
      icon: Smartphone,
      title: "Apple Pay & Google Pay Support",
      description: "Add to Apple Pay or Google Pay and pay anywhere.",
    },
    {
      icon: Globe,
      title: "Use It Anywhere",
      description: "Pay from anywhere in the world — no restrictions.",
    },
    {
      icon: Layers,
      title: "Unlimited Cards",
      description: "Order as many cards as you need, anytime.",
    },
    {
      icon: Lock,
      title: "Instant & Secure Conversion",
      description: "Real-time crypto to fiat conversion directly on your card.",
    },
  ];

  const pricingTiers = [
    {
      name: "Regular",
      price: "$50",
      period: "USD",
      popular: false,
      features: [
        "3D Secure Enabled",
        "$20 Starting Balance",
        "Reloadable",
        "Standard top-up fee",
      ],
    },
    {
      name: "Wave",
      price: "$50",
      period: "USD",
      popular: true,
      features: [
        "Apple & Google Pay Compatible",
        "$20 Starting Balance",
        "Reloadable",
        "6.8% Top-Up Fee",
      ],
    },
  ];

  return (
    <div className="w-full space-y-12 pb-8">
      {/* Hero Section */}
      <div className="relative overflow-hidden rounded-3xl bg-gradient-to-br from-amber-50 via-orange-50 to-yellow-50 dark:from-amber-950/20 dark:via-orange-950/20 dark:to-yellow-950/20 p-8 md:p-12">
        <div className="absolute inset-0 bg-grid-slate-900/[0.04] dark:bg-grid-slate-100/[0.03] [mask-image:linear-gradient(0deg,white,rgba(255,255,255,0.5))] dark:[mask-image:linear-gradient(0deg,rgba(255,255,255,0.1),rgba(255,255,255,0.05))]" />

        <div className="relative grid md:grid-cols-2 gap-8 items-center">
          {/* Left Column - Text Content */}
          <div className="space-y-6 z-10">
            <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-amber-100 dark:bg-amber-900/30 border border-amber-200 dark:border-amber-800">
              <Sparkles className="h-4 w-4 text-amber-600 dark:text-amber-400" />
              <span className="text-sm font-medium text-amber-900 dark:text-amber-100">
                Powered by 2fiat
              </span>
            </div>

            <div className="space-y-3">
              <h1 className="text-4xl md:text-5xl font-bold tracking-tight bg-gradient-to-br from-amber-600 via-orange-600 to-yellow-600 dark:from-amber-400 dark:via-orange-400 dark:to-yellow-400 bg-clip-text text-transparent">
                Virtual Prepaid Mastercard
              </h1>
              <p className="text-lg text-muted-foreground max-w-lg">
                Anonymous payments. Instant crypto top-ups. Apple & Google Pay
                ready.
              </p>
              <p className="text-base text-muted-foreground max-w-lg">
                Instantly convert Bitcoin, Monero, USDT or any crypto to fiat
                and make purchases online or in store — no identity verification
                required.
              </p>
            </div>

            <div className="flex flex-wrap gap-3">
              <ExternalLink to="https://2fiat.com">
                <Button variant="premium" size="lg" className="gap-2 shadow-xl">
                  <CreditCard className="h-5 w-5" />
                  Order Your Card Now
                  <ExternalLinkIcon className="h-4 w-4" />
                </Button>
              </ExternalLink>
              <ExternalLink to="https://2fiat.com">
                <Button variant="outline" size="lg" className="gap-2">
                  Learn More
                  <ExternalLinkIcon className="h-4 w-4" />
                </Button>
              </ExternalLink>
            </div>

            <div className="flex flex-wrap gap-2 pt-2">
              <Badge variant="outline" className="gap-1">
                <Check className="h-3 w-3" />
                No KYC
              </Badge>
              <Badge variant="outline" className="gap-1">
                <Check className="h-3 w-3" />
                Pay with Bitcoin or Crypto Instantly
              </Badge>
              <Badge variant="outline" className="gap-1">
                <Check className="h-3 w-3" />
                Use with Apple Pay & Google Pay
              </Badge>
            </div>
          </div>

          {/* Right Column - 3D Card */}
          <div className="flex justify-center items-center perspective-1000">
            <div
              className="relative w-full max-w-md transition-all duration-500 ease-out cursor-pointer"
              style={{
                transform: `perspective(1200px) rotateX(${rotateX}deg) rotateY(${rotateY}deg) ${
                  isHovering ? "scale(1.05)" : "scale(1)"
                }`,
                transformStyle: "preserve-3d",
              }}
              onMouseMove={handleMouseMove}
              onMouseLeave={handleMouseLeave}
            >
              <div
                className="relative aspect-[1.586/1] rounded-3xl p-8 shadow-2xl"
                style={{
                  background:
                    "linear-gradient(135deg, rgb(245, 158, 11) 0%, rgb(251, 191, 36) 50%, rgb(252, 211, 77) 100%)",
                  boxShadow: isHovering
                    ? "0 30px 80px -15px rgba(245, 158, 11, 0.5), 0 0 0 1px rgba(245, 158, 11, 0.2)"
                    : "0 20px 60px -10px rgba(245, 158, 11, 0.4), 0 0 0 1px rgba(245, 158, 11, 0.1)",
                }}
              >
                {/* Card Content */}
                <div className="h-full flex flex-col justify-between text-black relative z-10">
                  {/* Top Section */}
                  <div className="flex justify-between items-start">
                    <div className="flex items-center gap-3">
                      <UserAvatar className="h-14 w-14 ring-2 ring-black/10" />
                      <div className="text-left">
                        <div className="font-bold text-base">
                          {albyMe?.name || "Alby User"}
                        </div>
                        <div className="text-xs opacity-70 font-medium">
                          {albyMe?.lightning_address || "user@getalby.com"}
                        </div>
                      </div>
                    </div>
                  </div>

                  {/* Middle Section - Card Number */}
                  <div className="space-y-2">
                    <div className="font-mono text-xl tracking-[0.3em] font-semibold">
                      •••• •••• •••• ••••
                    </div>
                  </div>

                  {/* Bottom Section */}
                  <div className="flex justify-between items-end">
                    <div className="space-y-1">
                      <div className="text-[10px] font-semibold opacity-60 tracking-wider">
                        VALID THRU
                      </div>
                      <div className="font-mono text-base font-bold">••/••</div>
                    </div>
                    <div className="flex items-end">
                      <img
                        src="/images/2fiat.png"
                        alt="2fiat"
                        className="h-14"
                      />
                    </div>
                  </div>
                </div>

                {/* Holographic Shine Effect */}
                <div
                  className="absolute inset-0 rounded-3xl opacity-0 hover:opacity-100 transition-opacity duration-500 pointer-events-none"
                  style={{
                    background:
                      "linear-gradient(135deg, transparent 0%, rgba(255, 255, 255, 0.2) 45%, rgba(255, 255, 255, 0.4) 50%, rgba(255, 255, 255, 0.2) 55%, transparent 100%)",
                    transform: `translateX(${rotateY * 2}px) translateY(${-rotateX * 2}px)`,
                  }}
                />
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* What We Offer Section */}
      <div className="space-y-4">
        <h2 className="text-xl font-bold tracking-tight text-center">
          Why Choose 2fiat
        </h2>

        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3">
          {features.map((feature, index) => {
            const Icon = feature.icon;
            return (
              <div
                key={index}
                className="flex gap-3 p-3 rounded-lg border bg-card/50 h-full"
              >
                <div className="shrink-0">
                  <div className="w-9 h-9 rounded-lg bg-amber-500/10 dark:bg-amber-400/10 flex items-center justify-center">
                    <Icon className="h-4 w-4 text-amber-600 dark:text-amber-400" />
                  </div>
                </div>
                <div className="space-y-1 min-w-0">
                  <div className="font-semibold text-sm leading-tight">
                    {feature.title}
                  </div>
                  <p className="text-xs text-muted-foreground line-clamp-1">
                    {feature.description}
                  </p>
                </div>
              </div>
            );
          })}
        </div>
      </div>

      {/* Pricing Section */}
      <div className="space-y-6">
        <div className="text-center space-y-2">
          <h2 className="text-3xl font-bold tracking-tight">Our Products</h2>
          <p className="text-muted-foreground">
            Choose the card that fits your needs
          </p>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-6 max-w-4xl mx-auto">
          {pricingTiers.map((tier, index) => (
            <Card
              key={index}
              className={cn(
                "relative overflow-hidden transition-all duration-300 hover:shadow-xl hover:-translate-y-1",
                tier.popular
                  ? "border-amber-500 dark:border-amber-400 border-2 shadow-lg"
                  : "border-2"
              )}
            >
              {tier.popular && (
                <div className="absolute top-0 right-0">
                  <Badge
                    variant="default"
                    className="rounded-none rounded-bl-lg bg-gradient-to-r from-amber-500 to-amber-400 text-black font-bold px-4 py-1"
                  >
                    <Sparkles className="h-3 w-3 mr-1" />
                    Popular
                  </Badge>
                </div>
              )}

              <CardHeader className="space-y-4 pb-8">
                <div className="space-y-2">
                  <CardTitle className="text-2xl">{tier.name}</CardTitle>
                  <div className="flex items-baseline gap-2">
                    <span className="text-5xl font-bold tracking-tight">
                      {tier.price}
                    </span>
                    <span className="text-muted-foreground text-lg">
                      {tier.period}
                    </span>
                  </div>
                </div>
              </CardHeader>

              <CardContent className="space-y-6">
                <ul className="space-y-3">
                  {tier.features.map((feature, featureIndex) => (
                    <li key={featureIndex} className="flex items-start gap-3">
                      <div className="mt-0.5 rounded-full p-1">
                        <Check className="h-4 w-4 text-positive-foreground" />
                      </div>
                      <span className="text-sm leading-relaxed">{feature}</span>
                    </li>
                  ))}
                </ul>

                <ExternalLink to="https://2fiat.com">
                  <Button
                    variant={tier.popular ? "premium" : "outline"}
                    className="w-full"
                    size="lg"
                  >
                    Order Now
                    <ExternalLinkIcon className="h-4 w-4" />
                  </Button>
                </ExternalLink>
              </CardContent>
            </Card>
          ))}
        </div>
      </div>

      {/* About Section */}
      <Card className="border-2">
        <CardHeader>
          <div className="flex items-center gap-3">
            <div className="w-12 h-12 rounded-xl bg-gradient-to-br from-amber-500 to-orange-500 flex items-center justify-center shadow-lg">
              <CreditCard className="h-6 w-6 text-white" />
            </div>
            <div>
              <CardTitle className="text-2xl">About 2fiat</CardTitle>
              <CardDescription className="text-base mt-1">
                Your trusted anonymous crypto payment solution
              </CardDescription>
            </div>
          </div>
        </CardHeader>
        <CardContent className="space-y-6">
          <p className="text-muted-foreground leading-relaxed">
            Our virtual prepaid Mastercard cards let you instantly convert
            Bitcoin, Monero, USDT or any crypto to fiat and make purchases
            online or in store — no identity verification required. Experience
            total privacy, instant access, and worldwide freedom with 2fiat.
          </p>
          <div className="flex flex-wrap gap-3">
            <ExternalLink to="https://2fiat.com">
              <Button variant="default" size="lg">
                Visit 2fiat.com
                <ExternalLinkIcon className="h-4 w-4" />
              </Button>
            </ExternalLink>
            <ExternalLink to="mailto:support@2fiat.com">
              <Button variant="outline" size="lg">
                Contact Support
                <ExternalLinkIcon className="h-4 w-4" />
              </Button>
            </ExternalLink>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
