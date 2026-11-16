import { Link } from "react-router-dom";
import { T } from "../editable";
import Icon from "./Icon";
import ProductImage from "./ProductImage";
import SaveButton from "./SaveButton";
import { addToCart } from "../lib/cart";
import { formatPrice } from "../lib/format";
import type { Product } from "../lib/shop";

/**
 * A PRODUCT IN A GRID — the small card: the thing on a pale rounded tile
 * with a heart at its top right, then the name at the left and the price at
 * the right on one line, and under them an outlined pill that puts it in the
 * bag. Four to a row. The tinted card (TintCard) is for the few things a
 * page leads with; this is for all of them.
 */
export default function ProductCard({ product }: { product: Product }) {
  return (
    <article data-record={`products:${product.id}`} className="relative flex h-full flex-col">
      <Link to={`/shop/${product.id}`} data-testid={`card-${product.id}`} className="group block flex-1">
        <ProductImage src={product.image} hover={product.image_back} alt={product.name} aspect="1/1" className="rounded-[var(--radius-sm)]" />
        <span className="mt-3 flex items-baseline justify-between gap-3">
          <span className="text-sm font-medium text-ink underline-offset-4 group-hover:underline">{product.name}</span>
          <span className="shrink-0 text-sm font-semibold tabular-nums text-ink">{formatPrice(product.price)}</span>
        </span>
      </Link>
      <button type="button" data-testid={`add-${product.id}`}
        onClick={() => addToCart({ id: product.id, name: product.name, price: product.price, image: product.image })}
        className="mt-3 inline-flex w-full items-center justify-center gap-2 rounded-full border border-ink py-2 text-xs font-medium text-ink hover:bg-accent hover:text-accent-ink">
        <T k="home.add" /><Icon name="bag" size={14} />
      </button>
      <SaveButton id={product.id} name={product.name} className="absolute right-2 top-2" />
    </article>
  );
}
