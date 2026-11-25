import { useRef, useState, type FormEvent, type ReactNode } from "react";
import { useNavigate } from "react-router-dom";
import { T, useContent } from "../editable";
import { clearCart, type CartLine } from "../lib/cart";
import { formatPrice } from "../lib/format";
import { placeOrder, saveOrder, type Order } from "../lib/orders";
import type { ShopInfo } from "../lib/shop";
import { canSend, sendEnquiry, type Enquiry } from "../site";
import Icon, { type IconName } from "./Icon";

type State = "idle" | "sending" | "failed";
export type Delivery = Order["delivery"];
export type Payment = Order["payment"];

const DELIVERIES: { id: Delivery; icon: IconName }[] = [
  { id: "local", icon: "shipping" }, { id: "courier", icon: "package" }, { id: "pickup", icon: "store" },
];
const PAYMENTS: { id: Payment; icon: IconName }[] = [
  { id: "transfer", icon: "qr" }, { id: "cod", icon: "cash" }, { id: "shop", icon: "card" },
];

/** How long the owner's message may hold the customer here once the order is stored. */
const OWNER_MESSAGE_WAIT_MS = 10_000;

/** What the form sends: the words and choices and the bikes as id, size and count — never a price. */
function requestFrom(form: HTMLFormElement, delivery: Delivery, payment: Payment, lines: CartLine[]) {
  const data = new FormData(form);
  const value = (k: string) => String(data.get(k) ?? "").trim();
  return {
    name: value("name"), phone: value("phone"), delivery, address: delivery === "pickup" ? "" : value("address"),
    payment, note: value("note"), website: value("website"),
    lines: lines.map((l) => ({ id: l.id, variant: l.variant, qty: l.qty })),
  };
}

/** The owner's message, from what the server stored; the note is the form's, the answer has none. */
function ownerMessage(order: Order, note: string): Enquiry {
  const bikes = order.lines.map((l) => `${l.qty} × ${l.name}${l.variant ? ` (size ${l.variant})` : ""} — ${formatPrice(l.price * l.qty)}`).join("\n");
  return {
    name: order.name, reply_to: order.phone, message: [bikes, note].filter(Boolean).join("\n\n"),
    details: { order: order.number, delivery: order.delivery, address: order.address, payment: order.payment, total: formatPrice(order.total) },
  };
}

/** Tells the owner, waiting at most OWNER_MESSAGE_WAIT_MS. A message that fails or is slow never touches the stored order. */
async function tellOwner(order: Order, note: string): Promise<void> {
  if (!canSend) return;
  const sent = sendEnquiry(ownerMessage(order, note)).catch(() => undefined);
  let timer: ReturnType<typeof setTimeout> | undefined;
  const late = new Promise<void>((resolve) => { timer = setTimeout(resolve, OWNER_MESSAGE_WAIT_MS); });
  await Promise.race([sent, late]);
  clearTimeout(timer);
}

/**
 * THE ORDER, SENT. Who it is for, how it reaches them and how it is paid —
 * each choice a rounded card that fills with the tint when chosen, never a
 * box to type into. What happens next is said above the button. The order
 * goes to the server first, which prices it, holds the bikes and numbers it
 * (src/lib/orders.ts); only once it is stored does the owner get one message
 * (src/site.ts, the same inbox the contact form reaches). Nothing is charged
 * on this page — the shop calls first.
 */
