import { useState } from "react";
import { Link, useParams } from "react-router-dom";
import { T } from "../editable";
import { usePageMeta } from "../seo";
import { useCollection, useRecord } from "../data";
import { addToCart } from "../lib/cart";
import { formatPrice } from "../lib/format";
import { inStock, picturesOf, sizesOf, type Department, type Product as Bike, type Review, type ShopInfo } from "../lib/shop";
import AskForm from "../components/AskForm";
import Grid from "../components/Grid";
import Icon from "../components/Icon";
import ProductCard from "../components/ProductCard";
import ProductGallery from "../components/ProductGallery";
import ProductTabs from "../components/ProductTabs";
import QuantityStepper from "../components/QuantityStepper";
import Rating from "../components/Rating";
import RelatedRow from "../components/RelatedRow";
import ReviewCard from "../components/ReviewCard";
import SaveButton from "../components/SaveButton";
import SpecList from "../components/SpecList";
import StockLine from "../components/StockLine";
import { TINTS } from "../components/TintCard";

type Size = ReturnType<typeof sizesOf>[number];

/**
 * ONE BIKE, and everything somebody deciding on it asks, in the order they
 * ask it. The first screen: the maker's photograph large with its two
 * close-ups, and beside it the tinted head of the template's card (riding
 * type, brand and model), the price set large, the stock in the owner's
 * words, the frame sizes each with the rider height it fits, and the two ways
 * to buy — put it in the bag, or ask the shop to call.
 *
 * Then the specification beside the size guide, the longer things on tabs,
 * the form that asks the shop, and more bikes of the type.
 */
export default function Product() {
  const { id } = useParams();
  const { item } = useRecord<Bike>("products", id);
  const types = useCollection<Department>("categories", { pageSize: 24 }).items;
  usePageMeta("pages.product.title", "pages.product.description");
  if (!item) return <Missing />;
  const typeIndex = Math.max(0, types.findIndex((d) => d.id === item.category));
  const type = types[typeIndex];
  const sizes = sizesOf(item);
  return (
    <div className="pb-8 pt-8">
      <Crumbs bike={item} type={type} />
      <section className="flex flex-col gap-10 lg:flex-row lg:gap-12">
        <div className="lg:w-7/12">
          <ProductGallery images={picturesOf(item)} alt={`${item.brand} ${item.name}`} aspect="4/3" />
        </div>
        <div data-record={`products:${item.id}`} className="flex flex-col gap-6 lg:w-5/12">
          <Head bike={item} type={type} tint={typeIndex} />
          <Price bike={item} />
          <Buy bike={item} sizes={sizes} />
          <Promises />
        </div>
      </section>
      <Specification bike={item} sizes={sizes} />
      <Longer bike={item} />
      <section id="ask" className="mt-24 flex scroll-mt-24 flex-col gap-10 rounded bg-surface p-6 lg:flex-row lg:p-12">
        <div className="lg:w-4/12">
          <h2 className="text-[length:var(--size-h2)] leading-tight text-ink"><span className="font-light"><T k="pages.product.ask_light" /> </span><span className="font-bold">{item.name}</span></h2>
          <T k={inStock(item) ? "pages.product.ask_line" : "pages.product.ask_line_out"} as="p" className="mt-4 max-w-[40ch] text-sm leading-relaxed text-ink-soft" />
        </div>
        <div className="lg:w-8/12"><AskForm bike={`${item.brand} ${item.name}`} sizes={sizes.map((s) => s.size)} /></div>
      </section>
      <RelatedRow<Bike> collection="products" exclude={item.id} where={{ category: item.category, status: "live" }} max={3} phone={1} className="pt-24"
        title={<h2 className="mb-8 text-[length:var(--size-h2)] leading-tight text-ink"><span className="font-light"><T k="pages.product.related_light" /> </span><span className="font-bold">{type?.name}</span></h2>}>
        {(p) => <ProductCard product={p} />}
      </RelatedRow>
    </div>
  );
}

