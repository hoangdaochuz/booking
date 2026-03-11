"use client";

interface SeatLegendItem {
  label: string;
  color: string;
  border?: string;
}

interface SeatLegendProps {
  items: SeatLegendItem[];
}

export default function SeatLegend({ items }: SeatLegendProps) {
  return (
    <div className="flex flex-row items-center gap-4">
      {items.map((item) => (
        <div key={item.label} className="flex items-center gap-1.5">
          <span
            className="inline-block rounded-full"
            style={{
              width: 12,
              height: 12,
              backgroundColor: item.color,
              border: item.border ? `2px solid ${item.border}` : undefined,
            }}
          />
          <span className="text-xs text-muted">{item.label}</span>
        </div>
      ))}
    </div>
  );
}
