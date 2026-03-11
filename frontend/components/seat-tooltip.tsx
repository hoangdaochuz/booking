"use client";

interface SeatTooltipProps {
  content: {
    title: string;
    subtitle?: string;
    price?: number;
  };
  position: { x: number; y: number };
  visible: boolean;
}

export function SeatTooltip({ content, position, visible }: SeatTooltipProps) {
  if (!visible) return null;

  return (
    <div
      className="absolute z-50 pointer-events-none px-3 py-2 rounded-lg shadow-lg"
      style={{
        left: position.x + 12,
        top: position.y - 8,
        backgroundColor: "#1F2937",
      }}
    >
      <p className="text-white text-sm font-bold leading-tight">{content.title}</p>
      {content.subtitle && (
        <p className="text-gray-400 text-xs mt-0.5">{content.subtitle}</p>
      )}
      {content.price !== undefined && (
        <p className="text-xs font-semibold mt-1" style={{ color: "#FF8400" }}>
          ${content.price.toFixed(2)}
        </p>
      )}
    </div>
  );
}
