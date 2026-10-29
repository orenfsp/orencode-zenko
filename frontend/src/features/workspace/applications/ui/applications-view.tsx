"use client";

import { FormEvent, useCallback, useState } from "react";
import { AlertTriangle, LoaderCircle } from "lucide-react";
import Link from "next/link";

import { competitionApplicationApi, type CompetitionApplication } from "@/entities/competition-application";
import { judgeCategoryApi, type CurrentJudgeCategory, type JudgeCategory, type JudgeCategoryApplication } from "@/entities/judge-category";
import { profileApi } from "@/entities/profile";
import { rankApi, type Rank } from "@/entities/rank";
import { rankApplicationApi, type RankApplication, type RankApplicationBasis } from "@/entities/rank-application";
import { referenceDataApi, type Discipline } from "@/entities/reference-data";
import { userCoachApi, type UserCoachRelation } from "@/entities/user-coach";
import { useAuth } from "@/entities/session";
import {
  expiryDescription,
  formatDate,
  isExpiringSoon,
  StatusBadge,
} from "@/features/workspace/common";
import { useAsyncData } from "@/shared/lib/use-async-data";
import { useNotifications } from "@/shared/lib/notifications";
import { Button } from "@/shared/ui/button";
import { Card } from "@/shared/ui/card";
import { DataEmpty, DataError, DataLoading } from "@/shared/ui/data-state";
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "@/shared/ui/dialog";
import { Input } from "@/shared/ui/input";
import { Label } from "@/shared/ui/label";
import { DatePicker } from "@/shared/ui/date-picker";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/shared/ui/select";
import { WorkspaceToolbar } from "@/shared/ui/workspace-toolbar";
import { ConfirmAction } from "@/shared/ui/confirm-action";

interface ApplicationData {
  mode: "personal" | "review";
  activeRole: string | null;
  rank: RankApplication[];
  judge: JudgeCategoryApplication[];
  ranks: Rank[];
  disciplines: Discipline[];
  judgeCategories: JudgeCategory[];
  currentCategory: CurrentJudgeCategory | null;
  athletes: UserCoachRelation[];
  competitionApplications: CompetitionApplication[];
  userNames: Record<number, string>;
}

type ReviewTarget = { kind: "rank" | "judge"; id: number; status: "approved" | "rejected" };

