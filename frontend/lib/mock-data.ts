import { Event, VenueLayout, SeatRow } from "./types";

// ---------------------------------------------------------------------------
// Helper: generate rows + seats for a section
// ---------------------------------------------------------------------------
function generateRows(
  sectionId: string,
  rowCount: number,
  seatsPerRow: number,
  takenProbability: number = 0.3,
): SeatRow[] {
  return Array.from({ length: rowCount }, (_, ri) => ({
    id: `${sectionId}-row-${ri}`,
    label: `Row ${String.fromCharCode(65 + ri)}`,
    seats: Array.from({ length: seatsPerRow }, (_, si) => ({
      id: `${sectionId}-row-${ri}-seat-${si}`,
      label: `${si + 1}`,
      status:
        Math.random() < takenProbability
          ? (Math.random() < 0.5 ? ("booked" as const) : ("reserved" as const))
          : ("available" as const),
    })),
  }));
}

// ---------------------------------------------------------------------------
// Venue Layout: Concert Arena
// ---------------------------------------------------------------------------
// viewBox 0 0 800 600
// Stage at top center, curved sections wrapping around it.
// ---------------------------------------------------------------------------

function concertArenaLayout(prices: {
  vip: number;
  premium: number;
  ga: number;
}): VenueLayout {
  return {
    id: "concert-arena",
    name: "Concert Arena",
    sections: [
      // VIP Floor — curved section directly in front of stage
      {
        id: "vip-floor",
        name: "VIP Floor",
        tier: "VIP Floor",
        price: prices.vip,
        color: "#E8567F",
        path: "M 250,200 L 550,200 L 570,300 Q 400,340 230,300 Z",
        labelPosition: { x: 400, y: 260 },
        totalSeats: 30,
        availableSeats: 20,
        rows: generateRows("vip-floor", 3, 10, 0.35),
      },
      // Premium Left — side section, left
      {
        id: "premium-left",
        name: "Premium Left",
        tier: "Premium Seated",
        price: prices.premium,
        color: "#7C5CFC",
        path: "M 100,180 L 230,200 L 230,300 Q 180,340 100,320 Z",
        labelPosition: { x: 160, y: 255 },
        totalSeats: 24,
        availableSeats: 16,
        rows: generateRows("premium-left", 3, 8, 0.3),
      },
      // Premium Right — side section, right
      {
        id: "premium-right",
        name: "Premium Right",
        tier: "Premium Seated",
        price: prices.premium,
        color: "#7C5CFC",
        path: "M 570,200 L 700,180 L 700,320 Q 620,340 570,300 Z",
        labelPosition: { x: 640, y: 255 },
        totalSeats: 24,
        availableSeats: 16,
        rows: generateRows("premium-right", 3, 8, 0.3),
      },
      // GA Center — back section behind VIP
      {
        id: "ga-center",
        name: "GA Center",
        tier: "General Admission",
        price: prices.ga,
        color: "#38A3A5",
        path: "M 230,310 Q 400,350 570,310 L 600,440 Q 400,480 200,440 Z",
        labelPosition: { x: 400, y: 385 },
        totalSeats: 40,
        availableSeats: 30,
        rows: generateRows("ga-center", 4, 10, 0.25),
      },
      // GA Left — upper left
      {
        id: "ga-left",
        name: "GA Left",
        tier: "General Admission",
        price: prices.ga,
        color: "#38A3A5",
        path: "M 60,330 L 200,310 L 200,440 Q 140,460 60,440 Z",
        labelPosition: { x: 130, y: 385 },
        totalSeats: 24,
        availableSeats: 20,
        rows: generateRows("ga-left", 3, 8, 0.2),
      },
      // GA Right — upper right
      {
        id: "ga-right",
        name: "GA Right",
        tier: "General Admission",
        price: prices.ga,
        color: "#38A3A5",
        path: "M 600,310 L 740,330 L 740,440 Q 660,460 600,440 Z",
        labelPosition: { x: 670, y: 385 },
        totalSeats: 24,
        availableSeats: 20,
        rows: generateRows("ga-right", 3, 8, 0.2),
      },
    ],
    stage: {
      label: "STAGE",
      shape: "rectangle",
      position: { x: 200, y: 100, width: 400, height: 80 },
    },
  };
}

// ---------------------------------------------------------------------------
// Venue Layout: Sports Arena
// ---------------------------------------------------------------------------
// viewBox 0 0 800 600
// Central court / octagon in the middle, sections radiating outward.
// ---------------------------------------------------------------------------

