import { createElement } from "react";
import { flushSync } from "react-dom";
import { createRoot } from "react-dom/client";
import { afterEach, describe, expect, it, vi } from "vitest";
import {
  blank, columnsFromFields, fieldsOf, labelOf, missing, titleOf, type FieldSpec,
} from "../components/admin/fields";
import { deleteRow, newId, readRow, saveRow, useAttempt } from "../components/admin/store";
import { collections, schema, setCollection, type Row } from "../data";

// THE MANAGEMENT SCREENS' ONE PIECE OF LOGIC.
//
// Everything else in src/components/admin is markup, which the free walk and
// a reviewer can see. These four functions are what a screen DERIVES — the
// columns, the form, what is refused on save, which row a write lands on —
// and they are invisible: a wrong one ships a CMS that looks finished and
// edits the wrong record.
describe("the fields a management screen derives", () => {
  it("names a field a person can read", () => {
    expect(labelOf("created_at")).toBe("Created at");
    expect(labelOf("fullName")).toBe("Full name");
  });

  it("draws for a collection the schema does not describe", () => {
    // The scaffold ships schema.json with no collections at all, and a page
    // written before its data exists must still render rather than throw.
    expect(fieldsOf("nothing-here")).toEqual([]);
    expect(columnsFromFields([])).toEqual([]);
  });

  it("keeps long text out of the table and numbers to the right", () => {
    const fields: FieldSpec[] = [
      { key: "name", label: "Name", type: "text" },
      { key: "story", label: "Story", type: "longtext" },
      { key: "price", label: "Price", type: "price" },
    ];
    const cols = columnsFromFields(fields);
    expect(cols.map((c) => c.key)).toEqual(["name", "price"]);
    expect(cols.find((c) => c.key === "price")?.align).toBe("right");
  });

  it("refuses a required field that is empty, and only that", () => {
    const fields: FieldSpec[] = [
      { key: "name", label: "Name", type: "text", required: true },
      { key: "note", label: "Note", type: "text" },
      { key: "live", label: "Live", type: "bool", required: true },
    ];
    const values = blank(fields);
    // A false toggle is an answer, not an absence: only the text is missing.
    expect(Object.keys(missing(fields, values))).toEqual(["name"]);
    expect(missing(fields, { ...values, name: "Trà sen" })).toEqual({});
  });

  it("calls a record by its own name", () => {
    expect(titleOf({ id: "p1", name: "Nhà Gốm" })).toBe("Nhà Gốm");
    expect(titleOf({ id: "p1" })).toBe("p1");
    expect(titleOf(undefined)).toBe("");
  });
});

/** The answers the faked network gives, in order; the calls it was sent, for what they carried. */
function serve(...replies: (Response | Error)[]) {
  const calls: { url: string; init: RequestInit }[] = [];
  vi.stubGlobal("fetch", vi.fn(async (url: string, init: RequestInit = {}) => {
    calls.push({ url, init });
    const next = replies.shift();
    if (!next) throw new Error("no answer was queued");
    if (next instanceof Error) throw next;
    return next;
  }));
  return calls;
}

const json = (status: number, body: unknown, headers: Record<string, string> = {}) =>
  new Response(JSON.stringify(body), { status, headers: { "Content-Type": "application/json", ...headers } });

const sentBody = (call: { init: RequestInit }) => JSON.parse(String(call.init.body)) as Record<string, unknown>;
const headerOf = (call: { init: RequestInit }, name: string) => (call.init.headers as Record<string, string>)[name];

