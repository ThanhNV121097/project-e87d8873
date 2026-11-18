import { createElement } from "react";
import { flushSync } from "react-dom";
import { createRoot } from "react-dom/client";
import { afterEach, describe, expect, it, vi } from "vitest";

// THE OWNER'S SESSION, WITH THE NETWORK FAKED.
//
// The module keeps state of its own — whether the "who is this" question has
// been asked, and the sentence a session that ended leaves behind — so each
// test loads a fresh copy of it.
const KEY = "agentteam:session";
type Handler = (url: string, init: RequestInit) => Response | Error | Promise<Response>;

async function fresh() {
  vi.resetModules();
  return import("../lib/session");
}

function serve(handler: Handler) {
  const calls: { url: string; init: RequestInit }[] = [];
  vi.stubGlobal("fetch", vi.fn(async (url: string, init: RequestInit = {}) => {
    calls.push({ url, init });
    const answer = await handler(url, init);
    if (answer instanceof Error) throw answer;
    return answer;
  }));
  return calls;
}

const json = (status: number, body: unknown) =>
  new Response(JSON.stringify(body), { status, headers: { "Content-Type": "application/json" } });
const owner = (email = "owner@deo.test") => json(200, { email });
const notSignedIn = () => json(401, { error: "You are not signed in.", code: "UNAUTHENTICATED" });
const hint = (email = "owner@deo.test") => localStorage.setItem(KEY, JSON.stringify({ email }));
const kept = () => localStorage.getItem(KEY);

/**
 * Runs a hook in a component and hands back its latest value. The updates are
 * flushed by hand and waited for with a timer: react's act() is not in the
 * production build the test run may be given.
 */
const mounted: (() => void)[] = [];
function mount<T>(hook: () => T) {
  const root = createRoot(document.createElement("div"));
  const result = { current: undefined as T };
  const Probe = () => {
    result.current = hook();
    return null;
  };
  flushSync(() => root.render(createElement(Probe)));
  mounted.push(() => flushSync(() => root.unmount()));
  return result;
}
const settle = () => new Promise<void>((resolve) => setTimeout(resolve, 0));
async function until(check: () => boolean) {
  for (let tries = 0; tries < 50 && !check(); tries += 1) await settle();
  expect(check()).toBe(true);
}

afterEach(() => {
  mounted.splice(0).forEach((unmount) => unmount());
  vi.unstubAllGlobals();
  localStorage.clear();
});

describe("the session module", () => {
  it("keeps its four names, and the site has no roles", async () => {
    const m = await fresh();
    expect([m.useSession, m.signIn, m.signOut, m.useRole].every((f) => typeof f === "function")).toBe(true);
    serve(() => owner("lan.tran@deo.test"));
    hint("lan.tran@deo.test");
    const result = mount(() => ({ session: m.useSession(), role: m.useRole() }));
    expect(result.current.role).toBe("");
    // The name is made from the email: the foot of the sidebar shows "Lan Tran".
    expect(result.current.session).toEqual({ email: "lan.tran@deo.test", name: "Lan Tran" });
    expect(m.allowed(result.current.session)).toBe(true);
    expect(m.allowed(null)).toBe(false);
  });
});

describe("signing in", () => {
  it("sends the typed email and password, and keeps only the owner's email", async () => {
    const m = await fresh();
    const calls = serve(() => owner());
    await m.signIn("  Owner@deo.test ", "the-password");
    expect(calls[0].url).toBe("/api/v1/sessions");
    expect(calls[0].init.method).toBe("POST");
    expect(JSON.parse(String(calls[0].init.body))).toEqual({ email: "Owner@deo.test", password: "the-password" });
    expect(JSON.parse(kept() ?? "null")).toEqual({ email: "owner@deo.test" });
    expect(kept()).not.toContain("the-password");
  });

  it("rejects with the server's sentence for a refusal, and with the generic one for no usable answer", async () => {
    const m = await fresh();
    const answers = [
      json(401, { error: "That email and password do not match.", code: "UNAUTHENTICATED" }),
      json(429, { error: "Too many attempts. Try again in a few minutes.", code: "RATE_LIMITED" }),
      new TypeError("Failed to fetch"),
      new Response("<html>Bad gateway</html>", { status: 502 }),
    ];
    serve(() => answers.shift() ?? new Error("no answer was queued"));
    await expect(m.signIn("owner@deo.test", "wrong")).rejects.toThrow("That email and password do not match.");
    await expect(m.signIn("owner@deo.test", "right")).rejects.toThrow("Too many attempts. Try again in a few minutes.");
    await expect(m.signIn("owner@deo.test", "right")).rejects.toThrow("Could not sign in. Please try again.");
    await expect(m.signIn("owner@deo.test", "right")).rejects.toThrow("Could not sign in. Please try again.");
    expect(kept()).toBeNull();
  });
});

