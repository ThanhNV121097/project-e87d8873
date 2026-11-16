import type { CartLine } from "../cart";

/**
 * The design's mock of a placed order: what the checkout sent, kept on this
 * browser so the confirmation page can show it. The backend story replaces
 * this module with the order the server stored; no page changes.
 */
export interface Order {
  number: string;
  placed: string;
  name: string;
  phone: string;
  delivery: "local" | "courier" | "pickup";
  address: string;
  payment: "transfer" | "cod" | "shop";
  lines: CartLine[];
  subtotal: number;
  delivery_fee: number;
  total: number;
}

const KEY = "deo:last-order";

export function saveOrder(order: Order): void {
  try {
    sessionStorage.setItem(KEY, JSON.stringify(order));
  } catch {
    // Storage refused: the confirmation page shows its empty state.
  }
}

export function lastOrder(): Order | null {
  try {
    const raw = sessionStorage.getItem(KEY);
    return raw ? (JSON.parse(raw) as Order) : null;
  } catch {
    return null;
  }
}

/** An order number the shop can read out on the phone: DÈO-261116-4821. */
export function newOrderNumber(now = new Date()): string {
  const d = now.toISOString().slice(2, 10).replace(/-/g, "");
  return `DEO-${d}-${Math.floor(1000 + Math.random() * 9000)}`;
}

/** What the Studio, the walk and the owner's preview see on the confirmation page. */
export const EXAMPLE_ORDER: Order = {
  number: "DEO-261116-4821",
  placed: "2026-11-16T09:40:00",
  name: "Minh",
  phone: "09•• ••• •••",
  delivery: "local",
  address: "Delivered in town, built and set up",
  payment: "transfer",
  lines: [
    { id: "stance-0", name: "Stance 0", price: 66490000, image: "/images/ref/stance-0.jpg", variant: "M", qty: 1 },
  ],
  subtotal: 66490000,
  delivery_fee: 0,
  total: 66490000,
};

/** What the bag and the checkout show while being looked at, before anything is chosen. */
export const EXAMPLE_LINES: CartLine[] = [
  { id: "stance-0", name: "Stance 0", price: 66490000, image: "/images/ref/stance-0.jpg", variant: "M", qty: 1 },
  { id: "talon-2", name: "Talon 2", price: 18990000, image: "/images/ref/talon-2.jpg", variant: "S", qty: 1 },
];