describe("what Save does against the service", () => {
  const rows: Row[] = [
    { id: "a", name: "First" },
    { id: "b", name: "Second" },
    { id: "c", name: "Third" },
  ];
  const seed = () => setCollection("things", rows.map((r) => ({ ...r })));
  afterEach(() => {
    vi.unstubAllGlobals();
    localStorage.clear();
  });

  it("adds a record without an id, takes the id the server made, and puts it at the end", async () => {
    seed();
    const calls = serve(json(201, { id: "things-k3f9a2", name: "Fourth" }, { ETag: "\"1\"" }));
    const saved = await saveRow("things", { id: "", name: "Fourth" });
    expect(saved.id).toBe("things-k3f9a2");
    expect(calls[0].url).toBe("/api/v1/things");
    expect(calls[0].init.method).toBe("POST");
    expect(sentBody(calls[0])).toEqual({ name: "Fourth" });
    expect(headerOf(calls[0], "If-Match")).toBeUndefined();
    expect(collections().things.map((r) => r.name)).toEqual(["First", "Second", "Third", "Fourth"]);
  });

  it("changes a record with the token it was read with, and keeps the place and the next token", async () => {
    seed();
    const calls = serve(
      json(200, { id: "b", name: "Second", _source: "owner", price_listed: "19", wheel: "29" }, { ETag: "\"3\"" }),
      json(200, { id: "b", name: "Second, edited" }, { ETag: "\"4\"" }),
      json(200, { id: "b", name: "Second, again" }, { ETag: "\"5\"" }),
    );
    const read = await readRow("things", "b");
    expect(calls[0].url).toBe("/api/v1/things/b");
    await saveRow("things", { ...read, name: "Second, edited", wheel: 29 } as Row);
    expect(calls[1].url).toBe("/api/v1/things/b");
    expect(calls[1].init.method).toBe("PUT");
    expect(headerOf(calls[1], "If-Match")).toBe("\"3\"");
    // `_source` is never sent; what the form does not show goes back as read; wheel is text.
    expect(sentBody(calls[1])).toEqual({ id: "b", name: "Second, edited", price_listed: "19", wheel: "29" });
    expect(collections().things.map((r) => r.id)).toEqual(["a", "b", "c"]);
    expect(collections().things[1].name).toBe("Second, edited");
    await saveRow("things", { id: "b", name: "Second, again" });
    expect(headerOf(calls[2], "If-Match")).toBe("\"4\"");
  });

  it("removes exactly the row asked for, sending no token", async () => {
    seed();
    const calls = serve(new Response(null, { status: 204 }));
    await deleteRow("things", "b");
    expect(calls[0].url).toBe("/api/v1/things/b");
    expect(calls[0].init.method).toBe("DELETE");
    expect(headerOf(calls[0], "If-Match")).toBeUndefined();
    expect(collections().things.map((r) => r.id)).toEqual(["a", "c"]);
  });

  it("rejects with the server's own sentence and tells the screens nothing", async () => {
    seed();
    const before = collections().things;
    const stale = "This record changed while you were editing it — for example an order was placed. Reload and make your change again.";
    serve(
      json(409, { error: stale, code: "STALE_RECORD" }),
      json(409, { error: "Trail is still the riding type of 4 bikes. Give those bikes another riding type first.", code: "IN_USE" }),
      json(400, { error: "A live bike needs a price.", code: "VALIDATION_FAILED", field: "price" }),
    );
    await expect(saveRow("things", { id: "b", name: "Second" })).rejects.toThrow(stale);
    await expect(deleteRow("things", "b")).rejects.toThrow(/^Trail is still the riding type of 4 bikes\./);
    await expect(saveRow("things", { id: "", name: "No price" })).rejects.toThrow("A live bike needs a price.");
    expect(collections().things).toBe(before);
  });

  it("says the generic sentence when no usable answer came", async () => {
    seed();
    serve(
      new TypeError("Failed to fetch"),
      new TypeError("Failed to fetch"),
      new Response("<html>Bad gateway</html>", { status: 502 }),
      new Response("<html>Bad gateway</html>", { status: 502 }),
    );
    await expect(saveRow("things", { id: "b", name: "x" })).rejects.toThrow("The change could not be saved. Please try again.");
    await expect(deleteRow("things", "b")).rejects.toThrow("The record could not be deleted. Please try again.");
    await expect(saveRow("things", { id: "b", name: "x" })).rejects.toThrow("The change could not be saved. Please try again.");
    await expect(readRow("things", "b")).rejects.toThrow("The record could not be opened. Please try again.");
    expect(collections().things.map((r) => r.id)).toEqual(["a", "b", "c"]);
  });

  it("ends the session when the server says nobody is signed in", async () => {
    seed();
    localStorage.setItem("agentteam:session", JSON.stringify({ email: "owner@deo.test" }));
    serve(json(401, { error: "You are not signed in.", code: "UNAUTHENTICATED" }));
    await expect(saveRow("things", { id: "b", name: "x" })).rejects.toThrow("Your session has ended. Sign in again.");
    expect(localStorage.getItem("agentteam:session")).toBeNull();
  });

  it("reads a record that is gone as a 404 carrying the server's sentence", async () => {
    serve(json(404, { error: "That record does not exist.", code: "NOT_FOUND" }));
    await expect(readRow("things", "gone")).rejects.toMatchObject({ message: "That record does not exist.", status: 404 });
  });

  it("never gives two records the same id", () => {
    setCollection("things", []);
    expect(newId("things")).not.toBe(newId("things"));
  });
});

