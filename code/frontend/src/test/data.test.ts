// @vitest-environment jsdom
import { act, cleanup, renderHook, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

// The runner that executes `npm test` has NODE_ENV=production, which loads React's
// production build, and that build has no act(). Hoisted above the imports, so React
// is loaded as its development build.
vi.hoisted(() => {
  vi.stubEnv("NODE_ENV", "test");
});
import { collections, schema, setCollection, useCollection, useRecord } from "../data";
import type { Collection, Field, Filter, Query, Row, Source } from "../data";

// The data module against a stubbed fetch: every case records the addresses the
// module asks and answers with pages cut the way the real API cuts them.
const bikes: Row[] = Array.from({ length: 14 }, (_, i) => ({ id: `bike-${i + 1}`, name: `Bike ${i + 1}` }));
const ids = (rows: readonly { id: string }[]) => rows.map((r) => r.id);

const urls: string[] = [];
let respond: (url: string) => Response | Promise<Response>;

const reply = (body: unknown, status = 200) => new Response(JSON.stringify(body), { status });

function pageOf(rows: Row[], url: string) {
  const q = new URL(url, "http://shop.test").searchParams;
  const page = Number(q.get("page"));
  const limit = Number(q.get("limit"));
  return reply({ items: rows.slice((page - 1) * limit, page * limit), total: rows.length, page, limit });
}

/** An answer that waits until `release()` is called. */
function held(rows: Row[]) {
  const gate = { release: () => {} };
  const answer = (url: string) =>
    new Promise<Response>((resolve) => {
      gate.release = () => resolve(pageOf(rows, url));
    });
  return { gate, answer };
}

const settled = (r: { current: { loading: boolean } }) => waitFor(() => expect(r.current.loading).toBe(false));

beforeEach(() => {
  urls.length = 0;
  respond = (url) => pageOf(bikes, url);
  vi.stubGlobal("fetch", (url: string) => {
    urls.push(url);
    return Promise.resolve(respond(url));
  });
});

afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
  vi.unstubAllEnvs();
});

describe("what the module exports", () => {
  it("keeps every name, and the types still type-check", () => {
    const field: Field = { key: "name", type: "text" };
    const collection: Collection = { fields: [field] };
    const filter: Filter = { from: 1, to: 2 };
    const source: Source = "owner";
    const row: Row = { id: "x", _source: source };
    const query: Query = { where: { a: filter }, q: "x", search: ["name"], sort: "-name", page: 1, pageSize: 5 };
    expect([collection, row, query].every(Boolean)).toBe(true);
    expect([useCollection, useRecord, setCollection, collections].every((f) => typeof f === "function")).toBe(true);
    expect(Object.keys(schema.collections)).toContain("products");
  });
});

describe("the address", () => {
  it("is /api/v1/<collection>?page=&limit= when VITE_API_URL is not set", async () => {
    const { result } = renderHook(() => useCollection("products", { pageSize: 3 }));
    await settled(result);
    expect(urls[0]).toBe("/api/v1/products?page=1&limit=3");
  });

  it("is <VITE_API_URL>/v1/... when it is set, without its trailing slash", async () => {
    vi.stubEnv("VITE_API_URL", "https://x.test/");
    const { result } = renderHook(() => useCollection("products", { pageSize: 3 }));
    await settled(result);
    expect(urls[0]).toBe("https://x.test/v1/products?page=1&limit=3");
  });

  it("carries conditions, a list, ranges and no empty value, in the API's terms", async () => {
    const where: Record<string, Filter | undefined> = {
      status: "live",
      category: ["trail", "e-mtb"],
      stock: { from: 1 },
      price: { from: "", to: 70000000 },
      deal: false,
    };
    renderHook(() => useCollection("products", { where }));
    await waitFor(() => expect(urls).toHaveLength(1));
    expect(urls[0]).toBe("/api/v1/products?page=1&limit=12&status=live&category=trail,e-mtb&stock_from=1&price_to=70000000&deal=false");
  });

  it("sends a zero and a false, and nothing for an empty value", async () => {
    renderHook(() => useCollection("products", { where: { stock: 0, category: [], brand: "", month: undefined, price: {} } }));
    await waitFor(() => expect(urls).toHaveLength(1));
    expect(urls[0]).toBe("/api/v1/products?page=1&limit=12&stock=0");
  });

  it("carries the search words, the fields searched and the sort", async () => {
    renderHook(() => useCollection("products", { q: " cua lo ", search: ["name", "fact"], sort: "-price" }));
    await waitFor(() => expect(urls).toHaveLength(1));
    expect(urls[0]).toBe("/api/v1/products?page=1&limit=12&q=cua%20lo&search=name,fact&sort=-price");
  });

  it("leaves out blank search words, no fields and a blank sort", async () => {
    renderHook(() => useCollection("products", { q: "   ", search: [], sort: " " }));
    await waitFor(() => expect(urls).toHaveLength(1));
    expect(urls[0]).toBe("/api/v1/products?page=1&limit=12");
  });

  it("encodes a value, never a comma between values", async () => {
    renderHook(() => useCollection("products", { where: { name: "Đèo & co", category: ["a b", "c"] } }));
    await waitFor(() => expect(urls).toHaveLength(1));
    expect(urls[0]).toContain("&name=%C4%90%C3%A8o%20%26%20co&category=a%20b,c");
  });

  it("never asks for more than 100", async () => {
    renderHook(() => useCollection("products", { pageSize: 500 }));
    await waitFor(() => expect(urls).toHaveLength(1));
    expect(urls[0]).toBe("/api/v1/products?page=1&limit=100");
  });
});

