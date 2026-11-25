/**
 * The site's records — bikes, riding types, reviews, the team, gallery photographs,
 * workshop jobs, questions and the shop's own numbers — read from the shop's API.
 * schema.json describes every collection's fields; the records live in the database.
 *
 * Components read records ONLY through these hooks. Every request runs in the
 * browser, against `import.meta.env.VITE_API_URL ?? "/api"` (read when a request is
 * made), and speaks the service contract:
 *   GET <base>/v1/<name>?page=&limit=&<field>=<value>&q=&search=&sort= → { items, total, page, limit }
 *     — <field>=a,b is "any of"; <field>_from / <field>_to is an inclusive range; sort
 *     is "price" or "-price". `total` is every match, never the size of the page;
 *   GET <base>/v1/<name>/<id> → the record, or 404.
 * A page size above 100 is never asked: the API would answer 100 anyway.
 *
 * The page paints before its records arrive: `items` is empty and `loading` is true
 * until the first answer. There is no cache: every answer is `no-store`, and a request
 * that is already in flight for the same address is shared, nothing more.
 */
import { useCallback, useEffect, useRef, useState, useSyncExternalStore } from "react";
import schemaFile from "./schema.json";

/** Where a record came from: the owner's own data, a page on the web, or an assumption. */
export type Source = "owner" | "assumed" | `web:${string}`;

export interface Row {
  id: string;
  _source?: Source;
  [field: string]: unknown;
}

export interface Field {
  key: string;
  label?: string;
  /** text | longtext | number | price | image | url | bool | ref:<collection> | list */
  type: string;
  currency?: string;
  /**
   * A closed set of values — a status, a size, a kind: ["draft", "live"], or
   * [["draft", "Draft"], ["live", "On sale"]] where the label differs. The
   * management area draws it as a list to choose from and a badge in a table,
   * instead of a box to type "live" into.
   */
  options?: readonly (string | readonly [string, string])[];
  /** A value only the design reads — a swatch token, a layout hint. Never
   *  shown on a management form: it is not the owner's content. */
  hidden?: boolean;
  /** Which tab of the record's page the field sits on. */
  section?: string;
  /** Saving is refused while this is empty. */
  required?: boolean;
  /** One line under the control, for what a label cannot say. */
  hint?: string;
}

export interface Collection {
  label?: string;
  fields: Field[];
}

export const schema = schemaFile as { collections: Record<string, Collection> };

/**
 * One condition on a field. A plain value is an exact match; a list is "any
 * of" (a status filter with three boxes ticked); a range is from/to
 * inclusive, for dates and prices.
 *
 * An empty value — "", an empty list, a range with neither end — filters
 * nothing, and is not sent. A filter the person has not set yet must not
 * empty the table.
 */
export type Filter =
  | string
  | number
  | boolean
  | readonly (string | number)[]
  | { from?: string | number; to?: string | number };

export interface Query {
  /** Records whose field matches, e.g. { category: "trail", status: ["draft", "live"] }. */
  where?: Record<string, Filter | undefined>;
  /** Free text, matched against `search` fields — accents and case ignored. */
  q?: string;
  /** Which fields `q` searches. Every text field when not given. */
  search?: readonly string[];
  /** "price" ascending, "-price" descending. */
  sort?: string;
  /**
   * The page to return, 1-based. Given, `items` is THAT page alone — the
   * shape a table needs. Left out, `items` grows with loadMore(), which is
   * what a catalogue's "Show more" needs.
   */
  page?: number;
  /** How many a page holds: 12 when not given, at most 100. */
  pageSize?: number;
}

// What the module has seen, per collection, merged by id in first-seen order; and
// one counter per collection that hooks subscribe to, so that a collection reported
// changed is asked for again.
let loaded: Record<string, Row[]> = {};
const epochs: Record<string, number> = {};
const listeners = new Set<() => void>();
const inflight = new Map<string, Map<string, Promise<Answer>>>();

interface Answer {
  status: number;
  body: unknown;
}

/** The records the module has loaded so far, for code that looks one up by its id. */
export function collections(): Record<string, Row[]> {
  return loaded;
}

/**
 * Another module saved or deleted records of `name`: remember `rows` at once, and
 * make every list and record of that collection on screen ask the API again.
 */
export function setCollection(name: string, rows: Row[]) {
  loaded = { ...loaded, [name]: rows };
  epochs[name] = (epochs[name] ?? 0) + 1;
  inflight.delete(name);
  listeners.forEach((l) => l());
}

function remember(name: string, rows: Row[]) {
  const next = [...(loaded[name] ?? [])];
  for (const row of rows) {
    const at = next.findIndex((r) => r.id === row.id);
    if (at < 0) next.push(row);
    else next[at] = row;
  }
  loaded = { ...loaded, [name]: next };
}

function subscribe(l: () => void) {
  listeners.add(l);
  return () => {
    listeners.delete(l);
  };
}

function useEpoch(name: string): number {
  return useSyncExternalStore(subscribe, () => epochs[name] ?? 0);
}

