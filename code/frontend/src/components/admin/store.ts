/**
 * WHAT "SAVE" MEANS: THE SERVICE'S ANSWER, THEN THE SCREENS.
 *
 * Every write goes to the API (docs/architecture/services.md §3.4 to §3.6)
 * with the owner's session cookie, and the screens wait for the answer:
 * saveRow and deleteRow return promises that resolve when the server
 * accepted the change and reject with the sentence to show when it did not —
 * the server's own sentence where the service's error envelope came back, a
 * generic one where nothing usable did.
 *
 * ONLY AN ACCEPTED WRITE TELLS THE SCREENS. The data module (src/data) holds
 * what every list and record on screen shows; after the server said yes the
 * change is handed to it with setCollection, which redraws every
 * useCollection showing that collection. A refused or failed write sends no
 * signal, so what the screens show is what the server holds.
 *
 * THE VERSION TOKEN TRAVELS WITH THE VALUES IT WAS READ WITH. A list carries
 * no version, and a save must be refused when the record changed since the
 * screen loaded it (an order lowered its stock, or the owner saved it from
 * another window). So a screen that is about to edit a record reads it with
 * readRow, which keeps the answer's ETag next to the values it returns;
 * saveRow sends that ETag as If-Match on a change and keeps the one the
 * answer returns for the next save. It is never fetched at save time: a token
 * read after the person started typing would hide exactly the change it
 * exists to catch. An add and a delete send none.
 */
import { useCallback, useRef, useState } from "react";
import { collections, setCollection, type Row } from "../../data";
import { apiUrl, endSession, sentenceOf, SESSION_ENDED } from "../../lib/session";

const SAVE_FAILED = "The change could not be saved. Please try again.";
const DELETE_FAILED = "The record could not be deleted. Please try again.";
export const OPEN_FAILED = "The record could not be opened. Please try again.";

/** A refusal or a failure, carrying the sentence to show and the HTTP status (0: no answer). */
export class StoreError extends Error {
  status: number;

  constructor(message: string, status = 0) {
    super(message);
    this.name = "StoreError";
    this.status = status;
  }
}

/** The ETag of each record this browser read or wrote last: "products/talon-2" → `"7"`. */
const tokens = new Map<string, string>();

const keyOf = (collection: string, id: string) => `${collection}/${id}`;

function keep(collection: string, id: string, res: Response) {
  const tag = res.headers.get("ETag");
  if (tag) tokens.set(keyOf(collection, id), tag);
  else tokens.delete(keyOf(collection, id));
}

/** One request. A 401 ends the session; any other refusal is its sentence. */
async function call(path: string, init: RequestInit, generic: string): Promise<Response> {
  let res: Response;
  try {
    res = await fetch(apiUrl(path), init);
  } catch {
    throw new StoreError(generic);
  }
  if (res.ok) return res;
  if (res.status === 401) {
    endSession();
    throw new StoreError(SESSION_ENDED, 401);
  }
  throw new StoreError((await sentenceOf(res)) || generic, res.status);
}

async function rowOf(res: Response): Promise<Row | null> {
  try {
    const body = (await res.json()) as Row | null;
    return body && typeof body === "object" && typeof body.id === "string" ? body : null;
  } catch {
    return null;
  }
}

/**
 * What is sent for a record: everything the screen holds except `_source`,
 * which the service refuses as an undefined key and never touches. The
 * hidden fields the form does not show (`price_listed`, `_invented`) go back
 * as they were read, because a change replaces the record. An add sends no
 * `id`: the server makes one. `wheel` is the text "29" or "mixed".
 */
function bodyOf(row: Row): Record<string, unknown> {
  const { _source: _origin, ...rest } = row;
  const body: Record<string, unknown> = { ...rest };
  if (!body.id) delete body.id;
  if (typeof body.wheel === "number") body.wheel = String(body.wheel);
  return body;
}

/**
 * A new record's id.
 *
 * A COUNTER AND NOT THE LIST'S LENGTH, because two records added inside one
 * millisecond must not share an id. No screen needs it now — the server
 * makes the id of an add — but a text no other call returns is still the
 * contract of the name.
 */
let made = 0;

export function newId(collection: string): string {
  made += 1;
  return `${collection}-${Date.now().toString(36)}-${made}`;
}

/**
 * One record by its id, for a screen that is about to edit it: the values
 * and, kept beside them, the token saveRow will send with the change.
 * Rejects with the server's sentence ("That record does not exist." for a
 * 404 — the error's `status` says so) or OPEN_FAILED when no usable answer came.
 */
export async function readRow(collection: string, id: string): Promise<Row> {
  const res = await call(`/${collection}/${encodeURIComponent(id)}`, {}, OPEN_FAILED);
  const row = await rowOf(res);
  if (!row) throw new StoreError(OPEN_FAILED);
  keep(collection, row.id, res);
  return row;
}

/**
 * Adds a record (no id) or replaces one (its id), and resolves with the
 * record the server saved. A row keeps its place in the list.
 */
export async function saveRow(collection: string, row: Row): Promise<Row> {
  const adding = !row.id;
  const headers: Record<string, string> = { "Content-Type": "application/json" };
  const token = adding ? undefined : tokens.get(keyOf(collection, row.id));
  if (token) headers["If-Match"] = token;
  const res = await call(
    adding ? `/${collection}` : `/${collection}/${encodeURIComponent(row.id)}`,
    { method: adding ? "POST" : "PUT", headers, body: JSON.stringify(bodyOf(row)) },
    SAVE_FAILED,
  );
  const saved = (await rowOf(res)) ?? row;
  keep(collection, saved.id, res);
  const rows = collections()[collection] ?? [];
  const next = rows.some((r) => r.id === saved.id)
    ? rows.map((r) => (r.id === saved.id ? saved : r))
    : saved.id ? [...rows, saved] : [...rows];
  setCollection(collection, next);
  return saved;
}

/** Deletes a record. A delete sends no token and is not refused for staleness. */
export async function deleteRow(collection: string, id: string): Promise<void> {
  await call(`/${collection}/${encodeURIComponent(id)}`, { method: "DELETE" }, DELETE_FAILED);
  tokens.delete(keyOf(collection, id));
  setCollection(collection, (collections()[collection] ?? []).filter((r) => r.id !== id));
}

/**
 * A button that makes one request at a time. `run(job)` does nothing while
 * a job is in flight (a double press is one request), clears the last
 * problem when it starts, resolves true when the job succeeded and false when
 * it was refused — `problem` then holds the sentence — or was not started.
 */
export function useAttempt() {
  const flying = useRef(false);
  const [busy, setBusy] = useState(false);
  const [problem, setProblem] = useState("");
  const run = useCallback(async (job: () => Promise<unknown>): Promise<boolean> => {
    if (flying.current) return false;
    flying.current = true;
    setBusy(true);
    setProblem("");
    try {
      await job();
      return true;
    } catch (failure) {
      setProblem(failure instanceof Error ? failure.message : SAVE_FAILED);
      return false;
    } finally {
      flying.current = false;
      setBusy(false);
    }
  }, []);
  return { busy, problem, run };
}