export default function OrderForm({ lines, subtotal, fee, delivery, onDelivery, payment, onPayment, terms, className = "" }: {
  lines: CartLine[];
  /** The shop's own delivery and confirming terms (src/data/shop.json). */
  terms?: ShopInfo;
  subtotal: number;
  fee: number;
  delivery: Delivery;
  onDelivery: (d: Delivery) => void;
  payment: Payment;
  onPayment: (p: Payment) => void;
  className?: string;
}) {
  const [state, setState] = useState<State>("idle");
  const [problem, setProblem] = useState("");
  const offline = useContent("contact.form.offline");
  const go = useNavigate();
  const busy = useRef(false);

  async function submit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    if (busy.current) return;
    busy.current = true;
    setProblem("");
    setState("sending");
    const request = requestFrom(e.currentTarget, delivery, payment, lines);
    const placed = await placeOrder(request);
    if (!placed.ok) {
      setProblem(placed.problem);
      setState("failed");
      busy.current = false;
      return;
    }
    saveOrder(placed.order);
    await tellOwner(placed.order, request.note);
    clearCart();
    go("/checkout/done");
  }

  const input = "w-full appearance-none rounded-[var(--radius-sm)] border border-field bg-ground px-4 py-3 text-ink";
  return (
    <form onSubmit={(e) => void submit(e)} className={`flex flex-col gap-10 ${className}`}>
      <Group n="1" k="checkout.who">
        <div className="grid gap-4 sm:grid-cols-2">
          <Field k="checkout.name"><input id="order-name" name="name" autoComplete="name" required className={input} /></Field>
          <Field k="checkout.phone"><input id="order-phone" name="phone" type="tel" autoComplete="tel" required className={input} /></Field>
        </div>
      </Group>

      <Group n="2" k="checkout.how">
        <div className="grid gap-3 sm:grid-cols-3" role="radiogroup">
          {DELIVERIES.map((d) => (
            <Option key={d.id} on={delivery === d.id} onPick={() => { onDelivery(d.id); if (d.id !== "local" && payment === "cod") onPayment("transfer"); }}
              icon={d.icon} title={`checkout.delivery.${d.id}.title`} testid={`deliver-${d.id}`}
              line={d.id === "local" ? terms?.delivery_local : d.id === "courier" ? terms?.delivery_nationwide : terms?.pickup}
              extra={d.id === "courier" && fee > 0 ? formatPrice(fee) : undefined} />
          ))}
        </div>
        {delivery !== "pickup" && (
          <div className="mt-4">
            <Field k="checkout.address"><input id="order-address" name="address" autoComplete="street-address" required className={input} /></Field>
          </div>
        )}
      </Group>

      <Group n="3" k="checkout.pay">
        <div className="grid gap-3 sm:grid-cols-3" role="radiogroup">
          {PAYMENTS.map((p) => (
            <Option key={p.id} on={payment === p.id} disabled={p.id === "cod" && delivery !== "local"} onPick={() => onPayment(p.id)}
              icon={p.icon} title={`checkout.payment.${p.id}.title`} line={<T k={`checkout.payment.${p.id}.line`} />} testid={`pay-${p.id}`} />
          ))}
        </div>
        <div className="mt-4">
          <Field k="checkout.note"><textarea id="order-note" name="note" rows={2} className={input} /></Field>
        </div>
      </Group>

      <input name="website" tabIndex={-1} autoComplete="off" aria-hidden="true" className="absolute -left-[9999px] h-px w-px opacity-0" />

      <div className="flex flex-col gap-4 rounded bg-[var(--tint-2)] p-6">
        {terms && <p className="flex items-start gap-3 text-sm text-ink"><Icon name="phone" size={18} /><span>{terms.confirm}</span></p>}
        {state === "failed" && <p role="alert" className="text-sm text-danger">{problem}</p>}
        {!canSend && <p className="text-sm text-ink-soft">{offline}</p>}
        <button type="submit" disabled={state === "sending" || lines.length === 0} data-testid="place-order"
          className="inline-flex items-center justify-center gap-2 self-start whitespace-nowrap rounded-full bg-accent px-8 py-3.5 font-medium text-accent-ink disabled:opacity-60">
          <T k={state === "sending" ? "contact.form.sending" : "checkout.place"} />
          <span className="tabular-nums">· {formatPrice(subtotal + fee)}</span>
        </button>
      </div>
    </form>
  );
}

function Group({ n, k, children }: { n: string; k: string; children: ReactNode }) {
  return (
    <fieldset>
      <legend className="mb-4 flex items-center gap-3 text-lg text-ink">
        <span className="grid h-7 w-7 place-items-center rounded-full bg-accent text-xs font-semibold tabular-nums text-accent-ink">{n}</span>
        <T k={k} className="font-semibold" />
      </legend>
      {children}
    </fieldset>
  );
}

function Field({ k, children }: { k: string; children: ReactNode }) {
  return (
    <label className="flex flex-col gap-2 text-sm">
      <T k={k} className="text-ink-soft" />
      {children}
    </label>
  );
}

function Option({ on, onPick, icon, title, line, testid, extra, disabled = false }: {
  on: boolean; onPick: () => void; icon: IconName; title: string; line?: ReactNode; testid: string; extra?: string; disabled?: boolean;
}) {
  return (
    <button type="button" role="radio" aria-checked={on} disabled={disabled} onClick={onPick} data-testid={testid}
      className={`flex h-full flex-col items-start gap-2 rounded border p-4 text-left ${on ? "border-ink bg-[var(--tint-1)]" : "border-field bg-ground hover:border-ink"} disabled:cursor-not-allowed disabled:opacity-50`}>
      <span className="flex w-full items-center justify-between">
        <Icon name={icon} size={20} />
        <span className={`grid h-5 w-5 place-items-center rounded-full border ${on ? "border-ink bg-accent text-accent-ink" : "border-field"}`}>
          {on && <Icon name="check" size={12} />}
        </span>
      </span>
      <T k={title} className="text-sm font-semibold text-ink" />
      {line && <span className="text-xs leading-relaxed text-ink-soft">{line}</span>}
      {extra && <span className="text-xs font-semibold tabular-nums text-ink">{extra}</span>}
    </button>
  );
}
