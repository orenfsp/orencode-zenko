import { LoginForm } from "@/features/auth/login";
import { AuthShell } from "@/widgets/auth-shell";
import { AuthGuard } from "@/features/auth/auth-guard";

export function LoginPage() {
  return (
    <AuthGuard requireAuth={false} redirectTo="/workspace"><AuthShell>
      <LoginForm />
    </AuthShell></AuthGuard>
  );
}
