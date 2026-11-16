import { Link } from "react-router-dom";
import Icon from "./Icon";
import ProductImage from "./ProductImage";
import { addToCart } from "../lib/cart";
import { formatPrice } from "../lib/format";
import type { Product } from "../lib/shop";

// Written out whole so Tailwind finds every class in the source.
export const TINTS = ["bg-[var(--tint-1)]", "bg-[var(--tint-2)]", "bg-[var(--tint-3)]", "bg-[var(--tint-4)]"] as const;

/**
 * THE TINTED CARD: a tall rounded card filled with ONE pale tint, a photograph
 * of the thing on it, a small white pill at its top left saying what it
 * is, its name large beside that — and across its foot a bar of glass with
 * the price at the left and a round dark button at the right that puts it in
 * the bag.
 *
 * Pictures keep their backgrounds (nothing is cut out, 2026-11-16), so the
 * photograph FILLS its place on the card and the tint is the card round it.
 * Choose a photograph of the thing on a plain pale backdrop, so the picture
 * and the tint read as one card rather than a busy room set on a colour.
 */
export default function TintCard({ product, tint, label, compact = false }: {
  product: Product;
  /** Which of the four tints: the card's place in its row. */
  tint: number;
  /** The small pill at the top left — the department's name. */
  label?: string;
  compact?: boolean;
}) {
  return (
    <article data-record={`products:${product.id}`} data-tilt=""
      className={`relative overflow-hidden rounded ${TINTS[tint % TINTS.length]} [--product-ground:transparent] ${compact ? "aspect-[4/5]" : "aspect-[5/6]"}`}>
      <Link to={`/shop/${product.id}`} data-testid={`tint-${product.id}`} className="absolute inset-0 block">
        <span className="absolute bottom-0 right-0 block w-[78%]">
          <ProductImage src={product.image} hover={product.image_back} alt={product.name} aspect="1/1" className="rounded-none" />
        </span>
        <span className="relative flex flex-col items-start gap-3 p-5 lg:p-6">
          {label && <span className="rounded-full bg-ground px-3 py-1 text-xs text-ink">{label}</span>}
          {!compact && <span className="max-w-[9ch] text-2xl leading-[1.1] tracking-tight text-ink lg:text-[1.75rem]">{product.name}</span>}
        </span>
      </Link>
      <div className="absolute inset-x-3 bottom-3 flex items-center justify-between rounded-full bg-[var(--glass)] py-1.5 pl-5 pr-1.5 backdrop-blur-md">
        <span className="text-sm font-semibold tabular-nums text-ink">{formatPrice(product.price)}</span>
        <button type="button" aria-label={`Add ${product.name} to the bag`} data-testid={`add-${product.id}`}
          onClick={() => addToCart({ id: product.id, name: product.name, price: product.price, image: product.image })}
          className="grid h-10 w-10 place-items-center rounded-full bg-accent text-accent-ink hover:opacity-90">
          <Icon name="bag" size={17} />
        </button>
      </div>
    </article>
  );
}
