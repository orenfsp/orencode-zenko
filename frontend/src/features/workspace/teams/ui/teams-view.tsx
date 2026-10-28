"use client";

import { type FormEvent, useCallback, useState } from "react";
import Link from "next/link";
import { teamApi } from "@/entities/team";
import { registryApi } from "@/entities/registry";
import { userCoachApi } from "@/entities/user-coach";
import { profileApi } from "@/entities/profile";
import { useAuth } from "@/entities/session";
import { formatDate } from "@/features/workspace/common";
import { useAsyncData } from "@/shared/lib/use-async-data";
import { useNotifications } from "@/shared/lib/notifications";
import { Button } from "@/shared/ui/button";
import { Card } from "@/shared/ui/card";
import { Checkbox } from "@/shared/ui/checkbox";
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "@/shared/ui/dialog";
import { Input } from "@/shared/ui/input";
import { DataEmpty, DataError, DataLoading } from "@/shared/ui/data-state";
import { WorkspaceToolbar } from "@/shared/ui/workspace-toolbar";
import { ScrollMask } from "@/shared/ui/scroll-mask";

interface Candidate {
  id: number;
  name: string;
}
function nameOf(person: {
  first_name: string;
  last_name: string;
  middle_name?: string | null;
}) {
  return [person.last_name, person.first_name, person.middle_name]
    .filter(Boolean)
    .join(" ");
}

