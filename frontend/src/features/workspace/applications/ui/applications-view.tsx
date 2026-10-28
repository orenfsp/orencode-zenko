"use client";

import { useCallback } from "react";
import { Gavel, Medal } from "lucide-react";

import {
  judgeCategoryApi,
  type JudgeCategoryApplication,
} from "@/entities/judge-category";
import { profileApi } from "@/entities/profile";
import {
  rankApplicationApi,
  type RankApplication,
} from "@/entities/rank-application";
import { useAuth } from "@/entities/session";
import { formatDate, SectionHeading, StatusBadge } from "@/features/workspace/common";
import { useAsyncData } from "@/shared/lib/use-async-data";
import { Card } from "@/shared/ui/card";
import { DataEmpty, DataError, DataLoading } from "@/shared/ui/data-state";

interface ApplicationData {
  mode: "personal" | "review";
  rank: RankApplication[];
  judge: JudgeCategoryApplication[];
}

export function ApplicationsView() {
  const { identity } = useAuth();
  const permissions = identity?.permissions ?? [];
  const canReviewRank = permissions?.includes("rank_applications.review") ?? false;
  const canReviewJudge = permissions?.includes("judge_categories.review") ?? false;

  const load = useCallback(async (): Promise<ApplicationData> => {
    const profile = await profileApi.me();
    const activeRole = profile.roles.find((role) => role.id === identity?.roleId)?.code;
    const isFederationRole = activeRole === "federation" || activeRole === "super_admin";
    if (canReviewRank || canReviewJudge || isFederationRole) {
      const [rank, judge] = await Promise.all([
        canReviewRank || isFederationRole ? rankApplicationApi.pending({ limit: 100 }) : Promise.resolve([]),
        canReviewJudge || isFederationRole ? judgeCategoryApi.pending({ limit: 100 }) : Promise.resolve([]),
      ]);
      return { mode: "review", rank, judge };
    }

    if (activeRole === "judge") {
      return { mode: "personal", rank: [], judge: await judgeCategoryApi.mine() };
    }
    return { mode: "personal", rank: await rankApplicationApi.mine(), judge: [] };
  }, [canReviewJudge, canReviewRank, identity?.roleId]);

  const state = useAsyncData(load);
  const total = (state.data?.rank.length ?? 0) + (state.data?.judge.length ?? 0);

  return (
    <div className="mx-auto w-full max-w-6xl">
      <SectionHeading
        title={state.data?.mode === "review" ? "Заявки на рассмотрение" : "Мои заявки"}
        description={
          state.data?.mode === "review"
            ? "Очередь заявок, доступная активной роли."
            : "Статусы ваших заявок на разряд или судейскую категорию."
        }
      />
      {state.loading ? <DataLoading rows={4} /> : null}
      {state.error ? <DataError message={state.error} retry={state.reload} /> : null}
      {!state.loading && !state.error && !total ? (
        <DataEmpty title="Заявок пока нет" description="Поданные заявки и решения появятся на этой странице." />
      ) : null}

      <div className="grid gap-3">
        {state.data?.rank.map((application) => (
          <Card key={`rank-${application.id}`} className="bg-white/75 p-5 shadow-none">
            <div className="flex items-start justify-between gap-4">
              <div className="flex min-w-0 gap-3">
                <span className="flex size-10 shrink-0 items-center justify-center rounded-xl bg-[#eeebff] text-primary"><Medal className="size-5" /></span>
                <div>
                  <p className="font-medium">Заявка на спортивный разряд</p>
                  <p className="mt-1 text-sm text-muted-foreground">
                    Разряд #{application.rank_id} · дисциплина #{application.discipline_id}
                  </p>
                  <p className="mt-1 text-xs text-muted-foreground">Подана {formatDate(application.created_at, true)}</p>
                  {application.rejection_reason ? <p className="mt-2 text-sm text-destructive">{application.rejection_reason}</p> : null}
                </div>
              </div>
              <StatusBadge status={application.status} />
            </div>
          </Card>
        ))}

        {state.data?.judge.map((application) => (
          <Card key={`judge-${application.id}`} className="bg-white/75 p-5 shadow-none">
            <div className="flex items-start justify-between gap-4">
              <div className="flex min-w-0 gap-3">
                <span className="flex size-10 shrink-0 items-center justify-center rounded-xl bg-[#eeebff] text-primary"><Gavel className="size-5" /></span>
                <div>
                  <p className="font-medium">{application.category_name ?? `Судейская категория #${application.category_id}`}</p>
                  <p className="mt-1 text-sm text-muted-foreground">
                    {application.judge_name ? `${application.judge_name} · ` : ""}{application.application_type}
                  </p>
                  <p className="mt-1 text-xs text-muted-foreground">Подана {formatDate(application.submitted_at, true)}</p>
                  {application.rejection_reason ? <p className="mt-2 text-sm text-destructive">{application.rejection_reason}</p> : null}
                </div>
              </div>
              <StatusBadge status={application.status} />
            </div>
          </Card>
        ))}
      </div>
    </div>
  );
}
