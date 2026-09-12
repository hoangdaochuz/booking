"use client";

/**
 * Admin > Operations > Cron Jobs
 *
 * Manages scheduler_configs rows in scheduler-service. Toggling or
 * editing a job calls UpdateSchedulerById (gateway → scheduler gRPC)
 * with enable state, cron expression, and timeout; the update bumps
 * the config version and propagates via the scheduler-config Kafka
 * topic so the worker re-registers the job with the new schedule.
 */

import { useCallback, useEffect, useMemo, useState } from "react";
import {
  Check,
  ChevronRight,
  Search,
  X,
} from "lucide-react";
import { describeCron, nextTick, parseCron } from "@/lib/cron";
import { apiClient } from "@/lib/api/client";
import { ApiSchedulerJob } from "@/lib/api/types";

interface CronJobConfig {
  id: string;
  name: string;
  cronExpression: string;
  timeout: number; // seconds
  version: number;
  isEnabled: boolean;
  updatedAt: string;
}

/** ApiSchedulerJob (snake_case, RFC3339) → row model */
function toCronJobConfig(j: ApiSchedulerJob): CronJobConfig {
  const d = new Date(j.updated_at);
  return {
    id: j.id,
    name: j.name,
    cronExpression: j.cron_expression,
    timeout: j.timeout,
    version: j.version,
    isEnabled: j.is_enabled,
    updatedAt: isNaN(d.getTime())
      ? j.updated_at
      : d.toLocaleString(undefined, {
          year: "numeric",
          month: "2-digit",
          day: "2-digit",
          hour: "2-digit",
          minute: "2-digit",
        }),
  };
}

const CRON_PRESETS = [
  { label: "30s", value: "*/30 * * * * *" },
  { label: "1m", value: "0 * * * * *" },
  { label: "5m", value: "0 */5 * * * *" },
  { label: "15m", value: "0 */15 * * * *" },
  { label: "hourly", value: "0 0 * * * *" },
  { label: "daily 3am", value: "0 0 3 * * *" },
];

type Filter = "all" | "enabled" | "disabled";

