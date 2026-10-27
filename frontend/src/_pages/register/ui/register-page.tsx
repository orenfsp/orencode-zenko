import { RegisterForm } from "@/features/auth/register";
import { AuthShell } from "@/widgets/auth-shell";
import { AuthGuard } from "@/features/auth/auth-guard";

export function RegisterPage() {
  return (
    <AuthGuard requireAuth={false} redirectTo="/workspace"><AuthShell>
      <RegisterForm />
    </AuthShell></AuthGuard>
  );
}