describe("a catalogue that grows", () => {
  it("asks the next page on loadMore and appends it", async () => {
    const { result } = renderHook(() => useCollection("products", { pageSize: 5 }));
    await settled(result);
    expect(ids(result.current.items)).toEqual(ids(bikes.slice(0, 5)));
    expect(result.current).toMatchObject({ total: 14, pages: 3, page: 1, hasMore: true });

    act(() => result.current.loadMore());
    await waitFor(() => expect(result.current.items).toHaveLength(10));
    expect(urls[1]).toBe("/api/v1/products?page=2&limit=5");
    expect(result.current.hasMore).toBe(true);

    act(() => result.current.loadMore());
    await waitFor(() => expect(result.current.items).toHaveLength(14));
    expect(ids(result.current.items)).toEqual(ids(bikes));
    expect(result.current.hasMore).toBe(false);
  });

  it("ignores loadMore while a request is in flight and when nothing is left", async () => {
    const { result } = renderHook(() => useCollection("products", { pageSize: 10 }));
    await settled(result);
    const wait = held(bikes);
    respond = wait.answer;
    act(() => {
      result.current.loadMore();
      result.current.loadMore();
    });
    expect(urls).toHaveLength(2);
    expect(result.current.loading).toBe(true);
    act(() => wait.gate.release());
    await waitFor(() => expect(result.current.items).toHaveLength(14));
    act(() => result.current.loadMore());
    expect(urls).toHaveLength(2);
  });
});

describe("a table that shows a page", () => {
  it("holds that page alone and counts the pages", async () => {
    const { result } = renderHook(() => useCollection("products", { page: 2, pageSize: 5 }));
    await settled(result);
    expect(ids(result.current.items)).toEqual(ids(bikes.slice(5, 10)));
    expect(result.current).toMatchObject({ total: 14, pages: 3, page: 2, hasMore: false });
    expect(urls[0]).toBe("/api/v1/products?page=2&limit=5");
  });

  it("shows the last page when the page asked for is past it", async () => {
    const { result } = renderHook(() => useCollection("products", { page: 7, pageSize: 5 }));
    await settled(result);
    expect(result.current.page).toBe(3);
    expect(ids(result.current.items)).toEqual(ids(bikes.slice(10)));
    expect(urls).toEqual(["/api/v1/products?page=7&limit=5", "/api/v1/products?page=3&limit=5"]);
  });
});

describe("waiting and failing", () => {
  it("is loading, with no records, from the first render until the answer", async () => {
    const wait = held(bikes);
    respond = wait.answer;
    const { result } = renderHook(() => useCollection("products"));
    expect(result.current).toMatchObject({ loading: true, items: [], total: 0, error: null });
    await waitFor(() => expect(urls).toHaveLength(1));
    act(() => wait.gate.release());
    await settled(result);
    expect(result.current.items).toHaveLength(12);
  });

  it("sets error on a failed answer and keeps the records; a later success clears it", async () => {
    const { result, rerender } = renderHook(({ q }) => useCollection("products", { q }), { initialProps: { q: "a" } });
    await settled(result);
    respond = () => reply({ error: "Something went wrong. Please try again.", code: "INTERNAL" }, 500);
    rerender({ q: "b" });
    await waitFor(() => expect(result.current.error).toBeInstanceOf(Error));
    expect(result.current.error?.message).toBe("Something went wrong. Please try again.");
    expect(result.current.loading).toBe(false);
    expect(result.current.items).toHaveLength(12);

    respond = (url) => pageOf(bikes, url);
    rerender({ q: "c" });
    await waitFor(() => expect(result.current.error).toBeNull());
  });

  it("sets error when the network fails", async () => {
    respond = () => Promise.reject(new TypeError("Failed to fetch"));
    const { result } = renderHook(() => useCollection("products"));
    await settled(result);
    expect(result.current.error?.message).toBe("Failed to fetch");
    expect(result.current.items).toEqual([]);
  });

  it("sets error on an answer that is not a list", async () => {
    respond = () => reply({ something: "else" });
    const { result } = renderHook(() => useCollection("products"));
    await settled(result);
    expect(result.current.error).toBeInstanceOf(Error);
  });

  it("starts again from the first page of a new query, and drops the late answer of the old one", async () => {
    const slow = held([{ id: "from-a" }]);
    respond = (url) => (url.includes("q=a") ? slow.answer(url) : pageOf([{ id: "from-b" }], url));
    const { result, rerender } = renderHook(({ q }) => useCollection("products", { q }), { initialProps: { q: "a" } });
    await waitFor(() => expect(urls).toHaveLength(1));
    rerender({ q: "b" });
    await waitFor(() => expect(ids(result.current.items)).toEqual(["from-b"]));
    await act(async () => slow.gate.release());
    expect(ids(result.current.items)).toEqual(["from-b"]);
    expect(result.current.loading).toBe(false);
  });
});

