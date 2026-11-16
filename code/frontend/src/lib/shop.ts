/**
 * What this shop's records are, as the pages read them. The fields are the
 * ones src/data/schema.json declares; change both together.
 */
export interface Product {
  id: string;
  name: string;
  brand: string;
  category: string;
  year?: number;
  price: number;
  /** The price before a reduction; drawn struck through beside the price. */
  was?: number;
  /** Wheel, travel and gears: the line a rider compares by. */
  fact?: string;
  stock?: number;
  status?: string;
  /** Led with: shown on a tinted card on the home page (the first three). */
  deal?: boolean;
  /** The bike of the month on the home page. */
  month?: boolean;
  sizes?: string[];
  sizes_out?: string[];
  /** Rider height for each size, in the order of `sizes`. */
  size_fit?: string[];
  wheel?: string;
  travel?: string;
  frame?: string;
  fork?: string;
  shock?: string;
  motor?: string;
  drivetrain?: string;
  brakes?: string;
  wheels?: string;
  tyres?: string;
  weight?: string;
  colour?: string;
  rating?: number;
  review_count?: number;
  description?: string;
  ride?: string;
  image?: string;
  image_back?: string;
  image_detail?: string;
}

export interface Department {
  id: string;
  name: string;
  line?: string;
  terrain?: string;
}

export interface Review {
  id: string;
  name: string;
  rating: number;
  date: string;
  bike?: string;
  text: string;
}

export interface Member {
  id: string;
  name: string;
  role: string;
  line: string;
  rides?: string;
}

export interface Shot {
  id: string;
  image: string;
  title: string;
  caption?: string;
  type?: string;
}

export interface Service {
  id: string;
  name: string;
  line?: string;
  time?: string;
  price: number;
}

export interface Faq {
  id: string;
  q: string;
  a: string;
}

export interface ShopInfo {
  id: string;
  founded: number;
  bikes_built: number;
  rating: number;
  review_count: number;
  reply_hours: number;
  warranty_years: number;
  free_service_months: number;
  delivery_local: string;
  delivery_nationwide: string;
  delivery_fee: number;
  pickup: string;
  payment: string[];
  confirm: string;
  stock_rule: string;
}

/** One link of the header, as content.json holds it. */
export interface NavItem {
  label: string;
  href: string;
}

/** One of the three reassurances: a bold label over a line. */
export interface Feature {
  light: string;
  bold: string;
}

/** Every picture a product has, the main one first. */
export function picturesOf(p: Product): string[] {
  return [p.image, p.image_back, p.image_detail].filter((s): s is string => Boolean(s));
}

/** Whether a bike can be ordered at all. */
export function inStock(p: Product): boolean {
  return (p.stock ?? 0) > 0;
}

/** The sizes a bike comes in, each with the rider height it fits and whether it is in stock. */
export function sizesOf(p: Product): { size: string; fit: string; out: boolean }[] {
  const out = p.sizes_out ?? [];
  return (p.sizes ?? []).map((size, i) => ({ size, fit: p.size_fit?.[i] ?? "", out: !inStock(p) || out.includes(size) }));
}

/** The price bands the catalogue is narrowed by, in đồng. */
export const PRICE_BANDS: readonly { id: string; from?: number; to?: number }[] = [
  { id: "under-25", to: 25000000 },
  { id: "25-70", from: 25000000, to: 70000000 },
  { id: "70-100", from: 70000000, to: 100000000 },
  { id: "over-100", from: 100000000 },
];

export function inBand(price: number, band: string): boolean {
  const b = PRICE_BANDS.find((x) => x.id === band);
  if (!b) return true;
  return (b.from === undefined || price >= b.from) && (b.to === undefined || price < b.to);
}
