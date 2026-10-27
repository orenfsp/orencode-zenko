import { Separator } from "@/shared/ui/separator";

export function OAuthSeparator() {
  return (
    <span className="flex w-full items-center" aria-hidden="true">
      <Separator className="flex-1" />
      <span className="w-[138px] text-center text-xs leading-4 text-muted-foreground">
        ИЛИ
      </span>
      <Separator className="flex-1" />
    </span>
  );
}