export function TeamsView() {
  const { identity } = useAuth();
  const notify = useNotifications();
  const [memberIds, setMemberIds] = useState<number[]>([]);
  const [coachIds, setCoachIds] = useState<number[]>([]);
  const [pending, setPending] = useState(false);
  const [createOpen, setCreateOpen] = useState(false);
  const [search, setSearch] = useState("");
  const canReadRegistry =
    identity?.permissions?.includes("registry.read") ?? false;
  const load = useCallback(async () => {
    const teams = await teamApi.list();
    let members: Candidate[] = [];
    let coaches: Candidate[] = [];
    if (canReadRegistry) {
      const [athletes, coachUsers] = await Promise.all([
        registryApi.users({ role: "participant", limit: 100 }),
        registryApi.users({ role: "coach", limit: 100 }),
      ]);
      members = athletes.items.map((user) => ({
        id: user.id,
        name: nameOf(user),
      }));
      coaches = coachUsers.items.map((user) => ({
        id: user.id,
        name: nameOf(user),
      }));
    } else {
      const [relations, profile] = await Promise.all([
        userCoachApi.byCoach({ limit: 100 }),
        profileApi.me(),
      ]);
      const confirmedRelations = relations.filter((relation) => relation.status === "confirmed");
      const participantProfiles = await Promise.allSettled(confirmedRelations.map((relation) => profileApi.public(relation.participant_user_id)));
      members = confirmedRelations.map((relation, index) => ({
        id: relation.participant_user_id,
        name: participantProfiles[index]?.status === "fulfilled" ? nameOf(participantProfiles[index].value) : "Спортсмен",
      }));
      coaches = [{ id: profile.id, name: nameOf(profile) }];
    }
    return {
      teams,
      members: [...new Map(members.map((item) => [item.id, item])).values()],
      coaches,
    };
  }, [canReadRegistry]);
  const state = useAsyncData(load);
  const selectedCoachIds =
    !canReadRegistry && identity?.userId
      ? [...new Set([identity.userId, ...coachIds])]
      : coachIds;
  const filteredTeams = (state.data?.teams ?? []).filter((team) => team.name.toLocaleLowerCase("ru").includes(search.trim().toLocaleLowerCase("ru")));

  function toggle(
    id: number,
    selected: boolean,
    setter: (value: number[] | ((previous: number[]) => number[])) => void,
  ) {
    setter((previous) =>
      selected ? [...previous, id] : previous.filter((item) => item !== id),
    );
  }

  async function create(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const name = String(
      new FormData(event.currentTarget).get("name") ?? "",
    ).trim();
    if (!memberIds.length)
      return notify.error("Выберите хотя бы одного спортсмена");
    setPending(true);
    try {
      await teamApi.create({
        name,
        member_ids: memberIds,
        coach_ids: selectedCoachIds,
      });
      notify.success("Сборная создана");
      setCreateOpen(false);
      setMemberIds([]);
      setCoachIds([]);
      event.currentTarget.reset();
      state.reload();
    } catch (cause) {
      notify.error(
        cause instanceof Error ? cause.message : "Не удалось создать сборную",
      );
    } finally {
      setPending(false);
    }
  }

  return (
    <div className="mx-auto w-full max-w-6xl">
      <WorkspaceToolbar search={search} searchPlaceholder="Найти сборную" onSearchChange={setSearch} createLabel="Создать сборную" onCreate={() => setCreateOpen(true)} />
      <Dialog open={createOpen} onOpenChange={setCreateOpen}><DialogContent className="max-h-[90vh] overflow-y-auto sm:max-w-2xl"><DialogHeader><DialogTitle>Новая сборная</DialogTitle><DialogDescription>Выберите спортсменов и тренеров, которые войдут в состав.</DialogDescription></DialogHeader><form className="grid gap-4" onSubmit={(event) => void create(event)}>
          <Input
            name="name"
            required
            maxLength={200}
            placeholder="Название сборной"
          />
          <div className="grid gap-4 sm:grid-cols-2">
            <div>
              <p className="mb-2 font-medium">Спортсмены</p>
              <ScrollMask className="grid max-h-52 gap-2">
                {state.data?.members.map((candidate) => (
                  <label
                    key={candidate.id}
                    className="flex items-center gap-2 text-sm"
                  >
                    <Checkbox
                      checked={memberIds.includes(candidate.id)}
                      onCheckedChange={(checked) =>
                        toggle(candidate.id, checked === true, setMemberIds)
                      }
                    />
                    {candidate.name}
                  </label>
                ))}
              </ScrollMask>
            </div>
            <div>
              <p className="mb-2 font-medium">Тренеры</p>
              <ScrollMask className="grid max-h-52 gap-2">
                {state.data?.coaches.map((candidate) => (
                  <label
                    key={candidate.id}
                    className="flex items-center gap-2 text-sm"
                  >
                    <Checkbox
                      checked={selectedCoachIds.includes(candidate.id)}
                      disabled={
                        !canReadRegistry && candidate.id === identity?.userId
                      }
                      onCheckedChange={(checked) =>
                        toggle(candidate.id, checked === true, setCoachIds)
                      }
                    />
                    {candidate.name}
                    {!canReadRegistry && candidate.id === identity?.userId
                      ? " (вы)"
                      : ""}
                  </label>
                ))}
              </ScrollMask>
            </div>
          </div>
          {!state.data?.members.length && !state.loading && !state.error ? (
            <p className="text-sm text-muted-foreground">
              Подходящих спортсменов нет.{" "}
              {canReadRegistry
                ? "Проверьте реестр участников."
                : "Добавьте спортсмена и дождитесь подтверждения связи."}{" "}
              <Link
                className="underline"
                href={
                  canReadRegistry ? "/workspace/registry" : "/workspace/coaches"
                }
              >
                Перейти к списку
              </Link>
            </p>
          ) : null}
          {state.data?.members.length && !selectedCoachIds.length ? (
            <p className="text-sm text-muted-foreground">
              Выберите хотя бы одного тренера.
            </p>
          ) : null}
          <DialogFooter><Button
            disabled={
              pending || !state.data?.members.length || !selectedCoachIds.length
            }
          >
            {pending ? "Создание…" : "Создать сборную"}
          </Button></DialogFooter>
        </form>
      </DialogContent></Dialog>
      {state.loading ? <DataLoading rows={4} /> : null}
      {state.error ? (
        <DataError message={state.error} retry={state.reload} />
      ) : null}
      {!state.loading && !state.error && !filteredTeams.length ? (
        <DataEmpty
          title="Сборных пока нет"
          description="Выберите спортсменов и тренеров выше, чтобы создать первую сборную."
        />
      ) : null}
      <div className="grid gap-3">
        {filteredTeams.map((team) => (
          <Card key={team.id} className="bg-white/75 p-5 shadow-none">
            <p className="font-medium">{team.name}</p>
            <p className="mt-1 text-xs text-muted-foreground">
              Создана {formatDate(team.created_at)}
            </p>
            <p className="mt-3 text-sm">
              Спортсмены: {team.members?.map(nameOf).join(", ") || "—"}
            </p>
            <p className="mt-1 text-sm">
              Тренеры: {team.coaches?.map(nameOf).join(", ") || "—"}
            </p>
          </Card>
        ))}
      </div>
    </div>
  );
}