export function ApplicationsView() {
  const { identity } = useAuth();
  const notify = useNotifications();
  const permissions = identity?.permissions ?? [];
  const canReviewRank = permissions?.includes("rank_applications.review") ?? false;
  const canReviewJudge = permissions?.includes("judge_categories.review") ?? false;
  const canReviewCompetition = permissions?.includes("competition_applications.review") ?? false;
  const [createOpen, setCreateOpen] = useState(false);
  const [reviewTarget, setReviewTarget] = useState<ReviewTarget | null>(null);
  const [pending, setPending] = useState(false);
  const [search, setSearch] = useState("");
  const [selectedAthleteId, setSelectedAthleteId] = useState<number | null>(null);
  const [selectedDisciplineId, setSelectedDisciplineId] = useState<number | null>(null);
  const [selectedResultId, setSelectedResultId] = useState<number | null>(null);
  const [basisState, setBasisState] = useState<{ items: RankApplicationBasis[]; loading: boolean; error: string | null }>({ items: [], loading: false, error: null });

  const load = useCallback(async (): Promise<ApplicationData> => {
    const profile = await profileApi.me();
    const activeRole = profile.roles.find((role) => role.id === identity?.roleId)?.code ?? null;
    if (canReviewRank || canReviewJudge || canReviewCompetition) {
      const [rank, judge, competitionApplications, ranks, disciplines] = await Promise.all([
        canReviewRank ? rankApplicationApi.pending({ limit: 100 }) : Promise.resolve([]),
        canReviewJudge ? judgeCategoryApi.pending({ limit: 100 }) : Promise.resolve([]),
        canReviewCompetition ? competitionApplicationApi.pending({ status: "pending", limit: 100 }) : Promise.resolve([]),
        rankApi.list(),
        referenceDataApi.disciplines(),
      ]);
      const userIds = [...new Set(rank.map((item) => item.athlete_user_id))];
      const profiles = await Promise.allSettled(userIds.map((userId) => profileApi.public(userId)));
      const userNames = Object.fromEntries(profiles.flatMap((item, index) => item.status === "fulfilled" ? [[userIds[index], [item.value.last_name, item.value.first_name, item.value.middle_name].filter(Boolean).join(" ")]] : []));
      return { mode: "review", activeRole, rank, judge, ranks, disciplines, judgeCategories: [], currentCategory: null, athletes: [], competitionApplications, userNames };
    }
    const [ranks, disciplines, athletes, judgeCategories] = await Promise.all([
      rankApi.list(),
      referenceDataApi.disciplines(),
      activeRole === "coach" ? userCoachApi.byCoach({ limit: 100 }).then((items) => items.filter((item) => item.status === "confirmed")) : Promise.resolve([]),
      activeRole === "judge" ? judgeCategoryApi.categories() : Promise.resolve([]),
    ]);
    if (activeRole === "judge") {
      const [judge, currentCategory] = await Promise.all([judgeCategoryApi.mine(), judgeCategoryApi.current()]);
      return { mode: "personal", activeRole, rank: [], judge, ranks, disciplines, judgeCategories, currentCategory, athletes, competitionApplications: await competitionApplicationApi.mine(), userNames: {} };
    }
    const [rank, competitionApplications] = await Promise.all([rankApplicationApi.mine(), competitionApplicationApi.mine()]);
    const userIds = [...new Set(athletes.map((item) => item.participant_user_id))];
    const profiles = await Promise.allSettled(userIds.map((userId) => profileApi.public(userId)));
    const userNames = Object.fromEntries(profiles.flatMap((item, index) => item.status === "fulfilled" ? [[userIds[index], [item.value.last_name, item.value.first_name, item.value.middle_name].filter(Boolean).join(" ")]] : []));
    return { mode: "personal", activeRole, rank, judge: [], ranks, disciplines, judgeCategories, currentCategory: null, athletes, competitionApplications, userNames };
  }, [canReviewJudge, canReviewRank, canReviewCompetition, identity?.roleId]);

  const state = useAsyncData(load);
  const total = (state.data?.rank.length ?? 0) + (state.data?.judge.length ?? 0) + (state.data?.competitionApplications.length ?? 0);

  async function changeCompetitionApplication(application: CompetitionApplication, action: "approve" | "reject" | "withdraw") {
    setPending(true);
    try {
      if (action === "approve") await competitionApplicationApi.approve(application.id);
      else if (action === "reject") await competitionApplicationApi.reject(application.id);
      else await competitionApplicationApi.withdraw(application.id);
      notify.success(action === "approve" ? "Заявка одобрена" : action === "reject" ? "Заявка отклонена" : "Заявка отозвана");
      state.reload();
    } catch (cause) {
      notify.error(cause instanceof Error ? cause.message : "Не удалось изменить заявку");
    } finally { setPending(false); }
  }

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
        const basis = basisState.items.find((item) => item.result_id === selectedResultId);
        if (!basis || !selectedDisciplineId) {
          notify.error("Выберите подтверждённый результат-основание");
          return;
        }
        const athleteId = selectedAthleteId ?? 0;
        await rankApplicationApi.create({
          athlete_user_id: athleteId > 0 ? athleteId : undefined,
          rank_id: Number(form.get("rank_id")),
          discipline_id: selectedDisciplineId,
          application_type: String(form.get("application_type")),
          competition_id: basis.competition_id,
          result_id: basis.result_id,
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

  async function loadBases(disciplineId: number, athleteId: number | null) {
    if (state.data?.activeRole === "coach" && !athleteId) {
      setBasisState({ items: [], loading: false, error: null });
      return;
    }
    setSelectedResultId(null);
    setBasisState({ items: [], loading: true, error: null });
    try {
      const items = await rankApplicationApi.bases({
        discipline_id: disciplineId,
        athlete_user_id: state.data?.activeRole === "coach" ? athleteId ?? undefined : undefined,
      });
      setBasisState({ items, loading: false, error: null });
    } catch (cause) {
      setBasisState({ items: [], loading: false, error: cause instanceof Error ? cause.message : "Не удалось загрузить результаты" });
    }
  }

  function openCreate() {
    setSelectedAthleteId(null);
    setSelectedDisciplineId(null);
    setSelectedResultId(null);
    setBasisState({ items: [], loading: false, error: null });
    setCreateOpen(true);
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

  const canCreate = state.data?.mode === "personal" && (state.data.activeRole === "participant" || state.data.activeRole === "coach" || state.data.activeRole === "judge");
  const missingCoachAthletes = state.data?.activeRole === "coach" && state.data.athletes.length === 0;
  const query = search.trim().toLocaleLowerCase("ru");
  const matches = (value: string) => !query || value.toLocaleLowerCase("ru").includes(query);
  const visibleCompetitionApplications = (state.data?.competitionApplications ?? []).filter((item) => matches(`${item.competition_name} ${item.discipline_name} ${item.athlete_last_name} ${item.athlete_first_name} ${item.status}`));
  const visibleRankApplications = (state.data?.rank ?? []).filter((item) => matches(`${item.athlete_user_id} ${item.rank_id} ${item.discipline_id} ${item.status}`));
  const visibleJudgeApplications = (state.data?.judge ?? []).filter((item) => matches(`${item.judge_name ?? ""} ${item.category_name ?? item.category_id} ${item.status}`));
  const visibleTotal = visibleCompetitionApplications.length + visibleRankApplications.length + visibleJudgeApplications.length;
  const categoryExpiresSoon = isExpiringSoon(
    state.data?.currentCategory?.expires_at,
  );

  return (
    <div className="mx-auto w-full max-w-6xl">
      <WorkspaceToolbar search={search} onSearchChange={setSearch} searchPlaceholder="Найти заявку" createLabel="Подать заявку" onCreate={canCreate ? openCreate : undefined} />
      {state.data?.currentCategory ? (
        <Card className={`mb-4 p-4 shadow-none ${categoryExpiresSoon ? "border-amber-500/30 bg-amber-500/10" : "border-primary/10 bg-primary/5"}`}><div className="flex gap-3">{categoryExpiresSoon ? <AlertTriangle className="mt-0.5 size-4 shrink-0 text-amber-600" /> : null}<div><p className="font-medium">Текущая категория: {state.data.currentCategory.category_name}</p><p className="mt-1 text-xs text-muted-foreground">Назначена {formatDate(state.data.currentCategory.assigned_at)}{state.data.currentCategory.expires_at ? ` · до ${formatDate(state.data.currentCategory.expires_at)}` : ""}</p>{categoryExpiresSoon ? <p className="mt-1 text-sm text-amber-700 dark:text-amber-300">{expiryDescription(state.data.currentCategory.expires_at)}. Подайте заявку на подтверждение заранее.</p> : null}</div></div></Card>
      ) : null}
      {state.loading ? <DataLoading rows={4} /> : null}
      {state.error ? <DataError message={state.error} retry={state.reload} /> : null}
      {!state.loading && !state.error && !total ? <DataEmpty title={state.data?.mode === "review" ? "Очередь пуста" : "Заявок пока нет"} description={state.data?.mode === "review" ? "Новые заявки появятся здесь после подачи." : missingCoachAthletes ? "Подать заявку за спортсмена можно после подтверждения связи с ним." : "Подайте заявку, чтобы отслеживать её статус здесь."} /> : null}
      {!state.loading && !state.error && total > 0 && !visibleTotal ? <DataEmpty title="Заявки не найдены" description="Измените поисковый запрос." action={<Button type="button" variant="outline" size="sm" onClick={() => setSearch("")}>Сбросить поиск</Button>} /> : null}
      {missingCoachAthletes ? <Button asChild variant="outline" className="mb-4"><Link href="/workspace/coaches">Добавить спортсмена</Link></Button> : null}
      {!state.loading && !state.error && !total && state.data?.mode === "personal" ? <Button asChild variant="outline" className="mb-4"><Link href="/workspace/competitions">Найти соревнование</Link></Button> : null}

      <div className="grid gap-3">
        {visibleCompetitionApplications.map((application) => (
          <Card key={`competition-${application.id}`} variant="glass" className="p-5">
            <div className="flex flex-col justify-between gap-4 sm:flex-row sm:items-start">
              <div className="min-w-0"><div className="flex flex-wrap items-center gap-2"><p className="font-medium">Заявка на соревнование</p><StatusBadge status={application.status} /></div><p className="mt-1 text-sm">{application.competition_name} · {application.discipline_name}{application.age_group_name ? ` · ${application.age_group_name}` : ""}</p><p className="mt-1 text-sm text-muted-foreground">Спортсмен: {application.athlete_last_name} {application.athlete_first_name} · подана {formatDate(application.submitted_at, true)}</p><Button asChild variant="link" size="sm" className="mt-1 px-0"><Link href={`/workspace/competitions/${application.competition_id}`}>Открыть соревнование</Link></Button></div>
              {state.data?.mode === "review" ? <div className="flex w-full gap-2 sm:w-auto"><Button className="flex-1 sm:flex-none" size="sm" disabled={pending} onClick={() => void changeCompetitionApplication(application, "approve")}>Одобрить</Button><Button className="flex-1 sm:flex-none" size="sm" variant="outline" disabled={pending} onClick={() => void changeCompetitionApplication(application, "reject")}>Отклонить</Button></div> : application.can_withdraw ? <div className="flex w-full sm:w-auto"><ConfirmAction trigger={<Button className="w-full sm:w-auto" size="sm" variant="outline" disabled={pending}>Отозвать</Button>} title="Отозвать заявку?" description="Заявка перестанет участвовать в модерации. Это действие нельзя отменить." confirmLabel="Отозвать заявку" destructive onConfirm={() => changeCompetitionApplication(application, "withdraw")} /></div> : null}
            </div>
          </Card>
        ))}
        {visibleRankApplications.map((application) => (
          <Card key={`rank-${application.id}`} variant="glass" className="p-5">
            <div className="flex flex-col justify-between gap-4 sm:flex-row sm:items-start">
              <div className="flex min-w-0 gap-3"><div><div className="flex flex-wrap items-center gap-2"><p className="font-medium">Заявка на спортивный разряд</p><StatusBadge status={application.status} /></div><p className="mt-1 text-sm text-muted-foreground">{state.data?.userNames[application.athlete_user_id] ?? "Спортсмен"} · {state.data?.ranks.find((item) => item.id === application.rank_id)?.name ?? "Разряд"} · {state.data?.disciplines.find((item) => item.id === application.discipline_id)?.name ?? "Дисциплина"}</p><p className="mt-1 text-xs text-muted-foreground">Подана {formatDate(application.created_at, true)} · {application.application_type === "confirmation" ? "подтверждение" : "присвоение"}</p>{application.rejection_reason ? <p className="mt-2 text-sm text-destructive">{application.rejection_reason}</p> : null}</div></div>
              {state.data?.mode === "review" ? <div className="flex w-full shrink-0 gap-2 sm:w-auto"><Button className="flex-1 sm:flex-none" size="sm" onClick={() => setReviewTarget({ kind: "rank", id: application.id, status: "approved" })}>Одобрить</Button><Button className="flex-1 sm:flex-none" size="sm" variant="outline" onClick={() => setReviewTarget({ kind: "rank", id: application.id, status: "rejected" })}>Отклонить</Button></div> : null}
            </div>
          </Card>
        ))}

        {visibleJudgeApplications.map((application) => (
          <Card key={`judge-${application.id}`} variant="glass" className="p-5">
            <div className="flex flex-col justify-between gap-4 sm:flex-row sm:items-start">
              <div className="flex min-w-0 gap-3"><div><div className="flex flex-wrap items-center gap-2"><p className="font-medium">{application.category_name ?? `Судейская категория #${application.category_id}`}</p><StatusBadge status={application.status} /></div><p className="mt-1 text-sm text-muted-foreground">{application.judge_name ? `${application.judge_name} · ` : ""}{application.application_type}</p><p className="mt-1 text-xs text-muted-foreground">Подана {formatDate(application.submitted_at, true)}</p>{application.rejection_reason ? <p className="mt-2 text-sm text-destructive">{application.rejection_reason}</p> : null}</div></div>
              {state.data?.mode === "review" ? <div className="flex w-full shrink-0 gap-2 sm:w-auto"><Button className="flex-1 sm:flex-none" size="sm" onClick={() => setReviewTarget({ kind: "judge", id: application.id, status: "approved" })}>Одобрить</Button><Button className="flex-1 sm:flex-none" size="sm" variant="outline" onClick={() => setReviewTarget({ kind: "judge", id: application.id, status: "rejected" })}>Отклонить</Button></div> : null}
            </div>
          </Card>
        ))}
      </div>

      <Dialog open={createOpen} onOpenChange={setCreateOpen}>
        <DialogContent><DialogHeader><DialogTitle>Новая заявка</DialogTitle><DialogDescription>{state.data?.activeRole === "judge" ? "Укажите категорию и основание." : "Укажите разряд, дисциплину и результат-основание."}</DialogDescription></DialogHeader>
          <form className="grid gap-3" onSubmit={submitApplication}>
            {state.data?.activeRole === "judge" ? <><Label htmlFor="category-id">Судейская категория</Label><Select name="category_id" required><SelectTrigger id="category-id"><SelectValue placeholder="Выберите категорию" /></SelectTrigger><SelectContent>{state.data.judgeCategories.map((category) => <SelectItem key={category.id} value={String(category.id)}>{category.name}</SelectItem>)}</SelectContent></Select></> : <>
              <Label htmlFor="rank-id">Разряд</Label><Select name="rank_id" required><SelectTrigger id="rank-id"><SelectValue placeholder="Выберите разряд" /></SelectTrigger><SelectContent>{state.data?.ranks.map((rank) => <SelectItem key={rank.id} value={String(rank.id)}>{rank.name}</SelectItem>)}</SelectContent></Select>
              {state.data?.activeRole === "coach" ? <div><Label htmlFor="athlete-id">Спортсмен</Label><Select required value={selectedAthleteId ? String(selectedAthleteId) : ""} onValueChange={(value) => { const athleteId = Number(value); setSelectedAthleteId(athleteId); if (selectedDisciplineId) void loadBases(selectedDisciplineId, athleteId); }}><SelectTrigger id="athlete-id"><SelectValue placeholder="Выберите спортсмена" /></SelectTrigger><SelectContent>{state.data.athletes.map((relation) => <SelectItem key={relation.participant_user_id} value={String(relation.participant_user_id)}>{state.data?.userNames[relation.participant_user_id] ?? "Спортсмен"}</SelectItem>)}</SelectContent></Select></div> : null}
              <Label htmlFor="discipline-id">Дисциплина</Label><Select required value={selectedDisciplineId ? String(selectedDisciplineId) : ""} onValueChange={(value) => { const disciplineId = Number(value); setSelectedDisciplineId(disciplineId); void loadBases(disciplineId, selectedAthleteId); }}><SelectTrigger id="discipline-id"><SelectValue placeholder="Выберите дисциплину" /></SelectTrigger><SelectContent>{state.data?.disciplines.map((item) => <SelectItem key={item.id} value={String(item.id)}>{item.name}</SelectItem>)}</SelectContent></Select>
              <Label htmlFor="result-id">Результат-основание</Label><Select required disabled={!selectedDisciplineId || (state.data?.activeRole === "coach" && !selectedAthleteId) || basisState.loading || basisState.items.length === 0} value={selectedResultId ? String(selectedResultId) : ""} onValueChange={(value) => setSelectedResultId(Number(value))}><SelectTrigger id="result-id"><SelectValue placeholder={basisState.loading ? "Загружаем результаты…" : "Выберите результат"} /></SelectTrigger><SelectContent>{basisState.items.map((basis) => <SelectItem key={basis.result_id} value={String(basis.result_id)}>{basis.competition_name} · {formatDate(basis.start_at)}{basis.place ? ` · ${basis.place} место` : basis.score != null ? ` · ${basis.score} баллов` : ""}</SelectItem>)}</SelectContent></Select>
              {basisState.error ? <p className="text-xs text-destructive">{basisState.error}</p> : null}
              {!basisState.loading && selectedDisciplineId && (state.data?.activeRole !== "coach" || selectedAthleteId) && !basisState.items.length && !basisState.error ? <p className="text-xs text-muted-foreground">Нет утверждённых завершённых результатов по выбранной дисциплине.</p> : null}
            </>}
            <Label htmlFor="application-type">Тип заявки</Label><Select name="application_type" defaultValue="assignment"><SelectTrigger id="application-type"><SelectValue /></SelectTrigger><SelectContent><SelectItem value="assignment">Присвоение</SelectItem><SelectItem value="confirmation">Подтверждение</SelectItem></SelectContent></Select>
            <DialogFooter><Button type="submit" disabled={pending}>{pending ? <LoaderCircle className="animate-spin" /> : null}Подать</Button></DialogFooter>
          </form>
        </DialogContent>
      </Dialog>

      <Dialog open={Boolean(reviewTarget)} onOpenChange={(open) => !open && setReviewTarget(null)}>
        <DialogContent><DialogHeader><DialogTitle>{reviewTarget?.status === "approved" ? "Одобрить заявку" : "Отклонить заявку"}</DialogTitle><DialogDescription>Решение будет сохранено и отправлено заявителю.</DialogDescription></DialogHeader>
          <form className="grid gap-3" onSubmit={submitReview}>
            {reviewTarget?.status === "approved" ? <><Label htmlFor="assigned-at">Дата присвоения</Label><DatePicker id="assigned-at" name="assigned_at" defaultValue={new Date().toISOString().slice(0, 10)} /><Label htmlFor="expires-at">Действует до</Label><DatePicker id="expires-at" name="expires_at" /><Label htmlFor="order-number">Номер приказа</Label><Input id="order-number" name="order_number" /></> : <><Label htmlFor="rejection-reason">Причина отклонения</Label><Input id="rejection-reason" name="rejection_reason" required /></>}
            <DialogFooter><Button type="submit" variant={reviewTarget?.status === "rejected" ? "destructive" : "default"} disabled={pending}>{pending ? <LoaderCircle className="animate-spin" /> : null}Сохранить решение</Button></DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
    </div>
  );
}