describe("the question asked once per page load: who is this?", () => {
  it("answers with the hint while the server has not answered, then takes the server's email", async () => {
    const m = await fresh();
    serve(() => owner("owner@deo.test"));
    hint("old@deo.test");
    const result = mount(() => m.useSession());
    expect(result.current?.email).toBe("old@deo.test");
    await until(() => result.current?.email === "owner@deo.test");
    expect(JSON.parse(kept() ?? "null")).toEqual({ email: "owner@deo.test" });
  });

  it("clears the hint when the server says nobody is signed in", async () => {
    const m = await fresh();
    serve(notSignedIn);
    hint();
    const result = mount(() => m.useSession());
    expect(result.current?.email).toBe("owner@deo.test");
    await until(() => result.current === null);
    expect(kept()).toBeNull();
  });

  it("clears the hint when the server cannot be reached or answers anything but 200", async () => {
    const answers = () => [new TypeError("Failed to fetch"), json(503, { error: "The service is not ready.", code: "UNAVAILABLE" })];
    for (const index of [0, 1]) {
      const m = await fresh();
      serve(() => answers()[index]);
      hint();
      const result = mount(() => m.useSession());
      await until(() => result.current === null);
      expect(kept()).toBeNull();
    }
  });

  it("finds a valid cookie that has no hint", async () => {
    const m = await fresh();
    serve(() => owner());
    const result = mount(() => m.useSession());
    expect(result.current).toBeNull();
    await until(() => result.current?.email === "owner@deo.test");
  });

  it("does not undo a sign-in that finished while the question was still out", async () => {
    const m = await fresh();
    let reply: (r: Response) => void = () => {};
    const me = new Promise<Response>((resolve) => { reply = resolve; });
    serve((url) => (url.endsWith("/me") ? me : owner()));
    const result = mount(() => m.useSession());
    await m.signIn("owner@deo.test", "right");
    await settle();
    expect(result.current?.email).toBe("owner@deo.test");
    reply(notSignedIn());
    await me;
    await settle();
    expect(result.current?.email).toBe("owner@deo.test");
  });
});

describe("signing out", () => {
  it("forgets the session once the server answered 204", async () => {
    const m = await fresh();
    const calls = serve((url, init) => (init.method === "DELETE" ? new Response(null, { status: 204 }) : owner()));
    hint();
    const result = mount(() => m.useSession());
    await m.signOut();
    await settle();
    expect(calls.some((c) => c.url === "/api/v1/sessions/current" && c.init.method === "DELETE")).toBe(true);
    expect(result.current).toBeNull();
    expect(kept()).toBeNull();
  });

  it("leaves the owner signed in when the server does not say so", async () => {
    const m = await fresh();
    const answers: (Response | Error)[] = [
      new TypeError("Failed to fetch"),
      json(500, { error: "Something went wrong. Please try again.", code: "INTERNAL" }),
    ];
    serve((url) => (url.endsWith("/me") ? owner() : answers.shift() ?? new Error("no answer was queued")));
    hint();
    const result = mount(() => m.useSession());
    await m.signOut();
    await m.signOut();
    await settle();
    expect(result.current?.email).toBe("owner@deo.test");
    expect(kept()).not.toBeNull();
  });
});

describe("a session the server ended under a write", () => {
  it("drops the owner, leaves the sentence for the sign-in form, and clears it at the next attempt", async () => {
    const m = await fresh();
    serve((url) => (url.endsWith("/me") ? owner() : json(401, { error: "That email and password do not match.", code: "UNAUTHENTICATED" })));
    hint();
    const result = mount(() => ({ session: m.useSession(), ended: m.useSessionEnded() }));
    expect(result.current.ended).toBe("");
    m.endSession();
    await settle();
    expect(result.current.session).toBeNull();
    expect(result.current.ended).toBe("Your session has ended. Sign in again.");
    await m.signIn("owner@deo.test", "wrong").catch(() => undefined);
    await settle();
    expect(result.current.ended).toBe("");
  });
});
