import { ProfileView } from "@/features/profile/view";
import { BrandLogo } from "@/shared/ui/brand-logo";
import RibbonGlow from "@/widgets/auth-shell/ui/bg-effect";

export function PublicProfilePage({ userId }: { userId: number }) {
  return (
    <main className="relative min-h-svh min-w-0 overflow-hidden bg-workspace">
      <div className="fixed inset-0">
        <RibbonGlow />
      </div>
      <section className="relative z-10 min-h-svh bg-background/40 px-4 py-5 backdrop-blur-[20px] sm:px-6 sm:py-7 lg:px-10 lg:py-8">
        <header className="mx-auto mb-6 flex w-full max-w-6xl items-center justify-between sm:mb-10">
          <BrandLogo />
          <p className="text-xs text-muted-foreground">Публичный профиль</p>
        </header>
        <ProfileView mode="public" userId={userId} />
      </section>
    </main>
  );
}