function sportsArenaLayout(
  prices: { courtside: number; lower: number; upper: number },
  centerLabel: string,
  centerShape: "rectangle" | "circle",
): VenueLayout {
  return {
    id: "sports-arena",
    name: "Sports Arena",
    sections: [
      // Courtside / Cageside — four narrow strips around center
      {
        id: "courtside-north",
        name: `${centerLabel === "OCTAGON" ? "Cageside" : "Courtside"} North`,
        tier: centerLabel === "OCTAGON" ? "Cageside" : "Courtside",
        price: prices.courtside,
        color: "#E8567F",
        path: "M 280,160 L 520,160 L 540,200 L 260,200 Z",
        labelPosition: { x: 400, y: 185 },
        totalSeats: 20,
        availableSeats: 12,
        rows: generateRows("courtside-north", 2, 10, 0.4),
      },
      {
        id: "courtside-south",
        name: `${centerLabel === "OCTAGON" ? "Cageside" : "Courtside"} South`,
        tier: centerLabel === "OCTAGON" ? "Cageside" : "Courtside",
        price: prices.courtside,
        color: "#E8567F",
        path: "M 260,400 L 540,400 L 520,440 L 280,440 Z",
        labelPosition: { x: 400, y: 420 },
        totalSeats: 20,
        availableSeats: 12,
        rows: generateRows("courtside-south", 2, 10, 0.4),
      },
      // Lower Bowl — four sections around the courtside ring
      {
        id: "lower-north",
        name: "Lower Bowl North",
        tier: "Lower Bowl",
        price: prices.lower,
        color: "#7C5CFC",
        path: "M 220,90 L 580,90 L 540,155 L 280,155 Z",
        labelPosition: { x: 400, y: 125 },
        totalSeats: 30,
        availableSeats: 22,
        rows: generateRows("lower-north", 3, 10, 0.25),
      },
      {
        id: "lower-south",
        name: "Lower Bowl South",
        tier: "Lower Bowl",
        price: prices.lower,
        color: "#7C5CFC",
        path: "M 280,445 L 540,445 L 580,510 L 220,510 Z",
        labelPosition: { x: 400, y: 478 },
        totalSeats: 30,
        availableSeats: 22,
        rows: generateRows("lower-south", 3, 10, 0.25),
      },
      {
        id: "lower-west",
        name: "Lower Bowl West",
        tier: "Lower Bowl",
        price: prices.lower,
        color: "#7C5CFC",
        path: "M 140,120 L 250,160 L 250,440 L 140,480 Z",
        labelPosition: { x: 195, y: 300 },
        totalSeats: 32,
        availableSeats: 24,
        rows: generateRows("lower-west", 4, 8, 0.25),
      },
      {
        id: "lower-east",
        name: "Lower Bowl East",
        tier: "Lower Bowl",
        price: prices.lower,
        color: "#7C5CFC",
        path: "M 550,160 L 660,120 L 660,480 L 550,440 Z",
        labelPosition: { x: 605, y: 300 },
        totalSeats: 32,
        availableSeats: 24,
        rows: generateRows("lower-east", 4, 8, 0.25),
      },
      // Upper Deck — two large sections top and bottom
      {
        id: "upper-north",
        name: "Upper Deck North",
        tier: "Upper Deck",
        price: prices.upper,
        color: "#38A3A5",
        path: "M 160,30 L 640,30 L 580,85 L 220,85 Z",
        labelPosition: { x: 400, y: 60 },
        totalSeats: 36,
        availableSeats: 30,
        rows: generateRows("upper-north", 3, 12, 0.18),
      },
      {
        id: "upper-south",
        name: "Upper Deck South",
        tier: "Upper Deck",
        price: prices.upper,
        color: "#38A3A5",
        path: "M 220,515 L 580,515 L 640,570 L 160,570 Z",
        labelPosition: { x: 400, y: 545 },
        totalSeats: 36,
        availableSeats: 30,
        rows: generateRows("upper-south", 3, 12, 0.18),
      },
    ],
    stage: {
      label: centerLabel,
      shape: centerShape,
      position:
        centerShape === "circle"
          ? { x: 300, y: 230, width: 200, height: 140 }
          : { x: 270, y: 210, width: 260, height: 180 },
    },
  };
}

// ---------------------------------------------------------------------------
// Events
// ---------------------------------------------------------------------------

