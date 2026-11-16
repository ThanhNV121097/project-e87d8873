import { Link } from "react-router-dom";
import { T, useContent } from "../editable";
import { usePageMeta } from "../seo";
import { useCollection } from "../data";
import { lineKey, removeFromCart, setQuantity, useCart } from "../lib/cart";
import { useDemo } from "../lib/demo";
import { formatPrice } from "../lib/format";
import { EXAMPLE_LINES } from "../lib/mock/orders";
import type { Department, ShopInfo } from "../lib/shop";
import { Picture } from "../picture";
import Bleed from "../components/Bleed";
import DeliveryTerms from "../components/DeliveryTerms";
import Icon from "../components/Icon";
import ProductImage from "../components/ProductImage";
import QuantityStepper from "../components/QuantityStepper";
import Weights from "../components/Weights";

/**
 * THE BAG: each bike a row to read across — its picture, its name and size,
 * how many, what it comes to — and beside the rows a pink summary card with
 * what the total includes and the one button. Empty, it says so and offers
 * the riding types to start from.
 */
export default function Bag() {
  usePageMeta("pages.bag.title", "pages.bag.description");
  const cart = useCart();
  const demo = useDemo();
  const lines = demo && cart.lines.length === 0 ? EXAMPLE_LINES : cart.lines;
  const total = lines.reduce((s, l) => s + l.price * l.qty, 0);
  const remove = useContent("pages.bag.remove");
  const types = useCollection<Department>("categories", { pageSize: 12 }).items;
  const shop = useCollection<ShopInfo>("shop", { pageSize: 1 }).items[0];
  const image = useContent("pages.bag.image");
  const imageAlt = useContent("pages.bag.image_alt");
  return (
    <div className="pb-8">
      {/* The head on a band to both screen edges, a rider's-eye view of the
          bars off its right edge (on a phone, a strip across the top). */}
      <Bleed className="mb-10 bg-surface lg:flex lg:flex-row-reverse lg:gap-12">
        {image && (
          <div className="relative -mx-[var(--gutter)] aspect-[21/9] overflow-hidden lg:mx-0 lg:w-4/12 lg:shrink-0 lg:-mr-[calc((100vw-min(100vw,var(--page)))/2+var(--gutter))] lg:aspect-auto">
            <Picture src={image} alt={imageAlt} width={1280} height={852} loading="eager" className="absolute inset-0 h-full w-full object-cover" />
          </div>
        )}
        <div className="min-w-0 pb-10 pt-10 lg:flex-1">
          <h1 className="text-[length:var(--size-h1)] leading-[1.1] text-ink"><Weights k="pages.bag.heading" /></h1>
          <T k="pages.bag.lead" as="p" className="mt-4 max-w-[60ch] text-[0.9375rem] leading-relaxed text-ink-soft" />
        </div>
      </Bleed>
      {lines.length === 0 ? (
        <div className="flex flex-col items-start gap-6 rounded bg-[var(--tint-1)] px-8 py-14 lg:px-14">
          <T k="pages.bag.empty" as="p" className="text-2xl text-ink" />
          <T k="pages.bag.empty_line" as="p" className="max-w-[52ch] text-sm text-ink-soft" />
          <div className="flex flex-wrap gap-2">
            {types.map((d) => (
              <Link key={d.id} to={`/shop?category=${d.id}`} className="rounded-full border border-ink bg-ground px-5 py-2 text-sm text-ink hover:bg-accent hover:text-accent-ink">{d.name}</Link>
            ))}
          </div>
          <Link to="/shop" className="inline-flex items-center gap-2 rounded-full bg-accent px-7 py-3 text-sm font-medium text-accent-ink"><T k="pages.bag.browse" /><Icon name="arrow" size={16} /></Link>
        </div>
      ) : (
        <div className="flex flex-col gap-10 lg:flex-row lg:items-start lg:gap-14">
          <ul className="min-w-0 flex-1 divide-y divide-line border-y border-line">
            {lines.map((l) => (
              <li key={lineKey(l)} data-testid={`bag-line-${l.id}`} data-record={`products:${l.id}`} className="grid grid-cols-[7rem_1fr] items-center gap-x-5 gap-y-3 py-5 sm:grid-cols-[10rem_1fr_auto_8rem_auto]">
                <Link to={`/shop/${l.id}`} className="row-span-2 sm:row-span-1"><ProductImage src={l.image} alt={l.name} aspect="4/3" className="border border-line" /></Link>
                <div className="min-w-0">
                  <Link to={`/shop/${l.id}`} className="text-base font-semibold text-ink hover:underline">{l.name}</Link>
                  <p className="mt-1 text-sm text-ink-soft">
                    {l.variant && <><T k="pages.bag.size" /> {l.variant} · </>}<span className="tabular-nums">{formatPrice(l.price)}</span>
                  </p>
                </div>
                <div className="col-start-2 flex items-center justify-between gap-4 sm:col-start-auto sm:contents">
                  <QuantityStepper value={l.qty} onChange={(q) => setQuantity(lineKey(l), q)} />
                  <p className="text-right text-base font-bold tabular-nums text-ink">{formatPrice(l.price * l.qty)}</p>
                  <button type="button" onClick={() => removeFromCart(lineKey(l))} aria-label={`${remove} ${l.name}`}
                    className="grid h-10 w-10 place-items-center rounded-full text-ink-soft hover:bg-wash hover:text-ink"><Icon name="remove" size={17} /></button>
                </div>
              </li>
            ))}
          </ul>
          <aside className="rounded bg-[var(--tint-1)] p-6 lg:w-96 lg:shrink-0">
            <dl className="flex flex-col gap-2 text-sm">
              <div className="flex justify-between"><dt className="text-ink-soft"><T k="pages.bag.subtotal" /></dt><dd className="tabular-nums text-ink">{formatPrice(total)}</dd></div>
              <div className="flex justify-between gap-6"><dt className="text-ink-soft"><T k="pages.checkout.delivery" /></dt><dd className="text-right text-ink"><T k="pages.bag.delivery_note" /></dd></div>
            </dl>
            <div className="mt-5 flex items-center justify-between rounded-full border border-line bg-[var(--glass)] py-2 pl-5 pr-2">
              <T k="pages.bag.total" className="text-sm text-ink-soft" />
              <span data-testid="bag-total" className="text-xl font-bold tabular-nums text-ink">{formatPrice(total)}</span>
            </div>
            <Link to="/checkout" data-testid="bag-checkout"
              className="mt-5 flex items-center justify-center gap-2 rounded-full bg-accent px-6 py-3.5 font-medium text-accent-ink hover:opacity-90">
              <T k="pages.bag.checkout" /><Icon name="arrow" size={16} />
            </Link>
            {shop && (
              <ul className="mt-6 flex flex-col gap-2.5 text-sm text-ink">
                <li className="flex items-start gap-2.5"><Icon name="shipping" size={16} /><DeliveryTerms shop={shop} /></li>
                <li className="flex items-start gap-2.5"><Icon name="secure" size={16} /><span><span className="tabular-nums">{shop.warranty_years}</span> <T k="pages.product.warranty" /></span></li>
                <li className="flex items-start gap-2.5"><Icon name="phone" size={16} /><T k="pages.bag.call_note" /></li>
              </ul>
            )}
          </aside>
        </div>
      )}
    </div>
  );
}
