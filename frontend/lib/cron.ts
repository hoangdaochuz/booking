/**
 * 6-field cron expression utilities (seconds included), matching
 * robfig/cron v3 with the WithSeconds option used by scheduler-service.
 *
 * Fields: sec(0-59) min(0-59) hour(0-23) dom(1-31) month(1-12) dow(0-6)
 * Supported syntax per field: "*", "N", "N-M", star-over-step, "N-over-step",
 * range-over-step, and comma-separated lists of those.
 */

export interface ParsedCron {
  /** allowed values per field, ascending; null when invalid */
  fields: number[][];
}

export interface CronParseResult {
  error: string | null;
  fields: number[][] | null;
}

const FIELD_NAMES = ["sec", "min", "hour", "dom", "month", "dow"] as const;
const FIELD_RANGES: Array<[number, number]> = [
  [0, 59],
  [0, 59],
  [0, 23],
  [1, 31],
  [1, 12],
  [0, 6],
];

function parseField(
  part: string,
  min: number,
  max: number
): { values: number[] | null; error: string | null } {
  const vals = new Set<number>();
  for (const chunk of part.split(",")) {
    const m = chunk.match(/^(\*|\d+(?:-\d+)?)(?:\/(\d+))?$/);
    if (!m) {
      return { values: null, error: `“${chunk}” is not a valid cron value` };
    }
    let start = min;
    let end = max;
    if (m[1] !== "*") {
      if (m[1].includes("-")) {
        const [a, b] = m[1].split("-").map(Number);
        if (a < min || b > max || a > b) {
          return {
            values: null,
            error: `“${m[1]}” is out of range (${min}–${max})`,
          };
        }
        start = a;
        end = b;
      } else {
        const v = Number(m[1]);
        if (v < min || v > max) {
          return {
            values: null,
            error: `“${m[1]}” is out of range (${min}–${max})`,
          };
        }
        start = v;
        end = v;
      }
    }
    const step = m[2] ? Number(m[2]) : 1;
    if (step < 1) {
      return { values: null, error: `step “/${m[2]}” must be ≥ 1` };
    }
    for (let v = start; v <= end; v += step) vals.add(v);
  }
  if (!vals.size) return { values: null, error: "empty field" };
  return { values: [...vals].sort((a, b) => a - b), error: null };
}

/** Validate a 6-field cron expression. Never throws. */
export function parseCron(expr: string): CronParseResult {
  const parts = expr.trim().split(/\s+/);
  if (parts.length !== 6) {
    return {
      error: `Expected 6 fields, got ${parts.length}`,
      fields: null,
    };
  }
  const fields: number[][] = [];
  for (let i = 0; i < 6; i++) {
    const [min, max] = FIELD_RANGES[i];
    const { values, error } = parseField(parts[i], min, max);
    if (error) {
      return {
        error: `${error} in ${FIELD_NAMES[i]} field`,
        fields: null,
      };
    }
    fields.push(values!);
  }
  return { error: null, fields };
}

/** True when the expression is a valid 6-field cron. */
export function isValidCron(expr: string): boolean {
  return parseCron(expr).error === null;
}

const isEvery = (f: number[], lo: number, hi: number) =>
  f.length === hi - lo + 1;

/** Best-effort human-readable summary, e.g. "Every 30 seconds". */
export function describeCron(expr: string): string | null {
  const { error, fields } = parseCron(expr);
  if (error || !fields) return null;
  const [sec, min, hour, dom, month, dow] = fields;
  const restEvery =
    isEvery(min, 0, 59) &&
    isEvery(hour, 0, 23) &&
    isEvery(dom, 1, 31) &&
    isEvery(month, 1, 12) &&
    isEvery(dow, 0, 6);

  if (restEvery && sec.length === 1) {
    return sec[0] === 0
      ? "Every minute"
      : `Every minute at second :${String(sec[0]).padStart(2, "0")}`;
  }
  const stepOf = (f: number[], lo: number) => {
    if (f.length < 2) return null;
    const s = f[1] - f[0];
    return f[0] === lo && f.every((v, i) => i === 0 || v - f[i - 1] === s)
      ? s
      : null;
  };
  if (restEvery) {
    const secStep = stepOf(sec, 0);
    if (secStep) return secStep === 1 ? "Every second" : `Every ${secStep} seconds`;
  }
  const domEvery = isEvery(dom, 1, 31);
  const monthEvery = isEvery(month, 1, 12);
  const dowEvery = isEvery(dow, 0, 6);
  if (sec.length === 1 && sec[0] === 0 && domEvery && monthEvery && dowEvery) {
    if (min.length === 1 && isEvery(hour, 0, 23)) {
      return `At minute :${String(min[0]).padStart(2, "0")} of every hour`;
    }
    if (min.length === 1 && hour.length === 1) {
      return `Daily at ${String(hour[0]).padStart(2, "0")}:${String(min[0]).padStart(2, "0")}`;
    }
    const minStep = stepOf(min, 0);
    if (minStep && isEvery(hour, 0, 23)) {
      return `Every ${minStep} minutes`;
    }
  }
  return "Custom schedule";
}

/** Next fire time strictly after `from` (defaults to now), or null. */
export function nextTick(expr: string, from: Date = new Date()): Date | null {
  const { error, fields } = parseCron(expr);
  if (error || !fields) return null;
  const [sec, min, hour, dom, month] = fields;
  const d = new Date(from.getTime());
  d.setMilliseconds(0);
  d.setSeconds(d.getSeconds() + 1);
  // Bounded search: worst case ~2 years of seconds for pathological exprs.
  for (let i = 0; i < 500_000; i++) {
    if (!month.includes(d.getMonth() + 1)) {
      d.setMonth(d.getMonth() + 1, 1);
      d.setHours(0, 0, 0, 0);
      continue;
    }
    if (!dom.includes(d.getDate())) {
      d.setDate(d.getDate() + 1);
      d.setHours(0, 0, 0, 0);
      continue;
    }
    if (!hour.includes(d.getHours())) {
      d.setHours(d.getHours() + 1, 0, 0, 0);
      continue;
    }
    if (!min.includes(d.getMinutes())) {
      d.setMinutes(d.getMinutes() + 1, 0, 0);
      continue;
    }
    if (!sec.includes(d.getSeconds())) {
      d.setSeconds(d.getSeconds() + 1);
      continue;
    }
    return d;
  }
  return null;
}
