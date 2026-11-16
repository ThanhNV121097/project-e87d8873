import { Link } from "react-router-dom";
import { T } from "../editable";
import { usePageMeta } from "../seo";
import { useCollection } from "../data";
import { useDemo } from "../lib/demo";
import { formatDateTime, formatPrice } from "../lib/format";
import { EXAMPLE_ORDER, lastOrder } from "../lib/mock/orders";
import type { ShopInfo } from "../lib/shop";
import Bleed from "../components/Bleed";
import Icon, { type IconName } from "../components/Icon";
import ProductImage from "../components/ProductImage";
import Weights from "../components/Weights";

const STEPS: IconName[] = ["done", "phone", "settings", "shipping"];

/**
 * AFTER THE ORDER: the answer first and largest — the order is in — with its
 * number; then the four steps on one line on the shop's navy, run to the
 * screen's edge, where it is now lit in butter; then what was ordered with its picture, where it goes and how it is
 * paid, and who to ask.
 */
export default function CheckoutDone() {
  usePageMeta("pages.done.title", "pages.done.description");
  const demo = useDemo();
  const order = lastOrder() ?? (demo ? EXAMPLE_ORDER : null);
  const shop = useCollection<ShopInfo>("shop", { pageSize: 1 }).items[0];
  if (!order) {
    return (
      <div className="pb-8">
        {/* Nothing ordered on this browser yet: said once, on the shop's navy
            run to both screen edges, with the one way on. */}
        <Bleed className="flex flex-col items-start gap-6 bg-accent py-20 text-accent-ink">
          <h1 className="text-[length:var(--size-h1)] leading-[1.1] text-accent-ink"><Weights k="pages.done.none" /></h1>
          <T k="pages.done.none_line" as="p" className="max-w-[52ch] text-sm text-accent-ink/80" />
          <Link to="/shop" className="rounded-full bg-ground px-7 py-3 text-sm font-medium text-ink hover:opacity-90"><T k="pages.bag.browse" /></Link>
        </Bleed>
      </div>
    );
  }
  return (
    <div className="pb-8 pt-12">
      <div className="flex flex-col gap-6 lg:flex-row lg:items-end lg:justify-between">
        <div>
          <p className="mb-4 inline-flex items-center gap-2 rounded-full bg-[var(--tint-2)] px-4 py-1.5 text-sm text-ink"><Icon name="done" size={16} /><T k="pages.done.badge" /></p>
          <h1 data-testid="order-confirmed" className="text-[length:var(--size-h1)] leading-[1.1] text-ink"><Weights k="pages.done.heading" /></h1>
          <p className="mt-4 max-w-[56ch] text-sm leading-relaxed text-ink-soft">
            <T k="pages.done.call_a" /> <span className="font-semibold text-ink">{order.name}</span>{" "}
            <T k="pages.done.call_b" /> <span className="tabular-nums">{shop?.reply_hours ?? 2}</span> <T k="pages.done.call_c" />
          </p>
        </div>
        <dl className="shrink-0 text-sm lg:text-right">
          <dt className="text-ink-soft"><T k="pages.done.number" /></dt>
          <dd className="mt-1 text-2xl font-bold tabular-nums tracking-tight text-ink">{order.number}</dd>
          <dd className="mt-1 tabular-nums text-ink-soft">{formatDateTime(order.placed)}</dd>
        </dl>
      </div>

      <ol className="-mx-[var(--gutter)] mt-12 grid gap-6 bg-accent p-6 sm:mx-0 sm:grid-cols-4 sm:rounded lg:-mr-[calc((100vw-min(100vw,var(--page)))/2+var(--gutter))] lg:rounded-r-none lg:p-10 lg:pr-[calc((100vw-min(100vw,var(--page)))/2+var(--gutter)+2.5rem)]">
        {STEPS.map((icon, i) => (
          <li key={icon} className="flex items-start gap-3 sm:flex-col">
            <span className={`grid h-11 w-11 shrink-0 place-items-center rounded-full ${i === 0 ? "bg-[var(--tint-2)] text-ink" : "border border-accent-ink/40 text-accent-ink"}`}><Icon name={icon} size={19} /></span>
            <span>
              <T k={`pages.done.steps.${i}.title`} as="span" className="block text-sm font-semibold text-accent-ink" />
              <T k={`pages.done.steps.${i}.line`} as="span" className="mt-1 block text-xs leading-relaxed text-accent-ink/80" />
            </span>
          </li>
        ))}
      </ol>

      <section className="mt-12 grid gap-10 lg:grid-cols-12">
        <div className="lg:col-span-7">
          <T k="pages.done.ordered" as="h2" className="mb-5 block text-lg font-semibold text-ink" />
          <ul className="divide-y divide-line border-y border-line">
            {order.lines.map((l) => (
              <li key={`${l.id}-${l.variant ?? ""}`} className="flex items-center gap-5 py-4">
                <span className="w-32 shrink-0"><ProductImage src={l.image} alt={l.name} aspect="4/3" className="border border-line" /></span>
                <span className="min-w-0 flex-1 text-sm">
                  <span className="block font-semibold text-ink">{l.name}</span>
                  <span className="block text-xs text-ink-soft">{l.variant && <><T k="pages.bag.size" /> {l.variant} · </>}<span className="tabular-nums">{l.qty}</span> × <span className="tabular-nums">{formatPrice(l.price)}</span></span>
                </span>
                <span className="shrink-0 text-right text-sm font-semibold tabular-nums text-ink">{formatPrice(l.price * l.qty)}</span>
              </li>
            ))}
          </ul>
          <p className="mt-4 flex items-baseline justify-between text-sm">
            <T k="pages.checkout.total" className="text-ink-soft" />
            <span className="text-2xl font-bold tabular-nums text-ink">{formatPrice(order.total)}</span>
          </p>
        </div>
        <dl className="flex flex-col gap-5 rounded border border-line p-6 text-sm lg:col-span-5">
          <div><dt className="text-ink-soft"><T k="pages.done.goes" /></dt><dd className="mt-1 text-ink"><T k={`checkout.delivery.${order.delivery}.title`} />{order.address ? ` — ${order.address}` : ""}</dd></div>
          <div><dt className="text-ink-soft"><T k="pages.done.paid" /></dt><dd className="mt-1 text-ink"><T k={`checkout.payment.${order.payment}.title`} /></dd></div>
          <div className="border-t border-line pt-5">
            <dt className="text-ink-soft"><T k="pages.done.ask" /></dt>
            <dd className="mt-2"><Link to="/contact" className="inline-flex items-center gap-2 rounded-full bg-accent px-6 py-2.5 text-sm font-medium text-accent-ink"><Icon name="chat" size={16} /><T k="pages.done.message" /></Link></dd>
          </div>
        </dl>
      </section>
    </div>
  );
}
