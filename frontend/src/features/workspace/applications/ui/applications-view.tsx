"use client";

import { FormEvent, useCallback, useState } from "react";
import { Gavel, LoaderCircle, Medal, Plus } from "lucide-react";

import { judgeCategoryApi, type CurrentJudgeCategory, type JudgeCategoryApplication } from "@/entities/judge-category";
import { profileApi } from "@/entities/profile";
import { rankApi, type Rank } from "@/entities/rank";
import { rankApplicationApi, type RankApplication } from "@/entities/rank-application";
import { referenceDataApi, type Discipline } from "@/entities/reference-data";
import { useAuth } from "@/entities/session";
import { formatDate, SectionHeading, StatusBadge } from "@/features/workspace/common";
import { useAsyncData } from "@/shared/lib/use-async-data";
import { useNotifications } from "@/shared/lib/notifications";
import { Button } from "@/shared/ui/button";
import { Card } from "@/shared/ui/card";
import { DataEmpty, DataError, DataLoading } from "@/shared/ui/data-state";
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "@/shared/ui/dialog";
import { Input } from "@/shared/ui/input";
import { Label } from "@/shared/ui/label";

const selectClassName = "h-9 w-full rounded-[10px] border border-border bg-background px-2.5 text-sm outline-none focus:border-ring focus:ring-2 focus:ring-ring/25";

interface ApplicationData {
  mode: "personal" | "review";
  activeRole: string | null;
  rank: RankApplication[];
  judge: JudgeCategoryApplication[];
  ranks: Rank[];
  disciplines: Discipline[];
  currentCategory: CurrentJudgeCategory | null;
}

type ReviewTarget = { kind: "rank" | "judge"; id: number; status: "approved" | "rejected" };

