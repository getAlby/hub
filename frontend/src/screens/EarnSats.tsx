import { ExternalLinkIcon, StarIcon } from "lucide-react";
import AppHeader from "src/components/AppHeader";
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

interface ReviewOption {
  id: string;
  title: string;
  description: string;
  reward: string;
  platform: string;
  url: string;
  icon: React.ReactNode;
}

const reviewOptions: ReviewOption[] = [
  {
    id: "alby-go-play",
    title: "Alby Go on Google Play",
    description: "Write a review for Alby Go on the Google Play Store",
    reward: "1,000 sats",
    platform: "Google Play Store",
    url: "https://play.google.com/store/apps/details?id=com.getalby.mobile",
    icon: <StarIcon className="w-6 h-6 text-amber-500" />,
  },
  {
    id: "alby-go-apple",
    title: "Alby Go on App Store",
    description: "Write a review for Alby Go on the Apple App Store",
    reward: "1,000 sats",
    platform: "Apple App Store",
    url: "https://apps.apple.com/app/alby-go/id6476033149",
    icon: <StarIcon className="w-6 h-6 text-amber-500" />,
  },
  {
    id: "alby-extension-chrome",
    title: "Alby Extension on Chrome",
    description: "Rate the Alby Browser Extension on Chrome Web Store",
    reward: "1,000 sats",
    platform: "Chrome Web Store",
    url: "https://chrome.google.com/webstore/detail/alby-bitcoin-wallet-for-l/iokeahhehimjnekafflcihljlcjccdbe",
    icon: <StarIcon className="w-6 h-6 text-amber-500" />,
  },
  {
    id: "alby-extension-firefox",
    title: "Alby Extension on Firefox",
    description: "Rate the Alby Browser Extension on Firefox Add-ons",
    reward: "1,000 sats",
    platform: "Firefox Add-ons",
    url: "https://addons.mozilla.org/en-US/firefox/addon/alby/",
    icon: <StarIcon className="w-6 h-6 text-amber-500" />,
  },
  {
    id: "trustpilot",
    title: "Alby on Trustpilot",
    description: "Share your experience with Alby on Trustpilot",
    reward: "2,000 sats",
    platform: "Trustpilot",
    url: "https://www.trustpilot.com/review/getalby.com",
    icon: <StarIcon className="w-6 h-6 text-green-500" />,
  },
];

function EarnSats() {
  return (
    <>
      <AppHeader
        title="Earn Sats"
        description="Write reviews and earn satoshis for supporting Alby"
      />

      <div className="max-w-4xl mx-auto">
        {/* Introduction Card */}
        <Card className="mb-6">
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <StarIcon className="w-6 h-6 text-amber-500" />
              Support Alby & Earn Sats
            </CardTitle>
            <CardDescription>
              Help us grow by sharing your experience with Alby products. Write
              honest reviews on various platforms and earn satoshis as a thank
              you for your support! ⚡
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="bg-muted/50 p-4 rounded-lg">
              <p className="text-sm text-muted-foreground mb-2">
                <strong>How it works:</strong>
              </p>
              <ol className="text-sm text-muted-foreground space-y-1 ml-4">
                <li>1. Click on any review platform below</li>
                <li>2. Write an honest review about your Alby experience</li>
                <li>
                  3. Send a screenshot or link of your review to:{" "}
                  <strong>hello@getalby.com</strong>
                </li>
                <li>4. Receive your sats reward!</li>
              </ol>
            </div>
          </CardContent>
        </Card>

        {/* Review Options Grid */}
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {reviewOptions.map((option) => (
            <ExternalLink key={option.id} to={option.url}>
              <Card className="h-full hover:shadow-md transition-shadow">
                <CardHeader>
                  <div className="flex items-start justify-between">
                    <div className="flex items-center gap-3">
                      {option.icon}
                      <div>
                        <CardTitle className="text-lg">
                          {option.title}
                        </CardTitle>
                        <CardDescription className="text-sm text-muted-foreground">
                          {option.platform}
                        </CardDescription>
                      </div>
                    </div>
                    <Badge variant="secondary" className="shrink-0 font-mono">
                      {option.reward}
                    </Badge>
                  </div>
                </CardHeader>
                <CardContent>
                  <p className="text-sm text-muted-foreground mb-4">
                    {option.description}
                  </p>
                  <Button variant="outline" className="w-full">
                    Write Review
                    <ExternalLinkIcon className="w-4 h-4 ml-2" />
                  </Button>
                </CardContent>
              </Card>
            </ExternalLink>
          ))}
        </div>

        {/* Contact Information */}
        <Card className="mt-6">
          <CardHeader>
            <CardTitle>Ready to claim your sats?</CardTitle>
            <CardDescription>
              After writing your review, send us the details to receive your
              reward.
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="bg-primary/5 border border-primary/20 p-4 rounded-lg">
              <p className="text-sm mb-2">
                <strong>Send your review details to:</strong>
              </p>
              <p className="font-mono text-primary">hello@getalby.com</p>
              <p className="text-sm text-muted-foreground mt-2">
                Include a screenshot or direct link to your review, and we'll
                send your sats reward directly to your Alby Hub!
              </p>
            </div>
          </CardContent>
        </Card>
      </div>
    </>
  );
}

export default EarnSats;