function Missing() {
  return (
    <div className="flex flex-col items-start gap-6 py-24">
      <T k="pages.product.missing" as="h1" className="text-[length:var(--size-h2)] text-ink" />
      <Link to="/shop" className="rounded-full bg-accent px-6 py-2.5 text-sm font-medium text-accent-ink"><T k="pages.product.back" /></Link>
    </div>
  );
}

function Crumbs({ bike, type }: { bike: Bike; type?: Department }) {
  return (
    <nav aria-label="Breadcrumb" className="mb-6 flex items-center gap-1.5 text-sm text-ink-soft">
      <Link to="/shop" className="hover:text-ink"><T k="pages.shop.crumb" /></Link>
      {type && (<><Icon name="next" size={14} /><Link to={`/shop?category=${type.id}`} className="hover:text-ink">{type.name}</Link></>)}
      <Icon name="next" size={14} /><span className="text-ink">{bike.name}</span>
    </nav>
  );
}

/** The tinted head of the template's card: riding type, brand and model, the one line, the rating. */
function Head({ bike, type, tint }: { bike: Bike; type?: Department; tint: number }) {
  return (
    <div className={`rounded p-6 ${TINTS[tint % TINTS.length]}`}>
      <div className="flex items-start justify-between gap-4">
        {type && <span className="rounded-full bg-ground px-3 py-1 text-xs text-ink">{type.name}</span>}
        <SaveButton id={bike.id} name={bike.name} />
      </div>
      <h1 data-testid="product-name" className="mt-4 text-[length:var(--size-h2)] leading-[1.08] tracking-tight text-ink">
        <span className="font-light">{bike.brand} </span><span className="font-bold">{bike.name}</span>
      </h1>
      <p className="mt-3 text-sm text-ink-soft">{bike.fact}{bike.year ? ` · ${bike.year}` : ""}</p>
      {bike.rating !== undefined && (
        <p className="mt-3 flex items-center gap-2 text-sm text-ink">
          <Rating value={bike.rating} size={15} />
          <span className="font-semibold tabular-nums">{bike.rating}</span>
          <span className="text-ink-soft">(<span className="tabular-nums">{bike.review_count}</span> <T k="pages.product.ratings" />)</span>
        </p>
      )}
    </div>
  );
}

function Price({ bike }: { bike: Bike }) {
  const saving = bike.was !== undefined && bike.was > bike.price ? bike.was - bike.price : 0;
  return (
    <div>
      <p className="flex flex-wrap items-baseline gap-x-4 gap-y-1">
        <span className="text-4xl font-bold tracking-tight tabular-nums text-ink" data-testid="product-price">{formatPrice(bike.price)}</span>
        {saving > 0 && <span className="text-base tabular-nums text-ink-soft line-through">{formatPrice(bike.was)}</span>}
      </p>
      <div className="mt-2 flex flex-wrap items-center gap-3">
        <StockLine stock={bike.stock ?? 0} className="text-sm" />
        {saving > 0 && (
          <span className="rounded-full bg-[var(--accent-soft)] px-3 py-1 text-xs font-medium text-ink">
            <T k="pages.product.saved" /> <span className="tabular-nums">{formatPrice(saving)}</span>
          </span>
        )}
      </div>
    </div>
  );
}

