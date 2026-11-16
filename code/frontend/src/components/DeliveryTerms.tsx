import { T } from "../editable";
import { formatPrice } from "../lib/format";
import type { ShopInfo } from "../lib/shop";

/**
 * HOW A BIKE REACHES YOU, said the same way everywhere it is promised: in
 * town, and by courier elsewhere with its charge. Both terms are the shop's
 * own record (src/data/shop.json), so the home page, a bike's page, the bag
 * and the checkout can never promise two different things.
 */
export default function DeliveryTerms({ shop, className = "" }: { shop: ShopInfo; className?: string }) {
  return (
    <span className={`flex flex-col gap-1 ${className}`}>
      <span><T k="pages.product.d_local" className="font-semibold text-ink" />{" · "}{shop.delivery_local}</span>
      <span>
        <T k="pages.product.d_nation" className="font-semibold text-ink" />{" · "}{shop.delivery_nationwide}{" · "}
        <span className="whitespace-nowrap tabular-nums">{formatPrice(shop.delivery_fee)}</span>
      </span>
    </span>
  );
}