describe("one record", () => {
  it("is empty and loading until answered, then holds the record", async () => {
    respond = () => reply({ id: "stance-0", name: "Stance 0" });
    const { result } = renderHook(() => useRecord("products", "stance-0"));
    expect(result.current).toEqual({ item: undefined, loading: true, error: null });
    await settled(result);
    expect(result.current.item).toEqual({ id: "stance-0", name: "Stance 0" });
    expect(result.current.error).toBeNull();
    expect(urls).toEqual(["/api/v1/products/stance-0"]);
  });

  it("stays empty on a 404 and says nothing went wrong", async () => {
    respond = () => reply({ error: "That record does not exist.", code: "NOT_FOUND" }, 404);
    const { result } = renderHook(() => useRecord("products", "no-such-bike"));
    await settled(result);
    expect(result.current).toEqual({ item: undefined, loading: false, error: null });
  });

  it("reports any other failure as an Error", async () => {
    respond = () => reply({ error: "The service is not ready.", code: "UNAVAILABLE" }, 503);
    const { result } = renderHook(() => useRecord("products", "stance-0"));
    await settled(result);
    expect(result.current.item).toBeUndefined();
    expect(result.current.error?.message).toBe("The service is not ready.");
  });

  it("never shows the previous record for a new id", async () => {
    respond = (url) => reply({ id: url.split("/").pop() });
    const { result, rerender } = renderHook(({ id }) => useRecord("products", id), { initialProps: { id: "a" } });
    await settled(result);
    rerender({ id: "b" });
    expect(result.current).toEqual({ item: undefined, loading: true, error: null });
    await settled(result);
    expect(result.current.item?.id).toBe("b");
  });

  it("asks nothing without an id, and encodes the id it is given", async () => {
    const none = renderHook(() => useRecord("products", undefined));
    expect(none.result.current).toEqual({ item: undefined, loading: false, error: null });
    expect(urls).toHaveLength(0);
    respond = () => reply({ id: "a/b c" });
    const some = renderHook(() => useRecord("products", "a/b c"));
    await settled(some.result);
    expect(urls).toEqual(["/api/v1/products/a%2Fb%20c"]);
  });
});

describe("the refresh signal and collections()", () => {
  it("asks again, and shows the new records, when another module sets the collection", async () => {
    const { result } = renderHook(() => useCollection("products", { pageSize: 5 }));
    await settled(result);
    respond = (url) => pageOf([{ id: "new-bike" }, ...bikes], url);
    act(() => setCollection("products", [{ id: "new-bike" }]));
    expect(collections().products).toEqual([{ id: "new-bike" }]);
    await waitFor(() => expect(result.current.items[0]?.id).toBe("new-bike"));
    expect(urls).toHaveLength(2);
  });

  it("does not ask again for a list of another collection", async () => {
    const { result } = renderHook(() => useCollection("categories"));
    await settled(result);
    act(() => setCollection("products", []));
    await new Promise((resolve) => setTimeout(resolve, 20));
    expect(urls).toHaveLength(1);
  });

  it("holds the riding types, with their id and name, once a screen has loaded them", async () => {
    setCollection("categories", []);
    respond = (url) =>
      pageOf(
        [
          { id: "trail", name: "Trail" },
          { id: "e-mtb", name: "E-MTB" },
        ],
        url,
      );
    const { result } = renderHook(() => useCollection("categories"));
    await settled(result);
    expect((collections().categories ?? []).map((r) => [r.id, r.name])).toEqual([
      ["trail", "Trail"],
      ["e-mtb", "E-MTB"],
    ]);
  });
});