/** The sizes, how many, and the two ways to buy. */
function Buy({ bike, sizes }: { bike: Bike; sizes: Size[] }) {
  const [size, setSize] = useState("");
  const [qty, setQty] = useState(1);
  const [added, setAdded] = useState(false);
  const available = inStock(bike);
  const free = sizes.filter((s) => !s.out);
  const choice = size || (free.find((s) => s.size === "M") ?? free[0])?.size || "";
  const buy = () => {
    addToCart({ id: bike.id, name: `${bike.brand} ${bike.name}`, price: bike.price, image: bike.image, variant: choice || undefined }, qty);
    setAdded(true);
  };
  return (
    <>
      {sizes.length > 0 && <Sizes sizes={sizes} choice={choice} onPick={setSize} />}
      <div className="flex flex-wrap items-center gap-3">
        {available && <QuantityStepper value={qty} onChange={setQty} max={bike.stock} />}
        {available && (
          <button type="button" onClick={buy} data-testid="product-add"
            className="inline-flex flex-1 items-center justify-center gap-2 whitespace-nowrap rounded-full bg-accent px-7 py-3.5 font-medium text-accent-ink hover:opacity-90">
            <Icon name={added ? "check" : "bag"} size={18} />
            <T k={added ? "pages.product.added" : "pages.product.add"} />
          </button>
        )}
        <a href="#ask" data-testid="product-ask"
          className={`inline-flex items-center justify-center gap-2 whitespace-nowrap rounded-full px-6 py-3.5 font-medium ${available ? "border border-ink text-ink hover:bg-wash" : "flex-1 bg-accent text-accent-ink"}`}>
          <Icon name="phone" size={17} />
          <T k={available ? "pages.product.ask" : "pages.product.ask_back"} />
        </a>
      </div>
      {added && (
        <p className="-mt-2 flex items-center gap-2 text-sm text-ink">
          <Icon name="done" size={16} /><T k="pages.product.added_line" />
          <Link to="/bag" data-testid="product-to-bag" className="font-semibold underline underline-offset-4"><T k="pages.product.to_bag" /></Link>
        </p>
      )}
    </>
  );
}

/** Frame sizes as pills, each with the rider height it fits; a sold-out size is struck through. */
function Sizes({ sizes, choice, onPick }: { sizes: Size[]; choice: string; onPick: (s: string) => void }) {
  return (
    <fieldset id="sizes" className="scroll-mt-24">
      <legend className="mb-3 flex w-full items-baseline justify-between text-sm">
        <T k="pages.product.size" className="font-semibold text-ink" />
        <a href="#size-guide" className="text-xs text-ink-soft underline underline-offset-4 hover:text-ink"><T k="pages.product.size_guide" /></a>
      </legend>
      <div className="flex flex-wrap gap-2" role="radiogroup">
        {sizes.map((s) => <SizePill key={s.size} s={s} on={choice === s.size && !s.out} onPick={onPick} />)}
      </div>
    </fieldset>
  );
}

function SizePill({ s, on, onPick }: { s: Size; on: boolean; onPick: (s: string) => void }) {
  return (
    <button type="button" role="radio" aria-checked={on} disabled={s.out} onClick={() => onPick(s.size)} data-testid={`size-${s.size}`}
      className={`flex min-w-[3.6rem] flex-1 flex-col items-center rounded-[var(--radius-sm)] border px-2 py-2.5 text-center ${on ? "border-ink bg-accent text-accent-ink" : "border-field text-ink hover:border-ink"} disabled:cursor-not-allowed disabled:border-line disabled:text-ink-soft disabled:line-through`}>
      <span className="text-sm font-semibold">{s.size}</span>
      {s.fit && <span className={`text-[0.6875rem] tabular-nums ${on ? "text-accent-ink" : "text-ink-soft"}`}>{s.fit}</span>}
    </button>
  );
}

function Promises() {
  const shop = useCollection<ShopInfo>("shop", { pageSize: 1 }).items[0];
  if (!shop) return null;
  return (
    <ul className="flex flex-col gap-3 rounded border border-line p-5 text-sm text-ink">
      <li className="flex items-start gap-3"><Icon name="shipping" size={18} className="text-ink-soft" /><span>{shop.delivery_local}</span></li>
      <li className="flex items-start gap-3"><Icon name="secure" size={18} className="text-ink-soft" />
        <span><span className="tabular-nums">{shop.warranty_years}</span> <T k="pages.product.warranty" /></span></li>
      <li className="flex items-start gap-3"><Icon name="chat" size={18} className="text-ink-soft" />
        <span><T k="pages.product.reply" /> <span className="tabular-nums">{shop.reply_hours}</span> <T k="pages.product.reply_unit" /></span></li>
    </ul>
  );
}

const SPEC_KEYS = ["frame", "fork", "shock", "motor", "drivetrain", "brakes", "wheels", "tyres", "weight", "colour"] as const;

