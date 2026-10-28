import { PublicProfilePage } from "@/_pages/profile";

export default async function Page({ params }: { params: Promise<{ userId: string }> }) {
  const { userId } = await params;
  return <PublicProfilePage userId={Number(userId)} />;
}