/**
 * Runs a hook in a component and hands back its latest value. The updates are
 * flushed by hand and waited for with a timer: react's act() is not in the
 * production build the test run may be given.
 */
function mount<T>(hook: () => T) {
  const root = createRoot(document.createElement("div"));
  const result = { current: undefined as T };
  const Probe = () => {
    result.current = hook();
    return null;
  };
  flushSync(() => root.render(createElement(Probe)));
  return { result, unmount: () => flushSync(() => root.unmount()) };
}
const settle = () => new Promise<void>((resolve) => setTimeout(resolve, 0));

describe("a button that makes one request at a time", () => {
  it("ignores a second press while the first is in flight", async () => {
    const { result, unmount } = mount(() => useAttempt());
    let release = () => {};
    const job = vi.fn(() => new Promise<void>((resolve) => { release = resolve; }));
    const first = result.current.run(job);
    await settle();
    expect(result.current.busy).toBe(true);
    expect(await result.current.run(job)).toBe(false);
    expect(job).toHaveBeenCalledTimes(1);
    release();
    expect(await first).toBe(true);
    await settle();
    expect(result.current.busy).toBe(false);
    unmount();
  });

  it("keeps the refusal's sentence until the next attempt starts", async () => {
    const { result, unmount } = mount(() => useAttempt());
    expect(await result.current.run(() => Promise.reject(new Error("A live bike needs a price.")))).toBe(false);
    await settle();
    expect(result.current.problem).toBe("A live bike needs a price.");
    expect(await result.current.run(() => Promise.resolve())).toBe(true);
    await settle();
    expect(result.current.problem).toBe("");
    unmount();
  });
});

// WHAT A FIELD IS, READ OFF ITS ROWS.
//
// The form run bd4d1058 shipped: "Status" as a box to type `live` into,
// "Photograph" as a box holding a file path, "Backdrop colour" as a box
// holding `var(--swatch-2)`. Each of those is decidable from the data, and
// these are the rows that form was drawn from, in miniature.
describe("a field read for what it is", () => {
  const rows: Row[] = Array.from({ length: 8 }, (_, i) => ({
    id: `p${i}`,
    name: `Garment ${i}`,
    image: `/images/products-${i}.jpg`,
    status: i % 3 === 0 ? "draft" : "live",
    backdrop: `var(--swatch-${(i % 4) + 1})`,
    fabric: `A fabric nobody else has, number ${i}`,
  }));

  const specs = () => {
    schema.collections.garments = {
      fields: [
        { key: "name", type: "text" },
        { key: "image", type: "text" },
        { key: "status", type: "text" },
        { key: "backdrop", type: "text" },
        { key: "fabric", type: "text" },
        { key: "swatch", type: "text", hidden: true },
        { key: "size", type: "text", options: ["s", ["m", "Medium"]] },
      ],
    };
    setCollection("garments", rows);
    return fieldsOf("garments");
  };

  it("makes a choice of a text field whose rows share a few values", () => {
    const status = specs().find((f) => f.key === "status");
    expect(status?.options?.map(([v]) => v).sort()).toEqual(["draft", "live"]);
    // A field with a different value on every row is not a choice.
    expect(specs().find((f) => f.key === "name")?.options).toBeUndefined();
    expect(specs().find((f) => f.key === "fabric")?.options).toBeUndefined();
  });

  it("sees that a column of image paths is a picture", () => {
    expect(specs().find((f) => f.key === "image")?.type).toBe("image");
  });

  it("leaves the design's own plumbing off the form", () => {
    const keys = specs().map((f) => f.key);
    expect(keys).not.toContain("backdrop"); // every value is var(--…)
    expect(keys).not.toContain("swatch"); // the schema said hidden
  });

  it("believes the schema when it names the choices", () => {
    expect(specs().find((f) => f.key === "size")?.options).toEqual([["s", "S"], ["m", "Medium"]]);
  });
});