/** The specification as a two-column table, beside the size guide. */
function Specification({ bike, sizes }: { bike: Bike; sizes: Size[] }) {
  const rows = SPEC_KEYS.filter((k) => Boolean(bike[k])).map((k) => [<T key={k} k={`pages.product.spec.${k}`} />, bike[k]] as const);
  return (
    <section className="flex flex-col gap-12 pt-24 lg:flex-row">
      <div className="lg:w-7/12">
        <h2 className="mb-6 text-[length:var(--size-h2)] leading-tight text-ink"><span className="font-light"><T k="pages.product.spec_light" /> </span><span className="font-bold"><T k="pages.product.spec_bold" /></span></h2>
        <SpecList items={rows} />
      </div>
      {sizes.length > 0 && <SizeGuide sizes={sizes} />}
    </section>
  );
}

function SizeGuide({ sizes }: { sizes: Size[] }) {
  return (
    <div id="size-guide" className="scroll-mt-24 lg:w-5/12">
      <h2 className="mb-6 text-[length:var(--size-h2)] leading-tight text-ink"><span className="font-light"><T k="pages.product.guide_light" /> </span><span className="font-bold"><T k="pages.product.guide_bold" /></span></h2>
      <table className="w-full text-sm">
        <thead>
          <tr className="border-b border-line text-left text-xs uppercase tracking-wider text-ink-soft">
            <th className="py-3 font-medium"><T k="pages.product.guide_size" /></th>
            <th className="py-3 text-right font-medium"><T k="pages.product.guide_height" /></th>
            <th className="py-3 text-right font-medium"><T k="pages.product.guide_stock" /></th>
          </tr>
        </thead>
        <tbody>
          {sizes.map((s) => (
            <tr key={s.size} className="border-b border-line">
              <td className="py-3 font-semibold text-ink">{s.size}</td>
              <td className="py-3 text-right tabular-nums text-ink">{s.fit}</td>
              <td className="py-3 text-right text-ink-soft"><T k={s.out ? "stock.out" : "stock.in"} /></td>
            </tr>
          ))}
        </tbody>
      </table>
      <T k="pages.product.guide_note" as="p" className="mt-4 max-w-[48ch] text-xs leading-relaxed text-ink-soft" />
    </div>
  );
}

/** The longer things on tabs: what it is for, owning it, delivery and payment, reviews. */
function Longer({ bike }: { bike: Bike }) {
  const shop = useCollection<ShopInfo>("shop", { pageSize: 1 }).items[0];
  const reviews = useCollection<Review>("reviews", { where: { bike: bike.id }, pageSize: 6 }).items;
  const tabs = [
    { label: <T k="pages.product.tab_about" />, content: <p className="max-w-[65ch] leading-relaxed">{bike.description}</p> },
    { label: <T k="pages.product.tab_ride" />, content: <p className="max-w-[65ch] leading-relaxed">{bike.ride}</p> },
  ];
  if (shop) {
    tabs.push({ label: <T k="pages.product.tab_delivery" />, content: (
      <SpecList items={[
        [<T key="l" k="pages.product.d_local" />, shop.delivery_local],
        [<T key="n" k="pages.product.d_nation" />, `${shop.delivery_nationwide} · ${formatPrice(shop.delivery_fee)}`],
        [<T key="p" k="pages.product.d_pickup" />, shop.pickup],
        [<T key="y" k="pages.product.d_pay" />, shop.payment.join(" · ")],
        [<T key="c" k="pages.product.d_confirm" />, shop.confirm],
      ]} />
    ) });
  }
  if (reviews.length > 0) {
    tabs.push({ label: <><T k="pages.product.tab_reviews" /> <span className="tabular-nums">({reviews.length})</span></>, content: (
      <Grid max={2} phone={1}>{reviews.map((r) => <ReviewCard key={r.id} review={r} />)}</Grid>
    ) });
  }
  return <ProductTabs className="pt-24" tabs={tabs} />;
}
