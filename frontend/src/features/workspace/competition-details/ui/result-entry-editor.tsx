"use client";

import { useState, type FormEvent } from "react";

import type { CompetitionResultTable, ResultEntry, ResultEntryPayload, ResultTaskPayload, SaveResultsPayload } from "@/entities/competition-result";
import { Button } from "@/shared/ui/button";
import { Input } from "@/shared/ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/shared/ui/select";

function entryLabel(entry: ResultEntry) {
  if (entry.team) return entry.team.name;
  return entry.participants.map((person) => [person.last_name, person.first_name].join(" ")).join(", ") || `Строка №${entry.id}`;
}

function draftFor(entry: ResultEntry, taskCount: number): ResultEntryPayload {
  return {
    entry_id: entry.id,
    result_status: entry.result_status || "finished",
    tasks: Array.from({ length: taskCount }, (_, index): ResultTaskPayload => {
      const saved = entry.tasks.find((task) => task.task_number === index + 1);
      return { task_number: index + 1, status: saved?.status ?? "not_attempted", solve_time_deciseconds: saved?.solve_time_deciseconds ?? null, negative_attempts: saved?.negative_attempts ?? 0 };
    }),
  };
}

export function ResultEntryEditor({ table, pending, onSave }: { table: CompetitionResultTable; pending: boolean; onSave: (payload: SaveResultsPayload) => void }) {
  const [selectedId, setSelectedId] = useState(table.entries[0]?.id ?? 0);
  const [draft, setDraft] = useState<ResultEntryPayload | null>(() => table.entries[0] ? draftFor(table.entries[0], table.task_count) : null);
  const [error, setError] = useState<string | null>(null);

  function selectEntry(id: number) {
    const entry = table.entries.find((item) => item.id === id);
    if (!entry) return;
    setSelectedId(id);
    setDraft(draftFor(entry, table.task_count));
    setError(null);
  }

  function updateTask(number: number, change: Partial<ResultTaskPayload>) {
    setDraft((current) => current ? { ...current, tasks: current.tasks.map((task) => task.task_number === number ? { ...task, ...change } : task) } : current);
  }

  function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!draft) return;
    const tasks = draft.result_status === "finished" ? draft.tasks : [];
    if (tasks.some((task) => task.status === "solved" && (!task.solve_time_deciseconds || task.solve_time_deciseconds <= 0))) {
      setError("Для решённой задачи укажите положительное время в десятых долях секунды.");
      return;
    }
    setError(null);
    onSave({ entries: [{ ...draft, tasks }] });
  }

  if (!draft || !table.entries.length) return <p className="text-sm text-muted-foreground">Сначала настройте зачёт, чтобы появились участники с одобренными заявками.</p>;

  return <form className="grid gap-4 rounded-xl border p-4" onSubmit={submit}>
    <div className="grid gap-2"><label className="text-sm font-medium" htmlFor="result-entry">Участник или команда</label><Select value={String(selectedId)} onValueChange={(value) => selectEntry(Number(value))}><SelectTrigger id="result-entry"><SelectValue /></SelectTrigger><SelectContent>{table.entries.map((entry) => <SelectItem key={entry.id} value={String(entry.id)}>{entryLabel(entry)}</SelectItem>)}</SelectContent></Select></div>
    <div className="grid gap-2"><label className="text-sm font-medium" htmlFor="result-status">Статус выступления</label><Select value={draft.result_status} onValueChange={(value) => setDraft((current) => current ? { ...current, result_status: value } : current)}><SelectTrigger id="result-status"><SelectValue /></SelectTrigger><SelectContent><SelectItem value="finished">Завершил</SelectItem><SelectItem value="did_not_start">Не стартовал</SelectItem><SelectItem value="disqualified">Дисквалифицирован</SelectItem></SelectContent></Select></div>
    {draft.result_status === "finished" ? <div className="grid gap-3"><p className="text-sm font-medium">Задачи</p>{draft.tasks.map((task) => <div key={task.task_number} className="grid gap-2 rounded-lg bg-muted/50 p-3 sm:grid-cols-[4rem_minmax(0,1fr)_minmax(0,1fr)_minmax(0,1fr)] sm:items-end"><span className="text-sm font-medium">№{task.task_number}</span><div className="grid gap-1"><label className="text-xs" htmlFor={`task-status-${task.task_number}`}>Статус</label><Select value={task.status} onValueChange={(value) => updateTask(task.task_number, { status: value, solve_time_deciseconds: value === "solved" ? task.solve_time_deciseconds : null })}><SelectTrigger id={`task-status-${task.task_number}`}><SelectValue /></SelectTrigger><SelectContent><SelectItem value="not_attempted">Не решал</SelectItem><SelectItem value="attempted">Попытка</SelectItem><SelectItem value="solved">Решил</SelectItem></SelectContent></Select></div><div className="grid gap-1"><label className="text-xs" htmlFor={`task-time-${task.task_number}`}>Время (0,1 с)</label><Input id={`task-time-${task.task_number}`} type="number" min="1" step="1" inputMode="numeric" disabled={task.status !== "solved"} value={task.solve_time_deciseconds ?? ""} onChange={(event) => updateTask(task.task_number, { solve_time_deciseconds: event.target.value ? Number(event.target.value) : null })} /></div><div className="grid gap-1"><label className="text-xs" htmlFor={`task-attempts-${task.task_number}`}>Неудачные попытки</label><Input id={`task-attempts-${task.task_number}`} type="number" min="0" step="1" inputMode="numeric" value={task.negative_attempts} onChange={(event) => updateTask(task.task_number, { negative_attempts: Number(event.target.value) })} /></div></div>)}</div> : null}
    {error ? <p role="alert" className="text-sm text-destructive">{error}</p> : null}
    <Button type="submit" className="w-fit" disabled={pending}>Сохранить результат участника</Button>
  </form>;
}
