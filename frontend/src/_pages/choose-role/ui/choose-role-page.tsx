import { RequireAuth } from "@/features/auth/require-auth";
import { RoleSelection } from "@/features/auth/select-role";

export function ChooseRolePage() {
  return (
    <RequireAuth>
      <RoleSelection />
    </RequireAuth>
  );
}