export default function AdminOpsCronPage() {
  const [jobs, setJobs] = useState<CronJobConfig[]>([]);
  const [loading, setLoading] = useState(true);
  const [loadError, setLoadError] = useState<string | null>(null);
  const [busyId, setBusyId] = useState<string | null>(null);
  const [filter, setFilter] = useState<Filter>("all");
  const [query, setQuery] = useState("");
  const [editing, setEditing] = useState<CronJobConfig | null>(null);
  const [toast, setToast] = useState<string | null>(null);

  const fetchJobs = useCallback(async () => {
    try {
      const res = await apiClient.listSchedulerJobs();
      setJobs(res.schedulers.map(toCronJobConfig));
      setLoadError(null);
    } catch (err) {
      setLoadError(
        err instanceof Error ? err.message : "Failed to load scheduler jobs"
      );
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchJobs();
  }, [fetchJobs]);

  const visible = useMemo(
    () =>
      jobs.filter(
        (j) =>
          (filter === "all" ||
            (filter === "enabled" ? j.isEnabled : !j.isEnabled)) &&
          j.name.toLowerCase().includes(query.toLowerCase())
      ),
    [jobs, filter, query]
  );

  function notify(msg: string) {
    setToast(msg);
    window.setTimeout(() => setToast(null), 3200);
  }

  // PATCH /scheduler/jobs/:id { is_enable, cron_expression, timeout }
  async function toggleJob(job: CronJobConfig) {
    setBusyId(job.id);
    try {
      await apiClient.updateSchedulerJob({
        id: job.id,
        is_enable: !job.isEnabled,
        cron_expression: job.cronExpression,
        timeout: job.timeout,
      });
      notify(
        `${job.name} ${!job.isEnabled ? "scheduled" : "paused"} — propagating via Kafka`
      );
      await fetchJobs(); // version/updated_at are authoritative server-side
    } catch (err) {
      notify(err instanceof Error ? `Failed: ${err.message}` : "Update failed");
    } finally {
      setBusyId(null);
    }
  }

  // PATCH /scheduler/jobs/:id { is_enable, cron_expression, timeout }
  async function saveJob(updated: CronJobConfig) {
    setBusyId(updated.id);
    try {
      await apiClient.updateSchedulerJob({
        id: updated.id,
        is_enable: updated.isEnabled,
        cron_expression: updated.cronExpression,
        timeout: updated.timeout,
      });
      notify(`Saved ${updated.name} — propagating via Kafka`);
      setEditing(null);
      await fetchJobs();
    } catch (err) {
      notify(err instanceof Error ? `Failed: ${err.message}` : "Update failed");
    } finally {
      setBusyId(null);
    }
  }

  return (
    <div className="flex flex-col min-h-screen">
      {/* Topbar */}
      <header className="h-14 flex items-center justify-between px-6 bg-card border-b border-border sticky top-0 z-10">
        <div className="flex items-center gap-1.5 text-sm text-muted">
          <span>Ops</span>
          <ChevronRight size={12} />
          <strong className="text-foreground font-semibold">Cron Jobs</strong>
        </div>
        <div className="flex items-center gap-3">
          <span className="font-mono text-[11px] bg-tag-bg border border-border rounded-md px-2 py-0.5 text-muted">
            env: production
          </span>
          <div
            className="w-[30px] h-[30px] rounded-full bg-tag-bg border border-border grid place-items-center text-xs font-semibold"
            title="Ops admin"
          >
            OA
          </div>
        </div>
      </header>

      <main className="p-6 max-w-[1280px] w-full flex flex-col gap-4">
        {/* Page head */}
        <div className="flex items-start justify-between gap-4 flex-wrap">
          <div>
            <h1 className="text-[22px] font-bold tracking-tight">Cron Jobs</h1>
            <p className="text-muted text-[13px] mt-0.5 max-w-[560px]">
              Scheduled background jobs from{" "}
              <span className="font-mono text-xs">scheduler-service</span>.
              Changes propagate via the scheduler-config Kafka topic and take
              effect on the next tick.
            </p>
          </div>
          <button className="flex items-center gap-2 bg-primary hover:bg-primary-hover text-white font-semibold rounded-full h-[38px] px-[18px] text-[13px] transition-colors">
            Register job config
          </button>
        </div>

        {/* Toolbar */}
        <div className="flex items-center gap-2.5 flex-wrap">
          <div className="flex items-center gap-2 bg-card border border-border rounded-lg px-3 h-[38px] flex-1 min-w-[200px] max-w-[320px]">
            <Search size={14} className="text-muted shrink-0" />
            <input
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              placeholder="Search by job name…"
              className="border-none outline-none bg-transparent text-[13px] w-full text-foreground"
            />
          </div>
          <div className="flex bg-tag-bg rounded-full p-[3px]">
            {(["all", "enabled", "disabled"] as Filter[]).map((f) => (
              <button
                key={f}
                onClick={() => setFilter(f)}
                className={`px-3.5 py-[5px] rounded-full text-xs font-medium capitalize transition-colors ${
                  filter === f
                    ? "bg-card text-foreground font-semibold shadow-sm"
                    : "text-muted hover:text-foreground"
                }`}
              >
                {f}
              </button>
            ))}
          </div>
        </div>

        {/* Table */}
        <div className="bg-card border border-border rounded-xl overflow-hidden">
          {loadError && (
            <div className="px-4 py-3 bg-red-50 border-b border-red-100 text-[13px] text-red-700 flex items-center justify-between">
              <span>{loadError}</span>
              <button
                onClick={() => fetchJobs()}
                className="font-semibold underline underline-offset-2"
              >
                Retry
              </button>
            </div>
          )}
          <table className="w-full border-collapse">
            <thead>
              <tr className="bg-background">
                {["Job", "Schedule", "Timeout", "Version", "Last updated", "Status"].map(
                  (h) => (
                    <th
                      key={h}
                      className="text-left text-[11px] font-semibold uppercase tracking-wider text-muted px-4 py-3 border-b border-border whitespace-nowrap"
                    >
                      {h}
                    </th>
                  )
                )}
                <th className="text-right text-[11px] font-semibold uppercase tracking-wider text-muted px-4 py-3 border-b border-border">
                  Enabled
                </th>
              </tr>
            </thead>
            <tbody>
              {loading && (
                <tr>
                  <td
                    colSpan={7}
                    className="px-4 py-12 text-center text-muted text-[13px]"
                  >
                    Loading scheduler jobs…
                  </td>
                </tr>
              )}
              {!loading && visible.map((j) => (
                <tr key={j.id} className="hover:bg-tag-bg/60 transition-colors">
                  <td className="px-4 py-3.5 border-b border-border">
                    <div className="font-semibold text-[13px]">{j.name}</div>
                    <div className="font-mono text-[11px] text-muted mt-0.5">
                      {j.id.slice(0, 8)}
                    </div>
                  </td>
                  <td className="px-4 py-3.5 border-b border-border">
                    <span className="font-mono text-xs bg-tag-bg rounded-md px-2 py-[3px]">
                      {j.cronExpression}
                    </span>
                    <div className="text-xs text-muted mt-1">
                      {describeCron(j.cronExpression) ?? "—"}
                    </div>
                  </td>
                  <td className="px-4 py-3.5 border-b border-border tabular-nums">
                    {j.timeout}s
                  </td>
                  <td className="px-4 py-3.5 border-b border-border">
                    <span className="font-mono text-[11px] text-muted bg-tag-bg rounded-md px-1.5 py-0.5">
                      v{j.version}
                    </span>
                  </td>
                  <td className="px-4 py-3.5 border-b border-border text-xs text-muted whitespace-nowrap">
                    {j.updatedAt}
                  </td>
                  <td className="px-4 py-3.5 border-b border-border">
                    <span
                      className={`inline-flex items-center gap-1.5 text-xs font-semibold rounded-full px-2.5 py-[3px] ${
                        j.isEnabled
                          ? "bg-success-bg text-success"
                          : "bg-tag-bg text-muted"
                      }`}
                    >
                      <span className="w-1.5 h-1.5 rounded-full bg-current" />
                      {j.isEnabled ? "Scheduled" : "Paused"}
                    </span>
                  </td>
                  <td className="px-4 py-3.5 border-b border-border text-right whitespace-nowrap">
                    <button
                      onClick={() => setEditing(j)}
                      disabled={busyId === j.id}
                      className="text-muted hover:text-foreground hover:bg-tag-bg rounded-lg px-1.5 py-1.5 text-[13px] transition-colors disabled:opacity-40"
                    >
                      Edit
                    </button>
                    <Toggle
                      checked={j.isEnabled}
                      disabled={busyId === j.id}
                      onChange={() => toggleJob(j)}
                      label={`Toggle ${j.name}`}
                    />
                  </td>
                </tr>
              ))}
              {!loading && visible.length === 0 && (
                <tr>
                  <td
                    colSpan={7}
                    className="px-4 py-12 text-center text-muted text-[13px]"
                  >
                    {jobs.length === 0
                      ? "No scheduler job configs found."
                      : "No jobs match this filter."}
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      </main>

      {/* Edit drawer */}
      {editing && (
        <EditDrawer
          job={editing}
          onClose={() => setEditing(null)}
          onSave={saveJob}
        />
      )}

      {/* Toast */}
      {toast && (
        <div className="fixed bottom-6 left-1/2 -translate-x-1/2 bg-foreground text-white text-[13px] rounded-full px-5 py-2.5 flex items-center gap-2.5 shadow-lg z-50">
          <span className="w-[7px] h-[7px] rounded-full bg-primary" />
          {toast}
        </div>
      )}
    </div>
  );
}

function Toggle({
  checked,
  onChange,
  label,
  disabled = false,
}: {
  checked: boolean;
  onChange: () => void;
  label: string;
  disabled?: boolean;
}) {
  return (
    <button
      role="switch"
      aria-checked={checked}
      aria-label={label}
      onClick={onChange}
      disabled={disabled}
      className={`relative w-10 h-[22px] rounded-full border transition-colors align-middle ml-2 ${
        checked
          ? "bg-primary border-primary"
          : "bg-tag-bg border-border"
      } ${disabled ? "opacity-40 cursor-not-allowed" : ""}`}
    >
      <span
        className={`absolute top-[2px] left-[2px] w-4 h-4 rounded-full bg-white shadow-sm transition-transform ${
          checked ? "translate-x-[18px]" : ""
        }`}
      />
    </button>
  );
}

function EditDrawer({
  job,
  onClose,
  onSave,
}: {
  job: CronJobConfig;
  onClose: () => void;
  onSave: (job: CronJobConfig) => Promise<void>;
}) {
  const [enabled, setEnabled] = useState(job.isEnabled);
  const [cron, setCron] = useState(job.cronExpression);
  const [timeout, setTimeoutSecs] = useState(job.timeout);
  const [saving, setSaving] = useState(false);

  const parsed = parseCron(cron);
  const next = parsed.error ? null : nextTick(cron);
  const canSave = parsed.error === null;

  return (
    <>
      <div
        className="fixed inset-0 bg-foreground/30 z-40"
        onClick={onClose}
      />
      <aside className="fixed top-0 right-0 h-screen w-[min(440px,100vw)] bg-card border-l border-border z-50 flex flex-col">
        <div className="flex items-center justify-between px-[22px] py-4 border-b border-border">
          <div>
            <h2 className="text-base font-bold">Edit job</h2>
            <p className="text-xs text-muted mt-0.5">
              {job.name} · currently v{job.version}
            </p>
          </div>
          <button
            onClick={onClose}
            aria-label="Close"
            className="text-muted hover:text-foreground p-1.5"
          >
            <X size={18} />
          </button>
        </div>

        <div className="px-[22px] py-[22px] flex flex-col gap-[22px] overflow-y-auto flex-1">
          {/* Enabled switch */}
          <div className="flex items-center justify-between bg-background border border-border rounded-[10px] px-3.5 py-3">
            <div>
              <div className="text-[13px] font-semibold">Job enabled</div>
              <div className="text-xs text-muted">
                Disabled jobs keep their config but are removed from the cron
                runner
              </div>
            </div>
            <Toggle
              checked={enabled}
              onChange={() => setEnabled(!enabled)}
              label="Job enabled"
            />
          </div>

          {/* Cron expression */}
          <div className="flex flex-col gap-1.5">
            <div className="flex items-center justify-between">
              <label htmlFor="cron-input" className="text-xs font-semibold">
                Cron expression
              </label>
              <span className="text-[11px] text-muted font-normal">
                6-field · seconds included
              </span>
            </div>
            <input
              id="cron-input"
              value={cron}
              onChange={(e) => setCron(e.target.value)}
              spellCheck={false}
              className="border border-border rounded-lg px-3 py-2.5 text-[13px] font-mono bg-card text-foreground focus:outline-none focus:border-primary transition-colors"
            />
            <div className="flex gap-1.5 flex-wrap">
              {CRON_PRESETS.map((p) => (
                <button
                  key={p.value}
                  onClick={() => setCron(p.value)}
                  className={`border rounded-full px-3 py-1 text-xs font-mono transition-colors ${
                    cron === p.value
                      ? "bg-primary border-primary text-white"
                      : "border-border bg-card text-muted hover:border-primary hover:text-foreground"
                  }`}
                >
                  {p.label}
                </button>
              ))}
            </div>
            {parsed.error ? (
              <p className="text-xs text-red-600">{parsed.error}</p>
            ) : (
              <div className="bg-background border border-border rounded-lg px-3 py-2.5">
                <div className="text-[11px] uppercase tracking-wider text-muted font-semibold mb-1">
                  Interpreted as
                </div>
                <div className="text-[13px]">{describeCron(cron)}</div>
                {next && (
                  <div className="text-xs text-muted mt-1.5">
                    Next run: {next.toLocaleTimeString()} ·{" "}
                    {next.toLocaleDateString()}
                  </div>
                )}
              </div>
            )}
          </div>

          {/* Timeout */}
          <div className="flex flex-col gap-1.5">
            <div className="flex items-center justify-between">
              <label htmlFor="timeout-input" className="text-xs font-semibold">
                Timeout
              </label>
              <span className="text-[11px] text-muted font-normal">
                job context deadline
              </span>
            </div>
            <div className="flex border border-border rounded-lg overflow-hidden w-[200px]">
              <button
                type="button"
                aria-label="Decrease timeout"
                onClick={() =>
                  setTimeoutSecs((t) => Math.max(1, t - 5))
                }
                className="w-10 bg-tag-bg text-base hover:bg-border transition-colors"
              >
                −
              </button>
              <input
                id="timeout-input"
                type="number"
                min={1}
                max={600}
                value={timeout}
                onChange={(e) =>
                  setTimeoutSecs(
                    Math.max(1, Math.min(600, Number(e.target.value) || 30))
                  )
                }
                className="border-x border-border text-center text-[13px] py-2.5 w-full tabular-nums outline-none bg-card text-foreground"
              />
              <button
                type="button"
                aria-label="Increase timeout"
                onClick={() =>
                  setTimeoutSecs((t) => Math.min(600, t + 5))
                }
                className="w-10 bg-tag-bg text-base hover:bg-border transition-colors"
              >
                +
              </button>
            </div>
            <p className="text-xs text-muted">
              Seconds (1–600). Applied on the next job registration — when
              unset or ≤ 0, the worker falls back to its 30s default.
            </p>
          </div>

          {/* Propagation note */}
          <div className="bg-yellow-50 border border-yellow-100 rounded-[10px] px-3.5 py-3 text-xs text-yellow-800 leading-relaxed">
            Saving bumps the config <strong>version</strong> and emits a{" "}
            <span className="font-mono">scheduler-config-changed</span> event
            through the outbox. The worker re-registers the job when it
            consumes the event; stale lower-version events are ignored.
          </div>
        </div>

        <div className="px-[22px] py-4 border-t border-border flex gap-2.5 justify-end">
          <button
            onClick={onClose}
            className="bg-card border border-border text-foreground font-semibold rounded-full px-4 h-[38px] text-[13px] hover:bg-tag-bg transition-colors"
          >
            Cancel
          </button>
          <button
            disabled={!canSave || saving}
            onClick={async () => {
              setSaving(true);
              try {
                await onSave({
                  ...job,
                  isEnabled: enabled,
                  cronExpression: cron,
                  timeout,
                });
              } finally {
                setSaving(false);
              }
            }}
            className="bg-primary hover:bg-primary-hover disabled:opacity-40 disabled:cursor-not-allowed text-white font-semibold rounded-full px-[18px] h-[38px] text-[13px] transition-colors flex items-center gap-2"
          >
            <Check size={14} />
            {saving ? "Saving…" : "Save changes"}
          </button>
        </div>
      </aside>
    </>
  );
}
