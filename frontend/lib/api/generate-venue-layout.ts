import { VenueLayout, SeatRow } from "../types";
import { ApiTicketTier } from "./types";

// Generate rows + seats for a section, distributing seats from available/total quantities
function generateRows(
  sectionId: string,
  totalSeats: number,
  availableSeats: number,
): SeatRow[] {
  // Cap displayed seats for reasonable UI (max ~60 per section)
  const displayTotal = Math.min(totalSeats, 60);
  const seatsPerRow = Math.min(Math.ceil(Math.sqrt(displayTotal * 2)), 12);
  const rowCount = Math.ceil(displayTotal / seatsPerRow);
  const takenCount = displayTotal - Math.min(availableSeats, displayTotal);

  // Randomly assign which seats are taken (seeded by sectionId for consistency)
  let hash = 0;
  for (let i = 0; i < sectionId.length; i++) {
    hash = ((hash << 5) - hash + sectionId.charCodeAt(i)) | 0;
  }
  function seededRandom() {
    hash = (hash * 1103515245 + 12345) & 0x7fffffff;
    return hash / 0x7fffffff;
  }

  const allSeats: { row: number; seat: number }[] = [];
  for (let r = 0; r < rowCount; r++) {
    const seatsThisRow = r < rowCount - 1 ? seatsPerRow : displayTotal - seatsPerRow * (rowCount - 1);
    for (let s = 0; s < seatsThisRow; s++) {
      allSeats.push({ row: r, seat: s });
    }
  }

  // Mark some seats as taken
  const takenSet = new Set<string>();
  const shuffled = [...allSeats].sort(() => seededRandom() - 0.5);
  for (let i = 0; i < takenCount && i < shuffled.length; i++) {
    takenSet.add(`${shuffled[i].row}-${shuffled[i].seat}`);
  }

  return Array.from({ length: rowCount }, (_, ri) => {
    const seatsThisRow = ri < rowCount - 1 ? seatsPerRow : displayTotal - seatsPerRow * (rowCount - 1);
    return {
      id: `${sectionId}-row-${ri}`,
      label: `Row ${String.fromCharCode(65 + ri)}`,
      seats: Array.from({ length: Math.max(seatsThisRow, 1) }, (_, si) => ({
        id: `${sectionId}-row-${ri}-seat-${si}`,
        label: `${si + 1}`,
        status: takenSet.has(`${ri}-${si}`)
          ? ("taken" as const)
          : ("available" as const),
      })),
    };
  });
}

// Distribute a total count across N sections
function distribute(total: number, n: number): number[] {
  const base = Math.floor(total / n);
  const remainder = total % n;
  return Array.from({ length: n }, (_, i) => base + (i < remainder ? 1 : 0));
}

// Colors for tiers (ordered by price: most expensive first)
const TIER_COLORS = ["#E8567F", "#7C5CFC", "#38A3A5", "#F59E0B", "#6366F1"];

// Concert / Film arena layout — stage at top, curved sections
function concertLayout(tiers: ApiTicketTier[]): VenueLayout {
  const sorted = [...tiers].sort((a, b) => b.price_cents - a.price_cents);
  const sections: VenueLayout["sections"] = [];

  // Layout templates for up to 3 price tiers (or group extras into last)
  const layouts = [
    // Tier 0 (most expensive) — center front
    {
      templates: [
        { suffix: "center", path: "M 250,200 L 550,200 L 570,300 Q 400,340 230,300 Z", label: { x: 400, y: 260 } },
      ],
    },
    // Tier 1 — sides
    {
      templates: [
        { suffix: "left", path: "M 100,180 L 230,200 L 230,300 Q 180,340 100,320 Z", label: { x: 160, y: 255 } },
        { suffix: "right", path: "M 570,200 L 700,180 L 700,320 Q 620,340 570,300 Z", label: { x: 640, y: 255 } },
      ],
    },
    // Tier 2 (cheapest) — back
    {
      templates: [
        { suffix: "back-center", path: "M 230,310 Q 400,350 570,310 L 600,440 Q 400,480 200,440 Z", label: { x: 400, y: 385 } },
        { suffix: "back-left", path: "M 60,330 L 200,310 L 200,440 Q 140,460 60,440 Z", label: { x: 130, y: 385 } },
        { suffix: "back-right", path: "M 600,310 L 740,330 L 740,440 Q 660,460 600,440 Z", label: { x: 670, y: 385 } },
      ],
    },
  ];

  for (let i = 0; i < sorted.length && i < layouts.length; i++) {
    const tier = sorted[i];
    const layout = layouts[i];
    const sectionCount = layout.templates.length;
    const totals = distribute(tier.total_quantity, sectionCount);
    const avails = distribute(tier.available_quantity, sectionCount);
    const color = TIER_COLORS[i % TIER_COLORS.length];

    for (let j = 0; j < sectionCount; j++) {
      const t = layout.templates[j];
      const sectionId = `${tier.id}:${t.suffix}`;
      sections.push({
        id: sectionId,
        name: `${tier.name} ${sectionCount > 1 ? t.suffix.replace("back-", "").replace(/^\w/, c => c.toUpperCase()) : ""}`.trim(),
        tier: tier.name,
        price: tier.price_cents / 100,
        color,
        path: t.path,
        labelPosition: t.label,
        totalSeats: totals[j],
        availableSeats: avails[j],
        rows: generateRows(sectionId, totals[j], avails[j]),
      });
    }
  }

  return {
    id: "generated-concert",
    name: "Arena",
    sections,
    stage: {
      label: "STAGE",
      shape: "rectangle",
      position: { x: 200, y: 100, width: 400, height: 80 },
    },
  };
}

