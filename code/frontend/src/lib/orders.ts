import type { CartLine } from "./cart";

/**
 * A placed order: what the SERVER stored and answered (POST /v1/orders,
 * docs/architecture/services.md §3.10), kept on this browser so the
 * confirmation page can show it. The browser never makes a number, a price or
 * a total of its own — everything here is the server's answer.
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

/** What the checkout sends: who, how, and the bikes as id, size and count — no money. */
interface OrderRequest {
  name: string;
  phone: string;
  delivery: Order["delivery"];
  address: string;
  payment: Order["payment"];
  note: string;
  /** The hidden field, as it stands: a bot fills it, a person does not. */
  website: string;
  lines: { id: string; variant?: string; qty: number }[];
}

/** Said for every outcome the server did not explain in a sentence of its own. */
const COULD_NOT_SEND = "The order could not be sent. Please try again.";

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

/** The address of the backend: the environment's VITE_API_URL, else /api (the edge strips it). */
function apiBase(): string {
  const set = import.meta.env.VITE_API_URL as string | undefined;
  return (set ?? "/api").replace(/\/$/, "");
}

function isOrder(v: unknown): v is Order {
  if (typeof v !== "object" || v === null) return false;
  const o = v as Partial<Order>;
  return typeof o.number === "string" && typeof o.placed === "string" && Array.isArray(o.lines) && typeof o.total === "number";
}

/** The server's own sentence from a refusal, or the generic one when it gave none. */
function sentence(body: unknown): string {
  const said = typeof body === "object" && body !== null ? (body as { error?: unknown }).error : undefined;
  return typeof said === "string" && said.trim() !== "" ? said : COULD_NOT_SEND;
}

/**
 * Sends the order to the server and says what came of it. Only a 201 is an
 * order (it carries the server's own); a 400 or 409 brings the server's
 * sentence; every other outcome — no network, a 5xx, a 413, a 202 (the
 * server dropped it as a bot), a body that cannot be read — is the generic
 * sentence. It never throws.
 */
export async function placeOrder(request: OrderRequest): Promise<{ ok: true; order: Order } | { ok: false; problem: string }> {
  try {
    const res = await fetch(`${apiBase()}/v1/orders`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(request),
    });
    const body: unknown = await res.json().catch(() => null);
    if (res.status === 201 && isOrder(body)) return { ok: true, order: body };
    return { ok: false, problem: res.status === 400 || res.status === 409 ? sentence(body) : COULD_NOT_SEND };
  } catch {
    return { ok: false, problem: COULD_NOT_SEND };
  }
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