const base = () => String(import.meta.env.VITE_API_URL ?? "/api").replace(/\/+$/, "");
const enc = (v: string | number | boolean) => encodeURIComponent(String(v));
const asError = (e: unknown) => (e instanceof Error ? e : new Error(String(e)));
const given = (e: string | number | undefined): e is string | number => e !== undefined && e !== "";
const isRow = (b: unknown): b is Row => typeof b === "object" && b !== null && typeof (b as Row).id === "string";

/** GET <base>/v1/<name><path>. A request already in flight for the address is joined, never cached. */
function get(name: string, path: string): Promise<Answer> {
  const url = `${base()}/v1/${enc(name)}${path}`;
  let shared = inflight.get(name);
  if (!shared) {
    shared = new Map();
    inflight.set(name, shared);
  }
  const joined = shared.get(url);
  if (joined) return joined;
  const request: Promise<Answer> = fetch(url)
    .then(async (res) => ({ status: res.status, body: (await res.json().catch(() => null)) as unknown }))
    .finally(() => {
      if (inflight.get(name)?.get(url) === request) inflight.get(name)?.delete(url);
    });
  shared.set(url, request);
  return request;
}

/** The envelope's own sentence when the API gave one. */
function problem(a: Answer): Error {
  const sentence = (a.body as { error?: unknown } | null)?.error;
  return new Error(typeof sentence === "string" ? sentence : `The shop's records could not be read (${a.status}).`);
}

/** `field=value` terms, in the object's key order; what is empty is not sent, a false or a 0 is. */
function whereTerms(where: Query["where"]): string[] {
  const out: string[] = [];
  for (const [key, value] of Object.entries(where ?? {})) {
    if (value === undefined || value === null || value === "") continue;
    if (Array.isArray(value)) {
      if (value.length > 0) out.push(`${enc(key)}=${value.map((v) => enc(v)).join(",")}`);
    } else if (typeof value === "object") {
      const r = value as { from?: string | number; to?: string | number };
      if (given(r.from)) out.push(`${enc(`${key}_from`)}=${enc(r.from)}`);
      if (given(r.to)) out.push(`${enc(`${key}_to`)}=${enc(r.to)}`);
    } else {
      out.push(`${enc(key)}=${enc(value as string | number | boolean)}`);
    }
  }
  return out;
}

/** Everything of a query but the page and the size, as `&`-led terms in the API's own words. */
function terms(query: Query): string {
  const out = whereTerms(query.where);
  const q = (query.q ?? "").trim();
  if (q) out.push(`q=${enc(q)}`);
  const search = (query.search ?? []).filter((f) => f !== "");
  if (search.length > 0) out.push(`search=${search.map((f) => enc(f)).join(",")}`);
  const sort = (query.sort ?? "").trim();
  if (sort) out.push(`sort=${enc(sort)}`);
  return out.map((t) => `&${t}`).join("");
}

const pageLimit = (size: number | undefined) => {
  const n = Math.floor(size ?? 12);
  return Number.isFinite(n) ? Math.min(100, Math.max(1, n)) : 12;
};
const lastPage = (total: number, limit: number) => Math.max(1, Math.ceil(total / limit));

async function fetchPage(name: string, page: number, limit: number, rest: string) {
  const answer = await get(name, `?page=${page}&limit=${limit}${rest}`);
  if (answer.status < 200 || answer.status > 299) throw problem(answer);
  const body = answer.body as { items?: unknown; total?: unknown } | null;
  if (!body || !Array.isArray(body.items) || typeof body.total !== "number") {
    throw new Error("The shop's records came back in a form this page cannot read.");
  }
  remember(name, body.items as Row[]);
  return { items: body.items as Row[], total: body.total };
}

/**
 * What a list holds: `pages` are the records of the query `answered`; `settled` is the
 * latest query whose request finished, answered or failed; `pending` is a request for
 * the next page in flight.
 */
interface Held {
  name: string;
  answered: string;
  settled: string;
  pages: Row[][];
  total: number;
  error: Error | null;
  pending: boolean;
}
const fresh = (name: string): Held => ({ name, answered: "", settled: "", pages: [], total: 0, error: null, pending: false });

/**
 * A page at a time: `items` grows by `pageSize` with each `loadMore()` — wire
 * it to a "Show more" button or an IntersectionObserver for lazy loading.
 *
 * Returns { items, total, pages, page, hasMore, loading, error, loadMore }:
 * - `items` — the page asked for, or the records loaded so far;
 * - `total` — how many records match in all, loaded or not. A count the
 *   visitor reads ("24 products", "1–20 of 137") is `total`, never
 *   `items.length`, which is only the page;
 * - `pages` — how many pages of `pageSize` those matches make;
 * - `page` — the page `items` holds, clamped into range: a filter narrowed
 *   while standing on page 7 shows the last page, not an empty table;
 * - `loading` — true from the moment a request goes out until it is answered;
 * - `error` — an Error when a request failed, null again after one succeeds.
 *
 * When the query changes, `items` start again from its first page: until that answer
 * arrives the previous records of the same collection stay and `loading` is true, and
 * the answer of an earlier query that arrives late is discarded.
 *
 * Pass `page` for a dashboard list — a table's footer is "1–20 of 137" with
 * numbers to click — and leave it out for a catalogue, where `loadMore()`
 * appends the next page to the ones already shown.
 */
