import { useState } from "react";
import { Link } from "react-router-dom";
import { T } from "../editable";
import { usePageMeta } from "../seo";
import { useCollection } from "../data";
import { useCart } from "../lib/cart";
import { useDemo } from "../lib/demo";
import { formatPrice } from "../lib/format";
import { EXAMPLE_LINES } from "../lib/mock/orders";
import type { ShopInfo } from "../lib/shop";
import OrderForm, { type Delivery, type Payment } from "../components/OrderForm";
import ProductImage from "../components/ProductImage";
import Weights from "../components/Weights";

/**
 * CHECKING OUT: who it is for, how it reaches them and how it is paid at the
 * left, in three numbered steps; what is being bought at the right, with its
 * pictures, the delivery charge and the total, in view the whole time the
 * form is filled in.
 */
export default function Checkout() {
  usePageMeta("pages.checkout.title", "pages.checkout.description");
  const cart = useCart();
  const demo = useDemo();
  const shop = useCollection<ShopInfo>("shop", { pageSize: 1 }).items[0];
  const [delivery, setDelivery] = useState<Delivery>("local");
  const [payment, setPayment] = useState<Payment>("transfer");
  const lines = demo && cart.lines.length === 0 ? EXAMPLE_LINES : cart.lines;
  const subtotal = lines.reduce((s, l) => s + l.price * l.qty, 0);
  const fee = delivery === "courier" ? shop?.delivery_fee ?? 0 : 0;
  return (
    <div className="pb-8 pt-10">
      <h1 className="text-[length:var(--size-h1)] leading-[1.1] text-ink"><Weights k="pages.checkout.heading" /></h1>
      <T k="pages.checkout.lead" as="p" className="mt-4 max-w-[60ch] text-sm leading-relaxed text-ink-soft" />
      <div className="mt-10 flex flex-col gap-10 lg:flex-row lg:items-start lg:gap-14">
        <OrderForm className="min-w-0 flex-1" lines={lines} subtotal={subtotal} fee={fee}
          delivery={delivery} onDelivery={setDelivery} payment={payment} onPayment={setPayment} />
        <aside className="rounded border border-line p-6 lg:sticky lg:top-6 lg:w-[26rem] lg:shrink-0">
          <T k="pages.checkout.summary" as="h2" className="mb-5 block text-lg font-semibold text-ink" />
          {lines.length === 0 ? (
            <p className="text-sm text-ink-soft">
              <T k="pages.checkout.empty" />{" "}
              <Link to="/shop" className="font-medium text-ink underline underline-offset-4"><T k="pages.bag.browse" /></Link>
            </p>
          ) : (
            <>
              <ul className="flex flex-col gap-4">
                {lines.map((l) => (
                  <li key={`${l.id}-${l.variant ?? ""}`} className="flex items-center gap-4">
                    <span className="w-24 shrink-0"><ProductImage src={l.image} alt={l.name} aspect="4/3" className="border border-line" /></span>
                    <span className="min-w-0 flex-1 text-sm">
                      <span className="block font-semibold text-ink">{l.name}</span>
                      <span className="block text-xs text-ink-soft">
                        {l.variant && <><T k="pages.bag.size" /> {l.variant} · </>}<span className="tabular-nums">{l.qty}</span> × <span className="tabular-nums">{formatPrice(l.price)}</span>
                      </span>
                    </span>
                    <span className="shrink-0 text-right text-sm font-semibold tabular-nums text-ink">{formatPrice(l.price * l.qty)}</span>
                  </li>
                ))}
              </ul>
              <dl className="mt-6 flex flex-col gap-2 border-t border-line pt-5 text-sm">
                <div className="flex justify-between"><dt className="text-ink-soft"><T k="pages.bag.subtotal" /></dt><dd className="tabular-nums text-ink">{formatPrice(subtotal)}</dd></div>
                <div className="flex justify-between"><dt className="text-ink-soft"><T k="pages.checkout.delivery" /></dt>
                  <dd className="tabular-nums text-ink">{fee > 0 ? formatPrice(fee) : <T k="pages.checkout.free" />}</dd></div>
                <div className="flex items-baseline justify-between border-t border-line pt-3">
                  <dt className="font-semibold text-ink"><T k="pages.checkout.total" /></dt>
                  <dd data-testid="checkout-total" className="text-2xl font-bold tabular-nums text-ink">{formatPrice(subtotal + fee)}</dd>
                </div>
              </dl>
              <T k="pages.checkout.includes" as="p" className="mt-3 text-xs leading-relaxed text-ink-soft" />
            </>
          )}
        </aside>
      </div>
    </div>
  );
}
