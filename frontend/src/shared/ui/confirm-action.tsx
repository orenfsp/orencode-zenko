"use client";

import { useState, type ReactNode } from "react";

import { Button } from "@/shared/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/shared/ui/dialog";

export function ConfirmAction({
  trigger,
  title,
  description,
  confirmLabel = "Подтвердить",
  pending = false,
  destructive = false,
  onConfirm,
}: {
  trigger: ReactNode;
  title: string;
  description: string;
  confirmLabel?: string;
  pending?: boolean;
  destructive?: boolean;
  onConfirm: () => void | Promise<void>;
}) {
  const [open, setOpen] = useState(false);

  return (
    <Dialog open={open} onOpenChange={(next) => { if (!pending) setOpen(next); }}>
      <DialogTrigger asChild>{trigger}</DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{title}</DialogTitle>
          <DialogDescription>{description}</DialogDescription>
        </DialogHeader>
        <DialogFooter>
          <Button type="button" variant="outline" disabled={pending} onClick={() => setOpen(false)}>Отмена</Button>
          <Button type="button" variant={destructive ? "destructive" : "default"} disabled={pending} onClick={async () => { await onConfirm(); setOpen(false); }}>{confirmLabel}</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
