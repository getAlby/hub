import { AppStoreApp } from "src/components/connections/SuggestedAppData";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";

export function AboutAppCard({ appStoreApp }: { appStoreApp: AppStoreApp }) {
  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-2xl">About the App</CardTitle>
      </CardHeader>
      <CardContent className="flex flex-col gap-3">
        <p className="text-muted-foreground">
          {appStoreApp.extendedDescription}
        </p>
      </CardContent>
    </Card>
  );
}
