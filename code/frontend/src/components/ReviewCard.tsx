import { useCollection } from "../data";
import { formatDate } from "../lib/format";
import type { Product, Review } from "../lib/shop";
import Icon from "./Icon";
import Rating from "./Rating";

/**
 * What one rider said: the stars, their words, their first name and when,
 * and the bike they bought — a white card with a hairline edge, as the
 * template's testimonials are.
 */
export default function ReviewCard({ review }: { review: Review }) {
  const bikes = useCollection<Product>("products", { pageSize: 50 }).items;
  const bike = bikes.find((b) => b.id === review.bike);
  return (
    <figure data-record={`reviews:${review.id}`} className="flex h-full flex-col gap-4 rounded border border-line bg-ground p-6">
      <div className="flex items-center justify-between">
        <Rating value={review.rating} size={15} />
        <Icon name="quote" size={18} className="text-ink-soft" />
      </div>
      <blockquote className="flex-1 text-sm leading-relaxed text-ink">{review.text}</blockquote>
      <figcaption className="flex items-baseline justify-between gap-4 border-t border-line pt-4 text-xs">
        <span><span className="font-semibold text-ink">{review.name}</span>{bike && <span className="text-ink-soft"> · {bike.brand} {bike.name}</span>}</span>
        <span className="shrink-0 tabular-nums text-ink-soft">{formatDate(review.date, "medium")}</span>
      </figcaption>
    </figure>
  );
}
