"use client";

import { useState } from "react";
import { VenueLayout } from "@/lib/types";
import { SeatTooltip } from "@/components/seat-tooltip";

interface VenueMapProps {
  layout: VenueLayout;
  onSectionClick: (sectionId: string) => void;
  selectedSectionId?: string;
}

export function VenueMap({ layout, onSectionClick, selectedSectionId }: VenueMapProps) {
  const [hoveredSectionId, setHoveredSectionId] = useState<string | null>(null);
  const [tooltipPos, setTooltipPos] = useState({ x: 0, y: 0 });

  const hoveredSection = layout.sections.find((s) => s.id === hoveredSectionId);

  function handleMouseMove(e: React.MouseEvent) {
    const rect = e.currentTarget.getBoundingClientRect();
    setTooltipPos({
      x: e.clientX - rect.left,
      y: e.clientY - rect.top,
    });
  }

  return (
    <div className="relative w-full" onMouseMove={handleMouseMove}>
      <svg
        viewBox="0 0 800 600"
        className="w-full h-auto"
        xmlns="http://www.w3.org/2000/svg"
      >
        {/* Stage */}
        {layout.stage && (
          <g>
            {layout.stage.shape === "circle" ? (
              <ellipse
                cx={layout.stage.position.x + layout.stage.position.width / 2}
                cy={layout.stage.position.y + layout.stage.position.height / 2}
                rx={layout.stage.position.width / 2}
                ry={layout.stage.position.height / 2}
                fill="#374151"
              />
            ) : (
              <rect
                x={layout.stage.position.x}
                y={layout.stage.position.y}
                width={layout.stage.position.width}
                height={layout.stage.position.height}
                rx={8}
                fill="#374151"
              />
            )}
            <text
              x={layout.stage.position.x + layout.stage.position.width / 2}
              y={layout.stage.position.y + layout.stage.position.height / 2}
              textAnchor="middle"
              dominantBaseline="central"
              fill="white"
              fontSize={14}
              fontWeight={600}
            >
              {layout.stage.label}
            </text>
          </g>
        )}

        {/* Sections */}
        {layout.sections.map((section) => {
          const isUnavailable = section.availableSeats === 0;
          const isSelected = section.id === selectedSectionId;
          const isHovered = section.id === hoveredSectionId;

          return (
            <g key={section.id}>
              <path
                d={section.path}
                fill={isUnavailable ? "#D1D5DB" : section.color}
                opacity={isHovered && !isUnavailable ? 0.85 : 1}
                stroke={isSelected ? "white" : "transparent"}
                strokeWidth={isSelected ? 2 : 0}
                cursor={isUnavailable ? "default" : "pointer"}
                className="transition-opacity duration-150"
                onClick={() => {
                  if (!isUnavailable) onSectionClick(section.id);
                }}
                onMouseEnter={() => {
                  if (!isUnavailable) setHoveredSectionId(section.id);
                }}
                onMouseLeave={() => setHoveredSectionId(null)}
              />
              <text
                x={section.labelPosition.x}
                y={section.labelPosition.y}
                textAnchor="middle"
                dominantBaseline="central"
                fill="white"
                fontSize={11}
                fontWeight={600}
                pointerEvents="none"
              >
                {section.name}
              </text>
              <text
                x={section.labelPosition.x}
                y={section.labelPosition.y + 14}
                textAnchor="middle"
                dominantBaseline="central"
                fill="white"
                fontSize={9}
                opacity={0.85}
                pointerEvents="none"
              >
                {isUnavailable ? "Sold out" : `${section.availableSeats} left`}
              </text>
            </g>
          );
        })}
      </svg>

      <SeatTooltip
        content={{
          title: hoveredSection?.name ?? "",
          subtitle: hoveredSection ? `${hoveredSection.tier} tier` : undefined,
          price: hoveredSection?.price,
        }}
        position={tooltipPos}
        visible={!!hoveredSection}
      />
    </div>
  );
}
