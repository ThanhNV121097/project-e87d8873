import { useState, type FormEvent, type ReactNode } from "react";
import { useNavigate } from "react-router-dom";
import { T, useContent } from "../editable";
import { clearCart, type CartLine } from "../lib/cart";
import { formatPrice } from "../lib/format";
import { newOrderNumber, saveOrder, type Order } from "../lib/mock/orders";
import { canSend, sendEnquiry } from "../site";
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

/**
 * THE ORDER, SENT. Who it is for, how it reaches them and how it is paid —
 * each choice a rounded card that fills with the tint when chosen, never a
 * box to type into. What happens next is said above the button. The order
 * goes to the owner as one message (src/site.ts, the same inbox the contact
 * form reaches); nothing is charged on this page — the shop calls first.
 */
export default function OrderForm({ lines, subtotal, fee, delivery, onDelivery, payment, onPayment, className = "" }: {
  lines: CartLine[];
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

  async function submit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    const data = new FormData(e.currentTarget);
    const value = (k: string) => String(data.get(k) ?? "").trim();
    const order: Order = {
      number: newOrderNumber(), placed: new Date().toISOString(), name: value("name"), phone: value("phone"),
      delivery, address: value("address"), payment, lines, subtotal, delivery_fee: fee, total: subtotal + fee,
    };
    setState("sending");
    try {
      await sendEnquiry({
        name: order.name, reply_to: order.phone, website: value("website"),
        message: [lines.map((l) => `${l.qty} × ${l.name}${l.variant ? ` (size ${l.variant})` : ""} — ${formatPrice(l.price * l.qty)}`).join("\n"), value("note")].filter(Boolean).join("\n\n"),
        details: { order: order.number, delivery, address: order.address, payment, total: formatPrice(order.total) },
      });
      saveOrder(order);
      clearCart();
      go("/checkout/done");
    } catch (err) {
      setProblem(err instanceof Error ? err.message : "");
      setState("failed");
    }
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
              icon={d.icon} title={`checkout.delivery.${d.id}.title`} line={`checkout.delivery.${d.id}.line`} testid={`deliver-${d.id}`}
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
              icon={p.icon} title={`checkout.payment.${p.id}.title`} line={`checkout.payment.${p.id}.line`} testid={`pay-${p.id}`} />
          ))}
        </div>
        <div className="mt-4">
          <Field k="checkout.note"><textarea id="order-note" name="note" rows={2} className={input} /></Field>
        </div>
      </Group>

      <input name="website" tabIndex={-1} autoComplete="off" aria-hidden="true" className="absolute -left-[9999px] h-px w-px opacity-0" />

      <div className="flex flex-col gap-4 rounded bg-[var(--tint-2)] p-6">
        <p className="flex items-start gap-3 text-sm text-ink"><Icon name="phone" size={18} /><T k="checkout.after" /></p>
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
  on: boolean; onPick: () => void; icon: IconName; title: string; line: string; testid: string; extra?: string; disabled?: boolean;
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
      <T k={line} className="text-xs leading-relaxed text-ink-soft" />
      {extra && <span className="text-xs font-semibold tabular-nums text-ink">{extra}</span>}
    </button>
  );
}
