import { Input } from "@/shared/ui/input";
import { Label } from "@/shared/ui/label";

export function FormField({
  id,
  label,
  labelAddon,
  ...props
}: React.ComponentProps<typeof Input> & {
  id: string;
  label: string;
  labelAddon?: React.ReactNode;
}) {
  return (
    <fieldset className="space-y-2 border-0 p-0">
      <span className="flex items-center justify-between gap-2">
        <Label htmlFor={id}>{label}</Label>
        {labelAddon}
      </span>
      <Input id={id} {...props} />
    </fieldset>
  );
}
