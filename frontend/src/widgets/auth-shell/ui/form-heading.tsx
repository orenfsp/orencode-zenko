export function FormHeading({
  title,
  description,
}: {
  title: string;
  description: string;
}) {
  return (
    <header className="space-y-2 text-center">
      <h1 className="font-mono text-2xl font-medium leading-8 tracking-normal">
        {title}
      </h1>
      <p className="text-sm leading-5 text-muted-foreground">{description}</p>
    </header>
  );
}
