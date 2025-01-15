import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";
import { useMempoolApi } from "src/hooks/useMempoolApi";

export function BlockHeightWidget() {
  const { data: blocks } = useMempoolApi<{ height: number }[]>("/v1/blocks");

  if (!blocks?.length) {
    return null;
  }
  return (
    <Card>
      <CardHeader>
        <CardTitle>Block Height</CardTitle>
      </CardHeader>
      <CardContent className="mt-6">
        <p className="text-4xl font-semibold">{blocks[0].height}</p>
      </CardContent>
    </Card>
  );
}