export const events: Event[] = [
  {
    id: "weeknd-after-hours",
    title: "The Weeknd — After Hours World Tour",
    date: "Jul 15, 2026",
    time: "8:00 PM",
    venue: "Madison Square Garden",
    location: "New York",
    category: "Concerts",
    genre: "Pop / R&B",
    description:
      "Experience The Weeknd live on his After Hours World Tour. This concert promises an unforgettable night of music, stunning visuals, and the raw energy that has made him one of the biggest artists of his generation.",
    image: "https://images.unsplash.com/photo-1470229722913-7c0e2dbbafd3?w=800&q=80",
    featured: true,
    tiers: [
      { id: "weeknd-vip", name: "VIP Floor", description: "Front stage + meet & greet", price: 350, available: 5 },
      { id: "weeknd-premium", name: "Premium Seated", description: "Sections A-E, ground floor", price: 185, available: 20 },
      { id: "weeknd-ga", name: "General Admission", description: "Standing, upper sections", price: 130, available: 100 },
    ],
    venueLayout: concertArenaLayout({ vip: 350, premium: 185, ga: 130 }),
  },
  {
    id: "billie-eilish-tour",
    title: "Billie Eilish — Hit Me Hard and Soft Tour",
    date: "Apr 8, 2026",
    time: "7:30 PM",
    venue: "The O2 Arena",
    location: "London",
    category: "Concerts",
    genre: "Pop / Alternative",
    description:
      "Experience the magic of Billie Eilish live on her Hit Me Hard and Soft World Tour. This concert promises an unforgettable night of music, stunning visuals, and the raw energy that has made Billie one of the biggest artists of her generation.",
    image: "https://images.unsplash.com/photo-1493225457124-a3eb161ffa5f?w=800&q=80",
    tiers: [
      { id: "billie-vip", name: "VIP Floor", description: "First 5 rows + meet & greet", price: 350, available: 3 },
      { id: "billie-premium", name: "Premium Seated", description: "Sections A-E, ground floor", price: 185, available: 15 },
      { id: "billie-ga", name: "General Admission", description: "Standing, upper sections", price: 85, available: 80 },
    ],
    venueLayout: concertArenaLayout({ vip: 350, premium: 185, ga: 85 }),
  },
  {
    id: "nba-finals-game3",
    title: "NBA Finals 2026 — Game 3",
    date: "Jun 12, 2026",
    time: "6:00 PM",
    venue: "Chase Center",
    location: "San Francisco",
    category: "Sports",
    genre: "Basketball",
    description:
      "Watch the NBA Finals Game 3 live at Chase Center. Experience the intensity and excitement of professional basketball at its highest level as the top two teams battle for the championship.",
    image: "https://images.unsplash.com/photo-1504450758481-7338bbe75c8e?w=800&q=80",
    tiers: [
      { id: "nba-courtside", name: "Courtside", description: "Row 1, Seat D-1 Ticket", price: 500, available: 2 },
      { id: "nba-lower", name: "Lower Bowl", description: "Sections 101-120", price: 250, available: 30 },
      { id: "nba-upper", name: "Upper Deck", description: "Sections 201-230", price: 120, available: 150 },
    ],
    venueLayout: sportsArenaLayout(
      { courtside: 500, lower: 250, upper: 120 },
      "COURT",
      "rectangle",
    ),
  },
  {
    id: "ufc-310",
    title: "UFC 310 — Championship",
    date: "May 3, 2026",
    time: "7:00 PM",
    venue: "T-Mobile Arena",
    location: "Las Vegas",
    category: "Sports",
    genre: "MMA",
    description:
      "UFC 310 features a championship main event at T-Mobile Arena in Las Vegas. Witness the world's best fighters compete in the octagon for the ultimate prize in mixed martial arts.",
    image: "https://images.unsplash.com/photo-1579882392879-e04b8512a65e?w=800&q=80",
    tiers: [
      { id: "ufc-cageside", name: "Cageside", description: "Rows 1-3, premium view", price: 400, available: 4 },
      { id: "ufc-lower", name: "Lower Bowl", description: "Sections 1-10", price: 200, available: 25 },
      { id: "ufc-ga", name: "General Admission", description: "Upper sections", price: 95, available: 120 },
    ],
    venueLayout: sportsArenaLayout(
      { courtside: 400, lower: 200, upper: 95 },
      "OCTAGON",
      "circle",
    ),
  },
];
