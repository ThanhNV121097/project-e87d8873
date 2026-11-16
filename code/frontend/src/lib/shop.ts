/**
 * What this shop's records are, as the pages read them. The fields are the
 * ones src/data/schema.json declares; change both together.
 */
export interface Product {
  id: string;
  name: string;
  category: string;
  price: number;
  /** The price before a reduction; drawn struck through beside the price. */
  was?: number;
  /** The one fact a buyer compares by, under the name: a size, a capacity. */
  fact?: string;
  stock?: number;
  status?: string;
  /** Led with: shown on a tinted card on the home page (the first five). */
  deal?: boolean;
  image?: string;
  image_back?: string;
  image_detail?: string;
  options?: string[];
  material?: string;
  care?: string;
  description?: string;
}

export interface Department {
  id: string;
  name: string;
  image?: string;
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