// T is the record's shape as the page reads it. Any shape with an id will do —
// an interface included: `T extends Row` asked for Row's index signature,
// which no interface has, and a design turn spent 24 builds on the error.
export function useCollection<T extends { id: string } = Row>(name: string, query: Query = {}) {
  const epoch = useEpoch(name);
  const limit = pageLimit(query.pageSize);
  const rest = terms(query);
  const asked = query.page === undefined ? undefined : Math.max(1, Math.floor(query.page) || 1);
  const identity = JSON.stringify([name, rest, limit, asked ?? null, epoch]);
  const [held, setHeld] = useState<Held>(() => fresh(name));
  const generation = useRef(0);
  const busy = useRef(false);
  const live = useRef({ name, rest, limit, asked, identity, held });
  useEffect(() => {
    live.current = { name, rest, limit, asked, identity, held };
  });

  // An answer counts only while the query it answers is still the one on screen.
  const settle = useCallback((mine: number, next: (prev: Held) => Held) => {
    if (mine !== generation.current) return;
    busy.current = false;
    setHeld(next);
  }, []);

  useEffect(() => {
    const mine = ++generation.current;
    busy.current = true;
    (async () => {
      try {
        let page = asked ?? 1;
        let answer = await fetchPage(name, page, limit, rest);
        // A page past the last: ask for the last one, as many times as the total moves.
        while (mine === generation.current && answer.items.length === 0 && answer.total > 0 && page > lastPage(answer.total, limit)) {
          page = lastPage(answer.total, limit);
          answer = await fetchPage(name, page, limit, rest);
        }
        settle(mine, () => ({ name, answered: identity, settled: identity, pages: [answer.items], total: answer.total, error: null, pending: false }));
      } catch (e) {
        settle(mine, (prev) => ({ ...(prev.name === name ? prev : fresh(name)), name, settled: identity, error: asError(e), pending: false }));
      }
    })();
    return () => {
      generation.current += 1;
    };
  }, [name, rest, limit, asked, identity, settle]);

  const loadMore = useCallback(() => {
    const cur = live.current;
    const have = cur.held.pages.reduce((n, p) => n + p.length, 0);
    if (busy.current || cur.asked !== undefined || cur.held.answered !== cur.identity || have >= cur.held.total) return;
    const mine = generation.current;
    busy.current = true;
    setHeld((prev) => ({ ...prev, pending: true }));
    fetchPage(cur.name, cur.held.pages.length + 1, cur.limit, cur.rest).then(
      (more) => settle(mine, (prev) => ({ ...prev, pages: [...prev.pages, more.items], total: more.total, error: null, pending: false })),
      (e) => settle(mine, (prev) => ({ ...prev, error: asError(e), pending: false })),
    );
  }, [settle]);

  const same = held.name === name;
  const items = (same ? held.pages.flat() : []) as unknown as T[];
  const total = same ? held.total : 0;
  const pages = lastPage(total, limit);
  return {
    items,
    total,
    pages,
    page: asked === undefined ? 1 : Math.min(asked, pages),
    hasMore: asked === undefined && items.length < total,
    loading: held.settled !== identity || held.pending,
    error: same ? held.error : null,
    loadMore,
  };
}

/**
 * One record by id. `item` is empty until it is answered, and stays empty when the
 * answer is 404 (`error` is then null); `loading` is true from the first render for
 * an id until the answer; any other failure is an Error.
 */
export function useRecord<T extends { id: string } = Row>(
  name: string,
  id: string | undefined,
): { item: T | undefined; loading: boolean; error: Error | null } {
  const epoch = useEpoch(name);
  const key = JSON.stringify([name, id]);
  const [held, setHeld] = useState<{ key: string; item?: Row; error: Error | null } | null>(null);
  useEffect(() => {
    if (id === undefined) return undefined;
    let current = true;
    const settle = (item: Row | undefined, error: Error | null) => {
      if (current) setHeld({ key, item, error });
    };
    get(name, `/${enc(id)}`).then(
      (a) => {
        if (a.status === 200 && isRow(a.body)) {
          remember(name, [a.body]);
          settle(a.body, null);
        } else {
          settle(undefined, a.status === 404 ? null : problem(a));
        }
      },
      (e) => settle(undefined, asError(e)),
    );
    return () => {
      current = false;
    };
  }, [name, id, key, epoch]);
  if (id === undefined) return { item: undefined, loading: false, error: null };
  if (!held || held.key !== key) return { item: undefined, loading: true, error: null };
  return { item: held.item as T | undefined, loading: false, error: held.error };
}