export function ApplicationsView() {
  const { identity } = useAuth();
  const notify = useNotifications();
  const permissions = identity?.permissions ?? [];
  const canReviewRank = permissions?.includes("rank_applications.review") ?? false;
  const canReviewJudge = permissions?.includes("judge_categories.review") ?? false;
  const [createOpen, setCreateOpen] = useState(false);
  const [reviewTarget, setReviewTarget] = useState<ReviewTarget | null>(null);
  const [pending, setPending] = useState(false);

  const load = useCallback(async (): Promise<ApplicationData> => {
    const profile = await profileApi.me();
    const activeRole = profile.roles.find((role) => role.id === identity?.roleId)?.code ?? null;
    const isFederationRole = activeRole === "federation" || activeRole === "super_admin";
    if (canReviewRank || canReviewJudge || isFederationRole) {
      const [rank, judge] = await Promise.all([
        canReviewRank || isFederationRole ? rankApplicationApi.pending({ limit: 100 }) : Promise.resolve([]),
        canReviewJudge || isFederationRole ? judgeCategoryApi.pending({ limit: 100 }) : Promise.resolve([]),
      ]);
      return { mode: "review", activeRole, rank, judge, ranks: [], disciplines: [], currentCategory: null };
    }
    const [ranks, disciplines] = await Promise.all([rankApi.list(), referenceDataApi.disciplines()]);
    if (activeRole === "judge") {
      const [judge, currentCategory] = await Promise.all([judgeCategoryApi.mine(), judgeCategoryApi.current()]);
      return { mode: "personal", activeRole, rank: [], judge, ranks, disciplines, currentCategory };
    }
    return { mode: "personal", activeRole, rank: await rankApplicationApi.mine(), judge: [], ranks, disciplines, currentCategory: null };
  }, [canReviewJudge, canReviewRank, identity?.roleId]);

  const state = useAsyncData(load);
  const total = (state.data?.rank.length ?? 0) + (state.data?.judge.length ?? 0);

  async function submitApplication(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!state.data) return;
    const form = new FormData(event.currentTarget);
    setPending(true);
    try {
      if (state.data.activeRole === "judge") {
        await judgeCategoryApi.create({
          category_id: Number(form.get("category_id")),
          application_type: String(form.get("application_type")),
        });
      } else {
        const athleteId = Number(form.get("athlete_user_id"));
        await rankApplicationApi.create({
          athlete_user_id: athleteId > 0 ? athleteId : undefined,
          rank_id: Number(form.get("rank_id")),
          discipline_id: Number(form.get("discipline_id")),
          application_type: String(form.get("application_type")),
          competition_id: Number(form.get("competition_id")),
          result_id: Number(form.get("result_id")),
        });
      }
      notify.success("Заявка подана");
      setCreateOpen(false);
      state.reload();
    } catch (cause) {
      notify.error(cause instanceof Error ? cause.message : "Не удалось подать заявку");
    } finally {
      setPending(false);
    }
  }

  async function submitReview(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!reviewTarget) return;
    const form = new FormData(event.currentTarget);
    const assignedAt = String(form.get("assigned_at") ?? "");
    const expiresAt = String(form.get("expires_at") ?? "");
    const orderNumber = String(form.get("order_number") ?? "").trim();
    const rejectionReason = String(form.get("rejection_reason") ?? "").trim();
    const body = reviewTarget.status === "approved"
      ? { status: "approved" as const, assigned_at: assignedAt, expires_at: expiresAt || undefined, order_number: orderNumber || undefined }
      : { status: "rejected" as const, rejection_reason: rejectionReason };
    setPending(true);
    try {
      if (reviewTarget.kind === "rank") await rankApplicationApi.review(reviewTarget.id, body);
      else await judgeCategoryApi.review(reviewTarget.id, body);
      notify.success(reviewTarget.status === "approved" ? "Заявка одобрена" : "Заявка отклонена");
      setReviewTarget(null);
      state.reload();
    } catch (cause) {
      notify.error(cause instanceof Error ? cause.message : "Не удалось рассмотреть заявку");
    } finally {
      setPending(false);
    }
  }

  const canCreate = state.data?.mode === "personal";

  return (
    <div className="mx-auto w-full max-w-6xl">
      <SectionHeading
        title={state.data?.mode === "review" ? "Заявки на рассмотрение" : "Мои заявки"}
        description={state.data?.mode === "review" ? "Очередь заявок, доступная активной роли." : "Статусы заявок на разряд или судейскую категорию."}
        action={canCreate ? <Button onClick={() => setCreateOpen(true)}><Plus />Подать заявку</Button> : undefined}
      />
      {state.data?.currentCategory ? (
        <Card className="mb-4 bg-[#eeebff] p-4 shadow-none"><p className="font-medium">Текущая категория: {state.data.currentCategory.category_name}</p><p className="mt-1 text-xs text-muted-foreground">Назначена {formatDate(state.data.currentCategory.assigned_at)}{state.data.currentCategory.expires_at ? ` · до ${formatDate(state.data.currentCategory.expires_at)}` : ""}</p></Card>
      ) : null}
      {state.loading ? <DataLoading rows={4} /> : null}
      {state.error ? <DataError message={state.error} retry={state.reload} /> : null}
      {!state.loading && !state.error && !total ? <DataEmpty title="Заявок пока нет" description="Поданные заявки и решения появятся на этой странице." /> : null}

      <div className="grid gap-3">
        {state.data?.rank.map((application) => (
          <Card key={`rank-${application.id}`} className="bg-white/75 p-5 shadow-none">
            <div className="flex flex-col justify-between gap-4 sm:flex-row sm:items-start">
              <div className="flex min-w-0 gap-3"><span className="flex size-10 shrink-0 items-center justify-center rounded-xl bg-[#eeebff] text-primary"><Medal className="size-5" /></span><div><p className="font-medium">Заявка на спортивный разряд</p><p className="mt-1 text-sm text-muted-foreground">Спортсмен #{application.athlete_user_id} · разряд #{application.rank_id} · дисциплина #{application.discipline_id}</p><p className="mt-1 text-xs text-muted-foreground">Подана {formatDate(application.created_at, true)} · {application.application_type}</p>{application.rejection_reason ? <p className="mt-2 text-sm text-destructive">{application.rejection_reason}</p> : null}</div></div>
              <div className="flex shrink-0 items-center gap-2"><StatusBadge status={application.status} />{state.data?.mode === "review" ? <><Button size="sm" onClick={() => setReviewTarget({ kind: "rank", id: application.id, status: "approved" })}>Одобрить</Button><Button size="sm" variant="outline" onClick={() => setReviewTarget({ kind: "rank", id: application.id, status: "rejected" })}>Отклонить</Button></> : null}</div>
            </div>
          </Card>
        ))}

        {state.data?.judge.map((application) => (
          <Card key={`judge-${application.id}`} className="bg-white/75 p-5 shadow-none">
            <div className="flex flex-col justify-between gap-4 sm:flex-row sm:items-start">
              <div className="flex min-w-0 gap-3"><span className="flex size-10 shrink-0 items-center justify-center rounded-xl bg-[#eeebff] text-primary"><Gavel className="size-5" /></span><div><p className="font-medium">{application.category_name ?? `Судейская категория #${application.category_id}`}</p><p className="mt-1 text-sm text-muted-foreground">{application.judge_name ? `${application.judge_name} · ` : ""}{application.application_type}</p><p className="mt-1 text-xs text-muted-foreground">Подана {formatDate(application.submitted_at, true)}</p>{application.rejection_reason ? <p className="mt-2 text-sm text-destructive">{application.rejection_reason}</p> : null}</div></div>
              <div className="flex shrink-0 items-center gap-2"><StatusBadge status={application.status} />{state.data?.mode === "review" ? <><Button size="sm" onClick={() => setReviewTarget({ kind: "judge", id: application.id, status: "approved" })}>Одобрить</Button><Button size="sm" variant="outline" onClick={() => setReviewTarget({ kind: "judge", id: application.id, status: "rejected" })}>Отклонить</Button></> : null}</div>
            </div>
          </Card>
        ))}
      </div>

      <Dialog open={createOpen} onOpenChange={setCreateOpen}>
        <DialogContent><DialogHeader><DialogTitle>Новая заявка</DialogTitle><DialogDescription>{state.data?.activeRole === "judge" ? "Укажите категорию и основание." : "Укажите разряд, дисциплину и результат-основание."}</DialogDescription></DialogHeader>
          <form className="grid gap-3" onSubmit={submitApplication}>
            {state.data?.activeRole === "judge" ? <><Label htmlFor="category-id">ID судейской категории</Label><Input id="category-id" name="category_id" type="number" min="1" required /></> : <>
              <Label htmlFor="rank-id">Разряд</Label><select id="rank-id" name="rank_id" className={selectClassName} required defaultValue=""><option value="" disabled>Выберите разряд</option>{state.data?.ranks.map((rank) => <option key={rank.id} value={rank.id}>{rank.name}</option>)}</select>
              <Label htmlFor="discipline-id">Дисциплина</Label><select id="discipline-id" name="discipline_id" className={selectClassName} required defaultValue=""><option value="" disabled>Выберите дисциплину</option>{state.data?.disciplines.map((item) => <option key={item.id} value={item.id}>{item.name}</option>)}</select>
              {state.data?.activeRole === "coach" ? <><Label htmlFor="athlete-id">ID спортсмена</Label><Input id="athlete-id" name="athlete_user_id" type="number" min="1" required /></> : null}
              <div className="grid gap-3 sm:grid-cols-2"><div><Label htmlFor="competition-id">ID соревнования</Label><Input id="competition-id" name="competition_id" type="number" min="1" required /></div><div><Label htmlFor="result-id">ID результата</Label><Input id="result-id" name="result_id" type="number" min="1" required /></div></div>
            </>}
            <Label htmlFor="application-type">Тип заявки</Label><select id="application-type" name="application_type" className={selectClassName} defaultValue="assignment"><option value="assignment">Присвоение</option><option value="confirmation">Подтверждение</option></select>
            <DialogFooter><Button type="submit" disabled={pending}>{pending ? <LoaderCircle className="animate-spin" /> : null}Подать</Button></DialogFooter>
          </form>
        </DialogContent>
      </Dialog>

      <Dialog open={Boolean(reviewTarget)} onOpenChange={(open) => !open && setReviewTarget(null)}>
        <DialogContent><DialogHeader><DialogTitle>{reviewTarget?.status === "approved" ? "Одобрить заявку" : "Отклонить заявку"}</DialogTitle><DialogDescription>Решение будет сохранено и отправлено заявителю.</DialogDescription></DialogHeader>
          <form className="grid gap-3" onSubmit={submitReview}>
            {reviewTarget?.status === "approved" ? <><Label htmlFor="assigned-at">Дата присвоения</Label><Input id="assigned-at" name="assigned_at" type="date" defaultValue={new Date().toISOString().slice(0, 10)} required /><Label htmlFor="expires-at">Действует до</Label><Input id="expires-at" name="expires_at" type="date" /><Label htmlFor="order-number">Номер приказа</Label><Input id="order-number" name="order_number" /></> : <><Label htmlFor="rejection-reason">Причина отклонения</Label><Input id="rejection-reason" name="rejection_reason" required /></>}
            <DialogFooter><Button type="submit" variant={reviewTarget?.status === "rejected" ? "destructive" : "default"} disabled={pending}>{pending ? <LoaderCircle className="animate-spin" /> : null}Сохранить решение</Button></DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
    </div>
  );
}
