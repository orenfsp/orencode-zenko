import { AlertCircle, Inbox, RefreshCw } from "lucide-react";

import { Alert, AlertDescription, AlertTitle } from "@/shared/ui/alert";
import { Button } from "@/shared/ui/button";
import { Card } from "@/shared/ui/card";
import { Skeleton } from "@/shared/ui/skeleton";

export function DataLoading({ rows = 3 }: { rows?: number }) {
  return (
    <div className="grid gap-3" aria-label="Загрузка данных">
      {Array.from({ length: rows }, (_, index) => (
        <Skeleton key={index} className="h-24 w-full" />
      ))}
    </div>
  );
}

export function DataError({ message, retry }: { message: string; retry: () => void }) {
  return (
    <Alert variant="destructive" className="bg-white/80 shadow-none">
      <AlertCircle />
      <AlertTitle>Не удалось загрузить данные</AlertTitle>
      <AlertDescription>
        <p>{message}</p>
        <Button type="button" variant="outline" size="sm" onClick={retry}>
          <RefreshCw />
          Повторить
        </Button>
      </AlertDescription>
    </Alert>
  );
}

export function DataEmpty({
  title = "Здесь пока ничего нет",
  description = "Данные появятся после первой операции.",
}: {
  title?: string;
  description?: string;
}) {
  return (
    <Card className="items-center gap-2 bg-white/70 px-6 py-12 text-center shadow-none">
      <span className="flex size-11 items-center justify-center rounded-xl bg-muted text-muted-foreground">
        <Inbox className="size-5" />
      </span>
      <h2 className="font-medium">{title}</h2>
      <p className="max-w-md text-sm text-muted-foreground">{description}</p>
    </Card>
  );
}
