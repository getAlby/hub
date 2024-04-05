type Props = {
  title: string;
  description: string;
};

export default function TwoColumnLayoutHeader({ title, description }: Props) {
  return (
    <div className="grid gap-2 text-center">
      <h1 className="text-2xl font-semibold">{title}</h1>
      <p className="text-muted-foreground">{description}</p>
    </div>
  );
}