// Sports arena layout — center court/field, sections around it
function sportsLayout(tiers: ApiTicketTier[], category: string): VenueLayout {
  const sorted = [...tiers].sort((a, b) => b.price_cents - a.price_cents);
  const sections: VenueLayout["sections"] = [];

  const centerLabel = category.toLowerCase() === "sports" ? "FIELD" : "STAGE";

  const layouts = [
    // Tier 0 (most expensive) — courtside strips
    {
      templates: [
        { suffix: "north", path: "M 280,160 L 520,160 L 540,200 L 260,200 Z", label: { x: 400, y: 185 } },
        { suffix: "south", path: "M 260,400 L 540,400 L 520,440 L 280,440 Z", label: { x: 400, y: 420 } },
      ],
    },
    // Tier 1 — lower bowl
    {
      templates: [
        { suffix: "lower-north", path: "M 220,90 L 580,90 L 540,155 L 280,155 Z", label: { x: 400, y: 125 } },
        { suffix: "lower-south", path: "M 280,445 L 540,445 L 580,510 L 220,510 Z", label: { x: 400, y: 478 } },
        { suffix: "lower-west", path: "M 140,120 L 250,160 L 250,440 L 140,480 Z", label: { x: 195, y: 300 } },
        { suffix: "lower-east", path: "M 550,160 L 660,120 L 660,480 L 550,440 Z", label: { x: 605, y: 300 } },
      ],
    },
    // Tier 2 — upper deck
    {
      templates: [
        { suffix: "upper-north", path: "M 160,30 L 640,30 L 580,85 L 220,85 Z", label: { x: 400, y: 60 } },
        { suffix: "upper-south", path: "M 220,515 L 580,515 L 640,570 L 160,570 Z", label: { x: 400, y: 545 } },
      ],
    },
    // Tier 3 (cheapest fallback) — general admission
    {
      templates: [
        { suffix: "section-3", path: "M 60,30 L 140,30 L 140,570 L 60,570 Z", label: { x: 100, y: 300 } },
        { suffix: "section-4", path: "M 660,30 L 740,30 L 740,570 L 660,570 Z", label: { x: 700, y: 300 } },
      ],
    },
  ];

  for (let i = 0; i < sorted.length && i < layouts.length; i++) {
    const tier = sorted[i];
    const layout = layouts[i];
    const sectionCount = layout.templates.length;
    const totals = distribute(tier.total_quantity, sectionCount);
    const avails = distribute(tier.available_quantity, sectionCount);
    const color = TIER_COLORS[i % TIER_COLORS.length];

    for (let j = 0; j < sectionCount; j++) {
      const t = layout.templates[j];
      const sectionId = `${tier.id}:${t.suffix}`;
      const suffixLabel = t.suffix
        .split("-")
        .map((w) => w.charAt(0).toUpperCase() + w.slice(1))
        .join(" ");
      sections.push({
        id: sectionId,
        name: `${tier.name} ${suffixLabel}`,
        tier: tier.name,
        price: tier.price_cents / 100,
        color,
        path: t.path,
        labelPosition: t.label,
        totalSeats: totals[j],
        availableSeats: avails[j],
        rows: generateRows(sectionId, totals[j], avails[j]),
      });
    }
  }

  return {
    id: "generated-sports",
    name: "Stadium",
    sections,
    stage: {
      label: centerLabel,
      shape: "rectangle",
      position: { x: 270, y: 210, width: 260, height: 180 },
    },
  };
}

export function generateVenueLayout(
  tiers: ApiTicketTier[],
  category: string,
): VenueLayout {
  if (!tiers || tiers.length === 0) {
    return concertLayout([{ id: "default", event_id: "", name: "General", price_cents: 5000, total_quantity: 100, available_quantity: 100, version: 0 }]);
  }
  if (category.toLowerCase() === "sports") {
    return sportsLayout(tiers, category);
  }
  return concertLayout(tiers);
}
